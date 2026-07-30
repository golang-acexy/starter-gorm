package gormstarter

import (
	"net"
	"net/url"
	"strconv"
	"strings"

	drivermysql "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// OpenMysqlDB 创建Mysql数据库连接
func openMysqlDB(config *MySQLConfig, gormConfig *gorm.Config) (*gorm.DB, error) {
	params, err := url.ParseQuery(strings.TrimLeft(config.URLParams, "?&"))
	if err != nil {
		return nil, err
	}

	// parseTime 和 charset 由 Starter 统一管理，不允许额外连接参数覆盖。
	for key := range params {
		if strings.EqualFold(key, "parseTime") || strings.EqualFold(key, "charset") {
			delete(params, key)
		}
	}

	// 使用驱动默认配置，保留 mysql_native_password 等默认兼容行为。
	driverConfig := drivermysql.NewConfig()
	driverConfig.User = config.Username
	driverConfig.Passwd = config.Password
	driverConfig.Net = "tcp"
	driverConfig.Addr = net.JoinHostPort(config.Host, strconv.Itoa(int(config.Port)))
	driverConfig.DBName = config.Database
	driverConfig.ParseTime = true
	driverConfig.Params = map[string]string{"charset": config.Charset}

	dsn := driverConfig.FormatDSN()
	if encodedParams := params.Encode(); encodedParams != "" {
		dsn += "&" + encodedParams
	}

	// 由驱动解析完整 DSN，确保驱动级连接参数写入对应配置字段。
	driverConfig, err = drivermysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	return gorm.Open(gormmysql.New(gormmysql.Config{DSNConfig: driverConfig}), gormConfig)
}

func openPostgresDB(config *PostgresConfig, gormConfig *gorm.Config) (*gorm.DB, error) {
	sslMode := config.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	timeZone := "UTC"
	if config.Timezone != "" {
		timeZone = config.Timezone
	}
	dsn := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(config.Username, config.Password),
		Host:   net.JoinHostPort(config.Host, strconv.Itoa(int(config.Port))),
		Path:   config.Database,
	}
	query := dsn.Query()
	query.Set("sslmode", sslMode)
	query.Set("TimeZone", timeZone)
	dsn.RawQuery = query.Encode()
	return gorm.Open(postgres.Open(dsn.String()), gormConfig)
}
