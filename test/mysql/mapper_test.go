//go:build integration

package mysql

import (
	"fmt"
	"testing"
	"time"

	"github.com/acexy/golang-toolkit/util/json"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-gorm/test/model"
	"gorm.io/gorm"
)

func TestInsert(t *testing.T) {
	bm := model.TeacherMapper{}
	teacher := model.Teacher{Name: "mapper", Age: 12, Sex: 1, ClassNo: 12}
	fmt.Println(bm.Insert(&teacher, "ClassNo"))
	fmt.Println("saved id", teacher.ID)

}

func TestInsertWithoutZeroFields(t *testing.T) {
	bm := model.TeacherMapper{}
	teacher := model.Teacher{Name: "mapper", Age: 12, Sex: 0, ClassNo: 12}
	fmt.Println(bm.InsertWithoutZeroFields(&teacher))
	fmt.Println("saved id", teacher.ID)
}

func TestInsertWithMap(t *testing.T) {
	bm := model.TeacherMapper{}
	fmt.Println(bm.InsertWithMap(map[string]any{"name": "mapper", "age": 12, "sex": 1, "class_no": 12}))
}

func TestInsertVariants(t *testing.T) {
	bm := model.TeacherMapper{}
	teacher := model.Teacher{Name: "mapper", Age: 12, Sex: 0}
	fmt.Println(bm.Insert(&teacher))
	fmt.Println("saved id", teacher.ID)
	fmt.Println(bm.InsertWithoutZeroFields(&teacher))
	fmt.Println("saved id", teacher.ID)

	// 测试自动保存0值
	teacher1 := model.Teacher{Sex: 1}
	fmt.Println(bm.Insert(&teacher1))
	fmt.Println("saved id", teacher1.ID)

	// 测试排除指定的字段
	teacher3 := model.Teacher{Sex: 1}
	fmt.Println(bm.Insert(&teacher3, "name"))
	fmt.Println("saved id", teacher3.ID)

	// 测试主键冲突
	teacher4 := model.Teacher{
		Sex: 1,
	}
	teacher4.ID = 16
	fmt.Println(bm.Insert(&teacher4, "name"))
	fmt.Println("saved id", teacher4.ID)

	// updateAndUpdate
	teacher5 := model.Teacher{
		Sex:  1,
		Name: "name",
	}
	fmt.Println(bm.InsertOrUpdateByPrimaryKey(&teacher5, "create_time"))
	fmt.Println("saved id", teacher5.ID)
}

func TestBatch(t *testing.T) {
	teacher := model.Teacher{Name: "mapper", Age: 12, Sex: 1}
	teacher1 := model.Teacher{Sex: 1}
	v := []*model.Teacher{&teacher, &teacher1}
	bm := model.TeacherMapper{}
	bm.InsertBatch(v, "create_time")

}
func TestUpdateByID(t *testing.T) {
	bm := model.TeacherMapper{}
	updated := model.Teacher{Name: "update", Age: 21, Sex: 0}

	updated.ID = 47
	// 由于sex是零值并不会被用于更新的指定
	fmt.Println(bm.UpdateByID(&updated, updated.ID))
	// 通过指定字段更新 可以指定零值
	fmt.Println(bm.UpdateByID(&updated, updated.ID, "sex", "name", "age"))

	fmt.Println(bm.UpdateByIDWithoutZeroFields(&updated, updated.ID, "sex"))
}

func TestUpdateByIDWithMap(t *testing.T) {
	bm := model.TeacherMapper{}
	fmt.Println(bm.UpdateByIDWithMap(map[string]any{"name": "Miss A", "sex": 0}, 132))
}

func TestUpdateZeroFieldsSelection(t *testing.T) {
	bm := model.TeacherMapper{}
	teacher := model.Teacher{Name: "zf", Age: 21, Sex: 1, ClassNo: 3}
	if _, err := bm.Insert(&teacher); err != nil {
		t.Fatal(err)
	}
	defer bm.DeleteByID(teacher.ID)

	updatedByID := model.Teacher{ID: teacher.ID, Name: "zfid", Age: 0, Sex: 0}
	if _, err := bm.UpdateByIDWithoutZeroFields(&updatedByID, updatedByID.ID, "sex"); err != nil {
		t.Fatal(err)
	}
	var result model.Teacher
	if _, err := bm.SelectByID(teacher.ID, &result); err != nil {
		t.Fatal(err)
	}
	if result.Name != updatedByID.Name || result.Age != teacher.Age || result.Sex != 0 {
		t.Fatalf("unexpected ID update result: %+v", result)
	}

	updatedByCond := model.Teacher{Name: "zfcond", ClassNo: 0}
	condition := model.Teacher{ID: teacher.ID}
	if _, err := bm.UpdateByCondWithZeroFields(&updatedByCond, condition, "class_no"); err != nil {
		t.Fatal(err)
	}
	result = model.Teacher{}
	if _, err := bm.SelectByID(teacher.ID, &result); err != nil {
		t.Fatal(err)
	}
	if result.Name != updatedByCond.Name || result.Age != teacher.Age || result.ClassNo != 0 {
		t.Fatalf("unexpected condition update result: %+v", result)
	}
}

func TestUpdateByWhere(t *testing.T) {
	bm := model.TeacherMapper{}
	fmt.Println(bm.UpdateByWhere(&model.Teacher{Name: "Alex", Age: 0}, "name = ? and age > ?", "mapper", 5))
}

func TestDeleteByID(t *testing.T) {
	bm := model.TeacherMapper{}
	fmt.Println(bm.DeleteByID(1))
}

func TestDeleteByIDs(t *testing.T) {
	bm := model.TeacherMapper{}
	if _, err := bm.DeleteByIDs([]any{1, 2}); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteByWhere(t *testing.T) {
	bm := model.TeacherMapper{}
	fmt.Println(bm.DeleteByWhere("name = ? and age > ?", "Alex", 5))
}

func TestDeleteByCond(t *testing.T) {
	bm := model.TeacherMapper{}
	fmt.Println(bm.DeleteByCond(model.Teacher{
		Name: "mapper",
		Age:  12,
		Sex:  1,
	}))
}

func TestDeleteByMap(t *testing.T) {
	var bm model.TeacherMapper
	fmt.Println(bm.DeleteByMap(map[string]any{"name": "mapper", "sex": 1}))
}

func TestUpdateByCond(t *testing.T) {
	bm := model.TeacherMapper{}
	updated := model.Teacher{Name: "1", Age: 0}
	condition := model.Teacher{Name: "2", Age: 0}
	fmt.Println(bm.UpdateByCond(&updated, condition))
}

func TestSelectByID(t *testing.T) {
	bm := model.TeacherMapper{}
	var teacher model.Teacher
	fmt.Println(bm.SelectByID(1, &teacher))
	fmt.Println(json.ToString(teacher))
}

func TestSelectByIDs(t *testing.T) {
	bm := model.TeacherMapper{}
	var teachers []*model.Teacher
	fmt.Println(bm.SelectByIDs([]any{1, 2}, &teachers))
	fmt.Println(json.ToStringFormat(teachers))
}

func TestExistsByID(t *testing.T) {
	bm := model.TeacherMapper{}
	teacher := model.Teacher{Name: "exists", Age: 18}
	if _, err := bm.Insert(&teacher); err != nil {
		t.Fatal(err)
	}
	defer bm.DeleteByID(teacher.ID)

	exists, err := bm.ExistsByID(teacher.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected inserted teacher to exist")
	}

	if _, err = bm.DeleteByID(teacher.ID); err != nil {
		t.Fatal(err)
	}
	exists, err = bm.ExistsByID(teacher.ID)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("expected deleted teacher not to exist")
	}
}

func TestSelectByCond(t *testing.T) {
	bm := model.TeacherMapper{}
	var teachers []*model.Teacher
	// 由于Age是零值，不会用于查询
	//bm.SelectByCond(&Teacher{Sex: 1, Age: 0}, &teachers, "age")
	bm.SelectByCond(gormstarter.CondQuery[model.Teacher]{Condition: model.Teacher{Sex: 1, Age: 0}, QueryOptions: gormstarter.QueryOptions{OrderBySQL: "id desc"}}, &teachers)
	fmt.Println(json.ToStringFormat(teachers))
}

func TestSelectByWhere(t *testing.T) {
	bm := model.TeacherMapper{}
	teachers := new([]*model.Teacher)
	bm.SelectByWhere(gormstarter.WhereQuery{RawWhereSQL: "name =? and age > ?", Args: []any{"mapper", 5}}, teachers)
	fmt.Println(teachers)
}

func TestSelectByGorm(t *testing.T) {
	var bm model.TeacherMapper
	teachers := new([]*model.Teacher)
	row, _ := bm.SelectByGorm(teachers, func(db *gorm.DB) {
		db.Where("create_time < ?", time.Now())
	})
	fmt.Println(row)
	fmt.Println(json.ToStringFormat(teachers))
}

func TestSelectOneByGorm(t *testing.T) {
	var bm model.TeacherMapper
	var teacher model.Teacher
	row, _ := bm.SelectOneByGorm(&teacher, func(db *gorm.DB) {
		db.Where("id = 8")
	})
	fmt.Println(row)
	fmt.Println(json.ToStringFormat(teacher))
}

func TestSelectByMap(t *testing.T) {
	bm := model.TeacherMapper{}
	teachers := new([]*model.Teacher)
	bm.SelectByMap(gormstarter.MapQuery{Condition: map[string]any{"sex": 0}}, teachers)
	fmt.Println(json.ToStringFormat(teachers))
	for _, teacher := range *teachers {
		fmt.Printf("%+v\n", *teacher)
	}
}

func TestQueryWrapper(t *testing.T) {
	columns := struct {
		ID        gormstarter.Column[model.Teacher, uint64]
		Name      gormstarter.Column[model.Teacher, string]
		Age       gormstarter.Column[model.Teacher, uint]
		CreatedAt gormstarter.Column[model.Teacher, gormstarter.Timestamp]
	}{
		ID:        gormstarter.NewColumn(func(teacher *model.Teacher) *uint64 { return &teacher.ID }),
		Name:      gormstarter.NewColumn(func(teacher *model.Teacher) *string { return &teacher.Name }),
		Age:       gormstarter.NewColumn(func(teacher *model.Teacher) *uint { return &teacher.Age }),
		CreatedAt: gormstarter.NewColumn(func(teacher *model.Teacher) *gormstarter.Timestamp { return &teacher.CreatedAt }),
	}

	mapper := model.TeacherMapper{}
	query := mapper.Wrapper().
		Where(gormstarter.Or(columns.Age.Ge(0), columns.Name.Contains("mapper"))).
		Select(columns.ID, columns.Name).
		OrderByDesc(columns.CreatedAt).
		Limit(2)

	var teachers []*model.Teacher
	if _, err := mapper.SelectByWrapper(query, &teachers); err != nil {
		t.Fatal(err)
	}
	if len(teachers) > 2 {
		t.Fatalf("Wrapper Limit 未生效: %d", len(teachers))
	}
	total, err := mapper.CountByWrapper(query)
	if err != nil {
		t.Fatal(err)
	}
	if total < int64(len(teachers)) {
		t.Fatalf("Wrapper 计数不应应用 Limit: total=%d records=%d", total, len(teachers))
	}

	var pageRecords []*model.Teacher
	pageQuery := mapper.PageWrapper(1, 1).
		Where(gormstarter.Or(columns.Age.Ge(0), columns.Name.Contains("mapper"))).
		Select(columns.ID, columns.Name).
		OrderByDesc(columns.CreatedAt)
	pageTotal, err := mapper.SelectPageByWrapper(pageQuery, &pageRecords)
	if err != nil {
		t.Fatal(err)
	}
	if pageTotal != total || len(pageRecords) > 1 {
		t.Fatalf("Wrapper 分页结果异常: total=%d records=%d", pageTotal, len(pageRecords))
	}
}

func TestSelectPageByCond(t *testing.T) {
	bm := model.TeacherMapper{}
	teachers := new([]*model.Teacher)
	fmt.Println(bm.SelectPageByCond(gormstarter.PageQuery[model.Teacher]{
		Condition: model.Teacher{Sex: 1},
		PageOptions: gormstarter.PageOptions{
			Number: 2,
			Size:   3,
		},
	}, teachers))
	for _, teacher := range *teachers {
		fmt.Printf("%+v\n", *teacher)
	}
}

func TestSelectPageByMap(t *testing.T) {
	bm := model.TeacherMapper{}
	teachers := new([]*model.Teacher)
	fmt.Println(bm.SelectPageByMap(gormstarter.MapPageQuery{
		Condition:   map[string]any{"sex": 0},
		PageOptions: gormstarter.PageOptions{Number: 2, Size: 2},
	}, teachers))
	for _, teacher := range *teachers {
		fmt.Printf("%+v\n", *teacher)
	}
}

func TestSelectPageByGorm(t *testing.T) {
	bm := model.TeacherMapper{}
	teachers := new([]*model.Teacher)
	fmt.Println(bm.SelectPageByGorm(func(db *gorm.DB) {
		db.Where("name = 'mapper'")
	}, func(db *gorm.DB) {
		db.Where("name = 'mapper'").Limit(2)
	}, teachers))
	for _, teacher := range *teachers {
		fmt.Printf("%+v\n", *teacher)
	}
}

func TestUpdateByCondWithZeroFields(t *testing.T) {
	bm := model.TeacherMapper{}
	updated := model.Teacher{Name: "1", Age: 0}
	condition := model.Teacher{Name: "2", Age: 0}
	fmt.Println(bm.UpdateByCondWithZeroFields(&updated, condition, "ClassNo"))
}

func TestUpdateByMap(t *testing.T) {
	bm := model.TeacherMapper{}
	fmt.Println(bm.UpdateByMap(map[string]any{"age": 0}, map[string]any{"age": 12}))
}

func TestCount(t *testing.T) {
	var bm model.TeacherMapper
	fmt.Println(bm.CountByMap(gormstarter.MapQuery{Condition: map[string]any{"age": 0}}))
	fmt.Println(bm.CountByCond(gormstarter.CondQuery[model.Teacher]{Condition: model.Teacher{Age: 1}}))
}

func TestTransaction(t *testing.T) {
	var mp model.TeacherMapper
	db := gormstarter.RawMysqlGormDB()
	if db == nil {
		t.Fatal(gormstarter.ErrGormStarterNotStarted)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()

	mpTx := mp.WithTxMapper(tx)
	teacher := model.Teacher{Name: "mapper", Age: 12, Sex: 1, ClassNo: 12}
	if _, err := mpTx.Insert(&teacher); err != nil {
		t.Fatal(err)
	}
	var result model.Teacher
	rows, err := mpTx.SelectByID(teacher.ID, &result)
	if err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Fatalf("expected transaction to select 1 row, got %d", rows)
	}
	result = model.Teacher{}
	rows, err = mp.SelectByID(teacher.ID, &result)
	if err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatalf("uncommitted data is visible outside transaction")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatal(err)
	}
	result = model.Teacher{}
	rows, err = mp.SelectByID(teacher.ID, &result)
	if err != nil {
		t.Fatal(err)
	}
	if rows != 1 || result.ID != teacher.ID {
		t.Fatalf("committed data was not found")
	}
}
