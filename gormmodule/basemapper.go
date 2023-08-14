package gormmodule

import (
	"reflect"
	"time"
)

type BM interface {
	NewStruct() any
}

type BaseEntity struct {
	ID        uint64    `gorm:"<-:create,primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:create_time" gorm:"<-:false" json:"createTime"`
	UpdatedAt time.Time `gorm:"column:update_time" gorm:"<-:false" json:"updateTime"`
}

type BaseMapper[T any] struct {
	Value T
}

func (b BaseMapper[T]) Save(entity T) (int64, error) {
	rs := db.Save(&entity)
	if rs.Error != nil {
		return 0, rs.Error
	}
	return rs.RowsAffected, nil
}

func (b BaseMapper[T]) ModifyById(updated T) {
	idValue := reflect.ValueOf(updated).FieldByName("ID").Interface()
	println(idValue)
	db.Model(b.Value).Updates(updated)
}
