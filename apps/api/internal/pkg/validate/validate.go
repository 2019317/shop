package validate

import (
	"regexp"
	"strings"
)

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// Email 校验邮箱格式（宽松校验，拒绝空白与明显非法值）
func Email(s string) bool {
	s = strings.TrimSpace(s)
	return s != "" && len(s) <= 254 && emailRe.MatchString(s)
}

// Currency 校验 ISO 4217 三位大写货币码
func Currency(s string) bool {
	return len(s) == 3 && isAlphaUpper(s)
}

// Country 校验 ISO 3166-1 alpha-2 两位大写国家码
func Country(s string) bool {
	return len(s) == 2 && isAlphaUpper(s)
}

// MaxLen 限制自由文本长度，防止超长输入造成资源浪费
func MaxLen(s string, n int) bool {
	return len(s) <= n
}

func isAlphaUpper(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
