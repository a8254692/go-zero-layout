package cfgstatus

// admin 相关的接口返回响应码
const (
	Success = 0
	// 第三方api 相关
	FeiShuUserCode         = 11000
	FeiShuUserMarshalError = 11001

	FeiShuFeedbackMsgParamsIllegal = 11100
	FeiShuFeedbackMsgTokenError    = 11101
	FeiShuFeedbackMsgError         = 11102
	FeiShuFeedbackMsgRetError      = 11103
)
