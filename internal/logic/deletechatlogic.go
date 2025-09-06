package logic

import (
	"context"
	"rose/model"
	"strconv"

	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChatLogic {
	return &DeleteChatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteChatLogic) DeleteChat(req *types.ChatDeleteRequest) (resp *types.ChatDeleteResponse, err error) {
	chatModel := model.NewChatSessionModel(l.svcCtx.Mysql)
	sessionId, err := strconv.Atoi(req.SessionId)
	if err != nil {
		return nil, err
	}
	err = chatModel.UpdateIgnoreNull(l.ctx, &model.ChatSession{
		SessionId: int64(sessionId),
		IsDeleted: 1,
	})
	if err != nil {
		return nil, err
	}

	return &types.ChatDeleteResponse{}, nil
}
