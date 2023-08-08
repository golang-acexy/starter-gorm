package gormmodule

import (
	"gorm.io/gorm"
)

type autoTx struct {
	tx *gorm.DB
}

type DBExecutor func(tx *gorm.DB) error

func NewAutoTx() *autoTx {
	return &autoTx{
		tx: db.Begin(),
	}
}

func (a *autoTx) MuxTx(txPip ...DBExecutor) error {
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

type IBaseMapper interface {
	Save(db *gorm.DB) DBExecutor
}

type BaseMapper struct {
	autoTx *autoTx
}

func (b *BaseMapper) BaseSave() error {
	if b.autoTx == nil {
		b.autoTx = NewAutoTx()
	}
	rs := b.autoTx.tx.Save(b)
	if rs.Error != nil {
		return rs.Error
	}
	return nil
}
