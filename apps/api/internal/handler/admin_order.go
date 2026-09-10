package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/yourname/stationery-shop/apps/api/internal/logic/admin"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/response"
	"github.com/yourname/stationery-shop/apps/api/internal/svc"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

func AdminOrderList(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		data, err := svcCtx.AdminOrder.List(
			r.Context(),
			q.Get("status"),
			q.Get("keyword"),
			parseInt(r, "page", 1),
			parseInt(r, "page_size", 20),
		)
		if err != nil {
			response.ServerError(w, "query failed")
			return
		}
		response.OK(w, data)
	}
}

func AdminOrderDetail(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		data, err := svcCtx.AdminOrder.Detail(r.Context(), id)
		if err != nil {
			response.ServerError(w, "query failed")
			return
		}
		if data == nil {
			response.NotFound(w, "order not found")
			return
		}
		response.OK(w, data)
	}
}

func AdminOrderShip(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		var req types.ShipOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			response.BadRequest(w, "invalid params")
			return
		}
		if req.TrackingNo == "" {
			response.BadRequest(w, "tracking_no is required")
			return
		}

		if err := svcCtx.AdminOrder.Ship(r.Context(), id, req); err != nil {
			if err == admin.ErrOrderNotPaid {
				response.BadRequest(w, "only paid orders can be shipped")
				return
			}
			response.ServerError(w, "ship failed: "+err.Error())
			return
		}
		response.OK(w, map[string]string{"id": id, "status": "fulfilled"})
	}
}

func AdminOrderCancel(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			id = pathParam(r, "id")
		}
		var req types.CancelOrderReq
		_ = httpx.Parse(r, &req)
		if req.Reason == "" {
			req.Reason = "cancelled by admin"
		}

		if err := svcCtx.AdminOrder.Cancel(r.Context(), id, req.Reason); err != nil {
			if err == admin.ErrOrderNotPending {
				response.BadRequest(w, "only pending orders can be cancelled")
				return
			}
			response.ServerError(w, "cancel failed")
			return
		}
		response.OK(w, map[string]string{"id": id, "status": "cancelled"})
	}
}
