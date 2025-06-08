package svc

import (
	"forum/service/user/api/internal/config"
	"forum/service/user/rpc/userclient"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	UserRpc userclient.User // 关联user rpc,userclient是rpc中提供的客户端包
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		// 这里c.UserRpc取config中的配置此时已从配置中载入
		UserRpc: userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}
