package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

// VerifySignature 校验支付回调的 HMAC-SHA256 签名。
// 签名格式：hex(hmac_sha256(secret, rawBody))，兼容可选的 "sha256=" 前缀。
// 返回 true 表示签名有效。
func VerifySignature(secret, rawBody, provided string) bool {
	if secret == "" || provided == "" {
		return false
	}
	provided = strings.TrimPrefix(provided, "sha256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(rawBody))
	expected := mac.Sum(nil)

	got, err := hex.DecodeString(provided)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(expected, got) == 1
}
