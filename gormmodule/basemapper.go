package gormmodule

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel[IdType any] struct {
	ID        IdType    `gorm:"<-:create,primaryKey"`
	CreatedAt time.Time `gorm:"column:create_time" gorm:"<-:create"`
	UpdatedAt time.Time `gorm:"column:update_time" gorm:"<-:false"`
}

type IBaseModel interface {
	// EmptyStruct 返回空的Model
	EmptyStruct() any
	// IdStruct 返回带有Id的struct
	IdStruct() any
}

type BaseMapper[T any] struct {
	Value T
}

func checkResult(rs *gorm.DB) (int64, error) {
	if rs.Error != nil {
		return 0, rs.Error
	}
	return rs.RowsAffected, nil
}

func (b BaseMapper[T]) Save(entity IBaseModel) (int64, error) {
	return checkResult(db.Save(&entity))
}

//// ModifyById 通过ID更新非零值字段
//func (b BaseMapper[T]) ModifyById(updated IBaseModel) (int64, error) {
//	return checkResult(db.Model(updated.IdStruct()).Updates(updated))
//}
//
//// ModifyByCondition 通过条件更新非零值字段
//func (b BaseMapper[T]) ModifyByCondition(updated IBaseModel, where interface{}, args ...interface{}) (int64, error) {
//	return checkResult(db.Model(updated.EmptyStruct()).Where(where, args).Updates(updated))
//}
//
//func (b BaseMapper[T]) RemoveById(removed ...IBaseModel) (int64, error) {
//	if len(removed) == 0 {
//		return 0, errors.New("nil removed")
//	}
//	ids := make([]IBaseModel, len(removed))
//	for i, v := range removed {
//		ids[i] = v.IdStruct().(IBaseModel)
//	}
//	return checkResult(db.Delete(ids))
//}
