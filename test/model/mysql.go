package model

import (
	"time"

	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

// student

type Student struct {
	ID        uint      `gorm:"<-:create,primaryKey"`
	CreatedAt time.Time `gorm:"column:create_time" gorm:"<-:create"`
	UpdatedAt time.Time `gorm:"column:update_time" gorm:"<-:false"`
	Name      string
	Sex       uint
	TeacherID uint `gorm:"column:teacher_id"`
}

func (Student) TableName() string {
	return "demo_student"
}

// Teacher 继承BaseModel 并实现 Model
type Teacher struct {
	ID        uint64                `gorm:"<-:false;primaryKey" json:"id"`
	CreatedAt gormstarter.Timestamp `gorm:"column:create_time;<-:false" json:"createTime"`
	UpdatedAt gormstarter.Timestamp `gorm:"column:update_time;<-:update" json:"updateTime"` // 指定update时自动更新时间
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

func (t TeacherMapper) ByID(id uint64) *Teacher {
	r := new(Teacher)
	_, _ = t.BaseMapper.SelectByID(id, r)
	return r
}

var (
	_ gormstarter.Mapper[Teacher]       = TeacherMapper{}
	_ gormstarter.QueryMapper[Teacher]  = TeacherMapper{}
	_ gormstarter.InsertMapper[Teacher] = TeacherMapper{}
	_ gormstarter.UpdateMapper[Teacher] = TeacherMapper{}
	_ gormstarter.DeleteMapper[Teacher] = TeacherMapper{}
)

func (t TeacherMapper) WithTxMapper(tx *gorm.DB) TeacherMapper {
	return TeacherMapper{
		t.BaseMapper.GetBaseMapperWithTx(tx),
	}
}
