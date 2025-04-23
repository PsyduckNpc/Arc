package handler

import (
	"net/http"

	"Arc/front/internal/logic"
	"Arc/front/internal/svc"
	"Arc/front/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func queryCenterDataApiHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CenterDataApi
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewQueryCenterDataApiLogic(r.Context(), svcCtx)
		resp, err := l.QueryCenterDataApi(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
