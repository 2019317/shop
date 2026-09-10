package event

import (
	"context"
	"log"
	"sync"
)

// 领域事件总线
// 阶段 0：内存同步实现
// 阶段 1：替换为 asynq / RabbitMQ，业务代码无需改动
type Event struct {
	Type    string
	Payload map[string]interface{}
}

type Handler func(ctx context.Context, e Event) error

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewBus() *Bus {
	return &Bus{handlers: make(map[string][]Handler)}
}

// 事件名常量
const (
	OrderCreated  = "order.created"
	OrderPaid     = "order.paid"
	OrderShipped  = "order.shipped"
	OrderCanceled = "order.canceled"
)

func (b *Bus) Subscribe(eventType string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], h)
}

// Publish 同步派发；单个 handler 失败不影响其他订阅者
// 注意：切换到 MQ 后此处将变为异步投递，调用方不应依赖其同步语义
func (b *Bus) Publish(ctx context.Context, e Event) {
	b.mu.RLock()
	handlers := b.handlers[e.Type]
	b.mu.RUnlock()

	for _, h := range handlers {
		// 每个 handler 独立 recover，避免一个订阅者 panic 拖垮主流程
		func(h Handler) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("event handler panic: %v", r)
				}
			}()
			if err := h(ctx, e); err != nil {
				log.Printf("event %s handler error: %v", e.Type, err)
			}
		}(h)
	}
}
