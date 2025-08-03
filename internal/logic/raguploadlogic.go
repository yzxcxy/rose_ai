package logic

import (
	"context"

	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RagUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRagUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RagUploadLogic {
	return &RagUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RagUploadLogic) RagUpload(req *types.RagUploadRequest) (resp *types.RagUploadResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
