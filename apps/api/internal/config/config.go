package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

// Config 对应 etc/api.yaml
// 说明：领域相关配置按模块分组，后期拆分为独立服务时，各组配置随之迁移
type Config struct {
	rest.RestConf

	CorsOrigins string `json:",default=*"`

	// ---------- 数据库（本机容器 / Neon / 云 RDS）----------
	Database struct {
		Dsn          string // PostgreSQL 连接串
		MaxOpenConns int    `json:",default=20"`
		MaxIdleConns int    `json:",default=5"`
	}

	// ---------- 缓存 ----------
	Redis redis.RedisConf

	// ---------- 鉴权 ----------
	Auth struct {
		AccessSecret string
		AccessExpire int64 // 秒
	}
	AdminAuth struct {
		AccessSecret string
		AccessExpire int64
	}

	// ---------- 对象存储（Cloudflare R2，S3 协议）----------
	R2 struct {
		AccountId       string
		AccessKeyId     string
		SecretAccessKey string
		Bucket          string
		PublicUrl       string // 自定义域名，如 https://img.yourbrand.com
		PresignExpire   int64  `json:",default=600"`
	}

	// ---------- 支付（适配器：Stripe / PayPal）----------
	Payment struct {
		Driver          string `json:",default=stripe"` // stripe | paypal
		StripeSecretKey string
		StripeWebhookKey string
		Currency        string `json:",default=USD"`
	}

	// ---------- 邮件 ----------
	Mail struct {
		ApiKey string
		From   string
	}
}
