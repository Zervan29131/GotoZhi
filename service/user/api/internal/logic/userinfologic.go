package logic

import (
	"context"
	"encoding/json"
	"forum/service/user/rpc/userclient"

	"forum/service/user/api/internal/svc"
	"forum/service/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserInfoLogic) UserInfo() (resp *types.UserInfoResponse, err error) {
	// 注意这里ctx的uid是框架自动帮我们解析出来
	// uid是签发授权时存的user id,这里取出来的是interface{}类型
	uid, err := l.ctx.Value("uid").(json.Number).Int64()
	if err != nil {
		return nil, err
	}
	user, err := l.svcCtx.UserRpc.UserInfo(l.ctx, &userclient.UserInfoRequest{
		Id: uid,
	})
	if err != nil {
		return nil, err
	}

	return &types.UserInfoResponse{
		ID:     user.Id,
		Name:   user.Name,
		Mobile: user.Mobile,
		Gender: user.Gender,
	}, nil
}
