package test

import (
	"time"
)

type BaseModel struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"column:create_time"`
	UpdatedAt time.Time `gorm:"column:update_time"`
}

type Student struct {
	BaseModel
	Name string
}

func (Student) TableName() string {
	return "demo_student"
}

var StudentMux = func() {}
