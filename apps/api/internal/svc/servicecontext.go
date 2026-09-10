package svc

import (
	"log"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/yourname/stationery-shop/apps/api/internal/config"
	"github.com/yourname/stationery-shop/apps/api/internal/logic/admin"
	shopLogic "github.com/yourname/stationery-shop/apps/api/internal/logic/shop"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/event"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/payment"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/storage"
	"github.com/yourname/stationery-shop/apps/api/internal/repo"
)

// ServiceContext 依赖容器
// 约定：各业务模块只通过此容器获取依赖，禁止跨模块直接访问彼此的数据访问层
type ServiceContext struct {
	Config config.Config

	DB      sqlx.SqlConn
	Redis   *redis.Redis
	Storage *storage.Storage
	Bus     *event.Bus

	// 后台
	AdminAuth     *admin.AuthLogic
	AdminProduct  *admin.ProductLogic
	AdminCategory *admin.CategoryLogic
	AdminAsset    *admin.AssetLogic
	AdminOrder    *admin.OrderLogic

	// 前台
	ShopProduct *shopLogic.ProductLogic
	ShopOrder   *shopLogic.OrderLogic
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := sqlx.NewSqlConn("postgres", c.Database.Dsn)

	// go-zero v1.7.0 的 RedisConf.NewRedis 已内部完成连接，无需也不再支持 Start()
	rds := c.Redis.NewRedis()

	store := storage.New(storage.Config{
		AccountId:       c.R2.AccountId,
		AccessKeyId:     c.R2.AccessKeyId,
		SecretAccessKey: c.R2.SecretAccessKey,
		Bucket:          c.R2.Bucket,
		PublicUrl:       c.R2.PublicUrl,
	})

	// 支付渠道：按配置切换，接入 Stripe 只需新增实现
	gateway, err := payment.New(c.Payment.Driver)
	if err != nil {
		log.Fatalf("init payment gateway: %v", err)
	}

	productRepo := repo.NewProductRepo(db)
	categoryRepo := repo.NewCategoryRepo(db)
	adminUserRepo := repo.NewAdminUserRepo(db)
	orderRepo := repo.NewOrderRepo(db)
	inventoryRepo := repo.NewInventoryRepo(db)
	shippingRepo := repo.NewShippingRepo(db)

	bus := event.NewBus()
	registerEventHandlers(bus)

	return &ServiceContext{
		Config:  c,
		DB:      db,
		Redis:   rds,
		Storage: store,
		Bus:     bus,

		AdminAuth:     admin.NewAuthLogic(adminUserRepo, c.AdminAuth.AccessSecret, seconds(c.AdminAuth.AccessExpire)),
		AdminProduct:  admin.NewProductLogic(productRepo),
		AdminCategory: admin.NewCategoryLogic(categoryRepo, store),
		AdminAsset:    admin.NewAssetLogic(store),
		AdminOrder:    admin.NewOrderLogic(orderRepo, inventoryRepo, bus),

		ShopProduct: shopLogic.NewProductLogic(productRepo, categoryRepo, store.PublicURL),
		ShopOrder: shopLogic.NewOrderLogic(
			orderRepo, productRepo, inventoryRepo, shippingRepo, gateway, bus, store.PublicURL),
	}
}
