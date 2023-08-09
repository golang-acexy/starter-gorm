package gormmodule

import (
	"errors"
	"gorm.io/gorm"
)

// DBExecutor 定义基础数据库执行函数类型
// return 	bool: 是否需要回滚
//			error: 任何异常将中断执行链并回滚整个事务
type DBExecutor func(tx *gorm.DB) (bool, error)

type transaction struct {
	tx        *gorm.DB
	executors []DBExecutor
}

// NewTransaction 创建一个新的事务执行链
func NewTransaction() *transaction {
	return &transaction{
		tx:        db.Begin(),
		executors: make([]DBExecutor, 0),
	}
}

// Execute 执行所有装载的SQL链
func (b *transaction) Execute() error {
	if len(b.executors) == 0 {
		return errors.New("no executors")
	}
	for _, f := range b.executors {
		ok, err := f(b.tx)
		if err != nil {
			b.tx.Rollback()
			return err
		}
		if !ok {
			b.tx.Rollback()
			return errors.New("the execution result does not meet expectations")
		}
	}
	b.tx.Commit()
	return nil
}

// Save 预设的保存功能 传入变量指针
func (b *transaction) Save(entity any) *transaction {
	b.executors = append(b.executors, func(tx *gorm.DB) (bool, error) {
		result := tx.Save(entity)
		if result.Error != nil {
			return false, result.Error
		}
		return result.RowsAffected > 0, nil
	})
	return b
}

// Modify 预设的更新功能 传入变量指针
// request	condition 作为更新时条件
//			updated 作为需要更新数据 仅更新updated非零值字段数据
func (b *transaction) Modify(condition, updated any) *transaction {
	b.executors = append(b.executors, func(tx *gorm.DB) (bool, error) {
		result := tx.Model(condition).Updates(updated)
		if result.Error != nil {
			return false, result.Error
		}
		return result.RowsAffected > 0, nil
	})
	return b
}

// DBExecute 执行自定义的SQL逻辑
func (b *transaction) DBExecute(executors ...DBExecutor) *transaction {
	b.executors = append(b.executors, executors...)
	return b
}
