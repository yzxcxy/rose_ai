package handler

import (
	"context"
	xhttp "github.com/zeromicro/x/http"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"rose/internal/logic"
	"rose/internal/svc"
	"rose/internal/types"
)

func qaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QaRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, types.GetError(types.ErrorInvalidParamsCode))
			return
		}

		ctx := context.WithValue(r.Context(), "Authorization", r.Header.Get("Authorization"))
		l := logic.NewQaLogic(ctx, svcCtx)
		resp, err := l.Qa(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
