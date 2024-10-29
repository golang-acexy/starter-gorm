package gormstarter

import (
	"github.com/acexy/golang-toolkit/util/reflect"
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

// SelectById 通过主键查询数据
func (b *BaseMapper[T]) SelectById(id any, result *T) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Where("id = ?", id).Scan(result))
}

// SelectByIds 通过主键查询数据
func (b *BaseMapper[T]) SelectByIds(id []interface{}, result *[]T) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Where("id in ?", id).Scan(result))
}

// SelectOneByCond 通过条件查询 查询条件零值字段将被自动忽略
// specifyColumns 需要指定只查询的数据库字段
func (b *BaseMapper[T]) SelectOneByCond(condition *T, result *T, specifyColumns ...string) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Select(specifyColumns).Where(condition).Scan(result))
}

// SelectOneByCondMap 通过指定字段与值查询数据 解决查询条件零值问题
// specifyColumns 需要指定只查询的数据库字段
func (b *BaseMapper[T]) SelectOneByCondMap(condition map[string]any, result *T, specifyColumns ...string) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Select(specifyColumns).Where(condition).Scan(result))
}

// SelectOneByWhere 通过原始Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
func (b *BaseMapper[T]) SelectOneByWhere(rawWhereSql string, result *T, args ...interface{}) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Where(rawWhereSql, args...).Scan(result))
}

// SelectOneByGorm 通过原始Gorm查询单条数据
func (b *BaseMapper[T]) SelectOneByGorm(result *T, raw func(*gorm.DB)) (int64, error) {
	var db = gormDB.Table(b.Value.TableName())
	raw(db)
	return checkResult(db.Scan(result))
}

// SelectByCond 通过条件查询 查询条件零值字段将被自动忽略
// specifyColumns 需要指定只查询的数据库字段
func (b *BaseMapper[T]) SelectByCond(condition *T, orderBy string, result *[]*T, specifyColumns ...string) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Select(specifyColumns).Where(condition).Order(orderBy).Scan(result))
}

// SelectByCondMap 通过指定字段与值查询数据 解决零值条件问题
// specifyColumns 需要指定只查询的数据库字段
func (b *BaseMapper[T]) SelectByCondMap(condition map[string]any, orderBy string, result *[]*T, specifyColumns ...string) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Select(specifyColumns).Where(condition).Order(orderBy).Scan(result))
}

// SelectByWhere 通过原始Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
func (b *BaseMapper[T]) SelectByWhere(rawWhereSql, orderBy string, result *[]*T, args ...interface{}) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Where(rawWhereSql, args...).Order(orderBy).Scan(result))
}

// SelectByGorm 通过原始Gorm查询数据
func (b *BaseMapper[T]) SelectByGorm(result *[]*T, raw func(*gorm.DB)) (int64, error) {
	var db = gormDB.Table(b.Value.TableName())
	raw(db)
	return checkResult(db.Scan(result))
}

// CountByCond 通过条件查询数据总数 查询条件零值字段将被自动忽略
func (b *BaseMapper[T]) CountByCond(condition *T) (int64, error) {
	var count int64
	_, err := checkResult(gormDB.Table(b.Value.TableName()).Where(condition).Count(&count))
	return count, err
}

// CountByCondMap 通过指定字段与值查询数据总数 解决零值条件问题
func (b *BaseMapper[T]) CountByCondMap(condition map[string]any) (int64, error) {
	var count int64
	_, err := checkResult(gormDB.Table(b.Value.TableName()).Where(condition).Count(&count))
	return count, err
}

// CountByWhere 通过原始SQL查询数据总数
func (b *BaseMapper[T]) CountByWhere(rawWhereSql string, args ...interface{}) (int64, error) {
	var count int64
	_, err := checkResult(gormDB.Table(b.Value.TableName()).Where(rawWhereSql, args...).Count(&count))
	return count, err
}

// CountByGorm 通过原始Gorm查询数据总数
func (b *BaseMapper[T]) CountByGorm(raw func(*gorm.DB)) (int64, error) {
	var count int64
	var db = gormDB.Table(b.Value.TableName())
	raw(db)
	_, err := checkResult(db.Count(&count))
	return count, err
}

// SelectPageByCond 通过条件分页查询 零值字段将被自动忽略
// specifyColumns 需要指定只查询的数据库字段
func (b *BaseMapper[T]) SelectPageByCond(condition *T, orderBy string, pageNumber, pageSize int, result *[]*T, specifyColumns ...string) (total int64, err error) {
	_, err = checkResult(gormDB.Table(b.Value.TableName()).Where(condition).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	_, err = checkResult(gormDB.Table(b.Value.TableName()).Select(specifyColumns).Where(condition).Order(orderBy).Limit(pageSize).Offset((pageNumber - 1) * pageSize).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SelectPageByCondMap 通过指定字段与值查询数据分页查询 解决零值条件问题
// specifyColumns 需要指定只查询的数据库字段
func (b *BaseMapper[T]) SelectPageByCondMap(condition map[string]any, orderBy string, pageNumber, pageSize int, result *[]*T, specifyColumns ...string) (total int64, err error) {
	_, err = checkResult(gormDB.Table(b.Value.TableName()).Where(condition).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	_, err = checkResult(gormDB.Table(b.Value.TableName()).Select(specifyColumns).Where(condition).Order(orderBy).Limit(pageSize).Offset((pageNumber - 1) * pageSize).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SelectPageByWhere 通过原始SQL分页查询 rawWhereSql 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
func (b *BaseMapper[T]) SelectPageByWhere(rawWhereSql, orderBy string, pageNumber, pageSize int, result *[]*T, args ...interface{}) (total int64, err error) {
	_, err = checkResult(gormDB.Table(b.Value.TableName()).Where(rawWhereSql, args...).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	_, err = checkResult(gormDB.Table(b.Value.TableName()).Where(rawWhereSql, args...).Order(orderBy).Limit(pageSize).Offset((pageNumber - 1) * pageSize).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// Save 保存数据 零值也将参与保存
//
//	exclude 手动指定需要排除的字段名称 数据库字段/结构体字段
func (b *BaseMapper[T]) Save(entity *T, excludeColumns ...string) (int64, error) {
	var tx = gormDB
	if len(excludeColumns) > 0 {
		tx = tx.Omit(excludeColumns...)
	}
	return checkResult(tx.Create(entity))
}

// SaveBatch 批量新增 零值也将参与保存
//
//	exclude 手动指定需要排除的字段名称 数据库字段/结构体字段
func (b *BaseMapper[T]) SaveBatch(entities *[]*T, excludeColumns ...string) (int64, error) {
	var tx = gormDB
	if len(excludeColumns) > 0 {
		tx = tx.Omit(excludeColumns...)
	}
	return checkResult(tx.Create(entities))
}

// SaveOrUpdateByPrimaryKey 保存/更新数据 零值也将参与保存
//
// exclude 手动指定需要排除的字段名称 数据库字段/结构体字段 (如果触发的是update 创建时间可能会被错误的修改，可以通过excludeColumns来指定排除创建时间字段)
// 仅根据主键冲突默认支持update 更多操作需要参阅 https://gorm.io/zh_CN/docs/create.html#upsert
func (b *BaseMapper[T]) SaveOrUpdateByPrimaryKey(entity *T, excludeColumns ...string) (int64, error) {
	var tx = gormDB
	if len(excludeColumns) > 0 {
		tx = tx.Omit(excludeColumns...)
	}
	return checkResult(tx.Save(entity))
}

// UpdateById 通过ID更新含零值字段
// updateColumns 手动指定需要更新的列
func (b *BaseMapper[T]) UpdateById(updated *T, updateColumns ...string) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Select(updateColumns).Updates(updated))
}

// UpdateByIdWithoutZeroField 通过ID更新非零值字段
// zeroFiledColumns 额外指定需要更新零值字段
func (b *BaseMapper[T]) UpdateByIdWithoutZeroField(updated *T, zeroFiledColumns ...string) (int64, error) {
	nonZeroFields, err := reflect.NonZeroField(updated)
	if err != nil {
		return 0, err
	}
	nonZeroFields = append(nonZeroFields, zeroFiledColumns...)
	return checkResult(gormDB.Table(b.Value.TableName()).Select(nonZeroFields).Updates(updated))
}

// UpdateByIdUseMap 通过ID更新所有map中指定的列和值
func (b *BaseMapper[T]) UpdateByIdUseMap(updated map[string]any, id any) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Where("id = ?", id).Updates(updated))
}

// UpdateByCond 通过条件更新 条件：零值将自动忽略，更新：零值字段将被自动忽略
// updateColumns 需要指定更新的数据库字段 更新指定字段(支持零值字段)
func (b *BaseMapper[T]) UpdateByCond(updated, condition *T, updateColumns ...string) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Select(updateColumns).Where(condition).Updates(updated))
}

// UpdateByCondWithZeroField 通过条件更新，并指定零值字段用于更新 条件：零值将自动忽略
// zeroFiledColumns 额外指定需要更新的零值字段
func (b *BaseMapper[T]) UpdateByCondWithZeroField(updated, condition *T, zeroFiledColumns []string) (int64, error) {
	nonZeroFields, err := reflect.NonZeroField(updated)
	if err != nil {
		return 0, err
	}
	nonZeroFields = append(nonZeroFields, zeroFiledColumns...)
	return checkResult(gormDB.Table(b.Value.TableName()).Select(nonZeroFields).Where(condition).Updates(updated))
}

// UpdateByCondMap 通过Map类型条件更新
func (b *BaseMapper[T]) UpdateByCondMap(updated, condition map[string]any) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Where(condition).Updates(updated))
}

// UpdateByWhere 通过原始SQL查询条件，更新非零实体字段 Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
func (b *BaseMapper[T]) UpdateByWhere(updated *T, rawWhereSql string, args ...interface{}) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Where(rawWhereSql, args...).Updates(updated))
}

// DeleteById 通过ID删除相关数据
func (b *BaseMapper[T]) DeleteById(id ...any) (int64, error) {
	return checkResult(gormDB.Delete(b.Value, id))
}

// DeleteByCond 通过条件删除 零值字段将被自动忽略
func (b *BaseMapper[T]) DeleteByCond(condition *T) (int64, error) {
	return checkResult(gormDB.Table(b.Value.TableName()).Where(condition).Delete(b.Value))
}

// DeleteByWhere 通过原始SQL删除相关数据 Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
func (b *BaseMapper[T]) DeleteByWhere(rawWhereSql string, args ...interface{}) (int64, error) {
	return checkResult(gormDB.Where(rawWhereSql, args...).Delete(b.Value))
}
