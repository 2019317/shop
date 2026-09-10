package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Cloudflare R2 兼容 S3 协议，此处自行实现 AWS SigV4 预签名，
// 避免为了单一功能引入 aws-sdk-go-v2 整套依赖。
type Config struct {
	AccountId       string
	AccessKeyId     string
	SecretAccessKey string
	Bucket          string
	PublicUrl       string // 自定义域名，如 https://img.yourbrand.com
	Region          string // S3 默认 us-east-1
}

type Storage struct {
	cfg Config
}

func New(cfg Config) *Storage {
	if cfg.Region == "" {
		cfg.Region = "auto" // R2 使用 auto
	}
	return &Storage{cfg: cfg}
}

// Endpoint 形如 https://<account>.r2.cloudflarestorage.com
func (s *Storage) Endpoint() string {
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com", s.cfg.AccountId)
}

// PublicURL 返回对象的可访问地址
func (s *Storage) PublicURL(objectKey string) string {
	base := strings.TrimSuffix(s.cfg.PublicUrl, "/")
	if base == "" {
		return fmt.Sprintf("%s/%s/%s", s.Endpoint(), s.cfg.Bucket, objectKey)
	}
	return base + "/" + objectKey
}

// PresignPut 生成用于直传的预签名 PUT URL
func (s *Storage) PresignPut(objectKey, contentType string, ttl time.Duration) (string, error) {
	host := fmt.Sprintf("%s.r2.cloudflarestorage.com", s.cfg.AccountId)
	path := fmt.Sprintf("/%s/%s", s.cfg.Bucket, escapePath(objectKey))

	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, s.cfg.Region)

	query := url.Values{}
	query.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	query.Set("X-Amz-Credential", s.cfg.AccessKeyId+"/"+credentialScope)
	query.Set("X-Amz-Date", amzDate)
	query.Set("X-Amz-Expires", fmt.Sprintf("%d", int(ttl.Seconds())))
	query.Set("X-Amz-SignedHeaders", "content-type;host")

	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\n", contentType, host)
	signedHeaders := "content-type;host"
	canonicalRequest := strings.Join([]string{
		"PUT", path, canonicalQuery(query), canonicalHeaders, signedHeaders, "UNSIGNED-PAYLOAD",
	}, "\n")

	strToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256", amzDate, credentialScope,
		hexSHA256([]byte(canonicalRequest)),
	}, "\n")

	signingKey := hmacSHA256([]byte("AWS4"+s.cfg.SecretAccessKey), dateStamp)
	signingKey = hmacSHA256(signingKey, s.cfg.Region)
	signingKey = hmacSHA256(signingKey, "s3")
	signingKey = hmacSHA256(signingKey, "aws4_request")
	signature := fmt.Sprintf("%x", hmacSHA256(signingKey, strToSign))

	query.Set("X-Amz-Signature", signature)
	return fmt.Sprintf("https://%s%s?%s", host, path, canonicalQuery(query)), nil
}

func canonicalQuery(v url.Values) string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v.Get(k)))
	}
	return strings.Join(parts, "&")
}

func escapePath(key string) string {
	segs := strings.Split(key, "/")
	for i, seg := range segs {
		segs[i] = url.PathEscape(seg)
	}
	return strings.Join(segs, "/")
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func hexSHA256(b []byte) string {
	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum[:])
}
