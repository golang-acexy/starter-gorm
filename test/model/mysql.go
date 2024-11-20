package model

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
	CreatedAt gormstarter.Timestamp  `gorm:"column:create_time;<-:false" json:"createTime"`
	UpdatedAt *gormstarter.Timestamp `gorm:"column:update_time;<-:update" json:"updateTime"` // 指定update时自动更新时间
	Name      string
	Sex       uint
	Age       uint
	ClassNo   uint
}

func (Teacher) TableName() string {
	return "demo_teacher"
}

func (Teacher) DBType() gormstarter.DBType {
	return gormstarter.DBTypeMySQL
}

// TeacherMapper 声明Teacher 获取基于BaseMapper的能力
type TeacherMapper struct {
	gormstarter.BaseMapper[Teacher]
}
