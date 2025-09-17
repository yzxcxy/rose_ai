package logic

import (
	"context"
	"rose/internal/utils"
	"rose/model"
	"rose/pkg/snowflake"
	"strconv"

	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatLogic {
	return &CreateChatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateChatLogic) CreateChat(req *types.ChatCreateRequest) (resp *types.ChatCreateResponse, err error) {
	chatModel := model.NewChatSessionModel(l.svcCtx.Mysql)
	uid, _, _ := utils.GetUserIdAndUserNameFromContext(l.ctx)
	sessionId := snowflake.GenID()
	_, err = chatModel.Insert(l.ctx, &model.ChatSession{
		UserId:    uid,
		Title:     req.Title,
		SessionId: sessionId,
	})
	if err != nil {
		return nil, err
	}
	summaryIndexKey := strconv.FormatInt(uid, 10) + ":" + "summary_index:" + strconv.FormatInt(sessionId, 10)
	// 将summaryIndex预先定义为-1
	l.svcCtx.Redis.Set(l.ctx, summaryIndexKey, -1, 0)
	return &types.ChatCreateResponse{
		SessionId: strconv.FormatInt(sessionId, 10),
	}, nil
}
