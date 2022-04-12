package errorx

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"minicode.com/sirius/go-back-server/config/cfgstatus"
)

type CodeError struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
}

type codeErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
}

// api
func NewCodeError(code int, msg string) error {
	return &CodeError{Code: code, Msg: msg}
}

// rpc 专用
func NewCodeErrorRPC(code int, msg string) error {
	return status.Error(codes.Code(code), msg)
}

func (e *CodeError) Error() string {
	return e.Msg
}

func (e *CodeError) Data() *codeErrorResponse {
	return &codeErrorResponse{
		Code: e.Code,
		Msg:  e.Msg,
	}
}

func Parse(err error) *CodeError {
	s := status.Convert(err)
	return &CodeError{
		Code: int(s.Code()),
		Msg:  s.Message(),
	}
}

func NewDefaultError(msg string) error {
	return NewCodeError(cfgstatus.SystemDefault, msg)
}

func NewSystemUserAuthError(msg string) error {
	return NewCodeError(cfgstatus.SystemUserAuth, msg)
}

func NewSystemUserAuthSignError(msg string) error {
	return NewCodeError(cfgstatus.SystemUserAuthSign, msg)
}

func NewSystemUserAuthParamError(msg string) error {
	return NewCodeError(cfgstatus.SystemUserAuthParam, msg)
}

func NewSystemUserAuthChangeTypeError(msg string) error {
	return NewCodeError(cfgstatus.SystemUserAuthChangeType, msg)
}

func NewSystemFrequentRequestError(msg string) error {
	return NewCodeError(cfgstatus.SystemFrequentRequest, msg)
}

// 获取 Uin 失败
func NewSystemGetUinFromContextError() error {
	return status.Error(cfgstatus.SystemGetUinFromContext, "获取uin失败")
}

func NewParamUserBehaviorApiError(msg string) error {
	return NewCodeError(cfgstatus.ParamUserBehaviorApi, msg)
}

func NewParamUserBehaviorRpcError(msg string) error {
	return status.Error(cfgstatus.ParamUserBehaviorRpc, msg)
}

func NewGetDataUserBehaviorApiError(msg string) error {
	return NewCodeError(cfgstatus.GetDataUserBehaviorApi, msg)
}

func NewGetDataUserBehaviorRpcError(msg string) error {
	return NewCodeError(cfgstatus.GetDataUserBehaviorRpc, msg)
}

// 用户行为 rpc 获取db 数据
func NewGetDbDataUserBehaviorRpcError(msg string) error {
	return status.Error(cfgstatus.UserBehaviorDbGet, msg)
}

// 用户行为从缓存获取数据
func NewGetCacheDataUserBehaviorRpcError(msg string) error {
	return status.Error(cfgstatus.GetCacheDataUserBehaviorRpc, msg)
}

func NewSendRMQUserBehaviorRpcError(msg string) error {
	return status.Error(cfgstatus.UserBehaviorRMqSend, msg)
}

func NewUserBehaviorCommentRpcMarshal(msg string) error {
	return status.Error(cfgstatus.UserBehaviorCommentMarshalRPC, msg)
}

func NewUserBehaviorCommentRpcUnMarshal(msg string) error {
	return status.Error(cfgstatus.UserBehaviorCommentUnMarshalRPC, msg)
}

//// cube 错误
//
//func NewParamCubeIdError(msg string) error {
//	return NewCodeError(cfgstatus.ParamCubeIdError, msg)
//}
