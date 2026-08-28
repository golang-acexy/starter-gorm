package gormstarter

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/acexy/golang-toolkit/util/json"
	"gorm.io/gorm"
)

const (
	DBTypeMySQL    DBType = "mysql"
	DBTypePostgres DBType = "postgres"
)

// Timestamp 时间戳处理 接收数据库的时间类型
type Timestamp json.Timestamp

// NewTimestamp 创建时间戳
func NewTimestamp(time time.Time) Timestamp {
	return Timestamp{Time: time}
}

type DBType string

type Model interface {
	TableName() string
}

// ModelWithDBType 当gorm管理多个不同数据库类型时，需要实现此接口 以便指定该数据库类型 （若只有初始化了一种类型的数据库，不用指定，配置多种数据库类型时，默认缺省DBType 的默认为MySQL类型）
type ModelWithDBType interface {
	TableName() string
	DBType() DBType
}

type BaseMapper[M Model] struct {
	model M
	tx    *gorm.DB
}

// TimeRange 定义基于某个时间字段的起止范围过滤条件，使用左闭右开区间 [StartTime, EndTime)。
// StartTime / EndTime 为 nil 时表示该方向不限。
type TimeRange struct {
	Field     string     // 数据库时间字段名（列名），如 "created_at"
	StartTime *time.Time // 起始时间（含），nil 表示不限制下界
	EndTime   *time.Time // 结束时间（不含），nil 表示不限制上界
}

type QueryOptions struct {
	OrderBySQL    string
	SelectColumns []string
	TimeRanges    []TimeRange // 基于时间的过滤条件，支持多个时间字段
	Limit         int
}

type PageOptions struct {
	Number        int
	Size          int
	OrderBySQL    string
	SelectColumns []string
	TimeRanges    []TimeRange
}

type PageQuery[T Model] struct {
	Condition T
	PageOptions
}

type MapPageQuery struct {
	Condition map[string]any
	PageOptions
}

type WherePageQuery struct {
	RawWhereSQL string
	Args        []any
	PageOptions
}

type CondQuery[T Model] struct {
	Condition T
	QueryOptions
}

type MapQuery struct {
	Condition map[string]any
	QueryOptions
}

type WhereQuery struct {
	RawWhereSQL string
	Args        []any
	QueryOptions
}

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

func (t Timestamp) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil // 如果时间为零值，返回 nil
	}
	return t.Time, nil // 返回底层的 time.Time 类型
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return json.Time2Timestamp(t.Time)
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	formatTime, err := json.Timestamp2Time(data)
	if err != nil {
		return err
	}
	t.Time = formatTime
	return nil
}

// RawMapper 提供当前 Mapper 对应的原生 GORM 操作能力。
type RawMapper interface {
	// TableGormDB 获取已限定当前 Mapper 表名的原生 gorm.DB。
	TableGormDB() *gorm.DB

	// CurrentGormDB 获取当前 Mapper 使用的 gorm.DB；绑定事务时返回该事务，否则返回新的 gorm.DB。
	CurrentGormDB() *gorm.DB
}

// QueryMapper 提供查询能力。
type QueryMapper[T Model] interface {
	// SelectByID 通过主键查询数据
	SelectByID(id any, result *T) (int64, error)

	// SelectByIDs 通过主键查询数据
	SelectByIDs(ids []any, result *[]*T) (int64, error)

	// ExistsByID 判断指定主键的数据是否存在
	ExistsByID(id any) (bool, error)

	// SelectOneByCond 通过实体条件查询一条数据，条件中的零值字段将被自动忽略。
	SelectOneByCond(query CondQuery[T], result *T) (int64, error)

	// SelectByCond 通过条件查询 查询条件零值字段将被自动忽略
	// specifyColumns 指定只需要查询的数据库字段
	SelectByCond(query CondQuery[T], result *[]*T) (int64, error)

	// SelectOneByMap 通过指定字段与值查询数据 解决查询条件零值问题
	// specifyColumns 指定只需要查询的数据库字段
	SelectOneByMap(query MapQuery, result *T) (int64, error)

	// SelectByMap 通过指定字段与值查询数据 解决零值条件问题
	// specifyColumns 指定只需要查询的数据库字段
	SelectByMap(query MapQuery, result *[]*T) (int64, error)

	// SelectOneByWhere 通过原始 Where SQL 查询，只需要传入 rawWhereSQL 和参数
	SelectOneByWhere(query WhereQuery, result *T) (int64, error)

	// SelectByWhere 通过原始 Where SQL 查询，只需要传入 rawWhereSQL 和参数
	SelectByWhere(query WhereQuery, result *[]*T) (int64, error)

	// SelectOneByGorm 通过原始Gorm查询单条数据 构建Gorm查询条件
	SelectOneByGorm(result *T, rawDB func(*gorm.DB)) (int64, error)

	// SelectByGorm 通过原始Gorm查询数据
	SelectByGorm(result *[]*T, rawDB func(*gorm.DB)) (int64, error)

	// SelectOneByWrapper 使用类型安全的 Wrapper 查询一条数据。
	SelectOneByWrapper(query *QueryWrapper[T], result *T) (int64, error)

	// SelectByWrapper 使用类型安全的 Wrapper 查询数据。
	SelectByWrapper(query *QueryWrapper[T], result *[]*T) (int64, error)

	// CountByCond 通过实体条件统计数据，条件中的零值字段将被自动忽略。
	CountByCond(query CondQuery[T]) (int64, error)

	// CountByMap 通过 Map 条件统计数据，支持显式零值条件。
	CountByMap(query MapQuery) (int64, error)

	// CountByWhere 通过原始 SQL 条件统计数据。
	CountByWhere(query WhereQuery) (int64, error)

	// CountByGorm 通过原始Gorm查询数据总数
	CountByGorm(rawDB func(*gorm.DB)) (int64, error)

	// CountByWrapper 使用 Wrapper 条件统计数据总数，忽略投影、排序、Limit 和 Offset。
	CountByWrapper(query *QueryWrapper[T]) (int64, error)

	// SelectPageByCond 通过实体条件分页查询，条件中的零值字段将被自动忽略。
	SelectPageByCond(query PageQuery[T], result *[]*T) (total int64, err error)

	// SelectPageByMap 通过 Map 条件分页查询，支持显式查询零值字段。
	SelectPageByMap(query MapPageQuery, result *[]*T) (total int64, err error)

	// SelectPageByWhere 通过原始 SQL 分页查询
	SelectPageByWhere(query WherePageQuery, result *[]*T) (total int64, err error)

	// SelectPageByGorm 通过原始Gorm分页查询
	SelectPageByGorm(countRawDB func(*gorm.DB), pageRawDB func(*gorm.DB), result *[]*T) (total int64, err error)

	// SelectPageByWrapper 使用 Wrapper 条件分页查询。
	SelectPageByWrapper(query *PageWrapper[T], result *[]*T) (total int64, err error)
}

// InsertMapper 提供新增能力。
type InsertMapper[T Model] interface {
	// Insert 保存数据 零值也将参与保存
	//	exclude 手动指定需要排除的字段名称 数据库字段/结构体字段名称
	Insert(entity *T, excludeColumns ...string) (int64, error)

	// InsertBatch 批量新增 零值也将参与保存
	//	exclude 手动指定需要排除的字段名称 数据库字段/结构体字段
	InsertBatch(entities []*T, excludeColumns ...string) (int64, error)

	// InsertWithoutZeroFields 保存数据 零值字段将不会参与保存
	InsertWithoutZeroFields(entity *T) (int64, error)

	// InsertWithMap 通过 Map 类型保存数据
	InsertWithMap(entity map[string]any) (int64, error)

	// InsertOrUpdateByPrimaryKey 保存/更新数据 零值也将参与保存
	// exclude 手动指定需要排除的字段名称 数据库字段/结构体字段 (如果触发的是update 创建时间可能会被错误的修改，可以通过excludeColumns来指定排除创建时间字段)
	// 仅根据主键冲突默认支持update 更多操作需要参阅 https://gorm.io/zh_CN/docs/create.html#upsert
	InsertOrUpdateByPrimaryKey(entity *T, excludeColumns ...string) (int64, error)
}

// UpdateMapper 提供更新能力。
type UpdateMapper[T Model] interface {
	// UpdateByID 通过实体中的主键 ID 更新含零值字段
	// updateColumns 手动指定需要更新的列
	UpdateByID(updated *T, id any, updateColumns ...string) (int64, error)

	// UpdateByIDWithoutZeroFields 通过ID更新非零值字段
	// allowZeroFieldColumns 额外指定需要更新零值字段
	UpdateByIDWithoutZeroFields(updated *T, id any, allowZeroFieldColumns ...string) (int64, error)

	// UpdateByIDWithMap 通过 ID 更新 Map 中指定的列和值
	UpdateByIDWithMap(updated map[string]any, id any) (int64, error)

	// UpdateByCond 通过条件更新 条件：零值将自动忽略，更新：零值字段将被自动忽略
	// updateColumns 需要指定更新的数据库字段 更新指定字段(支持零值字段)
	UpdateByCond(updated *T, condition T, updateColumns ...string) (int64, error)

	// UpdateByCondWithZeroFields 通过条件更新，并指定可以更新的零值字段
	UpdateByCondWithZeroFields(updated *T, condition T, allowZeroFieldColumns ...string) (int64, error)

	// UpdateByMap 通过Map类型条件更新
	UpdateByMap(updated, condition map[string]any) (int64, error)

	// UpdateByWhere 通过原始 Where SQL 条件更新非零实体字段
	UpdateByWhere(updated *T, rawWhereSQL string, args ...any) (int64, error)

	// UpdateByWrapper 使用 Wrapper 条件和 Set 赋值更新数据。
	UpdateByWrapper(wrapper *UpdateWrapper[T]) (int64, error)
}

// DeleteMapper 提供删除能力。
type DeleteMapper[T Model] interface {
	// DeleteByID 通过 ID 删除单条数据
	DeleteByID(id any) (int64, error)

	// DeleteByIDs 通过多个 ID 删除数据
	DeleteByIDs(ids []any) (int64, error)

	// DeleteByCond 通过条件删除 零值字段将被自动忽略
	DeleteByCond(condition T) (int64, error)

	// DeleteByMap 通过Map类型条件删除
	DeleteByMap(condition map[string]any) (int64, error)

	// DeleteByWhere 通过原始 Where SQL 删除相关数据
	DeleteByWhere(rawWhereSQL string, args ...any) (int64, error)
}

// Mapper 聚合原生 GORM、查询、新增、更新和删除能力。
type Mapper[T Model] interface {
	RawMapper
	QueryMapper[T]
	InsertMapper[T]
	UpdateMapper[T]
	DeleteMapper[T]
}
