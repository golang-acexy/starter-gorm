package gormmodule

import (
	"gorm.io/gorm"
	"math"
)

type BaseModel[IdType any] struct {
	ID IdType `gorm:"<-:create,primaryKey" json:"id"`
}

type IBaseModel interface {
	TableName() string
}

type BaseMapper[T IBaseModel] struct {
	Value T
}

func checkResult(rs *gorm.DB, txCheck ...bool) (int64, error) {
	if rs.Error != nil {
		return 0, rs.Error
	}
	if len(txCheck) > 0 && txCheck[0] {
		// 兼容transaction的检查，如果是查询防止未命中数据时触发回滚
		return math.MaxInt64, nil
	}
	return rs.RowsAffected, nil
}

// QueryById 通过主键查询数据
func (b *BaseMapper[T]) QueryById(id any, result *T) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where("id = ?", id).Scan(result))
}

// QueryOneByCondition 通过非零条件查询
func (b *BaseMapper[T]) QueryOneByCondition(condition T, result *T) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where(condition).Scan(result))
}

// QueryOneByConditionMap 通过指定字段与值查询数据 解决零值条件问题
func (b *BaseMapper[T]) QueryOneByConditionMap(condition map[string]any, result *T) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where(condition).Scan(result))
}

// QueryOneByWhere 通过原始SQL查询
func (b *BaseMapper[T]) QueryOneByWhere(rawSql string, result *T, args ...interface{}) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where(rawSql, args...).Scan(result))
}

// QueryByCondition 通过非零条件查询
func (b *BaseMapper[T]) QueryByCondition(condition T, result *[]*T) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where(condition).Scan(result))
}

// QueryByConditionMap 通过指定字段与值查询数据 解决零值条件问题
func (b *BaseMapper[T]) QueryByConditionMap(condition map[string]any, result *[]*T) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where(condition).Scan(result))
}

// QueryByWhere 通过原始SQL查询
func (b *BaseMapper[T]) QueryByWhere(rawSql string, result *[]*T, args ...interface{}) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where(rawSql, args...).Scan(result))
}

// PageCondition 通过指定的非零值条件分页查询
func (b *BaseMapper[T]) PageCondition(condition T, pageNumber, pageSize int, result *[]*T) (total int64, err error) {
	_, err = checkResult(db.Table(b.Value.TableName()).Where(condition).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	_, err = checkResult(db.Table(b.Value.TableName()).Where(condition).Limit(pageSize).Offset((pageNumber - 1) * pageSize).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// PageConditionMap 通过指定字段与值查询数据分页查询  解决零值条件问题
func (b *BaseMapper[T]) PageConditionMap(condition map[string]any, pageNumber, pageSize int, result *[]*T) (total int64, err error) {
	_, err = checkResult(db.Table(b.Value.TableName()).Where(condition).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	_, err = checkResult(db.Table(b.Value.TableName()).Where(condition).Limit(pageSize).Offset((pageNumber - 1) * pageSize).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// PageWhere 通过原始SQL分页查询
func (b *BaseMapper[T]) PageWhere(rawSql string, pageNumber, pageSize int, result *[]*T, args ...interface{}) (total int64, err error) {
	_, err = checkResult(db.Table(b.Value.TableName()).Where(rawSql, args...).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	_, err = checkResult(db.Table(b.Value.TableName()).Where(rawSql, args...).Limit(pageSize).Offset((pageNumber - 1) * pageSize).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// Save 保存数据 零值也将参与保存
//
//	exclude 手动指定需要排除的字段
func (b *BaseMapper[T]) Save(entity *T, excludeColumns ...string) (int64, error) {
	var tx = db
	if len(excludeColumns) > 0 {
		tx = tx.Omit(excludeColumns...)
	}
	return checkResult(tx.Create(entity))
}

// SaveBatch 批量新增 零值也将参与保存
//
//	exclude 手动指定需要排除的字段
func (b *BaseMapper[T]) SaveBatch(entities *[]*T, excludeColumns ...string) (int64, error) {
	var tx = db
	if len(excludeColumns) > 0 {
		tx = tx.Omit(excludeColumns...)
	}
	return checkResult(tx.Create(entities))
}

// SaveOrUpdate 保存/更新数据 零值也将参与保存
//
// exclude 手动指定需要排除的字段(如果触发的是update 创建时间可能会被错误的修改，可以通过excludeColumns来指定排除创建时间字段)
// 仅根据主键冲突默认支持update 更多操作需要参阅 https://gorm.io/zh_CN/docs/create.html#upsert
func (b *BaseMapper[T]) SaveOrUpdate(entity *T, excludeColumns ...string) (int64, error) {
	var tx = db
	if len(excludeColumns) > 0 {
		tx = tx.Omit(excludeColumns...)
	}
	return checkResult(tx.Save(entity))
}

// ModifyById 通过ID更新非零值字段
func (b *BaseMapper[T]) ModifyById(updated T) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Updates(updated))
}

// ModifyMapById 通过ID更新所有map中指定的列和值
func (b *BaseMapper[T]) ModifyMapById(id any, updated map[string]any) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where("id = ?", id).Updates(updated))
}

// ModifyByCondition 通过非零实体条件，更新非零实体字段
func (b *BaseMapper[T]) ModifyByCondition(updated, condition T) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where(condition).Updates(updated))
}

// ModifyByWhere 通过原始SQL查询条件，更新非零实体字段
func (b *BaseMapper[T]) ModifyByWhere(updated T, rawWhereSql string, args ...interface{}) (int64, error) {
	return checkResult(db.Table(b.Value.TableName()).Where(rawWhereSql, args...).Updates(updated))
}

// RemoveById 通过ID删除相关数据
func (b *BaseMapper[T]) RemoveById(id ...any) (int64, error) {
	return checkResult(db.Delete(b.Value, id))
}

// RemoveByWhere 通过原始SQL删除相关数据
func (b *BaseMapper[T]) RemoveByWhere(rawSql string, args ...interface{}) (int64, error) {
	return checkResult(db.Where(rawSql, args...).Delete(b.Value))
}
