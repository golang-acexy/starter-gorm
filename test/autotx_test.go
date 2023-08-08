package test

import (
	"github.com/golang-acexy/starter-gorm/gormmodule"
	"testing"
)

func TestAutoTx(t *testing.T) {
	_ = gormmodule.NewAutoTx()
}
