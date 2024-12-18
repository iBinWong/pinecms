package helper

import (
	"github.com/xiusin/pine"
	"github.com/xiusin/pine/contracts"
	"github.com/xiusin/pine/di"
	"github.com/xiusin/pinecms/src/application/controllers"
)

// Inject 注入依赖
func Inject(key any, v any, single ...bool) {
	if len(single) == 0 {
		single = append(single, true)
	}
	if vi, ok := v.(di.BuildHandler); ok {
		di.Set(key, vi, single[0])
	} else {
		di.Set(key, func(builder di.AbstractBuilder) (i any, e error) {
			return v, nil
		}, single[0])
	}
}

// AbstractCache 获取缓存服务
func AbstractCache() contracts.Cache {
	return pine.Make(controllers.ServiceICache).(contracts.Cache)
}

// Cache 获取缓存服务
func Cache() contracts.Cache {
	return AbstractCache()
}

// App 获取应用实例
func App() *pine.Application {
	return pine.Make(controllers.ServiceApplication).(*pine.Application)
}
