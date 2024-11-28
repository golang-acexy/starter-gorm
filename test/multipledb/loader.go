package multipledb

import (
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-parent/parent"
)

var starterLoader *parent.StarterLoader

func init() {
	starterLoader = parent.NewStarterLoader([]parent.Starter{
		&gormstarter.GormStarter{
			Config: gormstarter.GormConfig{
				Username: "root",
				Password: "root",
				Database: "test",
				Host:     "127.0.0.1",
				Port:     13306,
			},
		},
		&gormstarter.GormStarter{
			LazyConfig: func() gormstarter.GormConfig {
				return gormstarter.GormConfig{
					Username: "postgres",
					Password: "tech-acexy",
					Database: "postgres",
					Host:     "127.0.0.1",
					Port:     5432,
					DBType:   gormstarter.DBTypePostgres,
				}
			},
		},
	})

	_ = starterLoader.Start()
}
