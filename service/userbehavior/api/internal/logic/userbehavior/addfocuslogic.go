package userbehavior

import (
    "context"
    "encoding/json"
    "minicode.com/sirius/go-back-server/utils/help"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "minicode.com/sirius/go-back-server/service/userbehavior/api/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/api/internal/types"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/rpcuserbehavior"
    "minicode.com/sirius/go-back-server/utils/errorx"

    "github.com/zeromicro/go-zero/core/logx"
)

type AddFocusLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewAddFocusLogic(ctx context.Context, svcCtx *svc.ServiceContext) AddFocusLogic {
    return AddFocusLogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *AddFocusLogic) AddFocus(req types.AddFocusReq) (resp *types.AddFocusResp, err error) {
    reqByte, _ := json.Marshal(req)
    reqStr := string(reqByte)

    uin, err := help.GetUinFromCtx(l.ctx)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取关Uin失败")
        return nil, errorx.NewDefaultError("获取关Uin失败")
    }

    if req.FocusUin == "" {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("参数校验失败")
        return nil, errorx.NewDefaultError("参数校验失败")
    }

    if uin == req.FocusUin {
        return nil, errorx.NewDefaultError("参数校验失败")
    }

    //增加关注关系
    _, err = l.svcCtx.UserBehaviorRpc.AddFocus(l.ctx, &rpcuserbehavior.AddFocusReq{
        AppId:    req.AppId,
        FocusUin: req.FocusUin,
    })
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("关注失败")
        return nil, errorx.NewDefaultError("关注失败")
    }

    ////调用关注计数方法
    //_, err = l.svcCtx.UserBehaviorRpc.AddFocusNum(l.ctx, &rpcuserbehavior.AddFocusNumReq{
    //    AppId:    req.AppId,
    //    FocusUin: req.FocusUin,
    //    Type:     cfgstatus.UserBehaviorOperationAddType,
    //})
    //if err != nil {
    //    filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //    l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("增加关注计数失败")
    //}
    //
    ////调用粉丝计数方法
    //_, err = l.svcCtx.UserBehaviorRpc.AddFollowNum(l.ctx, &rpcuserbehavior.AddFollowNumReq{
    //    AppId:    req.AppId,
    //    FocusUin: req.FocusUin,
    //    Type:     cfgstatus.UserBehaviorOperationAddType,
    //})
    //if err != nil {
    //    filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //    l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("增加关注计数失败")
    //}

    return
}
