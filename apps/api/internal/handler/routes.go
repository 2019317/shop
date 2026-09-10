package handler

import (
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest"

	"github.com/yourname/stationery-shop/apps/api/internal/middleware"
	"github.com/yourname/stationery-shop/apps/api/internal/svc"
)

// routePrefix 所有路由的统一前缀（由 AddRoutes 承载，不再传给 MustNewServer）
const routePrefix = "/api/v1"

// RegisterHandlers 注册路由
// 分层约定：handler 只做参数绑定与响应封装，业务逻辑一律下沉到 logic
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	cors := middleware.CORSHandlerFunc(strings.Split(serverCtx.Config.CorsOrigins, ","))

	// ---------------- 前台公开接口 ----------------
	server.AddRoutes(rest.WithMiddlewares(
		[]rest.Middleware{cors},
		rest.Route{Method: http.MethodGet, Path: "/healthz", Handler: Healthz()},
		rest.Route{Method: http.MethodGet, Path: "/products", Handler: ProductList(serverCtx)},
		rest.Route{Method: http.MethodGet, Path: "/products/:slug", Handler: ProductDetail(serverCtx)},
		rest.Route{Method: http.MethodGet, Path: "/categories", Handler: CategoryList(serverCtx)},

		// 下单与订单查询
		rest.Route{Method: http.MethodPost, Path: "/orders", Handler: CreateOrder(serverCtx)},
		rest.Route{Method: http.MethodGet, Path: "/orders/:order_no", Handler: OrderDetail(serverCtx)},
		// 支付回调（当前供 mock 渠道调用，接入 Stripe 后由其 Webhook 调用）
		rest.Route{Method: http.MethodPost, Path: "/payments/notify", Handler: PayNotify(serverCtx)},
	), rest.WithPrefix(routePrefix))

	// ---------------- 后台接口 ----------------
	adminAuth := middleware.AdminAuthHandlerFunc(serverCtx.Config.AdminAuth.AccessSecret)

	// 登录：无需鉴权
	server.AddRoutes(rest.WithMiddlewares(
		[]rest.Middleware{cors},
		rest.Route{Method: http.MethodPost, Path: "/admin/auth/login", Handler: AdminLogin(serverCtx)},
	), rest.WithPrefix(routePrefix))

	// 其余后台接口：CORS + JWT 鉴权
	server.AddRoutes(rest.WithMiddlewares(
		[]rest.Middleware{cors, adminAuth},
		rest.Route{Method: http.MethodGet, Path: "/admin/products", Handler: AdminProductList(serverCtx)},
		rest.Route{Method: http.MethodGet, Path: "/admin/products/:id", Handler: AdminProductDetail(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/admin/products", Handler: AdminProductCreate(serverCtx)},
		rest.Route{Method: http.MethodPut, Path: "/admin/products/:id", Handler: AdminProductUpdate(serverCtx)},
		rest.Route{Method: http.MethodDelete, Path: "/admin/products/:id", Handler: AdminProductDelete(serverCtx)},

		rest.Route{Method: http.MethodGet, Path: "/admin/categories", Handler: AdminCategoryList(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/admin/categories", Handler: AdminCategoryCreate(serverCtx)},
		rest.Route{Method: http.MethodPut, Path: "/admin/categories/:id", Handler: AdminCategoryUpdate(serverCtx)},
		rest.Route{Method: http.MethodDelete, Path: "/admin/categories/:id", Handler: AdminCategoryDelete(serverCtx)},

		rest.Route{Method: http.MethodPost, Path: "/admin/assets/presign", Handler: AdminAssetPresign(serverCtx)},

		// 订单管理
		rest.Route{Method: http.MethodGet, Path: "/admin/orders", Handler: AdminOrderList(serverCtx)},
		rest.Route{Method: http.MethodGet, Path: "/admin/orders/:id", Handler: AdminOrderDetail(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/admin/orders/:id/ship", Handler: AdminOrderShip(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/admin/orders/:id/cancel", Handler: AdminOrderCancel(serverCtx)},
	), rest.WithPrefix(routePrefix))
}
