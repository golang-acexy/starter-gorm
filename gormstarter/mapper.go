package gormstarter

import (
	"database/sql"
	"fmt"
	stdreflect "reflect"
	"strings"

	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/acexy/golang-toolkit/util/reflect"
	"gorm.io/gorm"
)

func (b BaseMapper[T]) rawDB() (*gorm.DB, error) {
	if b.tx != nil {
		return b.tx, nil
	}
	if v, flag := any(b.model).(ModelWithDBType); flag {
		dbType := v.DBType()
		db := RawGormDB(dbType)
		if db == nil {
			return nil, fmt.Errorf("%w: %s", ErrDatabaseNotRegistered, dbType)
		}
		return db, nil
	}
	db := RawGormDB()
	if db == nil {
		return nil, ErrGormStarterNotStarted
	}
	return db, nil
}

func (b BaseMapper[T]) tableDB() (*gorm.DB, error) {
	db, err := b.rawDB()
	if err != nil {
		return nil, err
	}
	return db.Table(b.model.TableName()), nil
}

func checkResult(rs *gorm.DB) (int64, error) {
	if rs.Error != nil {
		return 0, rs.Error
	}
	return rs.RowsAffected, nil
}

// validateStructCondition 防止结构体零值条件退化为全表更新或删除。
func validateStructCondition[T any](condition T) error {
	if stdreflect.ValueOf(condition).IsZero() {
		return ErrEmptyCondition
	}
	return nil
}

// effectiveUpdateFields 收集实际可更新字段，主键 ID 不单独视为更新内容。
func effectiveUpdateFields[T any](updated *T, explicitColumns ...string) ([]string, error) {
	if updated == nil {
		return nil, ErrNilEntity
	}
	fields, err := reflect.NonZeroFieldName(updated)
	if err != nil {
		return nil, err
	}
	fields = append(fields, explicitColumns...)
	fields = coll.SliceDistinct(fields)
	fields = coll.SliceFilter(fields, func(field string) bool {
		return strings.TrimSpace(field) != "" && !strings.EqualFold(field, "id")
	})
	if len(fields) == 0 {
		return nil, ErrNoFieldToUpdate
	}
	return fields, nil
}

// TableGormDB 获取已限定当前 Mapper 表名的原生 gorm.DB。
func (b BaseMapper[T]) TableGormDB() *gorm.DB {
	db, _ := b.tableDB()
	return db
}

// CurrentGormDB 获取当前 Mapper 使用的 gorm.DB；绑定事务时返回该事务，否则返回新的 gorm.DB。
func (b BaseMapper[T]) CurrentGormDB() *gorm.DB {
	db, _ := b.rawDB()
	return db
}

// GetBaseMapperWithTx 获取绑定指定事务的基础 Mapper
func (b BaseMapper[T]) GetBaseMapperWithTx(tx *gorm.DB) BaseMapper[T] {
	return BaseMapper[T]{
		model: b.model,
		tx:    tx,
	}
}

// NewBaseMapperWithTx 创建一个全新事务的基础 Mapper
func (b BaseMapper[T]) NewBaseMapperWithTx(opts ...*sql.TxOptions) BaseMapper[T] {
	baseMapper := BaseMapper[T]{
		model: b.model,
	}
	db, _ := baseMapper.rawDB()
	baseMapper.tx = db.Begin(opts...)
	return baseMapper
}

// SelectByID 通过主键查询数据
func (b BaseMapper[T]) SelectByID(id any, result *T) (int64, error) {
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Where("id = ?", id).Scan(result))
}

// SelectByIDs 通过主键查询数据
func (b BaseMapper[T]) SelectByIDs(ids []any, result *[]*T) (int64, error) {
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Where("id in ?", ids).Scan(result))
}

// ExistsByID 判断指定主键的数据是否存在
func (b BaseMapper[T]) ExistsByID(id any) (bool, error) {
	db, err := b.tableDB()
	if err != nil {
		return false, err
	}
	var count int64
	if err = db.Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// SelectOneByCond 通过条件查询 查询条件零值字段将被自动忽略
// specifyColumns 指定只需要查询的数据库字段
func (b BaseMapper[T]) SelectOneByCond(condition T, result *T, specifyColumns ...string) (int64, error) {
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Select(specifyColumns).Where(condition).Limit(1).Scan(result))
}

// SelectOneByMap 通过指定字段与值查询数据 解决查询条件零值问题
// specifyColumns 指定只需要查询的数据库字段
func (b BaseMapper[T]) SelectOneByMap(condition map[string]any, result *T, specifyColumns ...string) (int64, error) {
	if len(condition) == 0 {
		return 0, ErrEmptyCondition
	}
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Select(specifyColumns).Where(condition).Limit(1).Scan(result))
}

// SelectOneByWhere 通过原始 Where SQL 查询，只需要传入 rawWhereSQL 和参数
func (b BaseMapper[T]) SelectOneByWhere(rawWhereSQL string, result *T, args ...any) (int64, error) {
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Where(rawWhereSQL, args...).Scan(result))
}

// SelectOneByGorm 通过原始Gorm查询单条数据 构建Gorm查询条件
func (b BaseMapper[T]) SelectOneByGorm(result *T, rawDB func(*gorm.DB)) (int64, error) {
	db, err := b.tableDB()
	if err != nil { return 0, err }
	rawDB(db)
	return checkResult(db.Scan(result))
}

// SelectByCond 通过条件查询 查询条件零值字段将被自动忽略
// specifyColumns 指定只需要查询的数据库字段
func (b BaseMapper[T]) SelectByCond(condition T, orderBySQL string, result *[]*T, specifyColumns ...string) (int64, error) {
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Select(specifyColumns).Where(condition).Order(orderBySQL).Scan(result))
}

// SelectByMap 通过指定字段与值查询数据 解决零值条件问题
// specifyColumns 指定只需要查询的数据库字段
func (b BaseMapper[T]) SelectByMap(condition map[string]any, orderBySQL string, result *[]*T, specifyColumns ...string) (int64, error) {
	if len(condition) == 0 {
		return 0, ErrEmptyCondition
	}
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Select(specifyColumns).Where(condition).Order(orderBySQL).Scan(result))
}

// SelectByWhere 通过原始 Where SQL 查询，只需要传入 rawWhereSQL 和参数
func (b BaseMapper[T]) SelectByWhere(rawWhereSQL, orderBySQL string, result *[]*T, args ...any) (int64, error) {
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Where(rawWhereSQL, args...).Order(orderBySQL).Scan(result))
}

// SelectByGorm 通过原始Gorm查询数据
func (b BaseMapper[T]) SelectByGorm(result *[]*T, rawDB func(*gorm.DB)) (int64, error) {
	db, err := b.tableDB()
	if err != nil { return 0, err }
	rawDB(db)
	return checkResult(db.Scan(result))
}

// CountByCond 通过条件查询数据总数 查询条件零值字段将被自动忽略
func (b BaseMapper[T]) CountByCond(condition T) (int64, error) {
	var count int64
	db, err := b.tableDB()
	if err != nil { return 0, err }
	_, err = checkResult(db.Where(condition).Count(&count))
	return count, err
}

// CountByMap 通过指定字段与值查询数据总数 解决零值条件问题
func (b BaseMapper[T]) CountByMap(condition map[string]any) (int64, error) {
	if len(condition) == 0 {
		return 0, ErrEmptyCondition
	}
	var count int64
	db, err := b.tableDB()
	if err != nil { return 0, err }
	_, err = checkResult(db.Where(condition).Count(&count))
	return count, err
}

// CountByWhere 通过原始SQL查询数据总数
func (b BaseMapper[T]) CountByWhere(rawWhereSQL string, args ...any) (int64, error) {
	var count int64
	db, err := b.tableDB()
	if err != nil { return 0, err }
	_, err = checkResult(db.Where(rawWhereSQL, args...).Count(&count))
	return count, err
}

// CountByGorm 通过原始Gorm查询数据总数
func (b BaseMapper[T]) CountByGorm(rawDB func(*gorm.DB)) (int64, error) {
	var count int64
	db, err := b.tableDB()
	if err != nil { return 0, err }
	rawDB(db)
	_, err = checkResult(db.Count(&count))
	return count, err
}

// applyTimeRanges 将 TimeRanges 批量应用到 gorm.DB 的 WHERE 条件中
// 使用左闭右开区间 [StartTime, EndTime)，EndTime 为 nil 时无上界
func applyTimeRanges(db *gorm.DB, ranges []TimeRange) *gorm.DB {
	for _, tr := range ranges {
		if tr.StartTime != nil && tr.EndTime != nil {
			db = db.Where("("+tr.Field+" >= ? AND "+tr.Field+" < ?)", *tr.StartTime, *tr.EndTime)
		} else if tr.StartTime != nil {
			db = db.Where(tr.Field+" >= ?", *tr.StartTime)
		} else if tr.EndTime != nil {
			db = db.Where(tr.Field+" < ?", *tr.EndTime)
		}
	}
	return db
}

// SelectPageByCond 通过条件分页查询 零值字段将被自动忽略
// specifyColumns 指定只需要查询的数据库字段 pageNumber 页码 1开始
func (b BaseMapper[T]) SelectPageByCond(condition T, query PageQuery, result *[]*T) (total int64, err error) {
	if query.PageNumber <= 0 || query.PageSize <= 0 {
		return 0, ErrInvalidPage
	}
	countDB, err := b.tableDB()
	if err != nil { return 0, err }
	countDB = applyTimeRanges(countDB, query.TimeRanges)
	_, err = checkResult(countDB.Where(condition).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	selectDB, err := b.tableDB()
	if err != nil { return 0, err }
	selectDB = applyTimeRanges(selectDB, query.TimeRanges)
	_, err = checkResult(selectDB.Select(query.SpecifyColumns).Where(condition).Order(query.OrderBySQL).Limit(query.PageSize).Offset((query.PageNumber - 1) * query.PageSize).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SelectPageByMap 通过指定字段与值查询数据分页查询 解决零值条件问题
// specifyColumns 指定只需要查询的数据库字段 pageNumber 页码 1开始
func (b BaseMapper[T]) SelectPageByMap(condition map[string]any, query PageQuery, result *[]*T) (total int64, err error) {
	if len(condition) == 0 {
		return 0, ErrEmptyCondition
	}
	if query.PageNumber <= 0 || query.PageSize <= 0 {
		return 0, ErrInvalidPage
	}
	countDB, err := b.tableDB()
	if err != nil { return 0, err }
	countDB = applyTimeRanges(countDB, query.TimeRanges)
	_, err = checkResult(countDB.Where(condition).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	selectDB, err := b.tableDB()
	if err != nil { return 0, err }
	selectDB = applyTimeRanges(selectDB, query.TimeRanges)
	_, err = checkResult(selectDB.Select(query.SpecifyColumns).Where(condition).Order(query.OrderBySQL).Limit(query.PageSize).Offset((query.PageNumber - 1) * query.PageSize).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SelectPageByWhere 通过原始 SQL 分页查询
func (b BaseMapper[T]) SelectPageByWhere(rawWhereSQL string, query PageQuery, result *[]*T, args ...any) (total int64, err error) {
	if query.PageNumber <= 0 || query.PageSize <= 0 {
		return 0, ErrInvalidPage
	}
	countDB, err := b.tableDB()
	if err != nil { return 0, err }
	countDB = applyTimeRanges(countDB, query.TimeRanges)
	_, err = checkResult(countDB.Where(rawWhereSQL, args...).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	selectDB, err := b.tableDB()
	if err != nil { return 0, err }
	selectDB = applyTimeRanges(selectDB, query.TimeRanges)
	_, err = checkResult(selectDB.Select(query.SpecifyColumns).Where(rawWhereSQL, args...).Order(query.OrderBySQL).Limit(query.PageSize).Offset((query.PageNumber - 1) * query.PageSize).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SelectPageByGorm 通过原始Gorm分页查询
func (b BaseMapper[T]) SelectPageByGorm(countRawDB func(*gorm.DB), pageRawDB func(*gorm.DB), result *[]*T) (total int64, err error) {
	countDB, err := b.tableDB()
	if err != nil { return 0, err }
	countRawDB(countDB)
	_, err = checkResult(countDB.Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	selectDB, err := b.tableDB()
	if err != nil { return 0, err }
	pageRawDB(selectDB)
	_, err = checkResult(selectDB.Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// Insert 保存数据 零值也将参与保存
//
//	exclude 手动指定需要排除的字段名称 数据库字段/结构体字段名称
func (b BaseMapper[T]) Insert(entity *T, excludeColumns ...string) (int64, error) {
	if entity == nil {
		return 0, ErrNilEntity
	}
	db, err := b.rawDB()
	if err != nil { return 0, err }
	if len(excludeColumns) > 0 {
		db = db.Omit(excludeColumns...)
	}
	return checkResult(db.Create(entity))
}

// InsertWithoutZeroFields 保存数据 零值字段将不会参与保存
func (b BaseMapper[T]) InsertWithoutZeroFields(entity *T) (int64, error) {
	if entity == nil {
		return 0, ErrNilEntity
	}
	nonZeroFields, err := reflect.NonZeroFieldName(entity)
	if err != nil {
		return 0, err
	}
	if len(nonZeroFields) == 0 {
		return 0, ErrNoFieldToSave
	}
	db, err := b.tableDB()
	if err != nil { return 0, err }
	if len(nonZeroFields) == 1 {
		return checkResult(db.Select(nonZeroFields[0]).Create(entity))
	} else {
		nonZeroFieldsSlice := coll.SliceCollect(nonZeroFields[1:], func(t string) any {
			return t
		})
		return checkResult(db.Select(nonZeroFields[0], nonZeroFieldsSlice...).Create(entity))
	}
}

// InsertBatch 批量新增 零值也将参与保存
//
//	exclude 手动指定需要排除的字段名称 数据库字段/结构体字段
func (b BaseMapper[T]) InsertBatch(entities []*T, excludeColumns ...string) (int64, error) {
	if len(entities) == 0 {
		return 0, ErrNoFieldToSave
	}
	for _, entity := range entities {
		if entity == nil {
			return 0, ErrNilEntity
		}
	}
	db, err := b.rawDB()
	if err != nil { return 0, err }
	if len(excludeColumns) > 0 {
		db = db.Omit(excludeColumns...)
	}
	return checkResult(db.Create(entities))
}

// InsertWithMap 通过 Map 类型保存数据
func (b BaseMapper[T]) InsertWithMap(entity map[string]any) (int64, error) {
	if len(entity) == 0 {
		return 0, ErrNoFieldToSave
	}
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Create(entity))
}

// InsertOrUpdateByPrimaryKey 保存/更新数据 零值也将参与保存
// exclude 手动指定需要排除的字段名称 数据库字段/结构体字段 (如果触发的是update 创建时间可能会被错误的修改，可以通过excludeColumns来指定排除创建时间字段)
// 仅根据主键冲突默认支持update 更多操作需要参阅 https://gorm.io/zh_CN/docs/create.html#upsert
func (b BaseMapper[T]) InsertOrUpdateByPrimaryKey(entity *T, excludeColumns ...string) (int64, error) {
	if entity == nil {
		return 0, ErrNilEntity
	}
	db, err := b.rawDB()
	if err != nil { return 0, err }
	if len(excludeColumns) > 0 {
		db = db.Omit(excludeColumns...)
	}
	return checkResult(db.Save(entity))
}

// UpdateByID 通过实体中的主键 ID 更新含零值字段
// updateColumns 手动指定需要更新的列
func (b BaseMapper[T]) UpdateByID(updated *T, updateColumns ...string) (int64, error) {
	if _, err := effectiveUpdateFields(updated, updateColumns...); err != nil {
		return 0, err
	}
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Select(updateColumns).Updates(updated))
}

// UpdateByIDWithoutZeroFields 通过ID更新非零值字段
// allowZeroFieldColumns 额外指定需要更新零值字段
func (b BaseMapper[T]) UpdateByIDWithoutZeroFields(updated *T, allowZeroFieldColumns ...string) (int64, error) {
	nonZeroFields, err := effectiveUpdateFields(updated, allowZeroFieldColumns...)
	if err != nil { return 0, err }
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Select(nonZeroFields).Updates(updated))
}

// UpdateByIDWithMap 通过 ID 更新 Map 中指定的列和值
func (b BaseMapper[T]) UpdateByIDWithMap(updated map[string]any, id any) (int64, error) {
	if len(updated) == 0 {
		return 0, ErrNoFieldToUpdate
	}
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Where("id = ?", id).Updates(updated))
}

// UpdateByCond 通过条件更新 条件：零值将自动忽略，更新：零值字段将被自动忽略
// updateColumns 需要指定更新的数据库字段 更新指定字段(支持零值字段)
func (b BaseMapper[T]) UpdateByCond(updated *T, condition T, updateColumns ...string) (int64, error) {
	if _, err := effectiveUpdateFields(updated, updateColumns...); err != nil { return 0, err }
	if err := validateStructCondition(condition); err != nil { return 0, err }
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Select(updateColumns).Where(condition).Updates(updated))
}

// UpdateByCondWithZeroFields 通过条件更新，并指定可以更新的零值字段
func (b BaseMapper[T]) UpdateByCondWithZeroFields(updated *T, condition T, allowZeroFieldColumns ...string) (int64, error) {
	nonZeroFields, err := effectiveUpdateFields(updated, allowZeroFieldColumns...)
	if err != nil { return 0, err }
	if err = validateStructCondition(condition); err != nil { return 0, err }
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Select(nonZeroFields).Where(condition).Updates(updated))
}

// UpdateByMap 通过Map类型条件更新
func (b BaseMapper[T]) UpdateByMap(updated, condition map[string]any) (int64, error) {
	if len(updated) == 0 {
		return 0, ErrNoFieldToUpdate
	}
	if len(condition) == 0 {
		return 0, ErrEmptyCondition
	}
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Where(condition).Updates(updated))
}

// UpdateByWhere 通过原始 Where SQL 条件更新非零实体字段
func (b BaseMapper[T]) UpdateByWhere(updated *T, rawWhereSQL string, args ...any) (int64, error) {
	if _, err := effectiveUpdateFields(updated); err != nil { return 0, err }
	if strings.TrimSpace(rawWhereSQL) == "" { return 0, ErrEmptyWhereSQL }
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Where(rawWhereSQL, args...).Updates(updated))
}

// DeleteByID 通过 ID 删除单条数据
func (b BaseMapper[T]) DeleteByID(id any) (int64, error) {
	db, err := b.rawDB()
	if err != nil { return 0, err }
	return checkResult(db.Delete(b.model, id))
}

// DeleteByIDs 通过多个 ID 删除数据
func (b BaseMapper[T]) DeleteByIDs(ids []any) (int64, error) {
	if len(ids) == 0 {
		return 0, ErrEmptyIDs
	}
	db, err := b.rawDB()
	if err != nil { return 0, err }
	return checkResult(db.Delete(b.model, ids))
}

// DeleteByCond 通过条件删除 零值字段将被自动忽略
func (b BaseMapper[T]) DeleteByCond(condition T) (int64, error) {
	if err := validateStructCondition(condition); err != nil { return 0, err }
	db, err := b.tableDB()
	if err != nil { return 0, err }
	return checkResult(db.Where(condition).Delete(b.model))
}

// DeleteByWhere 通过原始 Where SQL 删除相关数据
func (b BaseMapper[T]) DeleteByWhere(rawWhereSQL string, args ...any) (int64, error) {
	if strings.TrimSpace(rawWhereSQL) == "" { return 0, ErrEmptyWhereSQL }
	db, err := b.rawDB()
	if err != nil { return 0, err }
	return checkResult(db.Where(rawWhereSQL, args...).Delete(b.model))
}

// DeleteByMap 通过Map类型条件删除
func (b BaseMapper[T]) DeleteByMap(condition map[string]any) (int64, error) {
	if len(condition) == 0 {
		return 0, ErrEmptyCondition
	}
	db, err := b.rawDB()
	if err != nil { return 0, err }
	return checkResult(db.Where(condition).Delete(b.model))
}
