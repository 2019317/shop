package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/response"
)

// serverError 统一处理服务端错误：
// 1. 把原始错误写入日志（便于线上排查，避免只回 "query failed" 无从定位）
// 2. 对客户端只返回通用文案，不泄露内部细节
func serverError(w http.ResponseWriter, r *http.Request, err error, msg string) {
	logx.Errorf("%s %s error: %v", r.Method, r.URL.Path, err)
	response.ServerError(w, msg)
}

// badRequest 参数解析失败时记录具体原因，便于定位前端提交的结构问题
func badRequest(w http.ResponseWriter, r *http.Request, err error, msg string) {
	logx.Errorf("%s %s bad request: %v", r.Method, r.URL.Path, err)
	response.BadRequest(w, msg)
}
