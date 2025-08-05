package logic

import (
	"context"
	"database/sql"
	"rose/model"
	"time"

	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"

	"rose/pkg/jwt"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	// 1. 验证用户名和密码
	userModel := model.NewUserModel(l.svcCtx.Mysql)
	user, err := userModel.FindOneByUsername(l.ctx, sql.NullString{String: req.Username, Valid: true})
	if err != nil {
		return nil, types.GetError(types.ErrorUserNotFound) // 用户不存在或查询错误
	}

	if l.svcCtx.Cipher.Verify(user.Password, req.Password) != nil {
		return nil, types.GetError(types.ErrorUserPasswordNotCorrect) // 密码不匹配
	}

	// 2. 生成 JWT Token
	secret := l.svcCtx.Config.Auth.AccessSecret
	expire := l.svcCtx.Config.Auth.AccessExpire

	token, _ := jwt.GetJwtToken(secret, time.Now().Unix(), expire, user.UserID, user.Username.String)

	return &types.LoginResponse{
		Token:  "Bearer " + token,
		UserId: user.Id,
	}, nil
}
