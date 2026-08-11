package mysql

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang-acexy/starter-gorm/gormstarter"
)

func TestWrapperAllCapabilities(t *testing.T) {
	mapper := TeacherMapper{}
	c := mapper.Columns()
	scope := fmt.Sprintf("w%x", time.Now().UnixNano()&0xff)
	cleanup := func() {
		if _, err := mapper.DeleteByWhere("name LIKE ?", scope+"%"); err != nil {
			t.Fatalf("清理 Wrapper 测试数据失败: %v", err)
		}
	}
	cleanup()
	defer cleanup()

	teachers := []*Teacher{
		{Name: scope + "_alpha", Sex: 0, Age: 10, ClassNo: 1},
		{Name: scope + "_beta", Sex: 1, Age: 20, ClassNo: 1},
		{Name: scope + "_gamma", Sex: 0, Age: 30, ClassNo: 2},
		{Name: scope + "_delta", Sex: 1, Age: 40, ClassNo: 2},
		{Name: scope + "_omega", Sex: 0, Age: 50, ClassNo: 3},
	}
	if rows, err := mapper.InsertBatch(teachers); err != nil || rows != int64(len(teachers)) {
		t.Fatalf("准备 Wrapper 测试数据失败: rows=%d err=%v", rows, err)
	}

	base := c.Name.HasPrefix(scope)
	assertCount := func(name string, expected int64, predicates ...gormstarter.Predicate[Teacher]) {
		t.Helper()
		query := mapper.Wrapper().Where(append([]gormstarter.Predicate[Teacher]{base}, predicates...)...)
		actual, err := mapper.CountByWrapper(query)
		if err != nil || actual != expected {
			t.Fatalf("%s: count=%d expected=%d err=%v", name, actual, expected, err)
		}
	}
	assertSimpleCount := func(name string, expected int64, build func(*gormstarter.QueryWrapper[Teacher])) {
		t.Helper()
		query := mapper.Wrapper().HasPrefix(c.Name, scope)
		build(query)
		actual, err := mapper.CountByWrapper(query)
		if err != nil || actual != expected {
			t.Fatalf("%s: count=%d expected=%d err=%v", name, actual, expected, err)
		}
	}

	assertSimpleCount("Eq", 1, func(query *gormstarter.QueryWrapper[Teacher]) { query.Eq(c.Age, 20) })
	assertSimpleCount("Ne", 4, func(query *gormstarter.QueryWrapper[Teacher]) { query.Ne(c.Age, 20) })
	assertSimpleCount("Gt", 2, func(query *gormstarter.QueryWrapper[Teacher]) { query.Gt(c.Age, 30) })
	assertSimpleCount("Ge", 3, func(query *gormstarter.QueryWrapper[Teacher]) { query.Ge(c.Age, 30) })
	assertSimpleCount("Lt", 2, func(query *gormstarter.QueryWrapper[Teacher]) { query.Lt(c.Age, 30) })
	assertSimpleCount("Le", 3, func(query *gormstarter.QueryWrapper[Teacher]) { query.Le(c.Age, 30) })
	assertSimpleCount("In", 2, func(query *gormstarter.QueryWrapper[Teacher]) { query.In(c.Age, 10, 50) })
	assertSimpleCount("NotIn", 3, func(query *gormstarter.QueryWrapper[Teacher]) { query.NotIn(c.Age, 10, 50) })
	assertSimpleCount("Between", 3, func(query *gormstarter.QueryWrapper[Teacher]) { query.Between(c.Age, 20, 40) })
	assertSimpleCount("IsNull", 0, func(query *gormstarter.QueryWrapper[Teacher]) { query.IsNull(c.Name) })
	assertSimpleCount("IsNotNull", 5, func(query *gormstarter.QueryWrapper[Teacher]) { query.IsNotNull(c.Name) })
	assertSimpleCount("Like", 1, func(query *gormstarter.QueryWrapper[Teacher]) { query.Like(c.Name, "%_alpha") })
	assertSimpleCount("NotLike", 4, func(query *gormstarter.QueryWrapper[Teacher]) { query.NotLike(c.Name, "%_alpha") })
	assertSimpleCount("HasSuffix", 1, func(query *gormstarter.QueryWrapper[Teacher]) { query.HasSuffix(c.Name, "omega") })
	assertSimpleCount("Contains", 1, func(query *gormstarter.QueryWrapper[Teacher]) { query.Contains(c.Name, "gamma") })
	assertCount("And", 2, gormstarter.And(c.Age.Ge(20), c.Age.Le(30)))
	assertCount("Or", 2, gormstarter.Or(c.Age.Eq(10), c.Age.Eq(50)))
	assertCount("Not", 4, gormstarter.Not(c.Age.Eq(10)))
	assertCount("Nested", 3, gormstarter.Or(
		gormstarter.And(c.Sex.Eq(0), c.Age.Lt(40)),
		gormstarter.And(c.Sex.Eq(1), c.ClassNo.Eq(2)),
	))

	listQuery := mapper.Wrapper().
		HasPrefix(c.Name, scope).
		Select(c.ID, c.Name).
		OrderByDesc(c.Age).
		Limit(2).
		Offset(1)
	var records []*Teacher
	if _, err := mapper.SelectByWrapper(listQuery, &records); err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || records[0].Name != scope+"_delta" || records[1].Name != scope+"_gamma" {
		t.Fatalf("投影、排序或范围查询结果异常: %+v", records)
	}
	if records[0].Age != 0 {
		t.Fatalf("Select 投影未生效: %+v", records[0])
	}

	oneQuery := mapper.Wrapper().HasPrefix(c.Name, scope).OrderByAsc(c.Age)
	var first Teacher
	if rows, err := mapper.SelectOneByWrapper(oneQuery, &first); err != nil || rows != 1 || first.Age != 10 {
		t.Fatalf("单条 Wrapper 查询异常: rows=%d teacher=%+v err=%v", rows, first, err)
	}

	count, err := mapper.CountByWrapper(listQuery)
	if err != nil || count != 5 {
		t.Fatalf("Wrapper 计数不应应用投影、排序、Limit 或 Offset: count=%d err=%v", count, err)
	}

	pageQuery := mapper.PageWrapper(2, 2).HasPrefix(c.Name, scope).OrderByAsc(c.Age)
	var pageRecords []*Teacher
	total, err := mapper.SelectPageByWrapper(pageQuery, &pageRecords)
	if err != nil || total != 5 || len(pageRecords) != 2 || pageRecords[0].Age != 30 || pageRecords[1].Age != 40 {
		t.Fatalf("Wrapper 分页异常: total=%d records=%+v err=%v", total, pageRecords, err)
	}

	// CreatedAt 使用 column:create_time，能正常排序说明列名来自当前 GORM Schema。
	var tagRecords []*Teacher
	tagQuery := mapper.Wrapper().HasPrefix(c.Name, scope).OrderByDesc(c.CreatedAt).Limit(1)
	if _, err := mapper.SelectByWrapper(tagQuery, &tagRecords); err != nil || len(tagRecords) != 1 {
		t.Fatalf("GORM column tag 动态解析失败: records=%d err=%v", len(tagRecords), err)
	}
}

func TestWrapperValidation(t *testing.T) {
	mapper := TeacherMapper{}
	c := mapper.Columns()
	var records []*Teacher
	if _, err := mapper.SelectByWrapper(nil, &records); !errors.Is(err, gormstarter.ErrNilQueryWrapper) {
		t.Fatalf("nil Wrapper 错误 = %v", err)
	}
	if _, err := mapper.SelectPageByWrapper(mapper.PageWrapper(0, 1), &records); !errors.Is(err, gormstarter.ErrInvalidPage) {
		t.Fatalf("非法分页错误 = %v", err)
	}
	if _, err := mapper.SelectByWrapper(mapper.Wrapper().Limit(-1), &records); !errors.Is(err, gormstarter.ErrInvalidQueryRange) {
		t.Fatalf("非法 Limit 错误 = %v", err)
	}
	if _, err := mapper.SelectByWrapper(mapper.Wrapper().Offset(-1), &records); !errors.Is(err, gormstarter.ErrInvalidQueryRange) {
		t.Fatalf("非法 Offset 错误 = %v", err)
	}
	if _, err := mapper.SelectByWrapper(mapper.Wrapper().Eq(c.Ignored, "ignored"), &records); !errors.Is(err, gormstarter.ErrNonPersistentColumn) {
		t.Fatalf("非持久化字段错误 = %v", err)
	}
}

func TestUpdateWrapper(t *testing.T) {
	mapper := TeacherMapper{}
	c := mapper.Columns()
	scope := fmt.Sprintf("u%x", time.Now().UnixNano()&0xff)
	cleanup := func() {
		if _, err := mapper.DeleteByWhere("name LIKE ?", scope+"%"); err != nil {
			t.Fatalf("清理 UpdateWrapper 测试数据失败: %v", err)
		}
	}
	cleanup()
	defer cleanup()

	teachers := []*Teacher{
		{Name: scope + "a", Sex: 1, Age: 10, ClassNo: 1},
		{Name: scope + "b", Sex: 1, Age: 20, ClassNo: 1},
		{Name: scope + "c", Sex: 1, Age: 30, ClassNo: 1},
	}
	if rows, err := mapper.InsertBatch(teachers); err != nil || rows != 3 {
		t.Fatalf("准备 UpdateWrapper 测试数据失败: rows=%d err=%v", rows, err)
	}

	update := mapper.UpdateWrapper().
		HasPrefix(c.Name, scope).
		Where(gormstarter.Or(
			c.Age.Eq(10),
			gormstarter.And(c.Age.Ge(20), c.Age.Lt(30)),
		)).
		Set(c.Sex, 0).
		Set(c.ClassNo, 8).
		Set(c.ClassNo, 9)
	rows, err := mapper.UpdateByWrapper(update)
	if err != nil || rows != 2 {
		t.Fatalf("复杂 UpdateWrapper 更新失败: rows=%d err=%v", rows, err)
	}

	var updated []*Teacher
	query := mapper.Wrapper().HasPrefix(c.Name, scope).Eq(c.ClassNo, 9).OrderByAsc(c.Age)
	if _, err = mapper.SelectByWrapper(query, &updated); err != nil {
		t.Fatal(err)
	}
	if len(updated) != 2 || updated[0].Age != 10 || updated[1].Age != 20 || updated[0].Sex != 0 || updated[1].Sex != 0 {
		t.Fatalf("UpdateWrapper 零值或重复 Set 未生效: %+v", updated)
	}
}

func TestUpdateWrapperValidation(t *testing.T) {
	mapper := TeacherMapper{}
	c := mapper.Columns()
	if _, err := mapper.UpdateByWrapper(nil); !errors.Is(err, gormstarter.ErrNilUpdateWrapper) {
		t.Fatalf("nil UpdateWrapper 错误 = %v", err)
	}
	if _, err := mapper.UpdateByWrapper(mapper.UpdateWrapper().Set(c.Sex, 0)); !errors.Is(err, gormstarter.ErrEmptyCondition) {
		t.Fatalf("空更新条件错误 = %v", err)
	}
	if _, err := mapper.UpdateByWrapper(mapper.UpdateWrapper().Eq(c.ID, 1)); !errors.Is(err, gormstarter.ErrNoFieldToUpdate) {
		t.Fatalf("空 Set 错误 = %v", err)
	}
	if _, err := mapper.UpdateByWrapper(mapper.UpdateWrapper().Eq(c.ID, 1).Set(c.Ignored, "ignored")); !errors.Is(err, gormstarter.ErrNonPersistentColumn) {
		t.Fatalf("非持久化更新字段错误 = %v", err)
	}
}
