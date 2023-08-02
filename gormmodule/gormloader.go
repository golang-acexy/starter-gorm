package gormmodule

import (
	"github.com/golang-acexy/starter-parent/parentmodule/declaration"
)

type GormModule struct {
	GormModuleConfig *declaration.ModuleConfig
	GormInterceptor  *func(instance interface{})
}

func (g *GormModule) ModuleConfig() *declaration.ModuleConfig {
	if g.GormModuleConfig != nil {
		return g.GormModuleConfig
	}
	return &declaration.ModuleConfig{
		ModuleName:               "Gorm",
		UnregisterPriority:       1,
		UnregisterAllowAsync:     true,
		UnregisterMaxWaitSeconds: 20,
	}
}

func (g *GormModule) Interceptor() *func(instance interface{}) {
	if g.GormInterceptor != nil {
		return g.GormInterceptor
	}
	return nil
}

func (g *GormModule) Register(interceptor *func(instance interface{})) error {
	return nil
}

func (g *GormModule) Unregister(maxWaitSeconds uint) (gracefully bool, err error) {
	return false, nil
}
