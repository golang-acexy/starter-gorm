package gormstarter

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/golang-acexy/starter-parent/parent"
	"gorm.io/gorm"
)

const defaultCharset = "utf8mb4"

var (
	gormRuntimeState atomic.Pointer[gormRuntime]
	gormLifecycleLock sync.Mutex
	gormState gormLifecycleState
)

type gormRuntime struct {
	databases     map[DBType]*gorm.DB
	defaultDBType DBType
}

type gormLifecycleState uint8

const (
	gormStopped gormLifecycleState = iota
	gormStarting
	gormRunning
	gormStopping
)

// DatabaseConfig selects a connection target and GORM behavior; it exposes no SQL pool limits or health-check deadline.
type DatabaseConfig struct {
	// Username is the driver connection username.
	Username string
	// Password is the driver connection password; keep it private.
	Password string
	// Host is the TCP database hostname or address.
	Host     string
	// Port is the TCP database port; no starter default is supplied.
	Port     uint
	// Database is the target database name.
	Database string
	// TimeUTC uses UTC for GORM-generated timestamps, not the connection session timezone.
	TimeUTC  bool
	// DryRun generates SQL without executing operations; startup still opens and checks the connection.
	DryRun   bool
}

// MySQLConfig extends common settings with charset and MySQL driver DSN parameters.
type MySQLConfig struct {
	// DatabaseConfig supplies the common connection settings.
	DatabaseConfig
	// Charset defaults to utf8mb4 when empty.
	Charset   string
	// URLParams supplies MySQL driver DSN options; parseTime and charset cannot override starter settings.
	URLParams string
}

// PostgresConfig extends common settings with server timezone and SSL mode.
type PostgresConfig struct {
	// DatabaseConfig supplies the common connection settings.
	DatabaseConfig
	// Timezone sets the PostgreSQL session timezone; empty defaults to UTC.
	Timezone string
	// SSLMode sets PostgreSQL SSL behavior; empty defaults to disable.
	SSLMode  string
}

// GormConfig initializes MySQL, PostgreSQL, or both; keep configuration fixed before first resolution.
type GormConfig struct {
	// SQLLoggerLevel selects trace logging; values below InfoLevel use DebugLevel.
	SQLLoggerLevel logger.Level
	// MySQL configures MySQL when nonnil.
	MySQL          *MySQLConfig
	// Postgres configures PostgreSQL when nonnil.
	Postgres       *PostgresConfig
	// InitFunc receives shared DB handles through the default parent hook after startup; pool tuning happens after connection checks.
	InitFunc       func(databases map[DBType]*gorm.DB)
}

// GormStarter owns the process-wide database registry; concurrent configuration mutation is unsupported.
type GormStarter struct {
	// Config supplies static settings when LazyConfig is nil.
	Config      GormConfig
	// LazyConfig takes precedence and is evaluated once on first Setting or Start resolution.
	LazyConfig  func() GormConfig
	config      *GormConfig
	configOnce  sync.Once
	// GormSetting replaces the default lifecycle setting, including its InitFunc hook.
	GormSetting *parent.Setting
}

func (g *GormStarter) getConfig() *GormConfig {
	g.configOnce.Do(func() {
		config := g.Config
		if g.LazyConfig != nil {
			config = g.LazyConfig()
		}
		if config.MySQL != nil {
			mysqlConfig := *config.MySQL
			config.MySQL = &mysqlConfig
		}
		if config.Postgres != nil {
			postgresConfig := *config.Postgres
			config.Postgres = &postgresConfig
		}
		g.config = &config
	})
	return g.config
}

// Setting returns a custom setting or resolves configuration for the default parent setting and post-start InitFunc hook.
func (g *GormStarter) Setting() *parent.Setting {
	if g.GormSetting != nil {
		return g.GormSetting
	}
	config := g.getConfig()
	return parent.NewSetting("Gorm-Starter", false, 20, true, time.Second*30, func(instance any) {
		if config.InitFunc != nil {
			config.InitFunc(instance.(map[DBType]*gorm.DB))
		}
	})
}

// Start initializes configured dialects and checks their SQL pools before publishing shared handles.
// The returned map is a copy, but DB handles are shared; overlapping lifecycle transitions are rejected.
// Connection health checks have no starter-supplied deadline; DryRun still requires initialization.
func (g *GormStarter) Start() (any, error) {
	config := g.getConfig()
	if config.MySQL == nil && config.Postgres == nil {
		return nil, ErrNoDatabaseConfigured
	}

	gormLifecycleLock.Lock()
	if gormState != gormStopped {
		gormLifecycleLock.Unlock()
		return nil, ErrGormStarterAlreadyStarted
	}
	gormState = gormStarting
	gormLifecycleLock.Unlock()
	started := false
	defer func() {
		if !started {
			gormLifecycleLock.Lock()
			gormState = gormStopped
			gormLifecycleLock.Unlock()
		}
	}()

	databases := make(map[DBType]*gorm.DB, 2)
	if config.MySQL != nil {
		if config.MySQL.Charset == "" {
			config.MySQL.Charset = defaultCharset
		}
		db, err := g.openDatabase(DBTypeMySQL, config.MySQL.DatabaseConfig, func(rawConfig *gorm.Config) (*gorm.DB, error) {
			return openMysqlDB(config.MySQL, rawConfig)
		})
		if err != nil {
			closeDatabases(databases)
			return nil, err
		}
		databases[DBTypeMySQL] = db
	}
	if config.Postgres != nil {
		db, err := g.openDatabase(DBTypePostgres, config.Postgres.DatabaseConfig, func(rawConfig *gorm.Config) (*gorm.DB, error) {
			return openPostgresDB(config.Postgres, rawConfig)
		})
		if err != nil {
			closeDatabases(databases)
			return nil, err
		}
		databases[DBTypePostgres] = db
	}

	defaultDBType := DBTypePostgres
	if databases[DBTypeMySQL] != nil {
		defaultDBType = DBTypeMySQL
	}
	runtime := &gormRuntime{databases: databases, defaultDBType: defaultDBType}
	gormLifecycleLock.Lock()
	gormRuntimeState.Store(runtime)
	gormState = gormRunning
	gormLifecycleLock.Unlock()
	started = true
	logger.Logrus().Infoln("Gorm-Starter databases started", "databaseTypes", databaseTypes(databases))
	return cloneDatabases(databases), nil
}

func (g *GormStarter) openDatabase(dbType DBType, config DatabaseConfig, open func(*gorm.Config) (*gorm.DB, error)) (*gorm.DB, error) {
	log := logger.Logrus().WithField("dbType", dbType)
	rawConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		DryRun:                                   config.DryRun,
		Logger:                                   &logrusLogger{level: effectiveSQLLoggerLevel(g.getConfig().SQLLoggerLevel)},
	}
	if config.TimeUTC {
		rawConfig.NowFunc = func() time.Time { return time.Now().UTC() }
	}
	db, err := open(rawConfig)
	if err != nil {
		log.WithError(err).Errorln("Database startup failed")
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.WithError(err).Errorln("Failed to get database connection")
		return nil, err
	}
	if err = sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		log.WithError(err).Errorln("Database connection check failed")
		return nil, err
	}
	return db, nil
}

// Stop withdraws the runtime, closes pools, then polls for drained connections.
// maxWaitTime bounds only post-close polling; SQL Close calls occur before the timer starts.
// Close errors or polling timeout report nongraceful completion with runtime still withdrawn.
func (g *GormStarter) Stop(maxWaitTime time.Duration) (gracefully, stopped bool, err error) {
	gormLifecycleLock.Lock()
	runtime := gormRuntimeState.Load()
	if gormState != gormRunning || runtime == nil {
		gormLifecycleLock.Unlock()
		return false, true, ErrGormStarterNotStarted
	}
	gormRuntimeState.Store(nil)
	gormState = gormStopping
	gormLifecycleLock.Unlock()
	databases := runtime.databases

	var closeErrors []error
	for dbType, db := range databases {
		log := logger.Logrus().WithField("dbType", dbType)
		sqlDB, dbErr := db.DB()
		if dbErr != nil {
			log.WithError(dbErr).Errorln("Failed to get database connection")
			closeErrors = append(closeErrors, dbErr)
			continue
		}
		if dbErr = sqlDB.Close(); dbErr != nil {
			log.WithError(dbErr).Errorln("Failed to close database")
			closeErrors = append(closeErrors, dbErr)
		}
	}
	if len(closeErrors) > 0 {
		gormLifecycleLock.Lock()
		gormState = gormStopped
		gormLifecycleLock.Unlock()
		return false, true, errors.Join(closeErrors...)
	}

	deadline := time.NewTimer(maxWaitTime)
	defer deadline.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if databasesClosed(databases) {
			gormLifecycleLock.Lock()
			gormState = gormStopped
			gormLifecycleLock.Unlock()
			logger.Logrus().Infoln("Gorm-Starter databases stopped", "databaseTypes", databaseTypes(databases))
			return true, true, nil
		}
		select {
		case <-deadline.C:
			logger.Logrus().WithError(ErrGormStopTimeout).Errorln("Gorm-Starter database shutdown timed out")
			gormLifecycleLock.Lock()
			gormState = gormStopped
			gormLifecycleLock.Unlock()
			return false, true, ErrGormStopTimeout
		case <-ticker.C:
		}
	}
}

func databasesClosed(databases map[DBType]*gorm.DB) bool {
	for _, db := range databases {
		sqlDB, err := db.DB()
		if err != nil {
			continue
		}
		stats := sqlDB.Stats()
		if stats.Idle != 0 || stats.InUse != 0 || stats.OpenConnections != 0 {
			return false
		}
	}
	return true
}

func closeDatabases(databases map[DBType]*gorm.DB) {
	for dbType, db := range databases {
		log := logger.Logrus().WithField("dbType", dbType)
		if sqlDB, err := db.DB(); err == nil {
			if err = sqlDB.Close(); err != nil {
				log.WithError(err).Errorln("Failed to close database during startup rollback")
				continue
			}
		} else {
			log.WithError(err).Errorln("Failed to get database connection during startup rollback")
		}
	}
}

func databaseTypes(databases map[DBType]*gorm.DB) string {
	types := make([]string, 0, len(databases))
	if databases[DBTypeMySQL] != nil {
		types = append(types, string(DBTypeMySQL))
	}
	if databases[DBTypePostgres] != nil {
		types = append(types, string(DBTypePostgres))
	}
	return strings.Join(types, ",")
}

func cloneDatabases(databases map[DBType]*gorm.DB) map[DBType]*gorm.DB {
	return coll.MapCollect(databases, func(dbType DBType, db *gorm.DB) (DBType, *gorm.DB) {
		return dbType, db
	})
}

// RawGormDB returns a borrowed DB, or nil when unavailable. Only the first optional dialect is used.
// Without an argument it prefers MySQL, otherwise PostgreSQL; callers must not close the shared pool.
func RawGormDB(dbType ...DBType) *gorm.DB {
	runtime := gormRuntimeState.Load()
	if runtime == nil {
		return nil
	}
	if len(dbType) == 0 {
		return runtime.databases[runtime.defaultDBType]
	}
	return runtime.databases[dbType[0]]
}

// RawMysqlGormDB returns the borrowed MySQL DB, or nil when unavailable.
func RawMysqlGormDB() *gorm.DB { return RawGormDB(DBTypeMySQL) }

// RawPostgresGormDB returns the borrowed PostgreSQL DB, or nil when unavailable.
func RawPostgresGormDB() *gorm.DB { return RawGormDB(DBTypePostgres) }
