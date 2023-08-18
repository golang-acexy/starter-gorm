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
	TableName() string
}

type BaseMapper[T IBaseModel] struct {
	Value T
}

func checkResult(rs *gorm.DB) (int64, error) {
	if rs.Error != nil {
		return 0, rs.Error
	}
	return rs.RowsAffected, nil
}

func (b BaseMapper[T]) Save(entity *T) (int64, error) {
	return checkResult(db.Save(entity))
}

// ModifyById 通过ID更新非零值字段
func (b BaseMapper[T]) ModifyById(updated T) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Updates(updated))
}

// ModifyMapById 通过ID更新所有map中指定的列和值
func (b BaseMapper[T]) ModifyMapById(id any, updated map[string]any) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where("id = ?", id).Updates(updated))
}

// RemoveById 通过ID删除相关数据
func (b BaseMapper[T]) RemoveById(id ...any) (int64, error) {
	return checkResult(db.Delete(b.Value, id))
}
