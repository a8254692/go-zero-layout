package userbehavior

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"minicode.com/sirius/go-back-server/service/userbehavior/api/internal/logic/userbehavior"
	"minicode.com/sirius/go-back-server/service/userbehavior/api/internal/svc"
	"minicode.com/sirius/go-back-server/service/userbehavior/api/internal/types"

	"minicode.com/sirius/go-back-server/utils/response"
)

func FocusFollowListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FocusFollowListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := userbehavior.NewFocusFollowListLogic(r.Context(), svcCtx)
		resp, err := l.FocusFollowList(req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			response.SuccessResponse(w, resp)
		}
	}
}
