package test

import (
	"fmt"
	"testing"
)

func TestBaseSave(t *testing.T) {
	bm := TeacherMapper{}
	teacher := Teacher{Name: "mapper", Age: 12, Sex: 1}
	fmt.Println(bm.Save(&teacher))
	fmt.Println("saved id", teacher.ID)
}

func TestModifyById(t *testing.T) {
	bm := TeacherMapper{}
	updated := Teacher{Name: "update", Age: 21, Sex: 0}
	updated.ID = 132
	fmt.Println(bm.ModifyById(updated))
}

func TestModifyMapById(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.ModifyMapById(132, map[string]any{"name": "Miss A", "sex": 0}))
}

func TestRemoveById(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.RemoveById(1))
}
