package cfgredis

import "time"

const (
	Expiration60M  = 60 * time.Minute
	ExpirationTenM = 10 * time.Minute
	ExpirationDay  = 24 * time.Hour

	// UserAssets 用户积分
	UserAssets = "sir@userAssets:%s:coin"

	// IntegrateUserInfo 积分商城用户信息
	IntegrateUserInfo = "sir@IntegrateUserInfo:%s"

	// IntegrateUserGoodsInfo 积分商城用户兑换商品信息
	IntegrateUserGoodsInfo = "sir@IntegrateUserGoodsInfo:%s:%d"

	// IntegrateGoodsInfo 积分商城单个商品信息
	IntegrateGoodsInfo = "sir@IntegrateGoodsInfo:%d"

	// IntegrateGoodsList 积分商城首页商品列表
	IntegrateGoodsList = "sir@IntegrateIndexGoodsList:%d"

	// IntegrateChangeCoinLock 积分变动锁key
	IntegrateChangeCoinLock = "sir@IntegrateChangeCoinLock:%s"

	// 反馈
	FeiShuFeedbackTokenKey = "sir@feiShuFeedbackAppToken" // 飞书反馈app token ==> string
	FeiShuFeedbackUserKey  = "sir@feiShuFeedbackAppUser"  // 飞书反馈app user  ===> string

	// =========================================== 用户行为 start ================================================

	// UserBehaviorCommentNum 用户行为，评论数
	UserBehaviorCommentNum = "sir@UserCommentNum:%s"

	// UserBehaviorPraiseNum 用户行为，点赞数
	UserBehaviorPraiseNum = "sir@UserPraiseNum:%s"

	// UserBehaviorShareNum 用户行为，分享数
	UserBehaviorShareNum = "sir@UserShareNum:%s"

	//// UserBehaviorFollowNum 用户行为，粉丝数
	//UserBehaviorFollowNum = "sir@UserFollowNum:%s"
	//// UserBehaviorFocusNum 用户行为，关注数
	//UserBehaviorFocusNum = "sir@UserFocusNum:%s"

	// UserBehaviorCountNumShow 用户行为，展示粉丝关注数量
	UserBehaviorCountNumShow            = "sir@UserCountNumShow:%s"
	UserBehaviorCountNumShowFieldFollow = "follow"
	UserBehaviorCountNumShowFieldFocus  = "focus"

	// UserBehaviorProduceCountShow 用户行为，展示分享点赞评论数量
	UserBehaviorProduceCountShow             = "sir@ProduceCountNumShow:%s"
	UserBehaviorProduceCountShowFieldComment = "comment"
	UserBehaviorProduceCountShowFieldPraise  = "praise"
	UserBehaviorProduceCountShowFieldShare   = "share"

	UserBehaviorCommentLatest       = "sir@commentLatest:%d:%d:%s"  // 用户行为 最新评论 app_id:topic_type:topic_id
	UserBehaviorCommentHot          = "sir@commentHot:%d:%d:%s:%d"  // 用户行为 最热评论 app_id:topic_type:topic_id:Hour
	UserBehaviorCommentCount        = "sir@commentCount:%d:%d:%s"   // 用户行为 评论数 app_id:topic_type:topic_id
	UserBehaviorUserUnReviewComment = "sir@commentUURC:%s:%d:%d:%s" // 用户行为 用户未被审核评论 uin:app_id:topic_type:topic_id
	// 用户未被审核评论数 app_id:topic_type:topic_id ，防止key 比数据还大，故用缩写
	UserBehaviorUserUnReviewCommentCount = "sir@commentUURCC:%s:%d:%d:%s" //UserUnReviewCommentCount 缩写 UURC 用户行为

	// UserBehaviorIsSendFocusMsg 用户行为，是否发送关注消息  uin:focusUin
	UserBehaviorIsSendFocusMsg = "sir@IsSendFocusMsg:%s:%s"
	// =========================================== 用户行为 end ================================================

	// 弹窗配置
	PopupConfigs         = "sir@popupConfigs"
	UserPopupCountPerDay = "sir@UserPopupCountPerday:%s:%s"
	AccountInfo          = "sir@account:%s"               // 用户信息
	UserAssetsVip        = "sir@userAssets:%s:vip"        // 用户vip
	UserOrderTheme       = "sir@userOrderTheme"           // 用户体验卡
	UserCourse           = "sir@userCourse:4:%s"          // 用户专项课
	UserTaskProgress     = "sir@user:task:progress:%s:%s" // 课时任务完成进度锁
	// 基础key
	UserInfo    = "sir@userinfo:%s"    // 用户基本信息key
	WorksInfo   = "sir@worksinfo:%s"   // 作品信息key
	SpecialInfo = "sir@specialinfo:%s" // 专题信息key
)
