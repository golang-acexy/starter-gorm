package test

import (
	"os"
	"testing"
	"time"

	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-parent/parent"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var starterLoader *parent.StarterLoader

func init() {
	starterLoader = parent.InitStarterLoader([]parent.Starter{
		&gormstarter.GormStarter{
			LazyConfig: func() gormstarter.GormConfig {
				return gormstarter.GormConfig{
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
						instance[gormstarter.DBTypePostgres].Logger.LogMode(logger.Info)
					},
				}
			},
		},
	})
}

func TestMain(m *testing.M) {
	err := starterLoader.Start()
	if err != nil {
		os.Exit(1)
	}
	code := m.Run()
	_, err = starterLoader.StopAllByRegisteredOrder(10 * time.Second)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(code)
}

func TestRegisterGorm(t *testing.T) {
	if gormstarter.RawPostgresGormDB() == nil {
		t.Fatal("postgres database is not initialized")
	}
}
