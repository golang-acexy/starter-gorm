package test

import (
	"fmt"
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
	fmt.Println(bm.UpdateById(&updated))
}

func TestModifyMapById(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.UpdateWithMapById(map[string]any{"name": "Miss A", "sex": 0}, 132))
}

func TestModifyByWhere(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.UpdateByWhere(&Teacher{Name: "Alex"}, "name = ? and age > ?", "mapper", 5))
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
	fmt.Println(bm.DeleteByCondition(&Teacher{
		Name: "mapper",
		Age:  12,
		Sex:  1,
	}))
}

func TestModifyByCondition(t *testing.T) {
	bm := TeacherMapper{}
	updated := Teacher{Name: "1", Age: 12}
	condition := Teacher{Name: "2", Age: 1}
	fmt.Println(bm.UpdateByCondition(&updated, &condition))
}

func TestQueryById(t *testing.T) {
	bm := TeacherMapper{}
	var teacher Teacher
	fmt.Println(bm.SelectById(1, &teacher))
	fmt.Println(teacher)
}

func TestQueryByCondition(t *testing.T) {
	bm := TeacherMapper{}
	teachers := new([]*Teacher)
	// 由于Age是零值，不会用于查询
	bm.SelectByCondition(&Teacher{Name: "王五", Sex: 1, Age: 0}, teachers)
	fmt.Println(teachers)
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
	bm.SelectByConditionMap(map[string]any{"sex": 0}, teachers)
	for _, teacher := range *teachers {
		fmt.Printf("%+v\n", *teacher)
	}
}

func TestPageCondition(t *testing.T) {
	bm := TeacherMapper{}
	teachers := new([]*Teacher)
	fmt.Println(bm.SelectPageByCondition(&Teacher{Name: "mapper"}, 3, 2, teachers))
	for _, teacher := range *teachers {
		fmt.Printf("%+v\n", *teacher)
	}
}

func TestPageConditionMap(t *testing.T) {
	bm := TeacherMapper{}
	teachers := new([]*Teacher)
	fmt.Println(bm.SelectPageByConditionMap(map[string]any{"sex": 0}, 2, 2, teachers))
	for _, teacher := range *teachers {
		fmt.Printf("%+v\n", *teacher)
	}
}
