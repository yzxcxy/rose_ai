package logic

import (
	"context"
	"encoding/json"
	"github.com/cloudwego/eino/schema"
	"rose/internal/utils"
	"rose/model"
	"strconv"

	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryChatLogic {
	return &QueryChatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryChatLogic) QueryChat(req *types.ChatQueryRequest) (resp *types.ChatQueryReponse, err error) {
	// 查询数据库
	chatModel := model.NewChatSessionModel(l.svcCtx.Mysql)
	uid, _, _ := utils.GetUserIdAndUserNameFromContext(l.ctx)
	uidStr := strconv.FormatInt(uid, 10)
	sessionId, err := strconv.Atoi(req.SessionId)
	if err != nil {
		return nil, err
	}

	session, err := chatModel.FindOneBySessionId(l.ctx, int64(sessionId))
	if err != nil || session.IsDeleted == 1 {
		return nil, types.GetError(types.ErrorChatNotFound)
	}

	// 查询redis
	historyKey := uidStr + "::" + req.SessionId
	n, err := l.svcCtx.Redis.Exists(l.ctx, historyKey).Result()
	if err != nil || n == 0 {
		return &types.ChatQueryReponse{
			SessionId: req.SessionId,
			Title:     session.Title,
		}, nil
	}
	length, err := l.svcCtx.Redis.LLen(l.ctx, historyKey).Result()
	if err != nil || length == 0 {
		return &types.ChatQueryReponse{
			SessionId: req.SessionId,
			Title:     session.Title,
		}, nil
	}

	vals, err := l.svcCtx.Redis.LRange(l.ctx, historyKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	// 封装数据
	var history []types.Message
	for idx, _ := range vals {
		var message schema.Message
		var val types.Message
		err = json.Unmarshal([]byte(vals[idx]), &message)
		if err != nil {
			return nil, err
		}
		val.Content = message.Content
		val.Role = string(message.Role)
		history = append(history, val)
	}
	return &types.ChatQueryReponse{
		SessionId: req.SessionId,
		Title:     session.Title,
		History:   history,
	}, nil
}
