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

type DelFocusLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewDelFocusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DelFocusLogic {
    return &DelFocusLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 取消关注详情
func (l *DelFocusLogic) DelFocus(in *userBehaviorProto.DelFocusReq) (*userBehaviorProto.DelFocusResp, error) {
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

    if uin == in.FocusUin {
        return nil, errors.New("参数校验失败")
    }

    msg, err := json.Marshal(AddFocusReq{
        OpType:   cfgstatus.UserBehaviorRmqCancelFocusType,
        AppId:    in.AppId,
        Uin:      uin,
        FocusUin: in.FocusUin,
    })
    err = l.svcCtx.AddFocusRmqQConn.PublishSimple(string(msg))
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": uin, "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("取消关注rmq发送消息失败")
        return nil, errors.New("取消关注rmq发送消息失败")
    }

    ////先查询是否是对方已关注
    //bothWayDbInfo, err := l.svcCtx.UserFocusModel.FindOneByUinFocusUin(in.FocusUin, uin)
    //if err != nil {
    //    filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //    l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取对方是否关注失败")
    //    return nil, errors.New("获取对方是否关注失败")
    //}
    ////如果对方是双向关注则更新对方状态
    //if bothWayDbInfo.Status == cfgstatus.UserBehaviorMutuallyFocus {
    //    err := l.svcCtx.UserFocusModel.UpdateStatus(in.FocusUin, uin, cfgstatus.UserBehaviorOneFocus)
    //    if err != nil {
    //        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("更新对方状态失败")
    //        return nil, err
    //    }
    //}
    //
    //err = l.svcCtx.UserFocusModel.DeleteByUinFocusUin(uin, in.FocusUin)
    //if err != nil {
    //    filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //    l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("取消关注详情失败")
    //    return nil, err
    //}

    return &userBehaviorProto.DelFocusResp{}, nil
}
