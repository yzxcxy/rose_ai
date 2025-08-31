package logic

import (
	"context"
	"os"
	"path/filepath"
	"rose/internal/utils"
	"strconv"

	"rose/internal/svc"
	"rose/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RagVectorizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRagVectorizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RagVectorizationLogic {
	return &RagVectorizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RagVectorizationLogic) RagVectorization(req *types.RagVectorizationRequest) (resp *types.RagVectorizationResponse, err error) {
	// 获得每个文件的绝对路径
	uid, _, err := utils.GetUserIdAndUserNameFromContext(l.ctx)
	if err != nil {
		logx.Error(err)
		return nil, types.GetError(types.ErrorUserNotFound)
	}
	wd, err := os.Getwd()
	if err != nil {
		logx.Error(err)
		return nil, types.GetError(types.ErrorInternalServer)
	}
	basePath := filepath.Join(wd, "uploads")
	var paths []string
	for _, file := range req.FileName {
		paths = append(paths, filepath.Join(basePath, strconv.Itoa(int(uid)), file))
	}
	// 检查每个文件是否存在
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			logx.Error(err)
			return nil, types.GetError(types.ErrorFileNotExist)
		}
	}
	// 进行向量化存储
	// 不能直接使用uid作为collectionName，因为其不能以数字作为开头
	ids, err := l.svcCtx.StoreFileToVector.Store(l.ctx, paths, "user"+strconv.Itoa(int(uid)))
	if err != nil {
		return nil, types.GetError(types.ErrorInternalServer)
	}
	return &types.RagVectorizationResponse{
		Ids: ids,
	}, nil
}
