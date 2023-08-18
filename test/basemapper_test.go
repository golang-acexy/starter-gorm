package test

import "testing"

func TestBaseSave(t *testing.T) {
	bm := TeacherMapper{}
	bm.Save(Teacher{Name: "zs"})
}

//func TestModifyById(t *testing.T) {
//	bm := TeacherMapper{}
//	updated := Teacher{Name: "ls"}
//	updated.ID = 12
//	bm.ModifyById(updated)
//}
//
//func TestModifyByCondition(t *testing.T) {
//	bm := TeacherMapper{}
//	updated := Teacher{Name: "ls"}
//
//	condition := Teacher{
//		Name: "xys1",
//	}
//	condition.ID = 1
//
//	bm.ModifyByCondition(updated, condition)
//
//	bm.ModifyByCondition(updated, "id = ?", 1)
//}

//func TestRemoveById(t *testing.T) {
//	bm := TeacherMapper{}
//	t1 := Teacher{}
//	t1.ID = 1
//	t2 := Teacher{}
//	t2.ID = 2
//	bm.RemoveById(t1, t2)
//}
