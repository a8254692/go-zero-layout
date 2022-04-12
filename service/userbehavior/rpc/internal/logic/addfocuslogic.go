package logic

import (
    "context"
    "encoding/json"
    "errors"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "github.com/zeromicro/go-zero/core/logx"

    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"
    "minicode.com/sirius/go-back-server/utils/help"
)

type AddFocusLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewAddFocusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddFocusLogic {
    return &AddFocusLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

//消息队列关注详情参数
type AddFocusReq struct {
    OpType   int64  `json:"opType"`
    AppId    int64  `json:"appId"`
    Uin      string `json:"uin"`
    FocusUin string `json:"focusUin"`
}

// 新增关注详情
func (l *AddFocusLogic) AddFocus(in *userBehaviorProto.AddFocusReq) (*userBehaviorProto.AddFocusResp, error) {
    reqByte, _ := json.Marshal(in)
    reqStr := string(reqByte)

    uin, err := help.GetRpcUinFromCtx(l.ctx)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取uin失败")
        return nil, errors.New("获取uin失败")
    }

    if in.FocusUin == "" {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("校验被关注者UIN失败")
        return nil, errors.New("校验被关注者UIN失败")
    }

    msg, err := json.Marshal(AddFocusReq{
        OpType:   cfgstatus.UserBehaviorRmqFocusType,
        AppId:    in.AppId,
        Uin:      uin,
        FocusUin: in.FocusUin,
    })
    err = l.svcCtx.AddFocusRmqQConn.PublishSimple(string(msg))
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": uin, "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("添加关注rmq发送消息失败")
        return nil, errors.New("添加关注rmq发送消息失败")
    }

    //status := cfgstatus.UserBehaviorOneFocus
    ////先查询是否是对方已关注
    //bothWayDbInfo, err := l.svcCtx.UserFocusModel.FindOneByUinFocusUin(in.FocusUin, uin)
    //if err != nil {
    //	filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //	l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取对方是否关注失败")
    //	return nil, errors.New("获取对方是否关注失败")
    //}
    //
    //if bothWayDbInfo != nil {
    //	if bothWayDbInfo.Status > cfgstatus.UserBehaviorCanNotFocus {
    //		status = cfgstatus.UserBehaviorMutuallyFocus
    //	}
    //}
    //
    //_, err = l.svcCtx.UserFocusModel.Insert(&userfocus.UserFocus{
    //	AppId:    in.AppId,
    //	Uin:      uin,
    //	FocusUin: in.FocusUin,
    //	Status:   int64(status),
    //})
    //if err != nil {
    //	filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //	l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("新增关注详情失败")
    //	return nil, err
    //}
    //
    ////双向关注则更新下对方的状态
    //if bothWayDbInfo.Status == cfgstatus.UserBehaviorOneFocus {
    //	err = l.svcCtx.UserFocusModel.UpdateStatus(in.FocusUin, uin, cfgstatus.UserBehaviorMutuallyFocus)
    //	if err != nil {
    //		filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //		l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("双向关注则更新下对方的状态失败")
    //		return nil, err
    //	}
    //}
    //
    //isSendKey := fmt.Sprintf(cfgredis.UserBehaviorIsSendFocusMsg, uin, in.FocusUin)
    //redisIsSet := l.svcCtx.Redis.SetNX(isSendKey, 1, cfgredis.ExpirationDay).Val()
    //if redisIsSet {
    //	//发送关注消息
    //	sendDataFrom := message.SendMessageRmqQFrom{
    //		AuthorId: uin,
    //	}
    //	sendDataExtra := message.SendMessageRmqQExtra{}
    //	sendData := message.SendMessageRmqQ{
    //		Type:    cfgstatus.SendMessageTypeFocus,
    //		Uin:     in.FocusUin,
    //		Content: cfgmsg.FocusMsg,
    //		From:    sendDataFrom,
    //		Extra:   sendDataExtra,
    //	}
    //	err := message.SendMessageUserCenter(l.svcCtx.SendMessageRmqQConn, sendData)
    //	if err != nil {
    //		filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //		l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("发送关注消息失败")
    //		return nil, err
    //	}
    //}

    return &userBehaviorProto.AddFocusResp{}, nil
}
