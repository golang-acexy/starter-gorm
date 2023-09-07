package gormmodule

import (
	"errors"
	"github.com/acexy/golang-toolkit/log"
	"gorm.io/gorm"
)

// DBExecutor 定义基础数据库执行函数类型
// return 	int64: 受影响行数
//			error: 任何异常将中断执行链并回滚整个事务
type DBExecutor func(tx *gorm.DB) (int64, error)

type transaction struct {
	tx              *gorm.DB
	executors       []DBExecutor
	allowZeroAffRow bool // 是否允许SQL执行结果未影响任何记录 default false (zero will rollback)

	execNow   bool
	execRow   int64
	execError error
	canCommit bool
}

// NewTransactionChain 创建一个新的事务执行链
// 该事务的执行方式将在执行Execute统一执行所有预设的SQL过程，任何在操作事务链中交互的参数在立即获取时并不能获取操作后结果，仅在调用Execute后事务链才异常执行
// allowZeroAffRow 是否允许执行影响行数为0 如果为false 则遇到执行行数为0是回滚整个事务
func NewTransactionChain(allowZeroAffRow ...bool) *transaction {
	tx := &transaction{
		tx:        db.Begin(),
		executors: make([]DBExecutor, 0),
	}
	if len(allowZeroAffRow) > 0 {
		tx.allowZeroAffRow = allowZeroAffRow[0]
	}
	return tx
}

func NewTransaction(allowZeroAffRow ...bool) *transaction {
	tx := &transaction{
		tx:      db.Begin(),
		execNow: true,
		execRow: -999,
	}
	if len(allowZeroAffRow) > 0 {
		tx.allowZeroAffRow = allowZeroAffRow[0]
	}
	return tx
}

func (t *transaction) exec(f func(tx *gorm.DB) (int64, error)) {
	if !t.execNow { // 如果非立即执行的模式，则保存到执行链中
		t.executors = append(t.executors, f)
		return
	}
	if t.execError != nil {
		t.canCommit = false
		return
	}
	if t.execRow != -999 {
		if !t.allowZeroAffRow && t.execRow <= 0 {
			t.canCommit = false
			return
		}
	}
	t.canCommit = true
	t.execRow, t.execError = f(t.tx)
}

// QueryById 通过Id查询数据
// model对象指针，用于指定数据表&接收返回结果
func (t *transaction) QueryById(model any, id any) *transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		return checkResult(tx.Find(model, id), true)
	})
	return t
}

// QueryByCondition 通过Id查询数据
// condition model非零参数条件
// result 返回数据指针
func (t *transaction) QueryByCondition(condition any, result any) *transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		return checkResult(tx.Model(condition).Where(condition).Scan(result), true)
	})
	return t
}

// QueryByConditionMap 通过Id查询数据
// model	实体
// condition 指定字段与值查询数据
// result 返回数据指针
func (t *transaction) QueryByConditionMap(model any, condition map[string]any, result any) *transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		return checkResult(tx.Model(model).Where(condition).Scan(result), true)
	})
	return t
}

// Save 预设的保存功能 传入变量指针
func (t *transaction) Save(entity any) *transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		result := tx.Save(entity)
		if result.Error != nil {
			return 0, result.Error
		}
		return result.RowsAffected, nil
	})
	return t
}

// ModifyById 预设的更新功能 通过Id更新
// request	condition 作为更新时条件 需要指定主键
//			updated 作为需要更新数据 仅更新updated非零值字段数据 零值会被自动忽略 可传入map[string]interface{}代替struct
func (t *transaction) ModifyById(condition, updated any) *transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		result := tx.Model(condition).Updates(updated)
		if result.Error != nil {
			return 0, result.Error
		}
		return result.RowsAffected, nil
	})
	return t
}

// ModifyByCondition 通过条件更新
// updated 作为需要更新数据 仅更新updated非零值字段数据 零值会被自动忽略 可传入map[string]interface{}代替struct
// where sql部分条件 也可以是一个model非零参数条件
func (t *transaction) ModifyByCondition(updated any, where interface{}, args ...interface{}) *transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		result := tx.Model(updated).Where(where, args...).Updates(updated)
		if result.Error != nil {
			return 0, result.Error
		}
		return result.RowsAffected, nil
	})
	return t
}

// ModifyByConditionMap 通过条件更新
// request	updated 作为需要更新数据 传入map[string]interface{}代替struct防止忽略零值
//			where	sql部分条件 也可以是一个model非零参数条件
func (t *transaction) ModifyByConditionMap(model any, updated map[string]interface{}, where interface{}, args ...interface{}) *transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		result := tx.Model(model).Where(where, args...).Updates(updated)
		if result.Error != nil {
			return 0, result.Error
		}
		return result.RowsAffected, nil
	})
	return t
}

// RemoveById 预设的删除功能 根据id或则ids删除
// request 	传入一个model，则其主键必须指定 调用通过主键删除
//			传入model切片(每个model需要指定主键) 批量通过主键删除
func (t *transaction) RemoveById(condition any) *transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		result := tx.Delete(condition)
		if result.Error != nil {
			return 0, result.Error
		}
		return result.RowsAffected, nil
	})
	return t
}

// RemoveByCondition 预设删除功能 根据条件删除
func (t *transaction) RemoveByCondition(model any, where interface{}, args ...interface{}) *transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		result := tx.Where(where, args...).Delete(model)
		if result.Error != nil {
			return 0, result.Error
		}
		return result.RowsAffected, nil
	})
	return t
}

// Customize 执行自定义的SQL逻辑
func (t *transaction) Customize(executors ...DBExecutor) *transaction {
	if len(executors) > 0 {
		for _, v := range executors {
			t.exec(v)
		}
	}
	return t
}

// Execute 执行所有装载的SQL链
func (t *transaction) Execute() error {
	if t.execNow {
		if t.canCommit {
			rs := t.tx.Commit()
			if rs.Error != nil {
				t.tx.Rollback()
				log.Logrus().Errorln("transaction commit error, exec rollback")
				return rs.Error
			}
			log.Logrus().Traceln("transaction commit")
		} else {
			if t.execError != nil {
				log.Logrus().WithError(t.execError).Error("rollback by error")
				t.tx.Rollback()
				return t.execError
			}
			if !t.allowZeroAffRow && t.execRow == 0 {
				t.tx.Rollback()
				err := errors.New("the execution result does not meet expectations")
				log.Logrus().WithError(err).Error("rollback by error")
				return err
			}
		}
	} else {
		if len(t.executors) == 0 {
			return errors.New("no executors")
		}
		for _, f := range t.executors {
			rows, err := f(t.tx)
			if err != nil {
				log.Logrus().WithError(err).Error("rollback by error")
				t.tx.Rollback()
				return err
			}
			if !t.allowZeroAffRow && rows == 0 {
				t.tx.Rollback()
				err = errors.New("the execution result does not meet expectations")
				log.Logrus().WithError(err).Error("rollback by error")
				return err
			}
		}
		rs := t.tx.Commit()
		if rs.Error != nil {
			t.tx.Rollback()
			log.Logrus().Errorln("transaction commit error, exec rollback")
			return rs.Error
		}
		log.Logrus().Traceln("transaction commit")
	}
	return nil
}
