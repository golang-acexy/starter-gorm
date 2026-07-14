package mysql

import (
	"os"
	"testing"
	"time"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-parent/parent"
)

var starterLoader *parent.StarterLoader

func init() {
	logger.EnableConsole(logger.DebugLevel)
	starterLoader = parent.InitStarterLoader([]parent.Starter{
		&gormstarter.GormStarter{
			LazyConfig: func() gormstarter.GormConfig {
				return gormstarter.GormConfig{
					MySQL: &gormstarter.MySQLConfig{
						DatabaseConfig: gormstarter.DatabaseConfig{
					Username:      "root",
					Password:      "root",
					Database:      "test",
					Host:          "127.0.0.1",
					Port:          13306,
						},
					},
					SQLLoggerLevel: logger.ErrorLevel,
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
	if gormstarter.RawMysqlGormDB() == nil {
		t.Fatal("mysql database is not initialized")
	}
}
