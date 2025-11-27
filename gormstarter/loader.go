package gormstarter

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-parent/parent"
	"gorm.io/gorm"
)

const defaultCharset = "utf8mb4"

// 管理多类型数据库操作实例
var gormDBs map[DBType]*gorm.DB
var defaultDBType DBType
var sqlLoggerLevel logger.Level

func init() {
	gormDBs = make(map[DBType]*gorm.DB)
}

type GormConfig struct {
	Username string
	Password string
	Host     string
	Port     uint
	Database string

	Charset string // default charset : utf8mb4
	DBType  DBType // 数据库类型 不指定时默认为 mysql

	TimeUTC       bool         // true: create/update UTC time; false LOCAL time
	DryRun        bool         // create sql not exec
	SQLoggerLevel logger.Level // 仅当不使用默认日志时，才生效 仅指定为InfoLevel	DebugLevel	TraceLevel 时才生效，默认为 DebugLevel

	// MYSQL 配置
	MySQLUrlParam string // more Param such as `allowNativePasswords=false&checkConnLiveness=false`  https://github.com/go-sql-driver/mysql?tab=readme-ov-file#dsn-data-source-name

	// Postgres 配置
	PostgresTimezone  string
	PostgresEnableSSl bool

	InitFunc func(instance *gorm.DB)
}

type GormStarter struct {
	// Config 配置
	Config GormConfig
	// 懒加载函数，用于在实际执行时动态获取配置 该权重高于Config的直接配置
	LazyConfig  func() GormConfig
	config      *GormConfig
	GormSetting *parent.Setting
}

func (g *GormStarter) getConfig() *GormConfig {
	if g.config == nil {
		if g.LazyConfig != nil {
			lazyGormConfig := g.LazyConfig()
			g.config = &lazyGormConfig
			g.config.DBType = lazyGormConfig.DBType
			if g.config.DBType == "" {
				g.config.DBType = DBTypeMySQL
			}
		} else {
			g.config = &g.Config
			if g.config.DBType == "" {
				g.config.DBType = DBTypeMySQL
			}
		}
	}
	if g.config.SQLoggerLevel < logger.InfoLevel {
		sqlLoggerLevel = logger.DebugLevel
	} else {
		sqlLoggerLevel = g.config.SQLoggerLevel
	}
	return g.config
}

func (g *GormStarter) Setting() *parent.Setting {
	if g.GormSetting != nil {
		return g.GormSetting
	}
	config := g.getConfig()
	return parent.NewSetting("Gorm-Starter: "+string(config.DBType), 20, true, time.Second*30, func(instance any) {
		if config.InitFunc != nil {
			config.InitFunc(instance.(*gorm.DB))
		}
	})
}

func (g *GormStarter) Start() (any, error) {
	var err error
	config := g.getConfig()
	rawGormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		DryRun:                                   config.DryRun,
	}
	rawGormConfig.Logger = &logrusLogger{}
	if config.TimeUTC {
		rawGormConfig.NowFunc = func() time.Time {
			return time.Now().UTC()
		}
	}
	if config.DBType == "" {
		config.DBType = DBTypeMySQL
	}
	if config.Charset == "" {
		config.Charset = defaultCharset
	}
	if defaultDBType == "" {
		defaultDBType = config.DBType
	}
	_, ok := gormDBs[config.DBType]
	if ok {
		return nil, errors.New("database type " + string(config.DBType) + " already exist")
	}
	gormDB, err := openDB(config, rawGormConfig)
	if err != nil {
		return nil, err
	}
	sqlDb, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	err = g.ping(sqlDb)
	if err != nil {
		return nil, err
	}
	gormDBs[config.DBType] = gormDB
	return gormDB, nil
}

func (g *GormStarter) ping(sqlDb *sql.DB) error {
	if sqlDb == nil {
		return nil
	}
	return sqlDb.Ping()
}

func (g *GormStarter) closedAllConn(sqlDb *sql.DB) bool {
	if sqlDb == nil {
		return true
	}
	s := sqlDb.Stats()
	if s.Idle == 0 && s.InUse == 0 && s.OpenConnections == 0 {
		return true
	}
	return false
}

func (g *GormStarter) Stop(maxWaitTime time.Duration) (gracefully, stopped bool, err error) {
	dbType := g.getConfig().DBType
	gormDB := gormDBs[dbType]
	sqlDb, err := gormDB.DB()
	if err != nil {
		return false, g.ping(sqlDb) != nil, err
	}
	err = sqlDb.Close()
	if err != nil {
		return false, g.ping(sqlDb) != nil, err
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		for {
			if g.closedAllConn(sqlDb) {
				cancelFunc()
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
	select {
	case <-ctx.Done():
		stopped = g.ping(sqlDb) != nil
		gracefully = true
	case <-time.After(maxWaitTime):
		stopped = g.ping(sqlDb) != nil
		gracefully = false
	}
	return
}

// RawGormDB 获取 gorm.DB原始能力，如果多数据库类型初始化后，不指定DBType默认返回最先加载的数据库类型
func RawGormDB(dbType ...DBType) *gorm.DB {
	if len(dbType) == 0 {
		return gormDBs[defaultDBType]
	}
	return gormDBs[dbType[0]]
}

// RawMysqlGormDB 获取 mysql 数据库类型的 gorm.DB
func RawMysqlGormDB() *gorm.DB {
	return RawGormDB(DBTypeMySQL)
}

// RawPostgresGormDB 获取 postgres 数据库类型的 gorm.DB
func RawPostgresGormDB() *gorm.DB {
	return RawGormDB(DBTypePostgres)
}
