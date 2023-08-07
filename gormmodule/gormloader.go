package gormmodule

import (
	"github.com/acexy/golang-toolkit/log"
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
const (
	DBTypeMySQL = "mysql"
)

type GormModule struct {
	Username string
	Password string
	Host     string
	Port     uint
	Database string

	// default charset : utf8mb4
	Charset string

	// only mysql now
	DBType DBType

	GormModuleConfig *declaration.ModuleConfig
	GormInterceptor  *func(instance interface{})
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
	}
}

// Interceptor 初始化gorm DB实例拦截器
// request instance: *gorm.DB
func (g *GormModule) Interceptor() *func(instance interface{}) {
	if g.GormInterceptor != nil {
		return g.GormInterceptor
	}
	return nil
}

func (g *GormModule) Register(interceptor *func(instance interface{})) error {
	var err error
	db, err = gorm.Open(mysql.Open(g.toDsn()))
	if err != nil {
		return err
	}
	if interceptor != nil {
		(*interceptor)(db)
	}
	db.Logger = &logrusLogger{
		log: log.Logrus(),
	}
	return nil
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
		// check stats
		for {
			s := sqlDb.Stats()
			log.Logrus().Tracef("check db stats %+v", s)
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
	if g.DBType == "" {
		g.DBType = DBTypeMySQL
	}
	if g.Charset == "" {
		g.Charset = defaultCharset
	}
	var builder strings.Builder
	switch g.DBType {
	case DBTypeMySQL:
		builder.WriteString(g.Username + ":" + g.Password + "@tcp(" + g.Host + ":" + strconv.Itoa(int(g.Port)) + ")/" + g.Database)
		builder.WriteString("?charset=" + g.Charset)
		builder.WriteString("&parseTime=True") // support time.Time
	}
	return builder.String()
}

func (g *GormModule) RawDB() *gorm.DB {
	return db
}
