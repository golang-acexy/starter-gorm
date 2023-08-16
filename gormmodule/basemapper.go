package gormmodule

import (
	"time"
)

type BaseEntity[IdType any] struct {
	ID        IdType    `gorm:"<-:create,primaryKey"`
	CreatedAt time.Time `gorm:"column:create_time" gorm:"<-:create"`
	UpdatedAt time.Time `gorm:"column:update_time" gorm:"<-:false"`
}

type IBaseEntity interface {
	// EmptyStruct 返回空的Model
	EmptyStruct() any
	// IdStruct 返回带有Id的struct
	IdStruct() any
}

type BaseMapper[T any] struct {
	Value T
}

func (b BaseMapper[T]) Save(entity IBaseEntity) (int64, error) {
	rs := db.Save(&entity)
	if rs.Error != nil {
		return 0, rs.Error
	}
	return rs.RowsAffected, nil
}

func (b BaseMapper[T]) ModifyById(updated IBaseEntity) {
	db.Model(updated.IdStruct()).Updates(updated)
}
