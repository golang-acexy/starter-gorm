package gormstarter

import (
	"errors"
	"github.com/acexy/golang-toolkit/logger"
	"gorm.io/gorm"
)

const notInvokeRowFlag = -9999

// DBExecutor 定义基础数据库执行函数类型
// return 	int64: 受影响行数
//
//	error: 任何异常将中断执行链并回滚整个事务
type DBExecutor func(tx *gorm.DB) (int64, error)

type Transaction struct {
	tx              *gorm.DB
	executors       []DBExecutor
	allowZeroAffRow bool // 是否允许SQL执行结果未影响任何记录 default false (zero will rollback)

	isPrepare        bool
	execAffRow       int64
	transactionError error
	allowCommit      bool
}

// NewTransactionPrepare 创建一个新的事务预执行链
// 该事务的执行方式将在执行Execute统一执行所有预设的DML过程
// 任何在该事务链中的对数据库操作均处于预备执行阶段，仅在调用Execute后才全部执行
// 整个事务链将在任何一个DML发生异常（或执行的结果不满足要求时）被标记为回滚，且在该事务链后面的DML操作将自动忽略执行
// allowZeroAffRow 是否允许执行影响行数为0 如果为false 则遇到执行行数为0时回滚整个事务
func NewTransactionPrepare(allowZeroAffRow ...bool) *Transaction {
	tx := &Transaction{
		tx:          gormDB.Begin(),
		executors:   make([]DBExecutor, 0),
		isPrepare:   true,
		allowCommit: true,
	}
	if len(allowZeroAffRow) > 0 {
		tx.allowZeroAffRow = allowZeroAffRow[0]
	}
	return tx
}

// NewTransaction 创建一个事务执行链
// 整个事务链将在任何一个DML发生异常（或执行的结果不满足要求时）被标记为回滚，且在该事务链后面的DML操作将自动忽略执行
// allowZeroAffRow 是否允许执行影响行数为0 如果为false 则遇到执行行数为0时回滚整个事务
func NewTransaction(allowZeroAffRow ...bool) *Transaction {
	tx := &Transaction{
		tx:          gormDB.Begin(),
		execAffRow:  notInvokeRowFlag,
		allowCommit: true,
	}
	if len(allowZeroAffRow) > 0 {
		tx.allowZeroAffRow = allowZeroAffRow[0]
	}
	return tx
}

func (t *Transaction) exec(f func(tx *gorm.DB) (int64, error)) {
	if t.isPrepare { // 如果预处理模式，则保存到执行链中
		t.executors = append(t.executors, f)
		return
	}

	if t.checkTransaction() {
		t.execAffRow, t.transactionError = f(t.tx)
		if t.transactionError != nil {
			t.allowCommit = false
			return
		}
		if t.execAffRow != notInvokeRowFlag {
			if !t.allowZeroAffRow && t.execAffRow <= 0 {
				t.allowCommit = false
			}
		}
	}
}

func (t *Transaction) checkTransaction() bool {
	if t.transactionError != nil {
		t.allowCommit = false
		return false
	}
	if t.execAffRow != notInvokeRowFlag {
		if !t.allowZeroAffRow && t.execAffRow <= 0 {
			t.allowCommit = false
			return false
		}
	}
	if !t.allowCommit {
		return false
	}
	return true
}

// SelectById 通过Id查询数据
// model对象指针，用于指定数据表&接收返回结果
func (t *Transaction) SelectById(model any, id any) *Transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		return checkResult(tx.Find(model, id), true)
	})
	return t
}

// SelectByCond 通过Id查询数据
// condition model非零参数条件
// result 返回数据指针
func (t *Transaction) SelectByCond(condition any, result any) *Transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		return checkResult(tx.Model(condition).Where(condition).Scan(result), true)
	})
	return t
}

// SelectByCondMap 通过Id查询数据
// model	实体
// condition 指定字段与值查询数据
// result 返回数据指针
func (t *Transaction) SelectByCondMap(model any, condition map[string]any, result any) *Transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		return checkResult(tx.Model(model).Where(condition).Scan(result), true)
	})
	return t
}

// Save 预设的保存功能 传入变量指针
func (t *Transaction) Save(entity any) *Transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		result := tx.Save(entity)
		if result.Error != nil {
			return 0, result.Error
		}
		return result.RowsAffected, nil
	})
	return t
}

// UpdateById 预设的更新功能 通过Id更新
// request	condition 作为更新时条件 需要指定主键
//
//	updated 作为需要更新数据 仅更新updated非零值字段数据 零值会被自动忽略 可传入map[string]interface{}代替struct
func (t *Transaction) UpdateById(condition, updated any) *Transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		result := tx.Model(condition).Updates(updated)
		if result.Error != nil {
			return 0, result.Error
		}
		return result.RowsAffected, nil
	})
	return t
}

// UpdateByCond 通过条件更新
// updated 作为需要更新数据 仅更新updated非零值字段数据 零值会被自动忽略 可传入map[string]interface{}代替struct
// where sql部分条件 也可以是一个model非零参数条件
func (t *Transaction) UpdateByCond(updated any, where interface{}, args ...interface{}) *Transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		result := tx.Model(updated).Where(where, args...).Updates(updated)
		if result.Error != nil {
			return 0, result.Error
		}
		return result.RowsAffected, nil
	})
	return t
}

// UpdateByCondMap 通过条件更新
// request	updated 作为需要更新数据 传入map[string]interface{}代替struct防止忽略零值
//
//	where	sql部分条件 也可以是一个model非零参数条件
func (t *Transaction) UpdateByCondMap(model any, updated map[string]interface{}, where interface{}, args ...interface{}) *Transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		result := tx.Model(model).Where(where, args...).Updates(updated)
		if result.Error != nil {
			return 0, result.Error
		}
		return result.RowsAffected, nil
	})
	return t
}

// DeleteById 预设的删除功能 根据id或则ids删除
// request 	传入一个model，则其主键必须指定 调用通过主键删除
//
//	传入model切片(每个model需要指定主键) 批量通过主键删除
func (t *Transaction) DeleteById(condition any) *Transaction {
	t.exec(func(tx *gorm.DB) (int64, error) {
		result := tx.Delete(condition)
		if result.Error != nil {
			return 0, result.Error
		}
		return result.RowsAffected, nil
	})
	return t
}

// DeleteByCond 预设删除功能 根据条件删除
func (t *Transaction) DeleteByCond(model any, where interface{}, args ...interface{}) *Transaction {
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
func (t *Transaction) Customize(executors ...DBExecutor) *Transaction {
	if len(executors) > 0 {
		for _, v := range executors {
			t.exec(v)
		}
	}
	return t
}

// Rollback 回滚事务
func (t *Transaction) Rollback() {
	t.exec(func(tx *gorm.DB) (int64, error) {
		t.allowCommit = false
		return notInvokeRowFlag, tx.Error
	})
}

// Execute 执行所有装载的SQL链
// bool true: committed false: rolled back
func (t *Transaction) Execute() (bool, error) {
	if !t.isPrepare {
		if t.allowCommit {
			rs := t.tx.Commit()
			if rs.Error != nil {
				t.tx.Rollback()
				logger.Logrus().Errorln("transaction commit error, exec rollback")
				return false, rs.Error
			}
			logger.Logrus().Traceln("transaction commit")
			return true, t.transactionError
		} else {
			rs := t.tx.Rollback()
			logger.Logrus().Warn("transaction rollback")
			return false, rs.Error
		}
	} else {
		if len(t.executors) == 0 {
			return false, errors.New("no executors")
		}
		for _, f := range t.executors {
			if t.checkTransaction() {
				t.execAffRow, t.transactionError = f(t.tx)
				if t.transactionError != nil {
					t.tx.Rollback()
					logger.Logrus().Warn("transaction rollback")
					return false, t.transactionError
				}
				if t.execAffRow != notInvokeRowFlag {
					if !t.allowZeroAffRow && t.execAffRow <= 0 {
						t.tx.Rollback()
						return false, t.tx.Error
					}
				}
			} else {
				t.tx.Rollback()
				logger.Logrus().Warn("transaction rollback")
				return false, t.tx.Error
			}
		}
		rs := t.tx.Commit()
		return true, rs.Error
	}
}
