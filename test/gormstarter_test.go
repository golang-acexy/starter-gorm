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
			LazyGromConfig: func() gormstarter.GromConfig {
				return gormstarter.GromConfig{
					Username: "root",
					Password: "root",
					Database: "test",
					Host:     "127.0.0.1",
					Port:     13306,
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
					var v int
					tx := db.Raw("SELECT SLEEP(5)").Scan(&v)
					if tx.Error != nil {
						fmt.Printf("%+v \n", tx.Error)
						return
					}
				}
			}()
		}
	}()

	time.Sleep(3 * time.Second)
	stopResult, err := starterLoader.Stop(time.Second * 10)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	fmt.Println(json.ToJsonFormat(stopResult))
}
