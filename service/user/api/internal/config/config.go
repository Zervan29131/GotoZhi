package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	// 这里直接定义一个字段就行 在config初始化时，会自动将etc中rpc配置加载到config中
	// zrpc.RpcClientConf 是一个结构体,在svc中初始化上下文时使用
	UserRpc zrpc.RpcClientConf

	// 授权配置信息
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
}
