package mysql

import (
	"fmt"
	"testing"
	"time"

	"github.com/acexy/golang-toolkit/logger"
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-parent/parent"
	"gorm.io/gorm"
)

var starterLoader *parent.StarterLoader

func init() {
	logger.EnableConsole(logger.TraceLevel, false)
	starterLoader = parent.NewStarterLoader([]parent.Starter{
		&gormstarter.GormStarter{
			LazyConfig: func() gormstarter.GormConfig {
				return gormstarter.GormConfig{
					Username:      "root",
					Password:      "root",
					Database:      "test",
					Host:          "127.0.0.1",
					Port:          13306,
					SQLoggerLevel: logger.InfoLevel,
					InitFunc: func(instance *gorm.DB) {
						fmt.Println(instance.Config)
					},
				}
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
					tx := db.Raw("SELECT SLEEP(1)").Scan(&v)
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
