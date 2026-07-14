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
	driverConfig := drivermysql.Config{
		User:      config.Username,
		Passwd:    config.Password,
		Net:       "tcp",
		Addr:      net.JoinHostPort(config.Host, strconv.Itoa(int(config.Port))),
		DBName:    config.Database,
		ParseTime: true,
		Params:    map[string]string{"charset": config.Charset},
	}
	for key, values := range params {
		if len(values) > 0 && !strings.EqualFold(key, "parseTime") && !strings.EqualFold(key, "charset") {
			driverConfig.Params[key] = values[len(values)-1]
		}
	}
	return gorm.Open(gormmysql.Open(driverConfig.FormatDSN()), gormConfig)
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
