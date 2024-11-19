package model

import (
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/lib/pq"
)

type Employee struct {
	gormstarter.BaseModel[uint64]

	CreatedAt gormstarter.Timestamp `gorm:"<-:false"`
	UpdatedAt gormstarter.Timestamp `gorm:"<-:false"`
	Name      string
	Sex       string
	LeaderId  pq.Int32Array `gorm:"type:integer[]"`
}

func (e Employee) DBType() gormstarter.DBType {
	return gormstarter.DBTypePostgres
}

func (e Employee) TableName() string {
	return "employee"
}

type EmployeeMapper struct {
	gormstarter.BaseMapper[Employee]
}
