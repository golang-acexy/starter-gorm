package test

import (
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
	"testing"
)

func init() {
	m := declaration.Module{ModuleLoaders: moduleLoaders}
	_ = m.Load()
}

func TestAutoTx(t *testing.T) {
	s := Student{}
	s.BaseSave()
}
