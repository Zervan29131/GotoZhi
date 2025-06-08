package logic

import (
	"context"
	"google.golang.org/grpc/status"

	"forum/service/user/rpc/internal/svc"
	"forum/service/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserInfoLogic) UserInfo(in *user.UserInfoRequest) (*user.UserInfoResponse, error) {
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return &user.UserInfoResponse{}, status.Error(400, err.Error())
	}

	return &user.UserInfoResponse{
		Id:     u.Id,
		Name:   u.Name,
		Mobile: u.Gender,
		Gender: u.Gender,
	}, nil
}
