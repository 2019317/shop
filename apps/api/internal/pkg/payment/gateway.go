package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/password"
)

// 支付渠道适配器
// 阶段 0：提供 mock 实现，不引入任何第三方 SDK
// 阶段 1：新增 stripe.go 实现同一接口，通过配置 Driver 切换，业务代码零改动

var (
	ErrUnsupportedDriver = errors.New("unsupported payment driver")
	ErrPaymentFailed     = errors.New("payment failed")
)

// CreateInput 创建支付意图的入参
type CreateInput struct {
	OrderId     string
	OrderNo     string
	AmountCents int64
	Currency    string
	Email       string
	Description string
}

// IntentResult 支付意图
type IntentResult struct {
	ProviderIntentId string
	Status           string // pending | succeeded | failed
	// 真实渠道为跳转地址；mock 实现为空
	CheckoutUrl string
	// 客户端所需的额外参数（如 Stripe client_secret）
	ClientSecret string
}

// NotifyResult 回调解析结果
type NotifyResult struct {
	ProviderIntentId string
	OrderNo          string
	Status           string // succeeded | failed | refunded
	AmountCents      int64
	EventId          string // 用于幂等去重
	Raw              string
}

type Gateway interface {
	// Name 渠道标识，对应配置中的 Driver
	Name() string
	// Create 创建支付意图
	Create(ctx context.Context, in CreateInput) (*IntentResult, error)
	// Query 查询支付状态
	Query(ctx context.Context, providerIntentId string) (*IntentResult, error)
}

// New 根据 driver 构造渠道实现
func New(driver string) (Gateway, error) {
	switch driver {
	case "", "mock":
		return NewMock(), nil
	default:
		// 后续接入 stripe / paypal 时在此扩展
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDriver, driver)
	}
}

// ---------------- mock 实现 ----------------

type Mock struct{}

func NewMock() *Mock { return &Mock{} }

func (m *Mock) Name() string { return "mock" }

func (m *Mock) Create(ctx context.Context, in CreateInput) (*IntentResult, error) {
	// 允许 0 元订单（如 100% 优惠券），仅拒绝非法负值
	if in.AmountCents < 0 {
		return nil, ErrPaymentFailed
	}
	return &IntentResult{
		ProviderIntentId: "mock_" + password.RandomString(16),
		Status:           "pending",
		ClientSecret:     "mock_secret_" + password.RandomString(12),
	}, nil
}

func (m *Mock) Query(ctx context.Context, providerIntentId string) (*IntentResult, error) {
	return &IntentResult{
		ProviderIntentId: providerIntentId,
		Status:           "pending",
	}, nil
}

// SimulateSuccess 模拟支付成功回调（仅 mock 渠道使用，便于联调下单闭环）
func (m *Mock) SimulateSuccess(orderNo string, amountCents int64) *NotifyResult {
	return &NotifyResult{
		ProviderIntentId: "mock_" + password.RandomString(16),
		OrderNo:          orderNo,
		Status:           "succeeded",
		AmountCents:      amountCents,
		EventId:          fmt.Sprintf("mock_%d", time.Now().UnixNano()),
	}
}
