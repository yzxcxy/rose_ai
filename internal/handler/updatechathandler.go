package handler

import (
	xhttp "github.com/zeromicro/x/http"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"rose/internal/logic"
	"rose/internal/svc"
	"rose/internal/types"
)

func updateChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatUpdateRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, types.GetError(types.ErrorInvalidParamsCode))
			return
		}

		l := logic.NewUpdateChatLogic(r.Context(), svcCtx)
		resp, err := l.UpdateChat(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
