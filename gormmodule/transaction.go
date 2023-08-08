package gormmodule

import (
	"gorm.io/gorm"
)

type AutoTx interface {
	Save(pip TxPip) error
}

type autoTx struct {
	tx *gorm.DB
}

type TxPip func(tx *gorm.DB) error

func NewAutoTx() *autoTx {
	return &autoTx{
		tx: db.Begin(),
	}
}

func (a *autoTx) MuxTx(txPip ...TxPip) error {
	for _, f := range txPip {
		err := f(a.tx)
		if a.tx.Error != nil {
			a.tx.Rollback()
			return err
		}
		if err != nil {
			a.tx.Rollback()
			return err
		}
	}
	a.tx.Commit()
	return nil
}
