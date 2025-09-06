package logic

import (
	"context"
	"rose/model"
	"strconv"

	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatLogic {
	return &UpdateChatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateChatLogic) UpdateChat(req *types.ChatUpdateRequest) (resp *types.ChatUpdateResponse, err error) {
	chatModel := model.NewChatSessionModel(l.svcCtx.Mysql)
	sessionId, err := strconv.Atoi(req.SessionId)
	if err != nil {
		return nil, err
	}
	err = chatModel.UpdateIgnoreNull(l.ctx, &model.ChatSession{
		SessionId: int64(sessionId),
		Title:     req.Title,
	})
	if err != nil {
		return nil, err
	}

	return &types.ChatUpdateResponse{}, nil
}
