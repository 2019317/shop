package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/storage"
	"github.com/yourname/stationery-shop/apps/api/internal/types"
)

// 上传限制：2GB 服务器不接收图片流，图片由浏览器直传 R2
const (
	MaxUploadSize = 5 * 1024 * 1024
	PresignTTL    = 10 * time.Minute
)

var allowedContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/avif": ".avif",
}

var ErrInvalidFileType = errors.New("unsupported file type")

type AssetLogic struct {
	store *storage.Storage
}

func NewAssetLogic(store *storage.Storage) *AssetLogic {
	return &AssetLogic{store: store}
}

// Presign 返回直传 R2 的预签名地址，图片不经过服务器
func (l *AssetLogic) Presign(ctx context.Context, req types.PresignReq) (*types.PresignResp, error) {
	ext, ok := allowedContentTypes[req.ContentType]
	if !ok {
		return nil, ErrInvalidFileType
	}
	if req.Size > MaxUploadSize {
		return nil, fmt.Errorf("file too large, max %d bytes", MaxUploadSize)
	}

	now := time.Now().UTC()
	objectKey := fmt.Sprintf("products/%s/%s/%s%s",
		now.Format("2006"), now.Format("01"), randomKey(), ext)

	uploadUrl, err := l.store.PresignPut(objectKey, req.ContentType, PresignTTL)
	if err != nil {
		return nil, err
	}

	return &types.PresignResp{
		UploadUrl: uploadUrl,
		ObjectKey: objectKey,
		PublicUrl: l.store.PublicURL(objectKey),
	}, nil
}

func randomKey() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
