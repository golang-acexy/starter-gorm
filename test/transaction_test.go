package test

import (
	"fmt"
	"github.com/golang-acexy/starter-gorm/gormmodule"
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"gorm.io/gorm"
	"testing"
)

func init() {
	m := declaration.Module{ModuleLoaders: moduleLoaders}
	_ = m.Load()
}

func TestStudentSave(t *testing.T) {
	tx := gormmodule.NewTransaction(true)
	tx.Save(&Student{Name: "张三"})
	tx.Save(&Teacher{Name: "王五"})
	tx.ModifyById(Student{ID: 1}, Student{Name: "赵老五", Sex: 1})
	tx.ModifyByCondition(Student{Name: "叶良辰", Sex: 0}, "name = ?", "王麻子") // sex 零值不能更新
	tx.ModifyByConditionMap(Student{}, map[string]interface{}{
		"name": "叶良辰",
		"sex":  0,
	}, "name = ?", "王麻子") // 使用map防止零值不更新
	tx.Customize(func(tx *gorm.DB) (bool, error) {
		_ = tx.Exec("update demo_student set name = '老赵哥'")
		return true, nil
	})
	tx.RemoveById(&Student{Name: "张三"})

	err := tx.Execute()
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}
