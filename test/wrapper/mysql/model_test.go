package mysql

import (
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"gorm.io/gorm"
)

// Teacher 在 Wrapper 测试包中独立定义，避免依赖公共测试 Model。
type Teacher struct {
	ID        uint64                `gorm:"<-:false;primaryKey"`
	CreatedAt gormstarter.Timestamp `gorm:"column:create_time;<-:false"`
	UpdatedAt gormstarter.Timestamp `gorm:"column:update_time;<-:update"`
	Name      string
	Sex       uint
	Age       uint
	ClassNo   uint
	Ignored   string `gorm:"-"`
}

func (Teacher) TableName() string { return "demo_teacher" }

func (Teacher) DBType() gormstarter.DBType { return gormstarter.DBTypeMySQL }

type TeacherMapper struct {
	gormstarter.BaseMapper[Teacher]

}

func (mapper TeacherMapper) WithTxMapper(tx *gorm.DB) TeacherMapper {
	return TeacherMapper{BaseMapper: mapper.BaseMapper.GetBaseMapperWithTx(tx)}
}

type TeacherColumns struct {
	ID        gormstarter.Column[Teacher, uint64]
	CreatedAt gormstarter.Column[Teacher, gormstarter.Timestamp]
	Name      gormstarter.Column[Teacher, string]
	Sex       gormstarter.Column[Teacher, uint]
	Age       gormstarter.Column[Teacher, uint]
	ClassNo   gormstarter.Column[Teacher, uint]
	Ignored   gormstarter.Column[Teacher, string]
}

var teacherColumns = TeacherColumns{
	ID:        gormstarter.NewColumn(func(teacher *Teacher) *uint64 { return &teacher.ID }),
	CreatedAt: gormstarter.NewColumn(func(teacher *Teacher) *gormstarter.Timestamp { return &teacher.CreatedAt }),
	Name:      gormstarter.NewColumn(func(teacher *Teacher) *string { return &teacher.Name }),
	Sex:       gormstarter.NewColumn(func(teacher *Teacher) *uint { return &teacher.Sex }),
	Age:       gormstarter.NewColumn(func(teacher *Teacher) *uint { return &teacher.Age }),
	ClassNo:   gormstarter.NewColumn(func(teacher *Teacher) *uint { return &teacher.ClassNo }),
	Ignored:   gormstarter.NewColumn(func(teacher *Teacher) *string { return &teacher.Ignored }),
}

// Columns 返回当前 Mapper 对应的只读字段元数据。
func (TeacherMapper) Columns() *TeacherColumns {
	return &teacherColumns
}

var _ gormstarter.Mapper[Teacher] = TeacherMapper{}
