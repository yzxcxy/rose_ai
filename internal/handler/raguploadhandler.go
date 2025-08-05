package handler

import (
	xhttp "github.com/zeromicro/x/http"
	"mime/multipart"
	"net/http"

	"rose/internal/logic"
	"rose/internal/svc"
	"rose/internal/types"
)

func ragUploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileData, handler, err := r.FormFile("file")
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, types.GetError(types.ErrorNotFile))
			return
		}
		// 关闭文件数据
		defer func(fileData multipart.File) {
			err := fileData.Close()
			if err != nil {

			}
		}(fileData)
		// 获取文件名
		filename := handler.Filename

		l := logic.NewRagUploadLogic(r.Context(), svcCtx)
		resp, err := l.RagUpload(&fileData, filename)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
