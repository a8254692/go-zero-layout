package response

import (
    "net/http"

    "github.com/zeromicro/go-zero/rest/httpx"
)

type body struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse 统一封装成功响应值
func SuccessResponse(w http.ResponseWriter, resp interface{}) {
    var body body

    body.Code = 0
    body.Message = "Succeed!"
    body.Data = resp

    httpx.OkJson(w, body)
}
