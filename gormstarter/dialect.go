package gormstarter

import (
	"errors"
	"fmt"
	"github.com/acexy/golang-toolkit/util/str"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

func openDB(config GormConfig, gormConfig *gorm.Config) (*gorm.DB, error) {
	switch config.DBType {
	case DBTypeMySQL:
		return openMysqlDB(config, gormConfig)
	case DBTypePostgres:
		return openPostgresDB(config, gormConfig)
	}
	return nil, errors.New("not supported database type now")
}

// OpenMysqlDB 创建Mysql数据库连接
func openMysqlDB(config GormConfig, gormConfig *gorm.Config) (*gorm.DB, error) {
	builder := str.NewBuilder(config.Username)
	builder.WriteString(":").WriteString(config.Password).WriteString("@tcp(").WriteString(config.Host).WriteString(":").WriteString(strconv.Itoa(int(config.Port)))
	builder.WriteString(")/").WriteString(config.Database).WriteString("?charset=" + config.Charset)
	builder.WriteString("&parseTime=True") // support time.Time
	if config.MySQLUrlParam != "" {
		if strings.HasPrefix(config.MySQLUrlParam, "&") || strings.HasPrefix(config.MySQLUrlParam, "?") {
			builder.WriteString("&").WriteString(str.Substring(config.MySQLUrlParam, 1, str.CharLength(config.MySQLUrlParam)))
		}
	}
	return gorm.Open(mysql.Open(builder.ToString()), gormConfig)
}

func openPostgresDB(config GormConfig, gormConfig *gorm.Config) (*gorm.DB, error) {
	sslModel := "disable"
	if config.PostgresEnableSSl {
		sslModel = "enable"
	}
	timeZone := "UTC"
	if config.PostgresTimezone != "" {
		timeZone = config.PostgresTimezone
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		config.Host,
		config.Username,
		config.Password,
		config.Database,
		config.Port,
		sslModel,
		timeZone,
	)
	return gorm.Open(postgres.Open(dsn), gormConfig)
}
