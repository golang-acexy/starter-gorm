package test

import (
	"github.com/golang-acexy/starter-gorm/gormmodule"
	"time"
)

// student

type Student struct {
	ID        uint      `gorm:"<-:create,primaryKey"`
	CreatedAt time.Time `gorm:"column:create_time" gorm:"<-:create"`
	UpdatedAt time.Time `gorm:"column:update_time" gorm:"<-:false"`
	Name      string
	Sex       uint
}

func (Student) TableName() string {
	return "demo_student"
}

// teacher

type Teacher struct {
	gormmodule.BaseEntity[uint64]
	Name string
}

func (Teacher) TableName() string {
	return "demo_teacher"
}

func (Teacher) EmptyStruct() any {
	return new(Teacher)
}

func (t Teacher) IdStruct() any {
	teacher := Teacher{}
	teacher.ID = t.ID
	return &teacher
}

type TeacherMapper struct {
	gormmodule.BaseMapper[Teacher]
}
