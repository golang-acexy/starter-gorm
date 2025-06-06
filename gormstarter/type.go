package gormstarter

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/acexy/golang-toolkit/util/json"
	"gorm.io/gorm"
	"time"
)

const (
	DBTypeMySQL    DBType = "mysql"
	DBTypePostgres DBType = "postgres"
)

// Timestamp 时间戳处理 接收数据库的时间类型
type Timestamp json.Timestamp

type DBType string

type BaseModel[IdType any] struct {
	ID IdType `gorm:"<-:false;primaryKey" json:"id"`
}

type IBaseModel interface {
	TableName() string
}

// IBaseModelWithDBType 当gorm管理多个不同数据库类型时，需要实现此接口 以便指定该数据库类型 （初始化加载的第一个数据库类型不需要指定）
type IBaseModelWithDBType interface {
	TableName() string
	DBType() DBType
}

type BaseMapper[M IBaseModel] struct {
	model M
	tx    *gorm.DB
}

func (t *Timestamp) Scan(value interface{}) error {
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

func (t Timestamp) UnmarshalJSON(data []byte) error {
	formatTime, err := json.Timestamp2Time(data)
	if err != nil {
		return err
	}
	t.Time = formatTime
	return nil
}

type IBaseMapper[M BaseMapper[T], T IBaseModel] interface {

	// GormWithTableName Mapper对应的原生Gorm操作能力 获取到的原始gorm.DB已经限定当前Mapper所对应的表名
	GormWithTableName() *gorm.DB

	// CurrentGorm 获取当前Mapper所使用的gorm.DB 如果当前Mapper已使用指定的事务，则返回当前Mapper所使用的事务，否则获取新的gorm.DB
	CurrentGorm() *gorm.DB

	// GetBaseMapperWithTx 获取携带指定事务的基础Mapper
	GetBaseMapperWithTx(tx *gorm.DB) BaseMapper[T]

	// NewBaseMapperWithTx 创建一个全新事务的基础Mapper
	NewBaseMapperWithTx(opts ...*sql.TxOptions) BaseMapper[T]

	// SelectById 通过主键查询数据
	SelectById(id any, result *T) (int64, error)

	// SelectByIds 通过主键查询数据
	SelectByIds(id []any, result *[]*T) (int64, error)

	// SelectOneByCond 通过条件查询 查询条件零值字段将被自动忽略
	// specifyColumns 指定只需要查询的数据库字段
	SelectOneByCond(condition, result *T, specifyColumns ...string) (int64, error)

	// SelectByCond 通过条件查询 查询条件零值字段将被自动忽略
	// specifyColumns 指定只需要查询的数据库字段
	SelectByCond(condition *T, orderBy string, result *[]*T, specifyColumns ...string) (int64, error)

	// SelectOneByMap 通过指定字段与值查询数据 解决查询条件零值问题
	// specifyColumns 指定只需要查询的数据库字段
	SelectOneByMap(condition map[string]any, result *T, specifyColumns ...string) (int64, error)

	// SelectByMap 通过指定字段与值查询数据 解决零值条件问题
	// specifyColumns 指定只需要查询的数据库字段
	SelectByMap(condition map[string]any, orderBy string, result *[]*T, specifyColumns ...string) (int64, error)

	// SelectOneByWhere 通过原始Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
	SelectOneByWhere(rawWhereSql string, result *T, args ...any) (int64, error)

	// SelectByWhere 通过原始Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
	SelectByWhere(rawWhereSql, orderBy string, result *[]*T, args ...any) (int64, error)

	// SelectOneByGorm 通过原始Gorm查询单条数据 构建Gorm查询条件
	SelectOneByGorm(result *T, rawDb func(*gorm.DB)) (int64, error)

	// SelectByGorm 通过原始Gorm查询数据
	SelectByGorm(result *[]*T, rawDb func(*gorm.DB)) (int64, error)

	// CountByCond 通过条件查询数据总数 查询条件零值字段将被自动忽略
	CountByCond(condition *T) (int64, error)

	// CountByMap 通过指定字段与值查询数据总数 解决零值条件问题
	CountByMap(condition map[string]any) (int64, error)

	// CountByWhere 通过原始SQL查询数据总数
	CountByWhere(rawWhereSql string, args ...any) (int64, error)

	// CountByGorm 通过原始Gorm查询数据总数
	CountByGorm(rawDb func(*gorm.DB)) (int64, error)

	// SelectPageByCond 通过条件分页查询 零值字段将被自动忽略
	// specifyColumns 指定只需要查询的数据库字段 pageNumber 页码 1开始
	SelectPageByCond(condition *T, orderBy string, pageNumber, pageSize int, result *[]*T, specifyColumns ...string) (total int64, err error)

	// SelectPageByMap 通过指定字段与值查询数据分页查询 解决零值条件问题
	// specifyColumns 指定只需要查询的数据库字段 pageNumber 页码 1开始
	SelectPageByMap(condition map[string]any, orderBy string, pageNumber, pageSize int, result *[]*T, specifyColumns ...string) (total int64, err error)

	// SelectPageByWhere 通过原始SQL分页查询 rawWhereSql 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
	// specifyColumns 指定只需要查询的数据库字段 pageNumber 页码 1开始
	SelectPageByWhere(rawWhereSql, orderBy string, pageNumber, pageSize int, result *[]*T, args []any, specifyColumns ...string) (total int64, err error)

	// SelectPageByGorm 通过原始Gorm分页查询
	SelectPageByGorm(countRawDb func(*gorm.DB), pageRawDb func(*gorm.DB), result *[]*T) (total int64, err error)

	// Insert 保存数据 零值也将参与保存
	//	exclude 手动指定需要排除的字段名称 数据库字段/结构体字段名称
	Insert(entity *T, excludeColumns ...string) (int64, error)

	// InsertBatch 批量新增 零值也将参与保存
	//	exclude 手动指定需要排除的字段名称 数据库字段/结构体字段
	InsertBatch(entities *[]*T, excludeColumns ...string) (int64, error)

	// InsertWithoutZeroField 保存数据 零值将不会参与保存
	InsertWithoutZeroField(entity *T) (int64, error)

	// InsertUseMap 通过Map类型保存数据
	InsertUseMap(entity map[string]any) (int64, error)

	// InsertOrUpdateByPrimaryKey 保存/更新数据 零值也将参与保存
	// exclude 手动指定需要排除的字段名称 数据库字段/结构体字段 (如果触发的是update 创建时间可能会被错误的修改，可以通过excludeColumns来指定排除创建时间字段)
	// 仅根据主键冲突默认支持update 更多操作需要参阅 https://gorm.io/zh_CN/docs/create.html#upsert
	InsertOrUpdateByPrimaryKey(entity *T, excludeColumns ...string) (int64, error)

	// UpdateById 通过ID更新含零值字段
	// updateColumns 手动指定需要更新的列
	UpdateById(updated *T, updateColumns ...string) (int64, error)

	// UpdateByIdWithoutZeroField 通过ID更新非零值字段
	// allowZeroFiledColumns 额外指定需要更新零值字段
	UpdateByIdWithoutZeroField(updated *T, allowZeroFiledColumns ...string) (int64, error)

	// UpdateByIdUseMap 通过ID更新所有map中指定的列和值
	UpdateByIdUseMap(updated map[string]any, id any) (int64, error)

	// UpdateByCond 通过条件更新 条件：零值将自动忽略，更新：零值字段将被自动忽略
	// updateColumns 需要指定更新的数据库字段 更新指定字段(支持零值字段)
	UpdateByCond(updated, condition *T, updateColumns ...string) (int64, error)

	// UpdateByCondWithZeroField 通过条件更新，并指定可以更新的零值字段
	UpdateByCondWithZeroField(updated, condition *T, allowZeroFiledColumns []string) (int64, error)

	// UpdateByMap 通过Map类型条件更新
	UpdateByMap(updated, condition map[string]any) (int64, error)

	// UpdateByWhere 通过原始SQL查询条件，更新非零实体字段 Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
	UpdateByWhere(updated *T, rawWhereSql string, args ...any) (int64, error)

	// DeleteById 通过ID删除相关数据
	DeleteById(id ...any) (int64, error)

	// DeleteByCond 通过条件删除 零值字段将被自动忽略
	DeleteByCond(condition *T) (int64, error)

	// DeleteByMap 通过Map类型条件删除
	DeleteByMap(condition map[string]any) (int64, error)

	// DeleteByWhere 通过原始SQL删除相关数据 Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
	DeleteByWhere(rawWhereSql string, args ...any) (int64, error)


}
