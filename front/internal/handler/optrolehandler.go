package handler

import (
	"net/http"

	"Arc/front/internal/logic"
	"Arc/front/internal/svc"
	"Arc/front/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func optRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OptRoleAO
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewOptRoleLogic(r.Context(), svcCtx)
		err := l.OptRole(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
