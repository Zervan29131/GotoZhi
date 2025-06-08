package logic

import (
	"context"
	"forum/service/user/rpc/userclient"

	"forum/service/user/api/internal/svc"
	"forum/service/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	// 注意这里是直接使用的userclient中的结构体组装的，并无单独的工厂方法
	res, err := l.svcCtx.UserRpc.Register(l.ctx, &userclient.RegisterRequest{
		Name:     req.Name,
		Mobile:   req.Mobile,
		Gender:   req.Gender,
		Password: req.Password,
	})

	if err != nil {
		return &types.RegisterResponse{}, err
	}

	return &types.RegisterResponse{
		ID:     res.Id,
		Name:   res.Name,
		Mobile: res.Mobile,
		Gender: res.Gender,
	}, nil
}
