package svc

import (
	"forum/service/user/model"
	"forum/service/user/rpc/internal/config"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	UserModel model.UsersModel // 添加model层关联
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		// 添加UserModel字段关联
		UserModel: model.NewUsersModel(sqlx.NewMysql(c.Mysql.DataSource)),
	}
}
