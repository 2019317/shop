package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/yourname/stationery-shop/apps/api/internal/pkg/oplog"
	"github.com/yourname/stationery-shop/apps/api/internal/pkg/response"
)

// serverError 统一处理服务端错误：
// 1. 把原始错误写入日志（便于线上排查，避免只回 "query failed" 无从定位）
// 2. 记录到请求上下文，供日志管理页面展示
// 3. 对客户端只返回通用文案，不泄露内部细节
func serverError(w http.ResponseWriter, r *http.Request, err error, msg string) {
	logx.Errorf("%s %s error: %v", r.Method, r.URL.Path, err)
	oplog.SetReason(r.Context(), err.Error())
	response.ServerError(w, msg)
}

// badRequest 参数解析/校验失败时记录具体原因，便于定位前端提交的结构问题
func badRequest(w http.ResponseWriter, r *http.Request, err error, msg string) {
	logx.Errorf("%s %s bad request: %v", r.Method, r.URL.Path, err)
	oplog.SetReason(r.Context(), err.Error())
	response.BadRequest(w, msg)
}

// badRequestMsg 无 error 对象时，直接记录原因
func badRequestMsg(w http.ResponseWriter, r *http.Request, reason, msg string) {
	logx.Errorf("%s %s bad request: %s", r.Method, r.URL.Path, reason)
	oplog.SetReason(r.Context(), reason)
	response.BadRequest(w, msg)
}
