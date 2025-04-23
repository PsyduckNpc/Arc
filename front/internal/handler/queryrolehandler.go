package handler

import (
	"net/http"

	"Arc/front/internal/logic"
	"Arc/front/internal/svc"
	"Arc/front/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func queryRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Role
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewQueryRoleLogic(r.Context(), svcCtx)
		resp, err := l.QueryRole(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
