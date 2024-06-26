package test

import (
	"context"
	"fmt"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"testing"
)

func init() {
	_ = starterLoader.Start()
}

func TestSelect(t *testing.T) {
	var sutdent Student
	gormstarter.RawGormDB().Model(Student{}).Where(&Student{Name: "1"}).Scan(&sutdent)
}

func TestSave(t *testing.T) {
	// 分别处于不通的事务中
	stu := &Student{Name: "王麻子"}
	result := gormstarter.RawGormDB().Create(stu)
	fmt.Println(result.Error, result.RowsAffected, stu.ID)

	//
	stu = &Student{Name: "王麻子1"}
	result = gormstarter.RawGormDB().Create(stu)
	fmt.Println(result.Error, result.RowsAffected, stu.ID)

	// withContext 分别处于不通的事务中
	db := gormstarter.RawGormDB().WithContext(context.Background())
	stu = &Student{Name: "王麻子2"}
	result = db.Create(stu)
	fmt.Println(result.Error, result.RowsAffected, stu.ID)

	stu = &Student{Name: "王麻子4"}
	result = db.Create(stu)
	fmt.Println(result.Error, result.RowsAffected, stu.ID)
}

func TestUpdate(t *testing.T) {
	result := gormstarter.RawGormDB().Model(Student{}).Where("name = ? and id = ?", "王麻子", 1).Update("name", "张三")
	fmt.Println(result.Error, result.RowsAffected)
	result = gormstarter.RawGormDB().Model(Student{ID: 1}).Updates(Student{ID: 1111, Name: "张三", Sex: 0}) // sex = 0 是零值，不会用于更新
	fmt.Println(result.Error, result.RowsAffected)
}
