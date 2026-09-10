package handler

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/rest"

	"github.com/yourname/stationery-shop/apps/api/internal/middleware"
	"github.com/yourname/stationery-shop/apps/api/internal/svc"
)

// routePrefix 所有路由的统一前缀（由 AddRoutes 承载，不再传给 MustNewServer）
const routePrefix = "/api/v1"

// RegisterHandlers 注册路由
// 分层约定：handler 只做参数绑定与响应封装，业务逻辑一律下沉到 logic
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	cors := middleware.NewCorsHandlerFunc(serverCtx.Config.CorsOrigins)
	security := middleware.NewSecurityHeadersHandlerFunc()
	adminAuth := middleware.AdminAuthHandlerFunc(serverCtx.Config.AdminAuth.AccessSecret)
	requireAdmin := middleware.RequireRole("admin")

	// ---------------- 前台公开读接口 ----------------
	server.AddRoutes(rest.WithMiddlewares(
		[]rest.Middleware{cors, security},
		rest.Route{Method: http.MethodGet, Path: "/healthz", Handler: Healthz()},
		rest.Route{Method: http.MethodGet, Path: "/readyz", Handler: Readyz(serverCtx)},
		rest.Route{Method: http.MethodGet, Path: "/products", Handler: ProductList(serverCtx)},
		rest.Route{Method: http.MethodGet, Path: "/products/:slug", Handler: ProductDetail(serverCtx)},
		rest.Route{Method: http.MethodGet, Path: "/categories", Handler: CategoryList(serverCtx)},
		// 优惠券：前台可查看可用券并试算
		rest.Route{Method: http.MethodGet, Path: "/coupons", Handler: CouponActive(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/coupons/validate", Handler: CouponValidate(serverCtx)},
	), rest.WithPrefix(routePrefix))

	// ---------------- 下单 / 订单查询 / 支付回调：限流 ----------------
	orderLimit := middleware.RateLimit(20, time.Minute)
	server.AddRoutes(rest.WithMiddlewares(
		[]rest.Middleware{cors, security, orderLimit},
		rest.Route{Method: http.MethodPost, Path: "/orders", Handler: CreateOrder(serverCtx)},
		rest.Route{Method: http.MethodGet, Path: "/orders/:order_no", Handler: OrderDetail(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/payments/notify", Handler: PayNotify(serverCtx)},
	), rest.WithPrefix(routePrefix))

	// ---------------- 后台登录：限流（防爆破） ----------------
	loginLimit := middleware.RateLimit(10, time.Minute)
	server.AddRoutes(rest.WithMiddlewares(
		[]rest.Middleware{cors, security, loginLimit},
		rest.Route{Method: http.MethodPost, Path: "/admin/auth/login", Handler: AdminLogin(serverCtx)},
	), rest.WithPrefix(routePrefix))

	// ---------------- 后台常规接口（需 JWT） ----------------
	server.AddRoutes(rest.WithMiddlewares(
		[]rest.Middleware{cors, security, adminAuth},
		rest.Route{Method: http.MethodGet, Path: "/admin/products", Handler: AdminProductList(serverCtx)},
		rest.Route{Method: http.MethodGet, Path: "/admin/products/:id", Handler: AdminProductDetail(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/admin/products", Handler: AdminProductCreate(serverCtx)},
		rest.Route{Method: http.MethodPut, Path: "/admin/products/:id", Handler: AdminProductUpdate(serverCtx)},

		rest.Route{Method: http.MethodGet, Path: "/admin/categories", Handler: AdminCategoryList(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/admin/categories", Handler: AdminCategoryCreate(serverCtx)},
		rest.Route{Method: http.MethodPut, Path: "/admin/categories/:id", Handler: AdminCategoryUpdate(serverCtx)},

		rest.Route{Method: http.MethodPost, Path: "/admin/assets/presign", Handler: AdminAssetPresign(serverCtx)},

		rest.Route{Method: http.MethodGet, Path: "/admin/orders", Handler: AdminOrderList(serverCtx)},
		rest.Route{Method: http.MethodGet, Path: "/admin/orders/:id", Handler: AdminOrderDetail(serverCtx)},

		// 优惠券管理
		rest.Route{Method: http.MethodGet, Path: "/admin/coupons", Handler: AdminCouponList(serverCtx)},
		rest.Route{Method: http.MethodGet, Path: "/admin/coupons/:id", Handler: AdminCouponDetail(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/admin/coupons", Handler: AdminCouponCreate(serverCtx)},
		rest.Route{Method: http.MethodPut, Path: "/admin/coupons/:id", Handler: AdminCouponUpdate(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/admin/coupons/:id/status", Handler: AdminCouponSetStatus(serverCtx)},

		// 运费规则管理
		rest.Route{Method: http.MethodGet, Path: "/admin/shipping/rules", Handler: AdminShippingRuleList(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/admin/shipping/rules", Handler: AdminShippingRuleCreate(serverCtx)},
		rest.Route{Method: http.MethodPut, Path: "/admin/shipping/rules/:id", Handler: AdminShippingRuleUpdate(serverCtx)},
		rest.Route{Method: http.MethodDelete, Path: "/admin/shipping/rules/:id", Handler: AdminShippingRuleDelete(serverCtx)},
	), rest.WithPrefix(routePrefix))

	// ---------------- 后台高危操作（仅 admin 角色） ----------------
	server.AddRoutes(rest.WithMiddlewares(
		[]rest.Middleware{cors, security, adminAuth, requireAdmin},
		rest.Route{Method: http.MethodDelete, Path: "/admin/products/:id", Handler: AdminProductDelete(serverCtx)},
		rest.Route{Method: http.MethodDelete, Path: "/admin/categories/:id", Handler: AdminCategoryDelete(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/admin/orders/:id/ship", Handler: AdminOrderShip(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/admin/orders/:id/cancel", Handler: AdminOrderCancel(serverCtx)},
		rest.Route{Method: http.MethodPost, Path: "/admin/orders/:id/delivered", Handler: AdminOrderMarkDelivered(serverCtx)},
	), rest.WithPrefix(routePrefix))
}
