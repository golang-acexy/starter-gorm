package mysql

import (
	"os"
	"testing"
	"time"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-parent/parent"
)

func TestMain(m *testing.M) {
	logger.EnableConsole(logger.ErrorLevel)
	loader := parent.InitStarterLoader([]parent.Starter{
		&gormstarter.GormStarter{LazyConfig: func() gormstarter.GormConfig {
			return gormstarter.GormConfig{
				MySQL: &gormstarter.MySQLConfig{DatabaseConfig: gormstarter.DatabaseConfig{
					Username: "root",
					Password: "root",
					Database: "test",
					Host:     "127.0.0.1",
					Port:     13306,
				}},
				SQLLoggerLevel: logger.ErrorLevel,
			}
		}},
	})
	if err := loader.Start(); err != nil {
		os.Exit(1)
	}
	code := m.Run()
	if _, err := loader.StopAllByRegisteredOrder(10 * time.Second); err != nil {
		os.Exit(1)
	}
	os.Exit(code)
}
