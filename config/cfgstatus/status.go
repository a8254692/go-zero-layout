package cfgstatus

const (
    // CoinGoodsDown 积分商城商品下架
    CoinGoodsDown = 1
    // CoinGoodsPutAway 积分商城商品上架
    CoinGoodsPutAway = 2
    // CoinGoodsExchangeIncome 积分商城积分收入
    CoinGoodsExchangeIncome = 1
    // CoinGoodsExchangeExpend 积分商城积分支出
    CoinGoodsExchangeExpend = 2
    // CoinGoodsTypeEntity 积分商品实物礼品
    CoinGoodsTypeEntity = 1
    // CoinGoodsTypeVirtual 积分商品虚拟物品
    CoinGoodsTypeVirtual = 2
    //CoinGoodsShowStatusIsOver 已兑完
    CoinGoodsShowStatusIsOver = 1
    //CoinGoodsShowStatusShortStock 库存紧张
    CoinGoodsShowStatusShortStock = 2
    //CoinGoodsSendVirtualMsgType RMQ消息队列发放物品消息
    CoinGoodsSendVirtualMsgType = 19
    //CoinGoodsSendVirtualPortrait RMQ消息队列发放虚拟物品ID
    CoinGoodsSendVirtualPortrait = 4

    //UserBehaviorCanNotFocus 用户行为，无法关注
    UserBehaviorCanNotFocus = 0
    //UserBehaviorOneFocus 用户行为，用户单向关注
    UserBehaviorOneFocus = 1
    //UserBehaviorMutuallyFocus 用户行为，用户互相关注
    UserBehaviorMutuallyFocus = 2
    //UserBehaviorNoFocus 用户行为，双方都没关注
    UserBehaviorNoFocus = 3
    //UserBehaviorOnlyMyFocus 用户行为，仅我关注了ta
    UserBehaviorOnlyMyFocus = 4
    //UserBehaviorOnlyHeFocus 用户行为，仅ta关注了我
    UserBehaviorOnlyHeFocus = 5

    //UserBehaviorOperationAddType 用户行为，接口新增操作类型
    UserBehaviorOperationAddType = 1
    //UserBehaviorOperationReduceType 用户行为，接口减少操作类型
    UserBehaviorOperationReduceType = 2
    //UserBehaviorFocusListFocusType 用户行为，关注/粉丝列表中关注列表
    UserBehaviorFocusListFocusType = 1
    //UserBehaviorFocusListFollowType 用户行为，关注/粉丝列表中粉丝列表
    UserBehaviorFocusListFollowType = 2

    //UserBehaviorRmqFocusType 用户行为，关注
    UserBehaviorRmqFocusType = 1
    //UserBehaviorRmqCancelFocusType 用户行为，取消关注
    UserBehaviorRmqCancelFocusType = 2
)
