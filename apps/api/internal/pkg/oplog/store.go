// Package oplog 提供进程内的请求日志环形缓冲，供后台「日志管理」页面查询。
// 说明：日志仅保存在内存中（进程重启即清空），适合单实例部署下的运维排查；
// 若后续需要持久化/多实例聚合，可在此基础上替换为写入 Redis/文件/ELK。
package oplog

import (
	"context"
	"strings"
	"sync"
	"time"
)

// reasonKey 请求上下文键：handler 写入真实错误原因，日志中间件读取后落库。
// 定义在此处供 handler 与 middleware 共享，避免两处定义不同类型而取不到值。
type reasonKey struct{}

// WithReason 在上下文中挂载一个错误原因载体，返回新上下文
func WithReason(ctx context.Context, holder *string) context.Context {
	return context.WithValue(ctx, reasonKey{}, holder)
}

// SetReason 将错误原因写入上下文（若存在载体）
func SetReason(ctx context.Context, reason string) {
	if holder, ok := ctx.Value(reasonKey{}).(*string); ok && holder != nil {
		*holder = reason
	}
}

// Entry 单条请求日志
type Entry struct {
	Id         int64     `json:"id"`
	Time       time.Time `json:"time"`
	Level      string    `json:"level"` // info | warn | error
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Status     int       `json:"status"`
	DurationMs int64     `json:"duration_ms"`
	Ip         string    `json:"ip"`
	Message    string    `json:"message"` // 简要信息（状态描述或错误原因）
	Error      string    `json:"error,omitempty"`
}

// Filter 查询条件
type Filter struct {
	Level    string
	Keyword  string
	Path     string
	Page     int
	PageSize int
}

// Store 环形缓冲日志存储（并发安全）
type Store struct {
	mu      sync.RWMutex
	buf     []Entry
	next    int
	count   int
	seq     int64
	cap     int
}

// NewStore 创建容量为 capacity 的日志存储
func NewStore(capacity int) *Store {
	if capacity <= 0 {
		capacity = 1000
	}
	return &Store{
		buf: make([]Entry, capacity),
		cap: capacity,
	}
}

// Add 追加一条日志，返回带自增 id 的条目
func (s *Store) Add(e Entry) Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.seq++
	e.Id = s.seq
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	s.buf[s.next] = e
	s.next = (s.next + 1) % s.cap
	if s.count < s.cap {
		s.count++
	}
	return e
}

// Query 按条件分页查询，返回（列表, 总数）。列表按时间倒序（最新在前）。
func (s *Store) Query(f Filter) ([]Entry, int) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 200 {
		f.PageSize = 50
	}

	s.mu.RLock()
	all := make([]Entry, 0, s.count)
	// 按时间正序取出所有有效条目
	start := (s.next - s.count + s.cap) % s.cap
	for i := 0; i < s.count; i++ {
		all = append(all, s.buf[(start+i)%s.cap])
	}
	s.mu.RUnlock()

	// 倒序 + 过滤
	matched := make([]Entry, 0, len(all))
	for i := len(all) - 1; i >= 0; i-- {
		e := all[i]
		if f.Level != "" && e.Level != f.Level {
			continue
		}
		if f.Path != "" && !strings.Contains(strings.ToLower(e.Path), strings.ToLower(f.Path)) {
			continue
		}
		if f.Keyword != "" {
			kw := strings.ToLower(f.Keyword)
			if !strings.Contains(strings.ToLower(e.Path), kw) &&
				!strings.Contains(strings.ToLower(e.Message), kw) &&
				!strings.Contains(strings.ToLower(e.Error), kw) {
				continue
			}
		}
		matched = append(matched, e)
	}

	total := len(matched)
	offset := (f.Page - 1) * f.PageSize
	if offset >= total {
		return []Entry{}, total
	}
	end := offset + f.PageSize
	if end > total {
		end = total
	}
	return matched[offset:end], total
}

// Clear 清空日志
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next = 0
	s.count = 0
}
