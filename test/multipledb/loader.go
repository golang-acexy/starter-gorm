package multipledb

import (
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
						Username: "root",
						Password: "root",
						Database: "test",
						Host:     "127.0.0.1",
						Port:     13306,
					},
				},
				Postgres: &gormstarter.PostgresConfig{
					DatabaseConfig: gormstarter.DatabaseConfig{
						Username: "postgres",
						Password: "tech-acexy",
						Database: "postgres",
						Host:     "127.0.0.1",
						Port:     5432,
					},
				},
				InitFunc: func(instance map[gormstarter.DBType]*gorm.DB) {
				},
			},
		},
	})
}
