package test

import (
	"fmt"
	"github.com/acexy/golang-toolkit/util/json"
	"testing"
)

func TestBaseSaveOne(t *testing.T) {
	bm := TeacherMapper{}
	teacher := Teacher{Name: "mapper", Age: 12, Sex: 1, ClassNo: 12}
	fmt.Println(bm.Save(&teacher, "ClassNo"))
	fmt.Println("saved id", teacher.ID)
}

func TestBaseSave(t *testing.T) {
	bm := TeacherMapper{}
	teacher := Teacher{Name: "mapper", Age: 12, Sex: 1}
	fmt.Println(bm.Save(&teacher))
	fmt.Println("saved id", teacher.ID)

	// 测试自动保存0值
	teacher1 := Teacher{Sex: 1}
	fmt.Println(bm.Save(&teacher1))
	fmt.Println("saved id", teacher1.ID)

	// 测试排除指定的字段
	teacher3 := Teacher{Sex: 1}
	fmt.Println(bm.Save(&teacher3, "name"))
	fmt.Println("saved id", teacher3.ID)

	// 测试主键冲突
	teacher4 := Teacher{
		Sex: 1,
	}
	teacher4.ID = 16
	fmt.Println(bm.Save(&teacher4, "name"))
	fmt.Println("saved id", teacher4.ID)

	// updateAndUpdate
	teacher5 := Teacher{
		Sex:  1,
		Name: "name",
	}
	fmt.Println(bm.SaveOrUpdateByPrimaryKey(&teacher5, "create_time"))
	fmt.Println("saved id", teacher5.ID)
}

func TestBatch(t *testing.T) {
	teacher := Teacher{Name: "mapper", Age: 12, Sex: 1}
	teacher1 := Teacher{Sex: 1}
	v := []*Teacher{&teacher, &teacher1}
	bm := TeacherMapper{}
	bm.SaveBatch(&v, "create_time")

}
func TestModifyById(t *testing.T) {
	bm := TeacherMapper{}
	updated := Teacher{Name: "update", Age: 21, Sex: 0}
	updated.ID = 47
	// 由于sex是零值并不会被用于更新的指定
	fmt.Println(bm.UpdateById(&updated))
	// 通过指定字段更新 可以指定零值
	fmt.Println(bm.UpdateById(&updated, "sex", "name", "age"))

	fmt.Println(bm.UpdateByIdWithoutZeroField(&updated, "sex"))
}

func TestModifyMapById(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.UpdateByIdUseMap(map[string]any{"name": "Miss A", "sex": 0}, 132))
}

func TestModifyByWhere(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.UpdateByWhere(&Teacher{Name: "Alex", Age: 0}, "name = ? and age > ?", "mapper", 5))
}

func TestRemoveById(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.DeleteById(1))
}

func TestRemoveByWhere(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.DeleteByWhere("name = ? and age > ?", "Alex", 5))
}

func TestRemoveByCondition(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.DeleteByCond(&Teacher{
		Name: "mapper",
		Age:  12,
		Sex:  1,
	}))
}

func TestModifyByCondition(t *testing.T) {
	bm := TeacherMapper{}
	updated := Teacher{Name: "1", Age: 0}
	condition := Teacher{Name: "2", Age: 0}
	fmt.Println(bm.UpdateByCond(&updated, &condition))
}

func TestQueryById(t *testing.T) {
	bm := TeacherMapper{}
	var teacher Teacher
	fmt.Println(bm.SelectById(1, &teacher))
	fmt.Println(teacher)
}

func TestQueryByIds(t *testing.T) {
	bm := TeacherMapper{}
	var teacher []Teacher
	fmt.Println(bm.SelectByIds([]interface{}{34, 36}, &teacher))
	fmt.Println(teacher)
}

func TestQueryByCondition(t *testing.T) {
	bm := TeacherMapper{}
	teachers := new([]*Teacher)
	// 由于Age是零值，不会用于查询
	bm.SelectByCond(&Teacher{Sex: 1, Age: 0}, teachers, "age")
	fmt.Println(json.ToJsonFormat(teachers))
}

func TestQueryByWhere(t *testing.T) {
	bm := TeacherMapper{}
	teachers := new([]*Teacher)
	bm.SelectByWhere("name =? and age > ?", teachers, "mapper", 5)
	fmt.Println(teachers)
}

func TestQueryByConditionMap(t *testing.T) {
	bm := TeacherMapper{}
	teachers := new([]*Teacher)
	bm.SelectByCondMap(map[string]any{"sex": 0}, teachers)
	for _, teacher := range *teachers {
		fmt.Printf("%+v\n", *teacher)
	}
}

func TestPageCondition(t *testing.T) {
	bm := TeacherMapper{}
	teachers := new([]*Teacher)
	fmt.Println(bm.SelectPageByCond(&Teacher{Name: "mapper"}, 3, 2, teachers))
	for _, teacher := range *teachers {
		fmt.Printf("%+v\n", *teacher)
	}
}

func TestPageConditionMap(t *testing.T) {
	bm := TeacherMapper{}
	teachers := new([]*Teacher)
	fmt.Println(bm.SelectPageByCondMap(map[string]any{"sex": 0}, 2, 2, teachers))
	for _, teacher := range *teachers {
		fmt.Printf("%+v\n", *teacher)
	}
}

func TestUpdateByCondWithZeroField(t *testing.T) {
	bm := TeacherMapper{}
	updated := Teacher{Name: "1", Age: 0}
	condition := Teacher{Name: "2", Age: 0}
	fmt.Println(bm.UpdateByCondWithZeroField(&updated, &condition, []string{"age"}))
}

func TestUpdateByCondMap(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.UpdateByCondMap(map[string]any{"age": 0}, map[string]any{"age": 12}))
}
