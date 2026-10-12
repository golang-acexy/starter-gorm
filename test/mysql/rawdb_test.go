//go:build integration

package mysql

import (
	"context"
	"testing"

	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-gorm/test/model"
)

func TestSelect(t *testing.T) {
	var student *model.Student
	result := gormstarter.RawGormDB().Model(model.Student{}).Where(&model.Student{Name: "王麻子"}).Scan(&student)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

func TestRawDBInsert(t *testing.T) {
	// Use separate transaction handles.
	stu := &model.Student{Name: "王麻子"}
	result := gormstarter.RawGormDB().Create(stu)
	if result.Error != nil {
		t.Fatal(result.Error)
	}

	//
	stu = &model.Student{Name: "王麻子1"}
	result = gormstarter.RawGormDB().Create(stu)
	if result.Error != nil {
		t.Fatal(result.Error)
	}

	// Use distinct contexts on separate transaction handles.
	db := gormstarter.RawGormDB().WithContext(context.Background())
	stu = &model.Student{Name: "王麻子2"}
	result = db.Create(stu)
	if result.Error != nil {
		t.Fatal(result.Error)
	}

	stu = &model.Student{Name: "王麻子4"}
	result = db.Create(stu)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

func TestUpdate(t *testing.T) {
	result := gormstarter.RawGormDB().Model(model.Student{}).Where("name = ? and id = ?", "王麻子", 1).Update("name", "张三")
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	result = gormstarter.RawGormDB().Model(model.Student{ID: 1}).Updates(model.Student{ID: 1111, Name: "张三", Sex: 0}) // GORM omits the zero-valued Sex field from this struct update.
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}
