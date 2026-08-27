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

type DatabaseConfig struct {
	Username string
	Password string
	Host     string
	Port     uint
	Database string
	TimeUTC  bool // 是否使用 UTC 时间生成创建和更新时间
	DryRun   bool // 是否只生成 SQL 而不执行
}

type MySQLConfig struct {
	DatabaseConfig
	Charset   string // 字符集，默认为 utf8mb4
	URLParams string // 额外连接参数，例如 `allowNativePasswords=false&checkConnLiveness=false`
}

type PostgresConfig struct {
	DatabaseConfig
	Timezone string
	SSLMode  string // SSL 模式，例如 disable、require、verify-ca 或 verify-full
}

type GormConfig struct {
	SQLLoggerLevel logger.Level // 全局 SQL 日志级别，默认为 DebugLevel
	MySQL          *MySQLConfig
	Postgres       *PostgresConfig
	InitFunc       func(databases map[DBType]*gorm.DB)
}

type GormStarter struct {
	Config      GormConfig
	LazyConfig  func() GormConfig
	config      *GormConfig
	configOnce  sync.Once
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

// RawGormDB 获取 gorm.DB 原始能力；不指定 DBType 时优先返回 MySQL
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

// RawMysqlGormDB 获取 MySQL 数据库类型的 gorm.DB
func RawMysqlGormDB() *gorm.DB { return RawGormDB(DBTypeMySQL) }

// RawPostgresGormDB 获取 PostgreSQL 数据库类型的 gorm.DB
func RawPostgresGormDB() *gorm.DB { return RawGormDB(DBTypePostgres) }
