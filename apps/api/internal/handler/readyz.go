package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/yourname/stationery-shop/apps/api/internal/svc"
)

// Readyz 就绪探针：检查数据库连通性与关键表是否存在。
// 排查「接口全部 500」时可直接访问，无需鉴权，快速定位是数据库/迁移问题还是业务问题。
func Readyz(svcCtx *svc.ServiceContext) http.HandlerFunc {
	// 关键表清单：任一缺失都意味着迁移未执行完整
	tables := []string{
		"catalog.products",
		"catalog.categories",
		"catalog.product_translations",
		"catalog.category_translations",
		"orders.orders",
		"promotion.coupons",
		"fulfillment.shipping_rules",
		"fulfillment.shipments",
	}

	return func(w http.ResponseWriter, r *http.Request) {
		result := map[string]interface{}{
			"status":  "ok",
			"db":      "ok",
			"missing": []string{},
		}
		missing := []string{}

		for _, t := range tables {
			var exists bool
			err := svcCtx.DB.QueryRowCtx(r.Context(), &exists, `SELECT to_regclass($1) IS NOT NULL`, t)
			if err != nil {
				result["status"] = "error"
				result["db"] = err.Error()
				httpx.OkJson(w, result)
				return
			}
			if !exists {
				missing = append(missing, t)
			}
		}

		if len(missing) > 0 {
			result["status"] = "degraded"
			result["missing"] = missing
		}
		httpx.OkJson(w, result)
	}
}
