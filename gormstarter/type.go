package gormstarter

import (
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
