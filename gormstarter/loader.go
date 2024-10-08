package gormstarter

import (
	"context"
	"database/sql"
	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/util/str"
	"github.com/golang-acexy/starter-parent/parent"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

type DBType string

var gormDB *gorm.DB

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

	TimeUTC       bool // true: create/update UTC time; false LOCAL time
	DryRun        bool // create sql not exec
	UseDefaultLog bool

	UrlParam string // more Param such as `allowNativePasswords=false&checkConnLiveness=false`  https://github.com/go-sql-driver/mysql?tab=readme-ov-file#dsn-data-source-name
}

type GormStarter struct {

	// GormConfig 配置
	GormConfig GormConfig

	// 懒加载函数，用于在实际执行时动态获取配置 该权重高于GormConfig的直接配置
	LazyGromConfig func() GormConfig

	GormSetting *parent.Setting
	InitFunc    func(instance *gorm.DB)
}

func (g *GormStarter) Setting() *parent.Setting {
	if g.GormSetting != nil {
		return g.GormSetting
	}
	return parent.NewSetting("Gorm-Starter", 20, false, time.Second*30, func(instance interface{}) {
		if g.InitFunc != nil {
			g.InitFunc(instance.(*gorm.DB))
		}
	})
}

func (g *GormStarter) Start() (interface{}, error) {
	var err error
	if g.LazyGromConfig != nil {
		g.GormConfig = g.LazyGromConfig()
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
	gormDB, err = gorm.Open(mysql.Open(g.toDsn()), config)
	if err != nil {
		return nil, err
	}
	sqlDb, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	return gormDB, g.ping(sqlDb)
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

func (g *GormStarter) toDsn() string {
	if g.GormConfig.Charset == "" {
		g.GormConfig.Charset = defaultCharset
	}
	builder := str.NewBuilder(g.GormConfig.Username)
	builder.WriteString(":").WriteString(g.GormConfig.Password).WriteString("@tcp(").WriteString(g.GormConfig.Host).WriteString(":").WriteString(strconv.Itoa(int(g.GormConfig.Port)))
	builder.WriteString(")/").WriteString(g.GormConfig.Database)
	builder.WriteString("?charset=" + g.GormConfig.Charset)
	builder.WriteString("&parseTime=True") // support time.Time
	if g.GormConfig.UrlParam != "" {
		if strings.HasPrefix(g.GormConfig.UrlParam, "&") || strings.HasPrefix(g.GormConfig.UrlParam, "?") {
			builder.WriteString("&" + str.Substring(g.GormConfig.UrlParam, 1, str.CharLength(g.GormConfig.UrlParam)))
		}
	}
	return builder.ToString()
}

func RawGormDB() *gorm.DB {
	return gormDB
}
