package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"rose/model"
	"rose/pkg/snowflake"

	"rose/internal/svc"
	"rose/internal/types"

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
	// 1. 检查用户是否存在
	userModel := model.NewUserModel(l.svcCtx.Mysql)
	_, err = userModel.FindOneByUsername(l.ctx, sql.NullString{String: req.Username, Valid: true})
	// 2. 如果用户不存在，创建新用户
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	} else {
		// 用户已存在
		return nil, errors.New(fmt.Sprintf("user %s already exists", req.Username))
	}
	// 3. 密码加密处理
	cipher := l.svcCtx.Cipher
	passwordCrypted, err := cipher.Encrypt(req.Password)
	if err != nil {
		return nil, err
	}

	// 4. 创建用户数据
	userId := snowflake.GenID()
	user := &model.User{
		UserID:   userId,
		Username: sql.NullString{String: req.Username, Valid: true},
		Password: string(passwordCrypted),
	}

	// 5. 保存用户信息到数据库
	_, err = userModel.Insert(l.ctx, user)
	if err != nil {
		return nil, err
	}

	// 6. 返回注册成功的响应
	return &types.RegisterResponse{
		UserId:   userId,
		UserName: req.Username,
	}, nil
}
