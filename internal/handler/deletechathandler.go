package handler

import (
	xhttp "github.com/zeromicro/x/http"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"rose/internal/logic"
	"rose/internal/svc"
	"rose/internal/types"
)

func deleteChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatDeleteRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, types.GetError(types.ErrorInvalidParamsCode))
			return
		}

		l := logic.NewDeleteChatLogic(r.Context(), svcCtx)
		resp, err := l.DeleteChat(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
