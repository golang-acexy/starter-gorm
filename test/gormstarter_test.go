package test

import (
	"fmt"
	"github.com/golang-acexy/starter-gorm/gormmodule"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"testing"
	"time"
)

var moduleLoaders []declaration.ModuleLoader
var gModule *gormmodule.GormModule

func init() {
	gModule = &gormmodule.GormModule{
		Username: "root",
		Password: "root",
		Database: "test",
		Host:     "127.0.0.1",
		Port:     13306,
		GormInterceptor: func(instance interface{}) {
			raw := instance.(*gorm.DB)
			raw.Logger.LogMode(logger.Info)
		},
	}
	moduleLoaders = []declaration.ModuleLoader{gModule}
}

func TestRegisterGorm(t *testing.T) {

	m := declaration.Module{ModuleLoaders: moduleLoaders}
	err := m.Load()
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	db := gormmodule.RawDB()

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

	time.Sleep(7 * time.Second)
	r := m.Unload(10)
	fmt.Printf("%+v\n", r)
}
