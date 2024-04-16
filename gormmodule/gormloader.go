package gormmodule

import (
	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

type DBType string

var db *gorm.DB

const (
	defaultCharset = "utf8mb4"
)

type GormModule struct {
	Username string
	Password string
	Host     string
	Port     uint
	Database string

	Charset string // default charset : utf8mb4

	TimeUTC       bool // true: create/update UTC time; false LOCAL time
	DryRun        bool // create sql not exec
	UseDefaultLog bool

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
	config := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		DryRun:                                   g.DryRun,
	}
	if !g.UseDefaultLog {
		config.Logger = &logrusLogger{logger.Logrus()}
	}
	if g.TimeUTC {
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
	if g.Charset == "" {
		g.Charset = defaultCharset
	}
	var builder strings.Builder
	builder.WriteString(g.Username + ":" + g.Password + "@tcp(" + g.Host + ":" + strconv.Itoa(int(g.Port)) + ")/" + g.Database)
	builder.WriteString("?charset=" + g.Charset)
	builder.WriteString("&parseTime=True") // support time.Time
	return builder.String()
}

func RawDB() *gorm.DB {
	return db
}
