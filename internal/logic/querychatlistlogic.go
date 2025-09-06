package logic

import (
	"context"
	"rose/model"
	"strconv"

	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryChatListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryChatListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryChatListLogic {
	return &QueryChatListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryChatListLogic) QueryChatList(req *types.ChatQueryUserListRequest) (resp *types.ChatQueryUserListResponse, err error) {
	chatModel := model.NewChatSessionModel(l.svcCtx.Mysql)
	chats, err := chatModel.FindChatSessionByUserId(l.ctx, req.UserId)
	if err != nil {
		return nil, types.GetError(types.ErrorInternalServer)
	}
	var chatSession []types.ChatSession
	for idx, _ := range chats {
		var session types.ChatSession
		session.SessionId = strconv.FormatInt(chats[idx].SessionId, 10)
		session.Title = chats[idx].Title
		chatSession = append(chatSession, session)
	}
	return &types.ChatQueryUserListResponse{
		UserId:   req.UserId,
		ChatList: chatSession,
	}, nil
}
