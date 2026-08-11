package gormstarter

import (
	"slices"
	"testing"
)

type WrapperEmbedded struct {
	Name string
}

type wrapperModel struct {
	ID int64
	WrapperEmbedded
	Ignored string `gorm:"-"`
}

type wrapperPointerModel struct {
	*WrapperEmbedded
}

func (wrapperPointerModel) TableName() string { return "wrapper_pointer_model" }

func (wrapperModel) TableName() string { return "wrapper_model" }

func TestColumnSelectorFieldPath(t *testing.T) {
	id := NewColumn(func(model *wrapperModel) *int64 { return &model.ID })
	if !slices.Equal(id.fieldPath, []int{0}) {
		t.Fatalf("ID 字段路径错误: %v", id.fieldPath)
	}
	name := NewColumn(func(model *wrapperModel) *string { return &model.Name })
	if !slices.Equal(name.fieldPath, []int{1, 0}) {
		t.Fatalf("嵌入字段路径错误: %v", name.fieldPath)
	}
}

func TestColumnSelectorEmbeddedPointer(t *testing.T) {
	name := NewColumn(func(model *wrapperPointerModel) *string { return &model.Name })
	if !slices.Equal(name.fieldPath, []int{0, 0}) {
		t.Fatalf("指针嵌入字段路径错误: %v", name.fieldPath)
	}
}

func TestColumnSelectorRejectsInvalidSelector(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered != ErrInvalidColumnSelector {
			t.Fatalf("无效字段选择器 panic = %v", recovered)
		}
	}()
	NewColumn(func(model *wrapperModel) *int64 {
		value := model.ID
		return &value
	})
}
