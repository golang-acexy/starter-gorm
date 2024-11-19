package gormstarter

import (
	"context"
	"database/sql"
	"errors"
	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-parent/parent"
	"gorm.io/gorm"
	"time"
)

// 管理多类型数据库操作实例
var gormDBs map[DBType]*gorm.DB
var defaultDBType DBType

func init() {
	gormDBs = make(map[DBType]*gorm.DB)
}

const (
	defaultCharset = "utf8mb4"
)

type GormConfig struct {
	Username string
	Password string
	Host     string
	Port     uint
	Database string

	Charset string // default charset : utf8mb4
	DBType  DBType // 数据库类型 不指定时默认为 mysql

	TimeUTC       bool // true: create/update UTC time; false LOCAL time
	DryRun        bool // create sql not exec
	UseDefaultLog bool

	// MYSQL 配置
	MySQLUrlParam string // more Param such as `allowNativePasswords=false&checkConnLiveness=false`  https://github.com/go-sql-driver/mysql?tab=readme-ov-file#dsn-data-source-name

	// Postgres 配置
	PostgresTimezone  string
	PostgresEnableSSl bool
}

type GormStarter struct {

	// GormConfig 配置
	GormConfig GormConfig

	// 懒加载函数，用于在实际执行时动态获取配置 该权重高于GormConfig的直接配置
	LazyGormConfig func() GormConfig

	GormSetting *parent.Setting
	InitFunc    func(instance *gorm.DB)

	lazyGormConfig *GormConfig
}

func (g *GormStarter) Setting() *parent.Setting {
	if g.GormSetting != nil {
		return g.GormSetting
	}
	if g.GormConfig.DBType == "" {
		if g.LazyGormConfig != nil {
			lazyGormConfig := g.LazyGormConfig()
			g.lazyGormConfig = &lazyGormConfig
			g.GormConfig.DBType = lazyGormConfig.DBType
			if g.GormConfig.DBType == "" {
				g.GormConfig.DBType = DBTypeMySQL
			}
		} else {
			g.GormConfig.DBType = DBTypeMySQL
		}
	}
	return parent.NewSetting("Gorm-Starter: "+string(g.GormConfig.DBType), 20, false, time.Second*30, func(instance interface{}) {
		if g.InitFunc != nil {
			g.InitFunc(instance.(*gorm.DB))
		}
	})
}

func (g *GormStarter) Start() (interface{}, error) {
	var err error
	if g.LazyGormConfig != nil {
		if g.lazyGormConfig == nil {
			g.GormConfig = g.LazyGormConfig()
		} else {
			g.GormConfig = *g.lazyGormConfig
		}
	}
	config := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		DryRun:                                   g.GormConfig.DryRun,
	}
	if !g.GormConfig.UseDefaultLog {
		config.Logger = &logrusLogger{logger.Logrus()}
	}
	if g.GormConfig.TimeUTC {
		config.NowFunc = func() time.Time {
			return time.Now().UTC()
		}
	}
	if g.GormConfig.DBType == "" {
		g.GormConfig.DBType = DBTypeMySQL
	}
	if g.GormConfig.Charset == "" {
		g.GormConfig.Charset = defaultCharset
	}
	if defaultDBType == "" {
		defaultDBType = g.GormConfig.DBType
	}
	_, ok := gormDBs[g.GormConfig.DBType]
	if ok {
		return nil, errors.New("database type " + string(g.GormConfig.DBType) + " already exist")
	}
	gormDB, err := openDB(g.GormConfig, config)
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
	gormDBs[g.GormConfig.DBType] = gormDB
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
	dbType := g.GormConfig.DBType
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
