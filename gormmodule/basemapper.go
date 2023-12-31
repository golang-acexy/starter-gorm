package gormmodule

import (
	"gorm.io/gorm"
	"math"
	"time"
)

type BaseModel[IdType any] struct {
	ID        IdType    `gorm:"<-:create,primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:create_time" gorm:"<-:create" json:"createTime"`
	UpdatedAt time.Time `gorm:"column:update_time" gorm:"<-:false" json:"updateTime"`
}

type IBaseModel interface {
	TableName() string
}

type BaseMapper[T IBaseModel] struct {
	Value T
}

func checkResult(rs *gorm.DB, isQuery ...bool) (int64, error) {
	if rs.Error != nil {
		return 0, rs.Error
	}
	if len(isQuery) > 0 {
		// 兼容transaction的检查，如果是查询防止未命中数据时触发回滚
		return math.MaxInt64, nil
	}
	return rs.RowsAffected, nil
}

// QueryById 通过主键查询数据
func (b BaseMapper[T]) QueryById(id any, result *T) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where("id = ?", id).Scan(result))
}

// QueryByCondition 通过非零条件查询
func (b BaseMapper[T]) QueryByCondition(condition T, result *[]*T) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where(condition).Scan(result))
}

// QueryByConditionMap 通过指定字段与值查询数据 解决零值条件问题
func (b BaseMapper[T]) QueryByConditionMap(condition map[string]any, result *[]*T) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where(condition).Scan(result))
}

// Save 保存数据
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

// ModifyByCondition 通过非零实体条件，更新非零实体字段
func (b BaseMapper[T]) ModifyByCondition(updated, condition T) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where(condition).Updates(updated))
}

// RemoveById 通过ID删除相关数据
func (b BaseMapper[T]) RemoveById(id ...any) (int64, error) {
	return checkResult(db.Delete(b.Value, id))
}
