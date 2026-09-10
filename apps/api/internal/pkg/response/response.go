package response

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// Result 统一响应体
type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Paged 分页包装
type Paged struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func OK(w http.ResponseWriter, data interface{}) {
	httpx.OkJson(w, Result{Code: 0, Msg: "ok", Data: data})
}

func Fail(w http.ResponseWriter, code int, msg string) {
	httpx.WriteJson(w, http.StatusOK, Result{Code: code, Msg: msg})
}

func BadRequest(w http.ResponseWriter, msg string) {
	Fail(w, 400, msg)
}

func Unauthorized(w http.ResponseWriter, msg string) {
	Fail(w, 401, msg)
}

func Forbidden(w http.ResponseWriter, msg string) {
	Fail(w, 403, msg)
}

func NotFound(w http.ResponseWriter, msg string) {
	Fail(w, 404, msg)
}

func ServerError(w http.ResponseWriter, msg string) {
	Fail(w, 500, msg)
}
