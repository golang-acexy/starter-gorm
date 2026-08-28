//go:build integration

package multipledb

import (
	"os"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-parent/parent"
	"gorm.io/gorm"
)

var starterLoader *parent.StarterLoader

func init() {
	logger.EnableConsole(logger.TraceLevel)
	starterLoader = parent.InitStarterLoader([]parent.Starter{
		&gormstarter.GormStarter{
			Config: gormstarter.GormConfig{
				MySQL: &gormstarter.MySQLConfig{
					DatabaseConfig: gormstarter.DatabaseConfig{
						Username: os.Getenv("STARTER_GORM_MYSQL_USERNAME"),
						Password: os.Getenv("STARTER_GORM_MYSQL_PASSWORD"),
						Database: os.Getenv("STARTER_GORM_MYSQL_DATABASE"),
						Host:     os.Getenv("STARTER_GORM_MYSQL_HOST"),
						Port:     13306,
					},
				},
				Postgres: &gormstarter.PostgresConfig{
					DatabaseConfig: gormstarter.DatabaseConfig{
						Username: os.Getenv("STARTER_GORM_POSTGRES_USERNAME"),
						Password: os.Getenv("STARTER_GORM_POSTGRES_PASSWORD"),
						Database: os.Getenv("STARTER_GORM_POSTGRES_DATABASE"),
						Host:     os.Getenv("STARTER_GORM_POSTGRES_HOST"),
						Port:     5432,
					},
				},
				InitFunc: func(instance map[gormstarter.DBType]*gorm.DB) {
				},
			},
		},
	})
}
