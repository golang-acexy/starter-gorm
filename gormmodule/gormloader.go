package gormmodule

import (
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type DBType string

const (
	defaultCharset = "utf8mb4"
)
const (
	DBTypeMySQL = "mysql"
)

type GormModule struct {
	db *gorm.DB

	Username string
	Password string
	Host     string
	Port     uint8
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
		UnregisterPriority:       1,
		UnregisterAllowAsync:     true,
		UnregisterMaxWaitSeconds: 20,
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
	db, err := gorm.Open(mysql.Open(g.toDsn()))
	if err != nil {
		return err
	}
	if interceptor != nil {
		(*interceptor)(db)
	}
	g.db = db
	return nil
}

func (g *GormModule) Unregister(maxWaitSeconds uint) (gracefully bool, err error) {
	return false, nil
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
