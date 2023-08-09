package gormmodule

import (
	"errors"
	"github.com/acexy/golang-toolkit/log"
	"gorm.io/gorm"
)

// DBExecutor 定义基础数据库执行函数类型
// return 	bool: 是否需要回滚
//			error: 任何异常将中断执行链并回滚整个事务
type DBExecutor func(tx *gorm.DB) (bool, error)

type transaction struct {
	tx              *gorm.DB
	executors       []DBExecutor
	allowZeroAffRow bool // 是否允许SQL执行结果未影响任何记录 default false (zero will rollback)
}

// NewTransaction 创建一个新的事务执行链
func NewTransaction(allowZeroAffRow ...bool) *transaction {
	tx := &transaction{
		tx:        db.Begin(),
		executors: make([]DBExecutor, 0),
	}
	if len(allowZeroAffRow) > 0 {
		tx.allowZeroAffRow = allowZeroAffRow[0]
	}
	return tx
}

// Execute 执行所有装载的SQL链
func (t *transaction) Execute() error {
	if len(t.executors) == 0 {
		return errors.New("no executors")
	}
	for _, f := range t.executors {
		ok, err := f(t.tx)
		if err != nil {
			log.Logrus().WithError(err).Error("rollback by error")
			t.tx.Rollback()
			return err
		}
		if !t.allowZeroAffRow && !ok {
			t.tx.Rollback()
			err = errors.New("the execution result does not meet expectations")
			log.Logrus().WithError(err).Error("rollback by error")
			return err
		}
	}
	t.tx.Commit()
	return nil
}

// Save 预设的保存功能 传入变量指针
func (t *transaction) Save(entity any) *transaction {
	t.executors = append(t.executors, func(tx *gorm.DB) (bool, error) {
		result := tx.Save(entity)
		if result.Error != nil {
			return false, result.Error
		}
		return result.RowsAffected > 0, nil
	})
	return t
}

// ModifyById 预设的更新功能 通过Id更新
// request	condition 作为更新时条件 需要指定主键
//			updated 作为需要更新数据 仅更新updated非零值字段数据 零值会被自动忽略 可传入map[string]interface{}代替struct
func (t *transaction) ModifyById(condition, updated any) *transaction {
	t.executors = append(t.executors, func(tx *gorm.DB) (bool, error) {
		result := tx.Model(condition).Updates(updated)
		if result.Error != nil {
			return false, result.Error
		}
		return result.RowsAffected > 0, nil
	})
	return t
}

// ModifyByCondition 通过条件更新
// request	updated 作为需要更新数据 仅更新updated非零值字段数据 零值会被自动忽略 可传入map[string]interface{}代替struct
//			where	sql部分条件
func (t *transaction) ModifyByCondition(updated any, where ...interface{}) *transaction {
	t.executors = append(t.executors, func(tx *gorm.DB) (bool, error) {
		//exec := tx.Table(updated.(schema.Tabler).TableName())
		exec := tx.Model(updated)
		if len(where) > 1 {
			exec.Where(where[0], where[1:])
		} else {
			exec.Where(where)
		}
		result := exec.Updates(updated)
		if result.Error != nil {
			return false, result.Error
		}
		return result.RowsAffected > 0, nil
	})
	return t
}

// ModifyByConditionMap 通过条件更新
// request	updated 作为需要更新数据 传入map[string]interface{}代替struct防止忽略零值
//			where	sql部分条件
func (t *transaction) ModifyByConditionMap(model any, updated map[string]interface{}, where ...interface{}) *transaction {
	t.executors = append(t.executors, func(tx *gorm.DB) (bool, error) {
		//exec := tx.Table(updated.(schema.Tabler).TableName())
		exec := tx.Model(model)
		if len(where) > 1 {
			exec.Where(where[0], where[1:])
		} else {
			exec.Where(where)
		}
		result := exec.Updates(updated)
		if result.Error != nil {
			return false, result.Error
		}
		return result.RowsAffected > 0, nil
	})
	return t
}

// RemoveById 预设的删除功能
// request 	传入一个model对象指针，则其主键必须指定 调用通过主键删除
//			传入model切片(每个model需要指定主键) 批量通过主键删除
func (t *transaction) RemoveById(condition any) *transaction {
	t.executors = append(t.executors, func(tx *gorm.DB) (bool, error) {
		result := tx.Delete(condition)
		if result.Error != nil {
			return false, result.Error
		}
		return result.RowsAffected > 0, nil
	})
	return t
}

// Customize 执行自定义的SQL逻辑
func (t *transaction) Customize(executors ...DBExecutor) *transaction {
	t.executors = append(t.executors, executors...)
	return t
}
