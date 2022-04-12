package cfgstatus

const (
	/*
	   错误码组成：错误类型+应用标识+错误编码
	   错误码位数：7位
	   错误码示例：1000000
	   使用规范：只增不改、避免混乱、先占先得、写好注释

	   错误类型(2位数字,10开始):
	   系统错误：10
	   参数错误：11
	   获取数据错误：12
	   缓存错误：13
	   数据库错误：14
	   RMQ错误：15

	   应用标识(2位数字)
	   公共：00
	   admin：01
	   integrateshop-api：02
	   integrateshop-rpc：03
	   userbehavior-api：04
	   userbehavior-rpc：05
	   cube-api: 06
	   cube-rpc: 07

	   错误编码(3位数字)
	*/

	SystemDefault            = 1000000
	SystemServerMaintenance  = 1000001 //服务器维护中
	SystemInvalidRequest     = 1000002 //非法请求
	SystemUserAuth           = 1000003 //用户校验错误
	SystemUserAuthSign       = 1000004 //用户校验签名错误
	SystemUserAuthParam      = 1000005 //用户校验参数错误
	SystemUserAuthChangeType = 1000006 //用户校验类型转换错误
	SystemFrequentRequest    = 1000007 //请求频繁
	SystemParamFormat        = 1000008 //请求参数格式错误
	SystemServiceUnavailable = 1000009 //服务不可用
	SystemNoObject           = 1000010 //访问对象不存在
	SystemUnauthorized       = 1000011 //未经授权的访问
	SystemNoLogin            = 1000012 //用户未登陆
	SystemGetUinFromContext  = 1000013 // 从Context 获取uin 失败

	UserBehaviorCommentMarshalRPC   = 1005001 // 用户行为RPC 评论 json marshal 错误
	UserBehaviorCommentUnMarshalRPC = 1005002 // 用户行为RPC 评论 json unmarshal 错误

	ParamDefault                      = 1100000
	ParamUserBehaviorApi              = 1104001 //用户行为API参数错误
	ParamUserBehaviorAddCommentFailed = 1104002 //用户行为添加评论失败
	ParamUserBehaviorRpc              = 1105001 //用户行为RPC参数错误

	GetDataDefault         = 1200000
	GetDataUserBehaviorApi = 1204001 //用户行为API获取数据错误
	GetDataUserBehaviorRpc = 1205001 //用户行为RPC获取数据错误

	GetCacheDataUserBehaviorRpc = 1305001 // 用户行为RPC 获取缓存数据

	UserBehaviorDbGet    = 1405001 // 用户行为rpc获取数据错误
	UserBehaviorDbInsert = 1405002 // 用户行为rpc写入数据库错误
	UserBehaviorDbUpdate = 1405003 // 用户行为rpc更新数据库错误

	UserBehaviorRMqSend = 1505001 // 用户行为rpc 发送rmq 错误
	UserBehaviorRMQGet  = 1505002 // 用户行为rpc 获取 rmq 错误

	ParamCubeIdError     = 1106000 // 问卷id错误
	UQCDbInsert          = 1407000 // 用户问卷选项
	UPCUpsert            = 1407001 // 用户弹窗上报
	PopupConfigError     = 1307000 // 获取缓存错误
	PopupConfigNotExists = 1307001 // 数据不存在
)
