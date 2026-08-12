package gormstarter

import (
	"database/sql"
	"fmt"
	stdreflect "reflect"
	"strings"

	"github.com/acexy/golang-toolkit/logger"
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

// validateID 防止空主键进入查询、更新或删除操作。
func validateID(id any) error {
	if id == nil {
		return ErrEmptyID
	}
	value := stdreflect.ValueOf(id)
	for value.Kind() == stdreflect.Interface || value.Kind() == stdreflect.Ptr {
		if value.IsNil() {
			return ErrEmptyID
		}
		value = value.Elem()
	}
	if value.IsZero() {
		return ErrEmptyID
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

// Wrapper 创建与当前 Mapper 模型类型绑定的查询 Wrapper。
func (b BaseMapper[T]) Wrapper() *QueryWrapper[T] {
	return NewQueryWrapper[T]()
}

// PageWrapper 创建与当前 Mapper 模型类型绑定的分页查询 Wrapper。
func (b BaseMapper[T]) PageWrapper(number, size int) *PageWrapper[T] {
	return newPageWrapper[T](number, size)
}

// UpdateWrapper 创建与当前 Mapper 模型类型绑定的更新 Wrapper。
func (b BaseMapper[T]) UpdateWrapper() *UpdateWrapper[T] {
	return newUpdateWrapper[T]()
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
	if err := validateID(id); err != nil {
		return 0, err
	}
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	return checkResult(db.Where("id = ?", id).Scan(result))
}

// SelectByIDs 通过主键查询数据
func (b BaseMapper[T]) SelectByIDs(ids []any, result *[]*T) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	return checkResult(db.Where("id in ?", ids).Scan(result))
}

// ExistsByID 判断指定主键的数据是否存在
func (b BaseMapper[T]) ExistsByID(id any) (bool, error) {
	if err := validateID(id); err != nil {
		return false, err
	}
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
func (b BaseMapper[T]) SelectOneByCond(query CondQuery[T], result *T) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db = applyTimeRanges(db, query.TimeRanges)
	return checkResult(db.Select(query.SelectColumns).Where(query.Condition).Order(query.OrderBySQL).Limit(1).Scan(result))
}

// SelectOneByMap 通过指定字段与值查询数据 解决查询条件零值问题
// specifyColumns 指定只需要查询的数据库字段
func (b BaseMapper[T]) SelectOneByMap(query MapQuery, result *T) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db = applyTimeRanges(db, query.TimeRanges)
	return checkResult(db.Select(query.SelectColumns).Where(query.Condition).Order(query.OrderBySQL).Limit(1).Scan(result))
}

// SelectOneByWhere 通过原始 Where SQL 查询，只需要传入 rawWhereSQL 和参数
func (b BaseMapper[T]) SelectOneByWhere(query WhereQuery, result *T) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db = applyTimeRanges(db, query.TimeRanges)
	rowsAffected, err := checkResult(db.Select(query.SelectColumns).Where(query.RawWhereSQL, query.Args...).Order(query.OrderBySQL).Scan(result))
	if err != nil {
		return 0, err
	}
	if rowsAffected > 1 {
		// 原始 SQL 的限制条件由调用方负责；返回多行时提示单条查询语义可能不明确。
		logger.Logrus().WithFields(map[string]any{
			"table":        b.model.TableName(),
			"rowsAffected": rowsAffected,
		}).Warnln("SelectOneByWhere queried more than one row")
	}
	return rowsAffected, nil
}

// SelectOneByGorm 通过原始Gorm查询单条数据 构建Gorm查询条件
func (b BaseMapper[T]) SelectOneByGorm(result *T, rawDB func(*gorm.DB)) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	rawDB(db)
	return checkResult(db.Scan(result))
}

// SelectByCond 通过条件查询 查询条件零值字段将被自动忽略
// specifyColumns 指定只需要查询的数据库字段
func (b BaseMapper[T]) SelectByCond(query CondQuery[T], result *[]*T) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db = applyTimeRanges(db, query.TimeRanges)
	db, err = applyQueryLimit(db, query.Limit)
	if err != nil {
		return 0, err
	}
	return checkResult(db.Select(query.SelectColumns).Where(query.Condition).Order(query.OrderBySQL).Scan(result))
}

// SelectByMap 通过指定字段与值查询数据 解决零值条件问题
// specifyColumns 指定只需要查询的数据库字段
func (b BaseMapper[T]) SelectByMap(query MapQuery, result *[]*T) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db = applyTimeRanges(db, query.TimeRanges)
	db, err = applyQueryLimit(db, query.Limit)
	if err != nil {
		return 0, err
	}
	return checkResult(db.Select(query.SelectColumns).Where(query.Condition).Order(query.OrderBySQL).Scan(result))
}

// SelectByWhere 通过原始 Where SQL 查询，只需要传入 rawWhereSQL 和参数
func (b BaseMapper[T]) SelectByWhere(query WhereQuery, result *[]*T) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db = applyTimeRanges(db, query.TimeRanges)
	db, err = applyQueryLimit(db, query.Limit)
	if err != nil {
		return 0, err
	}
	return checkResult(db.Select(query.SelectColumns).Where(query.RawWhereSQL, query.Args...).Order(query.OrderBySQL).Scan(result))
}

// SelectByGorm 通过原始Gorm查询数据
func (b BaseMapper[T]) SelectByGorm(result *[]*T, rawDB func(*gorm.DB)) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	rawDB(db)
	return checkResult(db.Scan(result))
}

// SelectOneByWrapper 使用类型安全的 Wrapper 查询一条数据。
func (b BaseMapper[T]) SelectOneByWrapper(query *QueryWrapper[T], result *T) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db, err = applyWrapper(db, query, true, true, false)
	if err != nil {
		return 0, err
	}
	return checkResult(db.Limit(1).Scan(result))
}

// SelectByWrapper 使用类型安全的 Wrapper 查询数据。
func (b BaseMapper[T]) SelectByWrapper(query *QueryWrapper[T], result *[]*T) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db, err = applyWrapper(db, query, true, true, true)
	if err != nil {
		return 0, err
	}
	return checkResult(db.Scan(result))
}

// CountByCond 通过实体条件统计数据，条件中的零值字段将被自动忽略。
func (b BaseMapper[T]) CountByCond(query CondQuery[T]) (int64, error) {
	var count int64
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db = applyTimeRanges(db, query.TimeRanges)
	_, err = checkResult(db.Where(query.Condition).Count(&count))
	return count, err
}

// CountByMap 通过 Map 条件统计数据，支持显式零值条件。
func (b BaseMapper[T]) CountByMap(query MapQuery) (int64, error) {
	var count int64
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db = applyTimeRanges(db, query.TimeRanges)
	_, err = checkResult(db.Where(query.Condition).Count(&count))
	return count, err
}

// CountByWhere 通过原始 SQL 条件统计数据。
func (b BaseMapper[T]) CountByWhere(query WhereQuery) (int64, error) {
	var count int64
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db = applyTimeRanges(db, query.TimeRanges)
	_, err = checkResult(db.Where(query.RawWhereSQL, query.Args...).Count(&count))
	return count, err
}

// CountByGorm 通过原始Gorm查询数据总数
func (b BaseMapper[T]) CountByGorm(rawDB func(*gorm.DB)) (int64, error) {
	var count int64
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	rawDB(db)
	_, err = checkResult(db.Count(&count))
	return count, err
}

// CountByWrapper 使用 Wrapper 条件统计数据总数。
func (b BaseMapper[T]) CountByWrapper(query *QueryWrapper[T]) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db, err = applyWrapper(db, query, false, false, false)
	if err != nil {
		return 0, err
	}
	var count int64
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

func applyQueryLimit(db *gorm.DB, limit int) (*gorm.DB, error) {
	if limit < 0 {
		return nil, fmt.Errorf("%w: limit", ErrInvalidQueryRange)
	}
	if limit > 0 {
		db = db.Limit(limit)
	}
	return db, nil
}

// SelectPageByCond 通过实体条件分页查询，条件中的零值字段将被自动忽略。
func (b BaseMapper[T]) SelectPageByCond(query PageQuery[T], result *[]*T) (total int64, err error) {
	if query.Number <= 0 || query.Size <= 0 {
		return 0, ErrInvalidPage
	}
	countDB, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	countDB = applyTimeRanges(countDB, query.TimeRanges)
	_, err = checkResult(countDB.Where(query.Condition).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	selectDB, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	selectDB = applyTimeRanges(selectDB, query.TimeRanges)
	_, err = checkResult(selectDB.Select(query.SelectColumns).Where(query.Condition).Order(query.OrderBySQL).Limit(query.Size).Offset((query.Number - 1) * query.Size).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SelectPageByMap 通过 Map 条件分页查询，支持显式查询零值字段。
func (b BaseMapper[T]) SelectPageByMap(query MapPageQuery, result *[]*T) (total int64, err error) {
	if query.Number <= 0 || query.Size <= 0 {
		return 0, ErrInvalidPage
	}
	countDB, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	countDB = applyTimeRanges(countDB, query.TimeRanges)
	_, err = checkResult(countDB.Where(query.Condition).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	selectDB, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	selectDB = applyTimeRanges(selectDB, query.TimeRanges)
	_, err = checkResult(selectDB.Select(query.SelectColumns).Where(query.Condition).Order(query.OrderBySQL).Limit(query.Size).Offset((query.Number - 1) * query.Size).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SelectPageByWhere 通过原始 SQL 分页查询
func (b BaseMapper[T]) SelectPageByWhere(query WherePageQuery, result *[]*T) (total int64, err error) {
	if query.Number <= 0 || query.Size <= 0 {
		return 0, ErrInvalidPage
	}
	countDB, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	countDB = applyTimeRanges(countDB, query.TimeRanges)
	_, err = checkResult(countDB.Where(query.RawWhereSQL, query.Args...).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	selectDB, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	selectDB = applyTimeRanges(selectDB, query.TimeRanges)
	_, err = checkResult(selectDB.Select(query.SelectColumns).Where(query.RawWhereSQL, query.Args...).Order(query.OrderBySQL).Limit(query.Size).Offset((query.Number - 1) * query.Size).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SelectPageByGorm 通过原始Gorm分页查询
func (b BaseMapper[T]) SelectPageByGorm(countRawDB func(*gorm.DB), pageRawDB func(*gorm.DB), result *[]*T) (total int64, err error) {
	countDB, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	countRawDB(countDB)
	_, err = checkResult(countDB.Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	selectDB, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	pageRawDB(selectDB)
	_, err = checkResult(selectDB.Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SelectPageByWrapper 使用 Wrapper 条件分页查询。
func (b BaseMapper[T]) SelectPageByWrapper(query *PageWrapper[T], result *[]*T) (total int64, err error) {
	if query == nil {
		return 0, ErrNilQueryWrapper
	}
	if query.number <= 0 || query.size <= 0 {
		return 0, ErrInvalidPage
	}
	countDB, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	countDB, err = applyWrapper(countDB, query.query, false, false, false)
	if err != nil {
		return 0, err
	}
	if _, err = checkResult(countDB.Count(&total)); err != nil || total == 0 {
		return total, err
	}
	selectDB, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	selectDB, err = applyWrapper(selectDB, query.query, true, true, false)
	if err != nil {
		return 0, err
	}
	offset := (query.number - 1) * query.size
	_, err = checkResult(selectDB.Limit(query.size).Offset(offset).Scan(result))
	return total, err
}

// Insert 保存数据 零值也将参与保存
//
//	exclude 手动指定需要排除的字段名称 数据库字段/结构体字段名称
func (b BaseMapper[T]) Insert(entity *T, excludeColumns ...string) (int64, error) {
	if entity == nil {
		return 0, ErrNilEntity
	}
	db, err := b.rawDB()
	if err != nil {
		return 0, err
	}
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
	if err != nil {
		return 0, err
	}
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
	if err != nil {
		return 0, err
	}
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
	if err != nil {
		return 0, err
	}
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
	if err != nil {
		return 0, err
	}
	if len(excludeColumns) > 0 {
		db = db.Omit(excludeColumns...)
	}
	return checkResult(db.Save(entity))
}

// UpdateByID 通过实体中的主键 ID 更新含零值字段
// updateColumns 手动指定需要更新的列
func (b BaseMapper[T]) UpdateByID(updated *T, id any, updateColumns ...string) (int64, error) {
	if _, err := effectiveUpdateFields(updated, updateColumns...); err != nil {
		return 0, err
	}
	if err := validateID(id); err != nil {
		return 0, err
	}
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	return checkResult(db.Select(updateColumns).Where("id = ?", id).Updates(updated))
}

// UpdateByIDWithoutZeroFields 通过ID更新非零值字段
// allowZeroFieldColumns 额外指定需要更新零值字段
func (b BaseMapper[T]) UpdateByIDWithoutZeroFields(updated *T, id any, allowZeroFieldColumns ...string) (int64, error) {
	nonZeroFields, err := effectiveUpdateFields(updated, allowZeroFieldColumns...)
	if err != nil {
		return 0, err
	}
	if err = validateID(id); err != nil {
		return 0, err
	}
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	return checkResult(db.Select(nonZeroFields).Where("id = ?", id).Updates(updated))
}

// UpdateByIDWithMap 通过 ID 更新 Map 中指定的列和值
func (b BaseMapper[T]) UpdateByIDWithMap(updated map[string]any, id any) (int64, error) {
	if len(updated) == 0 {
		return 0, ErrNoFieldToUpdate
	}
	if err := validateID(id); err != nil {
		return 0, err
	}
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	return checkResult(db.Where("id = ?", id).Updates(updated))
}

// UpdateByCond 通过条件更新 条件：零值将自动忽略，更新：零值字段将被自动忽略
// updateColumns 需要指定更新的数据库字段 更新指定字段(支持零值字段)
func (b BaseMapper[T]) UpdateByCond(updated *T, condition T, updateColumns ...string) (int64, error) {
	if _, err := effectiveUpdateFields(updated, updateColumns...); err != nil {
		return 0, err
	}
	if err := validateStructCondition(condition); err != nil {
		return 0, err
	}
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	return checkResult(db.Select(updateColumns).Where(condition).Updates(updated))
}

// UpdateByCondWithZeroFields 通过条件更新，并指定可以更新的零值字段
func (b BaseMapper[T]) UpdateByCondWithZeroFields(updated *T, condition T, allowZeroFieldColumns ...string) (int64, error) {
	nonZeroFields, err := effectiveUpdateFields(updated, allowZeroFieldColumns...)
	if err != nil {
		return 0, err
	}
	if err = validateStructCondition(condition); err != nil {
		return 0, err
	}
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
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
	if err != nil {
		return 0, err
	}
	return checkResult(db.Where(condition).Updates(updated))
}

// UpdateByWhere 通过原始 Where SQL 条件更新非零实体字段
func (b BaseMapper[T]) UpdateByWhere(updated *T, rawWhereSQL string, args ...any) (int64, error) {
	if _, err := effectiveUpdateFields(updated); err != nil {
		return 0, err
	}
	if strings.TrimSpace(rawWhereSQL) == "" {
		return 0, ErrEmptyWhereSQL
	}
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	return checkResult(db.Where(rawWhereSQL, args...).Updates(updated))
}

// UpdateByWrapper 使用 Wrapper 条件和 Set 赋值更新数据。
func (b BaseMapper[T]) UpdateByWrapper(wrapper *UpdateWrapper[T]) (int64, error) {
	if wrapper == nil {
		return 0, ErrNilUpdateWrapper
	}
	if wrapper.query == nil || len(wrapper.query.predicates) == 0 {
		return 0, ErrEmptyCondition
	}
	if len(wrapper.assignments) == 0 {
		return 0, ErrNoFieldToUpdate
	}
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db, err = applyWrapper(db, wrapper.query, false, false, false)
	if err != nil {
		return 0, err
	}
	updates := make(map[string]any, len(wrapper.assignments))
	for _, assignment := range wrapper.assignments {
		if assignment.field == nil {
			return 0, ErrInvalidColumnSelector
		}
		columnName, err := resolveColumnName[T](db, assignment.field.wrapperFieldPath())
		if err != nil {
			return 0, err
		}
		updates[columnName] = assignment.value
	}
	return checkResult(db.Updates(updates))
}

// DeleteByID 通过 ID 删除单条数据
func (b BaseMapper[T]) DeleteByID(id any) (int64, error) {
	if err := validateID(id); err != nil {
		return 0, err
	}
	db, err := b.rawDB()
	if err != nil {
		return 0, err
	}
	return checkResult(db.Delete(b.model, id))
}

// DeleteByIDs 通过多个 ID 删除数据
func (b BaseMapper[T]) DeleteByIDs(ids []any) (int64, error) {
	if len(ids) == 0 {
		return 0, ErrEmptyIDs
	}
	db, err := b.rawDB()
	if err != nil {
		return 0, err
	}
	return checkResult(db.Delete(b.model, ids))
}

// DeleteByCond 通过条件删除 零值字段将被自动忽略
func (b BaseMapper[T]) DeleteByCond(condition T) (int64, error) {
	if err := validateStructCondition(condition); err != nil {
		return 0, err
	}
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	return checkResult(db.Where(condition).Delete(b.model))
}

// DeleteByWhere 通过原始 Where SQL 删除相关数据
func (b BaseMapper[T]) DeleteByWhere(rawWhereSQL string, args ...any) (int64, error) {
	if strings.TrimSpace(rawWhereSQL) == "" {
		return 0, ErrEmptyWhereSQL
	}
	db, err := b.rawDB()
	if err != nil {
		return 0, err
	}
	return checkResult(db.Where(rawWhereSQL, args...).Delete(b.model))
}

// DeleteByMap 通过Map类型条件删除
func (b BaseMapper[T]) DeleteByMap(condition map[string]any) (int64, error) {
	if len(condition) == 0 {
		return 0, ErrEmptyCondition
	}
	db, err := b.rawDB()
	if err != nil {
		return 0, err
	}
	return checkResult(db.Where(condition).Delete(b.model))
}
