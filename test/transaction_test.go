package test

import (
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
	mapper := gormmodule.NewTransaction()
	mapper.Save(&Student{Name: "张三"}).Save(&Teacher{Name: "王五"}).Modify(&Student{
		ID: 1,
	}, Student{
		Name: "赵老五",
	}).Execute(func(tx *gorm.DB) (bool, error) {
		_ = tx.Exec("update demo_student set name = '老赵哥'")
		return true, nil
	}).Execute()
}
