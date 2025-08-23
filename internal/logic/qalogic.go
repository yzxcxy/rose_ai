package logic

import (
	"context"
	"github.com/cloudwego/eino/schema"
	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QaLogic {
	return &QaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QaLogic) Qa(req *types.QaRequest) (resp *types.QaResponse, err error) {
	messages := []*schema.Message{
		schema.UserMessage(req.Input),
	}

	generate, err := l.svcCtx.Agent.QA(l.ctx, messages)
	if err != nil {
		return nil, err
	}
	resp = &types.QaResponse{
		SessionID: req.SessionID,
		Output:    generate.Content,
		Memory:    req.Memory,
	}
	return
}
