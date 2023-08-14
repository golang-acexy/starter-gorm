package test

import (
	"github.com/golang-acexy/starter-gorm/gormmodule"
	"time"
)

type Student struct {
	ID        uint64    `gorm:"<-:false,primaryKey"`
	CreatedAt time.Time `gorm:"column:create_time"`
	UpdatedAt time.Time `gorm:"column:update_time"`
	Name      string
	Sex       uint
}

func (Student) TableName() string {
	return "demo_student"
}

type Teacher struct {
	gormmodule.BaseEntity
	Name string
}

func (Teacher) TableName() string {
	return "demo_teacher"
}

type TeacherBaseMapper struct {
	gormmodule.BaseMapper[Teacher]
}
