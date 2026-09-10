package svc

import (
	"context"
	"log"
	"time"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/event"
)

// seconds 将配置中的秒数转为 Duration，配置缺失时返回默认值
func seconds(v int64) time.Duration {
	if v <= 0 {
		return 12 * time.Hour
	}
	return time.Duration(v) * time.Second
}

// registerEventHandlers 订阅领域事件
// 阶段 0 仅记录日志；接入邮件 / MQ 后在此挂接真实处理逻辑，业务代码无需改动
func registerEventHandlers(bus *event.Bus) {
	bus.Subscribe(event.OrderCreated, func(ctx context.Context, e event.Event) error {
		log.Printf("[event] %s order=%v amount=%v", e.Type, e.Payload["order_no"], e.Payload["total"])
		return nil
	})

	bus.Subscribe(event.OrderPaid, func(ctx context.Context, e event.Event) error {
		log.Printf("[event] %s order=%v email=%v", e.Type, e.Payload["order_no"], e.Payload["email"])
		// TODO: 支付成功后发送确认邮件（接入 Resend 后实现）
		return nil
	})

	bus.Subscribe(event.OrderShipped, func(ctx context.Context, e event.Event) error {
		log.Printf("[event] %s order=%v tracking=%v", e.Type, e.Payload["order_no"], e.Payload["tracking_no"])
		// TODO: 发货通知邮件
		return nil
	})

	bus.Subscribe(event.OrderCanceled, func(ctx context.Context, e event.Event) error {
		log.Printf("[event] %s order=%v", e.Type, e.Payload["order_no"])
		return nil
	})
}
