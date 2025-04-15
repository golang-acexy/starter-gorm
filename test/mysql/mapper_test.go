package mysql

import (
	"fmt"
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-gorm/test/model"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestBaseSaveOne(t *testing.T) {
	bm := model.TeacherMapper{}
	teacher := model.Teacher{Name: "mapper", Age: 12, Sex: 1, ClassNo: 12}
	fmt.Println(bm.Save(&teacher, "ClassNo"))
	fmt.Println("saved id", teacher.ID)

}

func TestSaveWithoutZero(t *testing.T) {
	bm := model.TeacherMapper{}
	teacher := model.Teacher{Name: "mapper", Age: 12, Sex: 0, ClassNo: 12}
	fmt.Println(bm.SaveWithoutZeroField(&teacher))
	fmt.Println("saved id", teacher.ID)
}

func TestBaseSave(t *testing.T) {
	bm := model.TeacherMapper{}
	teacher := model.Teacher{Name: "mapper", Age: 12, Sex: 0}
	fmt.Println(bm.Save(&teacher))
	fmt.Println("saved id", teacher.ID)
	fmt.Println(bm.SaveWithoutZeroField(&teacher))
	fmt.Println("saved id", teacher.ID)

	// 测试自动保存0值
	teacher1 := model.Teacher{Sex: 1}
	fmt.Println(bm.Save(&teacher1))
	fmt.Println("saved id", teacher1.ID)

	// 测试排除指定的字段
	teacher3 := model.Teacher{Sex: 1}
	fmt.Println(bm.Save(&teacher3, "name"))
	fmt.Println("saved id", teacher3.ID)

	// 测试主键冲突
	teacher4 := model.Teacher{
		Sex: 1,
	}
	teacher4.ID = 16
	fmt.Println(bm.Save(&teacher4, "name"))
	fmt.Println("saved id", teacher4.ID)

	// updateAndUpdate
	teacher5 := model.Teacher{
		Sex:  1,
		Name: "name",
	}
	fmt.Println(bm.SaveOrUpdateByPrimaryKey(&teacher5, "create_time"))
	fmt.Println("saved id", teacher5.ID)
}

func TestBatch(t *testing.T) {
	teacher := model.Teacher{Name: "mapper", Age: 12, Sex: 1}
	teacher1 := model.Teacher{Sex: 1}
	v := []*model.Teacher{&teacher, &teacher1}
	bm := model.TeacherMapper{}
	bm.SaveBatch(&v, "create_time")

}
func TestModifyById(t *testing.T) {
	bm := model.TeacherMapper{}
	updated := model.Teacher{Name: "update", Age: 21, Sex: 0}
	updated.ID = 47
	// 由于sex是零值并不会被用于更新的指定
	fmt.Println(bm.UpdateById(&updated))
	// 通过指定字段更新 可以指定零值
	fmt.Println(bm.UpdateById(&updated, "sex", "name", "age"))

	fmt.Println(bm.UpdateByIdWithoutZeroField(&updated, "sex"))
}

func TestModifyMapById(t *testing.T) {
	bm := model.TeacherMapper{}
	fmt.Println(bm.UpdateByIdUseMap(map[string]any{"name": "Miss A", "sex": 0}, 132))
}

func TestModifyByWhere(t *testing.T) {
	bm := model.TeacherMapper{}
	fmt.Println(bm.UpdateByWhere(&model.Teacher{Name: "Alex", Age: 0}, "name = ? and age > ?", "mapper", 5))
}

func TestRemoveById(t *testing.T) {
	bm := model.TeacherMapper{}
	fmt.Println(bm.DeleteById(1))
}

func TestRemoveByWhere(t *testing.T) {
	bm := model.TeacherMapper{}
	fmt.Println(bm.DeleteByWhere("name = ? and age > ?", "Alex", 5))
}

func TestRemoveByCondition(t *testing.T) {
	bm := model.TeacherMapper{}
	fmt.Println(bm.DeleteByCond(&model.Teacher{
		Name: "mapper",
		Age:  12,
		Sex:  1,
	}))
}

func TestRemoveByMap(t *testing.T) {
	var bm model.TeacherMapper
	fmt.Println(bm.DeleteByMap(map[string]any{"name": "mapper", "sex": 1}))
}

func TestModifyByCondition(t *testing.T) {
	bm := model.TeacherMapper{}
	updated := model.Teacher{Name: "1", Age: 0}
	condition := model.Teacher{Name: "2", Age: 0}
	fmt.Println(bm.UpdateByCond(&updated, &condition))
}

func TestQueryById(t *testing.T) {
	bm := model.TeacherMapper{}
	var teacher model.Teacher
	fmt.Println(bm.SelectById(1, &teacher))
	fmt.Println(json.ToJson(teacher))
}

func TestQueryByIds(t *testing.T) {
	bm := model.TeacherMapper{}
	var teachers []*model.Teacher
	fmt.Println(bm.SelectByIds([]interface{}{34, 36}, &teachers))
	fmt.Println(json.ToJsonFormat(teachers))
}

func TestQueryByCondition(t *testing.T) {
	bm := model.TeacherMapper{}
	var teachers []*model.Teacher
	// 由于Age是零值，不会用于查询
	//bm.SelectByCond(&Teacher{Sex: 1, Age: 0}, &teachers, "age")
	bm.SelectByCond(&model.Teacher{Sex: 1, Age: 0}, "id desc", &teachers)
	fmt.Println(json.ToJsonFormat(teachers))
}

func TestQueryByWhere(t *testing.T) {
	bm := model.TeacherMapper{}
	teachers := new([]*model.Teacher)
	bm.SelectByWhere("name =? and age > ?", "", teachers, "mapper", 5)
	fmt.Println(teachers)
}

func TestQueryByGorm(t *testing.T) {
	var bm model.TeacherMapper
	teachers := new([]*model.Teacher)
	row, _ := bm.SelectByGorm(teachers, func(db *gorm.DB) {
		db.Where("create_time < ?", time.Now())
	})
	fmt.Println(row)
	fmt.Println(json.ToJsonFormat(teachers))
}

func TestQueryOneByGorm(t *testing.T) {
	var bm model.TeacherMapper
	var teacher model.Teacher
	row, _ := bm.SelectOneByGorm(&teacher, func(db *gorm.DB) {
		db.Where("id = 8")
	})
	fmt.Println(row)
	fmt.Println(json.ToJsonFormat(teacher))
}

func TestQueryByConditionMap(t *testing.T) {
	bm := model.TeacherMapper{}
	teachers := new([]*model.Teacher)
	bm.SelectByMap(map[string]any{"sex": 0}, "", teachers)
	fmt.Println(json.ToJsonFormat(teachers))
	for _, teacher := range *teachers {
		fmt.Printf("%+v\n", *teacher)
	}
}

func TestPageCondition(t *testing.T) {
	bm := model.TeacherMapper{}
	teachers := new([]*model.Teacher)
	fmt.Println(bm.SelectPageByCond(&model.Teacher{Sex: 1}, "", 2, 3, teachers))
	for _, teacher := range *teachers {
		fmt.Printf("%+v\n", *teacher)
	}
}

func TestPageConditionMap(t *testing.T) {
	bm := model.TeacherMapper{}
	teachers := new([]*model.Teacher)
	fmt.Println(bm.SelectPageByMap(map[string]any{"sex": 0}, "", 2, 2, teachers))
	for _, teacher := range *teachers {
		fmt.Printf("%+v\n", *teacher)
	}
}

func TestUpdateByCondWithZeroField(t *testing.T) {
	bm := model.TeacherMapper{}
	updated := model.Teacher{Name: "1", Age: 0}
	condition := model.Teacher{Name: "2", Age: 0}
	fmt.Println(bm.UpdateByCondWithZeroField(&updated, &condition, []string{"ClassNo"}))
}

func TestUpdateByCondMap(t *testing.T) {
	bm := model.TeacherMapper{}
	fmt.Println(bm.UpdateByMap(map[string]any{"age": 0}, map[string]any{"age": 12}))
}

func TestCount(t *testing.T) {
	var bm model.TeacherMapper
	fmt.Println(bm.CountByMap(map[string]any{"age": 0}))
	fmt.Println(bm.CountByCond(&model.Teacher{
		Age: 1,
	}))
}

func TestTransaction(t *testing.T) {
	var mp model.TeacherMapper
	mpTx := model.TeacherMapper{}
	tx := gormstarter.RawGormDB().Begin()
	mpTx.Tx = tx
	teacher := model.Teacher{Name: "mapper", Age: 12, Sex: 1, ClassNo: 12}
	fmt.Println(mp.Save(&teacher))
	teacher = model.Teacher{Name: "mapper", Age: 12, Sex: 1, ClassNo: 13}
	fmt.Println(mpTx.Save(&teacher))
	fmt.Println(mpTx.Save(&teacher))
	tx.Commit()
}
