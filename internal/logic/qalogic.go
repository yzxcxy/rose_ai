package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return
}
