package test

import (
	"context"
	"fmt"
	"github.com/golang-acexy/starter-gorm/gormmodule"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"testing"
)

func init() {
	m := declaration.Module{ModuleLoaders: moduleLoaders}
	_ = m.Load()
}

func TestSelect(t *testing.T) {
	var sutdent Student
	gormmodule.RawDB().Model(Student{}).Where(&Student{Name: "1"}).Scan(&sutdent)
}

func TestSave(t *testing.T) {
	// 分别处于不通的事务中
	stu := &Student{Name: "王麻子"}
	result := gormmodule.RawDB().Create(stu)
	fmt.Println(result.Error, result.RowsAffected, stu.ID)

	//
	stu = &Student{Name: "王麻子1"}
	result = gormmodule.RawDB().Create(stu)
	fmt.Println(result.Error, result.RowsAffected, stu.ID)

	// withContext 分别处于不通的事务中
	db := gormmodule.RawDB().WithContext(context.Background())
	stu = &Student{Name: "王麻子2"}
	result = db.Create(stu)
	fmt.Println(result.Error, result.RowsAffected, stu.ID)

	stu = &Student{Name: "王麻子4"}
	result = db.Create(stu)
	fmt.Println(result.Error, result.RowsAffected, stu.ID)
}

func TestUpdate(t *testing.T) {
	result := gormmodule.RawDB().Model(Student{}).Where("name = ? and id = ?", "王麻子", 1).Update("name", "张三")
	fmt.Println(result.Error, result.RowsAffected)
	result = gormmodule.RawDB().Model(Student{ID: 1}).Updates(Student{ID: 1111, Name: "张三", Sex: 0}) // sex = 0 是零值，不会用于更新
	fmt.Println(result.Error, result.RowsAffected)
}
