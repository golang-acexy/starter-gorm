package test

import "testing"

func TestBaseSave(t *testing.T) {
	bm := TeacherBaseMapper{}
	bm.Save(Teacher{Name: "zs"})

	updated := Teacher{Name: "ls"}
	updated.ID = 12
	bm.ModifyById(updated)
}
