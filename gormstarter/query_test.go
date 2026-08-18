package gormstarter

import (
	"reflect"
	"testing"
)

func TestQueryBuilders(t *testing.T) {
	rangeFilter := TimeRange{Field: "created_at"}
	cond := NewCondQuery(wrapperModel{ID: 1}).OrderBy("id desc").Select("id", "name").WithTimeRanges(rangeFilter).WithLimit(10)
	if cond.Condition.ID != 1 || cond.OrderBySQL != "id desc" || !reflect.DeepEqual(cond.SelectColumns, []string{"id", "name"}) || len(cond.TimeRanges) != 1 || cond.Limit != 10 {
		t.Fatalf("实体条件查询构建异常: %+v", cond)
	}

	where := NewWhereQuery("id = ?", 1).OrderBy("id").Select("id").WithLimit(1)
	if where.RawWhereSQL != "id = ?" || !reflect.DeepEqual(where.Args, []any{1}) || where.OrderBySQL != "id" || where.Limit != 1 {
		t.Fatalf("原始 SQL 查询构建异常: %+v", where)
	}

	page := NewPageQuery(wrapperModel{ID: 2}, 2, 20).OrderBy("id desc").Select("id").WithTimeRanges(rangeFilter)
	if page.Condition.ID != 2 || page.Number != 2 || page.Size != 20 || page.OrderBySQL != "id desc" || len(page.TimeRanges) != 1 {
		t.Fatalf("实体条件分页查询构建异常: %+v", page)
	}

	mapPage := NewMapPageQuery(map[string]any{"status": 0}, 1, 10).OrderBy("id").Select("id")
	if mapPage.Number != 1 || mapPage.Size != 10 || mapPage.OrderBySQL != "id" || !reflect.DeepEqual(mapPage.SelectColumns, []string{"id"}) {
		t.Fatalf("Map 分页查询构建异常: %+v", mapPage)
	}
}
