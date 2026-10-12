package gormstarter

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/acexy/golang-toolkit/util/json"
	"gorm.io/gorm"
)

const (
	// DBTypeMySQL identifies the MySQL registry entry.
	DBTypeMySQL    DBType = "mysql"
	// DBTypePostgres identifies the PostgreSQL registry entry.
	DBTypePostgres DBType = "postgres"
)

// Timestamp scans SQL time.Time/null and uses toolkit global JSON timestamp units; synchronize shared mutation.
type Timestamp json.Timestamp

// NewTimestamp wraps a time value without changing its location.
func NewTimestamp(time time.Time) Timestamp {
	return Timestamp{Time: time}
}

// DBType identifies a registered database dialect.
type DBType string

// Model supplies a stable table name and must support TableName on its zero value.
type Model interface {
	// TableName returns stable model metadata and must support the model zero value.
	TableName() string
}

// ModelWithDBType selects a database dialect; a bound transaction takes precedence over routing.
type ModelWithDBType interface {
	// TableName returns stable model metadata and must support the model zero value.
	TableName() string
	// DBType returns stable model metadata and must support the model zero value.
	DBType() DBType
}

// BaseMapper supplies driver-facing operations with caller-owned result destinations.
// Its model zero value must support TableName and optional DBType. Shared DB handles do not
// make mutable result buffers, filters, query options, or chained sessions concurrency-safe.
type BaseMapper[M Model] struct {
	model M
	tx    *gorm.DB
}

// TimeRange filters a trusted database field over [StartTime, EndTime); nil endpoints omit that bound.
type TimeRange struct {
	// Field is a trusted SQL field expression; it is not quoted or allowlisted by this range API.
	Field     string
	// StartTime is an inclusive lower bound; nil omits it.
	StartTime *time.Time
	// EndTime is an exclusive upper bound; nil omits it.
	EndTime   *time.Time
}

// QueryOptions groups trusted ordering/projection, time filters, and a list-only limit.
type QueryOptions struct {
	// OrderBySQL is a trusted ordering expression; do not pass unvalidated external SQL.
	OrderBySQL    string
	// SelectColumns supplies trusted projection columns or expressions.
	SelectColumns []string
	// TimeRanges adds half-open filters; endpoint pointers are borrowed and range ordering is not validated.
	TimeRanges    []TimeRange
	// Limit applies only to list queries; zero is unrestricted and negatives are rejected.
	Limit         int
}

// PageOptions contains positive page inputs plus trusted ordering/projection and time filters.
type PageOptions struct {
	// Number is the positive one-based page input.
	Number        int
	// Size is the positive page size; no shared maximum is imposed.
	Size          int
	// OrderBySQL is a trusted ordering expression; do not pass unvalidated external SQL.
	OrderBySQL    string
	// SelectColumns supplies trusted projection columns or expressions.
	SelectColumns []string
	// TimeRanges adds half-open filters; endpoint pointers are borrowed and range ordering is not validated.
	TimeRanges    []TimeRange
}

// PageQuery combines struct predicates, whose zero fields are omitted, with page options.
type PageQuery[T Model] struct {
	// Condition supplies predicates; struct zero fields are omitted and maps remain borrowed.
	Condition T
	// PageOptions supplies page inputs, ordering, projection, and time filters.
	PageOptions
}

// MapPageQuery combines a borrowed predicate map with page options.
type MapPageQuery struct {
	// Condition supplies predicates; struct zero fields are omitted and maps remain borrowed.
	Condition map[string]any
	// PageOptions supplies page inputs, ordering, projection, and time filters.
	PageOptions
}

// WherePageQuery combines trusted SQL, bound arguments, and page options.
type WherePageQuery struct {
	// RawWhereSQL is a trusted SQL filter; runtime values must use Args.
	RawWhereSQL string
	// Args supplies bound values; referenced argument values remain borrowed.
	Args        []any
	// PageOptions supplies page inputs, ordering, projection, and time filters.
	PageOptions
}

// CondQuery combines struct predicates, whose zero fields are omitted, with query options.
type CondQuery[T Model] struct {
	// Condition supplies predicates; struct zero fields are omitted and maps remain borrowed.
	Condition T
	// QueryOptions supplies ordering, projection, time filters, and list limits.
	QueryOptions
}

// MapQuery combines a borrowed predicate map preserving zero values with query options.
type MapQuery struct {
	// Condition supplies predicates; struct zero fields are omitted and maps remain borrowed.
	Condition map[string]any
	// QueryOptions supplies ordering, projection, time filters, and list limits.
	QueryOptions
}

// WhereQuery combines trusted SQL and bound arguments with query options.
type WhereQuery struct {
	// RawWhereSQL is a trusted SQL filter; runtime values must use Args.
	RawWhereSQL string
	// Args supplies bound values; referenced argument values remain borrowed.
	Args        []any
	// QueryOptions supplies ordering, projection, time filters, and list limits.
	QueryOptions
}

// Scan accepts time.Time or nil on a nonnil receiver; null resets it and unsupported types preserve it.
func (t *Timestamp) Scan(value any) error {
	if value == nil {
		*t = Timestamp{Time: time.Time{}}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*t = Timestamp{Time: v}
	default:
		return fmt.Errorf("cannot scan type %T into Timestamp", v)
	}
	return nil
}

// Value returns SQL null for zero time, otherwise the underlying time.Time.
func (t Timestamp) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil // Represent zero time as SQL null.
	}
	return t.Time, nil // Return the underlying SQL-compatible time value.
}

// MarshalJSON uses toolkit timestamp units; zero time encodes as numeric zero.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return json.Time2Timestamp(t.Time)
}

// UnmarshalJSON uses toolkit timestamp units on a nonnil receiver; null resets it and errors preserve it.
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	formatTime, err := json.Timestamp2Time(data)
	if err != nil {
		return err
	}
	t.Time = formatTime
	return nil
}

// RawMapper exposes borrowed GORM handles; caller must not close their shared pools.
type RawMapper interface {
	// TableGormDB returns a table-scoped handle, or nil on routing failure; caller must not close the shared pool.
	TableGormDB() *gorm.DB
	// CurrentGormDB returns the bound transaction or shared routed DB, not a new isolated session; errors become nil.
	CurrentGormDB() *gorm.DB
}

// QueryMapper scans into caller-owned nonnil destinations; zero rows generally do not return record-not-found errors.
type QueryMapper[T Model] interface {
	// SelectByID binds the ID to column id and scans into a nonnil destination; no match returns zero rows.
	SelectByID(id any, result *T) (int64, error)
	// SelectByIDs binds an ID batch to column id; an empty batch is rejected and results are not ordered by input IDs.
	SelectByIDs(ids []any, result *[]*T) (int64, error)
	// ExistsByID counts column-id matches; empty IDs are rejected and driver errors are propagated.
	ExistsByID(id any) (bool, error)
	// SelectOneByCond scans at most one match; GORM omits zero-valued struct condition fields.
	SelectOneByCond(query CondQuery[T], result *T) (int64, error)
	// SelectByCond scans struct-filter matches; zero condition fields are omitted and zero Limit is unrestricted.
	SelectByCond(query CondQuery[T], result *[]*T) (int64, error)
	// SelectOneByMap scans at most one match using explicit map predicates, including zero values.
	SelectOneByMap(query MapQuery, result *T) (int64, error)
	// SelectByMap scans map-filter matches, preserving zero predicates; negative Limit is rejected.
	SelectByMap(query MapQuery, result *[]*T) (int64, error)
	// SelectOneByWhere scans a trusted SQL filter without adding LIMIT 1 and warns if multiple rows are scanned.
	// Callers own cardinality and must bind runtime values through Args.
	SelectOneByWhere(query WhereQuery, result *T) (int64, error)
	// SelectByWhere scans a trusted SQL filter with bound Args; negative Limit is rejected.
	SelectByWhere(query WhereQuery, result *[]*T) (int64, error)
	// SelectOneByGorm invokes a nonnil callback on the table-scoped DB, then scans; the callback owns cardinality.
	SelectOneByGorm(result *T, rawDB func(*gorm.DB)) (int64, error)
	// SelectByGorm invokes a nonnil callback on the table-scoped DB and scans into the caller-owned slice.
	SelectByGorm(result *[]*T, rawDB func(*gorm.DB)) (int64, error)
	// SelectOneByWrapper scans at most one match; wrapper Limit and Offset are not applied.
	SelectOneByWrapper(query *QueryWrapper[T], result *T) (int64, error)
	// SelectByWrapper scans wrapper matches with projection, ordering, and validated range settings.
	SelectByWrapper(query *QueryWrapper[T], result *[]*T) (int64, error)
	// CountByCond applies struct predicates and time ranges; zero fields are omitted and projection/order/Limit are ignored.
	CountByCond(query CondQuery[T]) (int64, error)
	// CountByMap applies map predicates and time ranges; projection/order/Limit are ignored.
	CountByMap(query MapQuery) (int64, error)
	// CountByWhere counts a trusted SQL filter with bound Args and time ranges, ignoring projection/order/Limit.
	CountByWhere(query WhereQuery) (int64, error)
	// CountByGorm invokes a nonnil callback on the table-scoped DB; the callback owns count filters.
	CountByGorm(rawDB func(*gorm.DB)) (int64, error)
	// CountByWrapper counts wrapper predicates while ignoring projection, order, Limit, and Offset.
	CountByWrapper(query *QueryWrapper[T]) (int64, error)
	// SelectPageByCond counts then scans a struct-filter page; positive page inputs are required and zero fields are omitted.
	// Count and scan are separate statements; zero count leaves the destination untouched. No snapshot is created.
	SelectPageByCond(query PageQuery[T], result *[]*T) (total int64, err error)
	// SelectPageByMap counts then scans a map-filter page; positive page inputs are required.
	// Count and scan are separate statements; zero count leaves the destination untouched. No snapshot is created.
	SelectPageByMap(query MapPageQuery, result *[]*T) (total int64, err error)
	// SelectPageByWhere counts then scans a trusted SQL-filter page using bound Args and positive page inputs.
	// Count and scan are separate statements; zero count leaves the destination untouched. No snapshot is created.
	SelectPageByWhere(query WherePageQuery, result *[]*T) (total int64, err error)
	// SelectPageByGorm invokes nonnil count/data callbacks separately; callers supply matching filters and pagination.
	// Zero count skips the data callback and leaves the result destination untouched.
	SelectPageByGorm(countRawDB func(*gorm.DB), pageRawDB func(*gorm.DB), result *[]*T) (total int64, err error)
	// SelectPageByWrapper counts then scans a page wrapper; nil wrappers and invalid page inputs are rejected.
	// Count and scan are separate statements; zero count leaves the destination untouched. No snapshot is created.
	SelectPageByWrapper(query *PageWrapper[T], result *[]*T) (total int64, err error)
}

// InsertMapper writes model or map data using GORM defaults and returns driver-reported affected rows.
type InsertMapper[T Model] interface {
	// Insert applies GORM Create to a nonnil entity; excluded columns are omitted and model defaults follow GORM semantics.
	Insert(entity *T, excludeColumns ...string) (int64, error)
	// InsertBatch applies GORM Create to a nonempty batch of nonnil entities; excluded columns are omitted.
	InsertBatch(entities []*T, excludeColumns ...string) (int64, error)
	// InsertWithoutZeroFields selects reflected nonzero Go fields; nil or empty effective data is rejected.
	InsertWithoutZeroFields(entity *T) (int64, error)
	// InsertWithMap creates an explicit nonempty column/value map without model-field zero omission.
	InsertWithMap(entity map[string]any) (int64, error)
	// InsertOrUpdateByPrimaryKey delegates to GORM Save, not an arbitrary conflict-target upsert.
	// Excluded columns are omitted; full-field updates can overwrite creation timestamps unless excluded.
	InsertOrUpdateByPrimaryKey(entity *T, excludeColumns ...string) (int64, error)
}

// UpdateMapper applies guarded updates; affected-row semantics depend on the database driver and connection settings.
type UpdateMapper[T Model] interface {
	// UpdateByID updates column-id matches; without selected columns GORM omits zero-valued struct updates.
	// Explicitly selected columns include zero values. Empty IDs or effective update data are rejected.
	UpdateByID(updated *T, id any, updateColumns ...string) (int64, error)
	// UpdateByIDWithoutZeroFields selects nonzero fields plus allowZeroFieldColumns; the ID itself is not sufficient update data.
	UpdateByIDWithoutZeroFields(updated *T, id any, allowZeroFieldColumns ...string) (int64, error)
	// UpdateByIDWithMap applies explicit nonempty map values, including zero/nil, to a validated column-id match.
	UpdateByIDWithMap(updated map[string]any, id any) (int64, error)
	// UpdateByCond omits zero struct predicates and, unless columns are selected, zero struct update values.
	UpdateByCond(updated *T, condition T, updateColumns ...string) (int64, error)
	// UpdateByCondWithZeroFields omits zero struct predicates and selects nonzero updates plus allowZeroFieldColumns.
	UpdateByCondWithZeroFields(updated *T, condition T, allowZeroFieldColumns ...string) (int64, error)
	// UpdateByMap requires nonempty predicate and update maps and preserves explicit zero/nil values.
	UpdateByMap(updated, condition map[string]any) (int64, error)
	// UpdateByWhere requires trusted nonblank SQL with bound args and nonempty effective struct update data.
	UpdateByWhere(updated *T, rawWhereSQL string, args ...any) (int64, error)
	// UpdateByWrapper requires predicates and assignments; zero/nil assignments are kept and the last repeated field wins.
	UpdateByWrapper(wrapper *UpdateWrapper[T]) (int64, error)
}

// DeleteMapper applies guarded deletion and follows model soft-delete settings; DeleteByID has string-condition caveats.
type DeleteMapper[T Model] interface {
	// DeleteByID delegates ID conditions to GORM Delete and follows model soft-delete behavior.
	// Nonnumeric strings are interpreted as SQL clauses, not bound IDs; do not pass arbitrary string IDs.
	// Use a bound Where/Map filter for ordinary string keys until this shortcut is changed.
	DeleteByID(id any) (int64, error)
	// DeleteByIDs delegates a nonempty ID batch to GORM Delete; individual IDs are not validated and soft deletion follows the model.
	DeleteByIDs(ids []any) (int64, error)
	// DeleteByCond rejects a reflected zero condition; GORM omits zero-valued struct predicates and applies model soft deletion.
	DeleteByCond(condition T) (int64, error)
	// DeleteByMap requires a nonempty predicate map and follows model soft deletion.
	DeleteByMap(condition map[string]any) (int64, error)
	// DeleteByWhere requires a trusted nonblank SQL filter with bound args and follows model soft deletion.
	DeleteByWhere(rawWhereSQL string, args ...any) (int64, error)
}

// Mapper aggregates raw handles and query, insert, update, and delete APIs.
type Mapper[T Model] interface {
	RawMapper
	QueryMapper[T]
	InsertMapper[T]
	UpdateMapper[T]
	DeleteMapper[T]
}
