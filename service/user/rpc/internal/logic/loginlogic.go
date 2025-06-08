package logic

import (
	"context"
	"forum/common/bcryptx"
	"google.golang.org/grpc/status"

	"forum/service/user/rpc/internal/svc"
	"forum/service/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user.LoginRequest) (*user.LoginResponse, error) {
	// 先查用户
	u, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.Mobile)
	if err != nil {
		return &user.LoginResponse{}, status.Error(400, err.Error())
	}

	// 判断密码对么
	if err = bcryptx.ValidatePassword(u.Password, in.Password); err != nil {
		return &user.LoginResponse{}, status.Error(400, "无效密码")
	}
	return &user.LoginResponse{
		Id:     u.Id,
		Name:   u.Name,
		Mobile: u.Mobile,
		Gender: u.Gender,
	}, nil
}
