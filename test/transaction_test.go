package test

import (
	"fmt"
	"github.com/golang-acexy/starter-gorm/gormmodule"
	"gorm.io/gorm"
	"testing"
)

func TestTransaction(t *testing.T) {

	// 开启一个事务 该事务的每一步都将立即执行，通过tx.Execute() 最终抉择是否需要提交
	tx := gormmodule.NewTransaction(false)

	i := new([]Teacher)
	tx.QueryByCondition(Teacher{Name: "王五"}, i)
	fmt.Printf("%+v\n", i)

	i = new([]Teacher)
	tx.QueryByConditionMap(Teacher{}, map[string]any{"sex": 0}, i)
	fmt.Printf("%+v\n", i)

	tx.Save(&Student{Name: "张三"})
	teacher := &Teacher{Name: "王五"}
	tx.Save(teacher)
	fmt.Printf("%+v\n", teacher)

	queryTeacher := new(Teacher)
	tx.QueryById(queryTeacher, teacher.ID)
	fmt.Printf("%+v\n", queryTeacher)

	tx.ModifyById(Student{ID: 1}, Student{Name: "赵老五", Sex: 1})
	tx.ModifyByCondition(Student{Name: "叶良辰", Sex: 0}, "name = ? and id = ?", "王麻子", 1) // sex 零值不能更新
	tx.ModifyByConditionMap(Student{}, map[string]interface{}{
		"name": "叶良辰",
		"sex":  0,
	}, "name = ? and id = ?", "王麻子", 11111) // 使用map防止零值不更新

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

func TestTransactionChain(t *testing.T) {

	// 开启一个事务 该事务的每一步事务操作并不会立即执行 通过tx.Execute() 最终执行所有事务链步骤，并抉择是否提交
	tx := gormmodule.NewTransactionChain(true)

	tx.Save(&Student{Name: "张三"})
	teacher := &Teacher{Name: "王五"}
	tx.Save(teacher)
	fmt.Printf("由于此时tx.Save并没有执行，所以只能获取零值 teacher %+v\n", teacher)

	queryTeacher := new(Teacher)
	tx.QueryById(queryTeacher, teacher.ID)
	fmt.Printf("%+v\n", queryTeacher)

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
