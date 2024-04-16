package gormmodule

import (
	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/util/str"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type DBType string

var db *gorm.DB

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
}

type GormModule struct {

	// GromConfig 配置
	GromConfig GromConfig

	// 懒加载函数，用于在实际执行时动态获取配置 该权重高于GormConfig的直接配置
	LazyGromConfig func() GromConfig

	GormModuleConfig *declaration.ModuleConfig
	GormInterceptor  func(instance *gorm.DB)
}

func (g *GormModule) ModuleConfig() *declaration.ModuleConfig {
	if g.GormModuleConfig != nil {
		return g.GormModuleConfig
	}
	return &declaration.ModuleConfig{
		ModuleName:               "Gorm",
		UnregisterPriority:       20,
		UnregisterAllowAsync:     false,
		UnregisterMaxWaitSeconds: 30,
		LoadInterceptor: func(instance interface{}) {
			if g.GormInterceptor != nil {
				g.GormInterceptor(instance.(*gorm.DB))
			}
		},
	}
}

func (g *GormModule) Register() (interface{}, error) {
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
	db, err = gorm.Open(mysql.Open(g.toDsn()), config)
	if err != nil {
		return nil, err
	}
	logger.Logrus().Traceln(g.ModuleConfig().ModuleName, "started")
	return db, nil
}

func (g *GormModule) Unregister(maxWaitSeconds uint) (gracefully bool, err error) {
	sqlDb, err := db.DB()
	if err != nil {
		return false, err
	}

	err = sqlDb.Close()
	if err != nil {
		return false, err
	}

	done := make(chan bool)

	go func() {
		for {
			s := sqlDb.Stats()
			logger.Logrus().Tracef("check db stats %+v", s)
			if s.Idle == 0 && s.InUse == 0 && s.OpenConnections == 0 {
				done <- true
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	select {
	case <-done:
		gracefully = true
	case <-time.After(time.Second * time.Duration(maxWaitSeconds)):
		gracefully = false
	}
	return
}

func (g *GormModule) toDsn() string {
	if g.GromConfig.Charset == "" {
		g.GromConfig.Charset = defaultCharset
	}
	builder := str.NewBuilder(g.GromConfig.Username + ":" + g.GromConfig.Password + "@tcp(" + g.GromConfig.Host + ":" + strconv.Itoa(int(g.GromConfig.Port)) + ")/" + g.GromConfig.Database)
	builder.WriteString("?charset=" + g.GromConfig.Charset)
	builder.WriteString("&parseTime=True") // support time.Time
	return builder.ToString()
}

func RawDB() *gorm.DB {
	return db
}
