package i18n

import (
	"net/http"
	"strings"
)

// 支持的语言与默认语言
const (
	DefaultLocale = "en"
	ZhLocale      = "zh"
	EnLocale      = "en"
)

var supported = map[string]bool{
	"en":   true,
	"zh":   true,
	"en-US": true,
	"zh-CN": true,
}

// Normalize 将语言标签归一化为受支持的 locale，非法值回落默认语言
// 例：zh-CN → zh，en-US → en，fr → en
func Normalize(raw string) string {
	if raw == "" {
		return DefaultLocale
	}
	tag := strings.ToLower(strings.TrimSpace(raw))
	if supported[tag] {
		return tag[:2]
	}
	// 取主语言部分：zh-Hans-CN → zh
	if idx := strings.Index(tag, "-"); idx > 0 {
		base := tag[:idx]
		if base == "zh" || base == "en" {
			return base
		}
	}
	if strings.HasPrefix(tag, "zh") {
		return ZhLocale
	}
	return DefaultLocale
}

// FromRequest 优先取查询参数 locale，其次 Accept-Language
func FromRequest(r *http.Request) string {
	if r == nil {
		return DefaultLocale
	}

	if q := r.URL.Query().Get("locale"); q != "" {
		return Normalize(q)
	}

	accept := r.Header.Get("Accept-Language")
	if accept == "" {
		return DefaultLocale
	}

	// Accept-Language: zh-CN,zh;q=0.9,en;q=0.8
	for _, part := range strings.Split(accept, ",") {
		tag := strings.TrimSpace(strings.Split(part, ";")[0])
		if normalized := Normalize(tag); normalized != DefaultLocale {
			return normalized
		}
		if strings.HasPrefix(strings.ToLower(tag), "en") {
			return EnLocale
		}
	}
	return DefaultLocale
}
