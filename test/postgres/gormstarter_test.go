package test

import (
	"fmt"
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-parent/parent"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"testing"
	"time"
)

var starterLoader *parent.StarterLoader

func init() {
	starterLoader = parent.NewStarterLoader([]parent.Starter{
		&gormstarter.GormStarter{
			LazyGormConfig: func() gormstarter.GormConfig {
				return gormstarter.GormConfig{
					Username: "postgres",
					Password: "tech-acexy",
					Database: "postgres",
					Host:     "127.0.0.1",
					Port:     5432,
					DBType:   gormstarter.DBTypePostgres,
				}
			},

			InitFunc: func(instance *gorm.DB) {
				instance.Logger.LogMode(logger.Info)
			},
		},
	})
}

func TestRegisterGorm(t *testing.T) {

	err := starterLoader.Start()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	db := gormstarter.RawGormDB()

	// 启动一批协程，并执行延迟sql，模拟并发多连接执行中场景
	go func() {
		for i := 1; i <= 10; i++ {
			go func() {
				for {
					var v string
					tx := db.Raw("SELECT pg_sleep(1)").Scan(&v)
					if tx.Error != nil {
						fmt.Printf("%+v \n", tx.Error)
						return
					}
					fmt.Println(v)
				}
			}()
		}
	}()

	time.Sleep(5 * time.Second)
	stopResult, err := starterLoader.Stop(time.Second * 10)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	fmt.Println(json.ToJsonFormat(stopResult))
}
