package test

import (
	"github.com/golang-acexy/starter-gorm/gormmodule"
	"time"
)

type BaseModel struct {
	gormmodule.BaseMapper

	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"column:create_time"`
	UpdatedAt time.Time `gorm:"column:update_time"`
}

type Student struct {
	gormmodule.BaseMapper
	BaseModel
	Name string
}

func (*Student) TableName() string {
	return "demo_student"
}

func (s *Student) Save() error {
	rs := gormmodule.RawDB().Save(s)
	if rs.Error != nil {
		return rs.Error
	}
	return nil
}
