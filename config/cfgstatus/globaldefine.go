package cfgstatus

import (
	"database/sql"
)

// 用户行为相关 - 主体类型定义
const (
	UserBehaviorWorkType    = 1
	UserBehaviorProjectType = 2
	UserBehaviorCommentType = 3
)

var UserBehaviorTypeMap = map[int]string{
	UserBehaviorWorkType:    "作品",
	UserBehaviorProjectType: "专题",
	UserBehaviorCommentType: "评论",
}

// 用户行为可评论的类型
var UserBehaviorCanCommentTopicType = map[int]string{
	UserBehaviorWorkType:    "作品",
	UserBehaviorProjectType: "专题",
}

// 评论审核错误信息
var CommentReviewErrMsg = map[int]string{
	1: "参数错误",
	2: "验证码错误",
	3: "ww错误",
	4: "网络错误",
	5: "服务器错误",
	6: "内容不合格",
}

const (
	CommentContentMaxLength = 100
)

// 发消息类型常量
const (
	SendMessageTypeFocus               = 21 // 21关注
	SendMessageTypeComment             = 22 // 22评论
	SendMessageTypeCommentReviewFailed = 23 // 23评论未通过
	SendMessageTypeCommentReviewOk     = 24 // 24评论审核通过

	SendMessageNextTypeHome   = 0 // 打开首页
	SendMessageNextTypeLink   = 1 // link跳转
	SendMessageNextTypeScheme = 2 // scheme 跳转

	SendMessageUrlMsgCenter = "minicode://main_activity/?page=newMsgCenter" // scheme消息中心地址
)

// 评论封禁原因map
var CommentBanReasonMap = map[int]string{
	0: "",
	1: "色情内容",
	2: "政治敏感",
	3: "暴力恐怖",
	4: "广告",
	5: "欺诈",
	6: "价值观",
}

// 用于生成缓存的member，rpc 与 crontab 共用
// 使用 time.Time 会有时区问题，数据库和 redis 存储的时间格式会有问题，导致更新的时候会有异常
type CommentRInfo struct {
	Id        int64          `db:"id"`
	AppId     int64          `db:"app_id"`     // app
	TopicId   string         `db:"topic_id"`   // 主题id
	TopicType int64          `db:"topic_type"` // 主题类型
	Content   sql.NullString `db:"content"`    // 评论内容
	Uin       string         `db:"uin"`        // 评论用户id
	//CreatedAt        time.Time      `db:"created_at"`
	CreatedTs int64 `db:"created_ts"`
}

// app 一级页tab类型
var TabMap = map[int64]string{
	1: "首页",
	2: "课程",
	3: "我的",
}
