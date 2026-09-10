package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// 轻量 HS256 JWT 实现（仅标准库）
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpired      = errors.New("token expired")
)

type Claims struct {
	Sub   string `json:"sub"`   // 用户 ID
	Role  string `json:"role"`  // admin | operator | member
	Scope string `json:"scope"` // admin | shop
	Exp   int64  `json:"exp"`
}

func Sign(claims Claims, secret string, ttl time.Duration) (string, error) {
	claims.Exp = time.Now().Add(ttl).Unix()
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	h := base64.RawURLEncoding.EncodeToString(header)
	p := base64.RawURLEncoding.EncodeToString(payload)
	signingInput := h + "." + p
	sig := signHS256(signingInput, secret)
	return signingInput + "." + sig, nil
}

func Parse(token, secret string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	// 恒定时间比较签名，避免时序攻击泄露校验结果
	expected := signHS256(parts[0]+"."+parts[1], secret)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[2])) != 1 {
		return nil, ErrInvalidToken
	}

	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if claims.Exp < time.Now().Unix() {
		return nil, ErrExpired
	}
	return &claims, nil
}

func signHS256(input, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
