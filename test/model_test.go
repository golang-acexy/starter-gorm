package test

import (
	"fmt"
	"github.com/golang-acexy/starter-gorm/gormmodule"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"testing"
)

func init() {
	var moduleLoaders []declaration.ModuleLoader
	var gModule *gormmodule.GormModule
	gModule = &gormmodule.GormModule{
		Username: "test",
		Password: "test",
		Database: "test",
		Host:     "127.0.0.1",
		Port:     3306,
		TimeUTC:  true,
	}
	moduleLoaders = []declaration.ModuleLoader{gModule}
	m := declaration.Module{ModuleLoaders: moduleLoaders}
	_ = m.Load()
}

func TestStudentSave(t *testing.T) {

	// 分别处于不通的事务中
	//stu := &Student{Name: "王麻子"}
	//result := gormmodule.RawDB().Create(stu)
	//fmt.Println(result.Error, result.RowsAffected, stu.ID)
	//
	//stu = &Student{Name: "王麻子1"}
	//result = gormmodule.RawDB().Create(stu)
	//fmt.Println(result.Error, result.RowsAffected, stu.ID)

	// withContext 分别处于不通的事务中
	//db := gormmodule.RawDB().WithContext(context.Background())
	//stu := &Student{Name: "王麻子"}
	//result := db.Create(stu)
	//fmt.Println(result.Error, result.RowsAffected, stu.ID)
	//stu = &Student{Name: "王麻子1"}
	//result = db.Create(stu)
	//fmt.Println(result.Error, result.RowsAffected, stu.ID)
}

func TestStudentUpdate(t *testing.T) {
	result := gormmodule.RawDB().Model(&Student{}).Where("name = ?", "王麻子").Update("name", "张三")
	fmt.Println(result.Error, result.RowsAffected)
}
