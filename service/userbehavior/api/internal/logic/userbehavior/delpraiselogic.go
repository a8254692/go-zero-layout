package userbehavior

import (
    "context"
    "encoding/json"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/rpcuserbehavior"
    "minicode.com/sirius/go-back-server/utils/errorx"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "minicode.com/sirius/go-back-server/service/userbehavior/api/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/api/internal/types"

    "github.com/zeromicro/go-zero/core/logx"
)

type DelPraiseLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewDelPraiseLogic(ctx context.Context, svcCtx *svc.ServiceContext) DelPraiseLogic {
    return DelPraiseLogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *DelPraiseLogic) DelPraise(req types.DelPraiseReq) (resp *types.DelPraiseResp, err error) {
    reqByte, _ := json.Marshal(req)
    reqStr := string(reqByte)

    if req.TopicId == "" {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("参数校验失败")
        return nil, errorx.NewDefaultError("参数校验失败")
    }

    //增加点赞
    _, err = l.svcCtx.UserBehaviorRpc.DelPraise(l.ctx, &rpcuserbehavior.DelPraiseReq{
        AppId:     req.AppId,
        TopicType: int32(req.TopicType),
        TopicId:   req.TopicId,
    })
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("取消点赞失败")
        return nil, errorx.NewDefaultError("取消点赞失败")
    }

    //调用点赞计数方法
    _, err = l.svcCtx.UserBehaviorRpc.AddPraiseNum(l.ctx, &rpcuserbehavior.AddPraiseNumReq{
        AppId:     req.AppId,
        TopicType: int32(req.TopicType),
        TopicId:   req.TopicId,
        Type:      cfgstatus.UserBehaviorOperationReduceType,
    })
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("取消点赞计数失败")
    }

    return
}
