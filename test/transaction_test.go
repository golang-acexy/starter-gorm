package test

import (
	"fmt"
	"github.com/golang-acexy/starter-gorm/gormmodule"
	"gorm.io/gorm"
	"testing"
)

func TestTransaction(t *testing.T) {

	// 开启一个事务
	tx := gormmodule.NewTransaction(true)

	tx.Save(&Student{Name: "张三"})
	teacher := &Teacher{Name: "王五"}
	//teacher.ID = 123
	tx.Save(teacher)

	tx.ModifyById(Student{ID: 1}, Student{Name: "赵老五", Sex: 1})
	tx.ModifyByCondition(Student{Name: "叶良辰", Sex: 0}, "name = ? and id = ?", "王麻子", 1) // sex 零值不能更新
	tx.ModifyByConditionMap(Student{}, map[string]interface{}{
		"name": "叶良辰",
		"sex":  0,
	}, "name = ? and id = ?", "王麻子", 1) // 使用map防止零值不更新

	// 自定义事务逻辑
	tx.Customize(func(tx *gorm.DB) (int64, error) {
		_ = tx.Exec("update demo_student set name = '老赵哥'")
		return 1, nil
	})

	// 移出单条
	tx.RemoveById(Student{ID: 1})

	// 移出多条
	tx.RemoveById([]Student{{ID: 1}, {ID: 2}})
	tx.RemoveByCondition(Teacher{}, "name = ? and id = ?", "张三", 1)

	// 执行事务
	err := tx.Execute()

	if err != nil {
		fmt.Printf("%+v\n", err)
	}
}
