package gormstarter

import (
	"database/sql"
	"fmt"
	"math"
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

// validateStructCondition rejects reflected zero values, not every condition that GORM could omit.
func validateStructCondition[T any](condition T) error {
	if stdreflect.ValueOf(condition).IsZero() {
		return ErrEmptyCondition
	}
	return nil
}

// validateID rejects nil, typed nil, and reflected zero IDs; it does not sanitize GORM string conditions.
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

// effectiveUpdateFields collects reflected nonzero and explicit fields, excluding ID as the sole payload.
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

// TableGormDB returns a table-scoped handle, or nil on routing failure; caller must not close the shared pool.
func (b BaseMapper[T]) TableGormDB() *gorm.DB {
	db, _ := b.tableDB()
	return db
}

// CurrentGormDB returns the bound transaction or shared routed DB, not a new isolated session; errors become nil.
func (b BaseMapper[T]) CurrentGormDB() *gorm.DB {
	db, _ := b.rawDB()
	return db
}

// GetBaseMapperWithTx returns a mapper copy borrowing tx; nil selects ordinary routing. Caller owns commit/rollback.
func (b BaseMapper[T]) GetBaseMapperWithTx(tx *gorm.DB) BaseMapper[T] {
	return BaseMapper[T]{
		model: b.model,
		tx:    tx,
	}
}

// Wrapper creates a mutable query builder; do not construct or mutate it concurrently.
func (b BaseMapper[T]) Wrapper() *QueryWrapper[T] {
	return NewQueryWrapper[T]()
}

// PageWrapper creates a mutable page builder; positive page inputs are validated at execution.
func (b BaseMapper[T]) PageWrapper(number, size int) *PageWrapper[T] {
	return newPageWrapper[T](number, size)
}

// UpdateWrapper creates a mutable update builder requiring predicates and assignments at execution.
func (b BaseMapper[T]) UpdateWrapper() *UpdateWrapper[T] {
	return newUpdateWrapper[T]()
}

// NewBaseMapperWithTx begins an independent transaction on the routed database; caller owns commit/rollback.
// Routing errors are currently ignored and can cause a nil-pointer panic. Check availability first
// and inspect the returned CurrentGormDB().Error for Begin failures; this does not create a nested transaction.
func (b BaseMapper[T]) NewBaseMapperWithTx(opts ...*sql.TxOptions) BaseMapper[T] {
	baseMapper := BaseMapper[T]{
		model: b.model,
	}
	db, _ := baseMapper.rawDB()
	baseMapper.tx = db.Begin(opts...)
	return baseMapper
}

// SelectByID binds the ID to column id and scans into a nonnil destination; no match returns zero rows.
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

// SelectByIDs binds an ID batch to column id; an empty batch is rejected and results are not ordered by input IDs.
func (b BaseMapper[T]) SelectByIDs(ids []any, result *[]*T) (int64, error) {
	if len(ids) == 0 {
		return 0, ErrEmptyIDs
	}
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	return checkResult(db.Where("id in ?", ids).Scan(result))
}

// ExistsByID counts column-id matches; empty IDs are rejected and driver errors are propagated.
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

// SelectOneByCond scans at most one match; GORM omits zero-valued struct condition fields.
func (b BaseMapper[T]) SelectOneByCond(query CondQuery[T], result *T) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db = applyTimeRanges(db, query.TimeRanges)
	return checkResult(db.Select(query.SelectColumns).Where(query.Condition).Order(query.OrderBySQL).Limit(1).Scan(result))
}

// SelectOneByMap scans at most one match using explicit map predicates, including zero values.
func (b BaseMapper[T]) SelectOneByMap(query MapQuery, result *T) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	db = applyTimeRanges(db, query.TimeRanges)
	return checkResult(db.Select(query.SelectColumns).Where(query.Condition).Order(query.OrderBySQL).Limit(1).Scan(result))
}

// SelectOneByWhere scans a trusted SQL filter without adding LIMIT 1 and warns if multiple rows are scanned.
// Callers own cardinality and must bind runtime values through Args.
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
		// The caller owns raw SQL cardinality; warn when a single-result request scans multiple rows.
		logger.Logrus().WithFields(map[string]any{
			"table":        b.model.TableName(),
			"rowsAffected": rowsAffected,
		}).Warnln("SelectOneByWhere queried more than one row")
	}
	return rowsAffected, nil
}

// SelectOneByGorm invokes a nonnil callback on the table-scoped DB, then scans; the callback owns cardinality.
func (b BaseMapper[T]) SelectOneByGorm(result *T, rawDB func(*gorm.DB)) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	rawDB(db)
	return checkResult(db.Scan(result))
}

// SelectByCond scans struct-filter matches; zero condition fields are omitted and zero Limit is unrestricted.
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

// SelectByMap scans map-filter matches, preserving zero predicates; negative Limit is rejected.
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

// SelectByWhere scans a trusted SQL filter with bound Args; negative Limit is rejected.
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

// SelectByGorm invokes a nonnil callback on the table-scoped DB and scans into the caller-owned slice.
func (b BaseMapper[T]) SelectByGorm(result *[]*T, rawDB func(*gorm.DB)) (int64, error) {
	db, err := b.tableDB()
	if err != nil {
		return 0, err
	}
	rawDB(db)
	return checkResult(db.Scan(result))
}

// SelectOneByWrapper scans at most one match; wrapper Limit and Offset are not applied.
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

// SelectByWrapper scans wrapper matches with projection, ordering, and validated range settings.
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

// CountByCond applies struct predicates and time ranges; zero fields are omitted and projection/order/Limit are ignored.
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

// CountByMap applies map predicates and time ranges; projection/order/Limit are ignored.
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

// CountByWhere counts a trusted SQL filter with bound Args and time ranges, ignoring projection/order/Limit.
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

// CountByGorm invokes a nonnil callback on the table-scoped DB; the callback owns count filters.
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

// CountByWrapper counts wrapper predicates while ignoring projection, order, Limit, and Offset.
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

// applyTimeRanges uses trusted field expressions and bound values with inclusive starts and exclusive ends.
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

// pageOffset rejects nonpositive page inputs and integer multiplication overflow.
func pageOffset(number, size int) (int, error) {
	if number <= 0 || size <= 0 {
		return 0, ErrInvalidPage
	}
	if number-1 > math.MaxInt/size {
		return 0, ErrInvalidPage
	}
	return (number - 1) * size, nil
}

// SelectPageByCond counts then scans a struct-filter page; positive page inputs are required and zero fields are omitted.
// Count and scan are separate statements; zero count leaves the destination untouched. No snapshot is created.
func (b BaseMapper[T]) SelectPageByCond(query PageQuery[T], result *[]*T) (total int64, err error) {
	offset, err := pageOffset(query.Number, query.Size)
	if err != nil {
		return 0, err
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
	_, err = checkResult(selectDB.Select(query.SelectColumns).Where(query.Condition).Order(query.OrderBySQL).Limit(query.Size).Offset(offset).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SelectPageByMap counts then scans a map-filter page; positive page inputs are required.
// Count and scan are separate statements; zero count leaves the destination untouched. No snapshot is created.
func (b BaseMapper[T]) SelectPageByMap(query MapPageQuery, result *[]*T) (total int64, err error) {
	offset, err := pageOffset(query.Number, query.Size)
	if err != nil {
		return 0, err
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
	_, err = checkResult(selectDB.Select(query.SelectColumns).Where(query.Condition).Order(query.OrderBySQL).Limit(query.Size).Offset(offset).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SelectPageByWhere counts then scans a trusted SQL-filter page using bound Args and positive page inputs.
// Count and scan are separate statements; zero count leaves the destination untouched. No snapshot is created.
func (b BaseMapper[T]) SelectPageByWhere(query WherePageQuery, result *[]*T) (total int64, err error) {
	offset, err := pageOffset(query.Number, query.Size)
	if err != nil {
		return 0, err
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
	_, err = checkResult(selectDB.Select(query.SelectColumns).Where(query.RawWhereSQL, query.Args...).Order(query.OrderBySQL).Limit(query.Size).Offset(offset).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SelectPageByGorm invokes nonnil count/data callbacks separately; callers supply matching filters and pagination.
// Zero count skips the data callback and leaves the result destination untouched.
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

// SelectPageByWrapper counts then scans a page wrapper; nil wrappers and invalid page inputs are rejected.
// Count and scan are separate statements; zero count leaves the destination untouched. No snapshot is created.
func (b BaseMapper[T]) SelectPageByWrapper(query *PageWrapper[T], result *[]*T) (total int64, err error) {
	if query == nil {
		return 0, ErrNilQueryWrapper
	}
	offset, err := pageOffset(query.number, query.size)
	if err != nil {
		return 0, err
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
	_, err = checkResult(selectDB.Limit(query.size).Offset(offset).Scan(result))
	return total, err
}

// Insert applies GORM Create to a nonnil entity; excluded columns are omitted and model defaults follow GORM semantics.
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

// InsertWithoutZeroFields selects reflected nonzero Go fields; nil or empty effective data is rejected.
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

// InsertBatch applies GORM Create to a nonempty batch of nonnil entities; excluded columns are omitted.
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

// InsertWithMap creates an explicit nonempty column/value map without model-field zero omission.
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

// InsertOrUpdateByPrimaryKey delegates to GORM Save, not an arbitrary conflict-target upsert.
// Excluded columns are omitted; full-field updates can overwrite creation timestamps unless excluded.
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

// UpdateByID updates column-id matches; without selected columns GORM omits zero-valued struct updates.
// Explicitly selected columns include zero values. Empty IDs or effective update data are rejected.
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

// UpdateByIDWithoutZeroFields selects nonzero fields plus allowZeroFieldColumns; the ID itself is not sufficient update data.
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

// UpdateByIDWithMap applies explicit nonempty map values, including zero/nil, to a validated column-id match.
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

// UpdateByCond omits zero struct predicates and, unless columns are selected, zero struct update values.
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

// UpdateByCondWithZeroFields omits zero struct predicates and selects nonzero updates plus allowZeroFieldColumns.
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

// UpdateByMap requires nonempty predicate and update maps and preserves explicit zero/nil values.
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

// UpdateByWhere requires trusted nonblank SQL with bound args and nonempty effective struct update data.
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

// UpdateByWrapper requires predicates and assignments; zero/nil assignments are kept and the last repeated field wins.
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

// DeleteByID delegates ID conditions to GORM Delete and follows model soft-delete behavior.
// Nonnumeric strings are interpreted as SQL clauses, not bound IDs; do not pass arbitrary string IDs.
// Use a bound Where/Map filter for ordinary string keys until this shortcut is changed.
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

// DeleteByIDs delegates a nonempty ID batch to GORM Delete; individual IDs are not validated and soft deletion follows the model.
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

// DeleteByCond rejects a reflected zero condition; GORM omits zero-valued struct predicates and applies model soft deletion.
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

// DeleteByWhere requires a trusted nonblank SQL filter with bound args and follows model soft deletion.
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

// DeleteByMap requires a nonempty predicate map and follows model soft deletion.
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
