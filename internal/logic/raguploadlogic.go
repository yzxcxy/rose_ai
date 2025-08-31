package logic

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

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

func (l *RagUploadLogic) RagUpload(data *multipart.File, fileName string) (resp *types.RagUploadResponse, err error) {
	// 获得用户ID
	userId, ok := l.ctx.Value("uid").(json.Number)
	if !ok {
		err = types.GetError(types.ErrorUserNotFound)
		return
	}
	uid := userId.String()

	allowedExt := map[string]bool{
		".txt":  true,
		".pdf":  true,
		".doc":  true,
		".html": true,
		".md":   true,
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	if !allowedExt[ext] {
		err = types.GetError(types.ErrorNotSupportFileType)
		return
	}

	// 创建用户文件夹
	userDir := filepath.Join("uploads", uid)
	if err = os.MkdirAll(userDir, os.ModePerm); err != nil {
		return
	}

	// 保存文件
	filePath := filepath.Join(userDir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer func(dst *os.File) {
		err := dst.Close()
		if err != nil {

		}
	}(dst)

	// data 是 *multipart.File
	if _, err = io.Copy(dst, *data); err != nil {
		return nil, types.GetError(types.ErrorUploadFailure)
	}

	resp = &types.RagUploadResponse{Message: "上传成功"}
	return
}
