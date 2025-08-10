package handler

import (
	"github.com/zeromicro/go-zero/core/logx"
	xhttp "github.com/zeromicro/x/http"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"rose/internal/logic"
	"rose/internal/svc"
	"rose/internal/types"
)

func createTodoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateTodoReq
		if err := httpx.Parse(r, &req); err != nil {
			logx.Error(err)
			xhttp.JsonBaseResponseCtx(r.Context(), w, types.GetError(types.ErrorInvalidParamsCode))
			return
		}

		l := logic.NewCreateTodoLogic(r.Context(), svcCtx)
		resp, err := l.CreateTodo(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
