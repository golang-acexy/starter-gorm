package test

import (
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"time"
)

// student

type Student struct {
	ID        uint      `gorm:"<-:create,primaryKey"`
	CreatedAt time.Time `gorm:"column:create_time" gorm:"<-:create"`
	UpdatedAt time.Time `gorm:"column:update_time" gorm:"<-:false"`
	Name      string
	Sex       uint
	TeacherId uint
}

func (Student) TableName() string {
	return "demo_student"
}

// Teacher 继承BaseModel 并实现 IBaseModel
type Teacher struct {
	gormstarter.BaseModel[uint64]
	CreatedAt time.Time `gorm:"column:create_time;<-:false" json:"createTime"`
	UpdatedAt time.Time `gorm:"column:update_time;<-:update" json:"updateTime"`
	Name      string
	Sex       uint
	Age       uint
}

func (Teacher) TableName() string {
	return "demo_teacher"
}

// TeacherMapper 声明Teacher 获取基于BaseMapper的能力
type TeacherMapper struct {
	gormstarter.BaseMapper[Teacher]
}
