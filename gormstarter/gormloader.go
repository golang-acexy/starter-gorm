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

type GromConfig struct {
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

	// GromConfig 配置
	GromConfig GromConfig

	// 懒加载函数，用于在实际执行时动态获取配置 该权重高于GormConfig的直接配置
	LazyGromConfig func() GromConfig

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
		g.GromConfig = g.LazyGromConfig()
	}

	config := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		DryRun:                                   g.GromConfig.DryRun,
	}
	if !g.GromConfig.UseDefaultLog {
		config.Logger = &logrusLogger{logger.Logrus()}
	}
	if g.GromConfig.TimeUTC {
		config.NowFunc = func() time.Time {
			return time.Now().UTC()
		}
	}
	gormDB, err = gorm.Open(mysql.Open(g.toDsn()), config)
	if err != nil {
		return nil, err
	}
	return gormDB, nil
}

func (g *GormStarter) isClosed(sqlDb *sql.DB) bool {
	if sqlDb == nil {
		return true
	}
	if pingErr := sqlDb.Ping(); pingErr != nil {
		return true
	}
	return false
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
		return false, g.isClosed(sqlDb), err
	}
	err = sqlDb.Close()
	if err != nil {
		return false, g.isClosed(sqlDb), err
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
		stopped = g.isClosed(sqlDb)
		gracefully = true
	case <-time.After(maxWaitTime):
		stopped = g.isClosed(sqlDb)
		gracefully = false
	}
	return
}

func (g *GormStarter) toDsn() string {
	if g.GromConfig.Charset == "" {
		g.GromConfig.Charset = defaultCharset
	}
	builder := str.NewBuilder(g.GromConfig.Username + ":" + g.GromConfig.Password + "@tcp(" + g.GromConfig.Host + ":" + strconv.Itoa(int(g.GromConfig.Port)) + ")/" + g.GromConfig.Database)
	builder.WriteString("?charset=" + g.GromConfig.Charset)
	builder.WriteString("&parseTime=True") // support time.Time
	if g.GromConfig.UrlParam != "" {
		if strings.HasPrefix(g.GromConfig.UrlParam, "&") || strings.HasPrefix(g.GromConfig.UrlParam, "?") {
			builder.WriteString("&" + str.Substring(g.GromConfig.UrlParam, 1, str.CharLength(g.GromConfig.UrlParam)))
		}
	}
	return builder.ToString()
}

func RawGormDB() *gorm.DB {
	return gormDB
}
