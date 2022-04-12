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

type AddPraiseLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewAddPraiseLogic(ctx context.Context, svcCtx *svc.ServiceContext) AddPraiseLogic {
    return AddPraiseLogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *AddPraiseLogic) AddPraise(req types.AddPraiseReq) (resp *types.AddPraiseResp, err error) {
    reqByte, _ := json.Marshal(req)
    reqStr := string(reqByte)

    if req.TopicId == "" {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("参数校验失败")
        return nil, errorx.NewDefaultError("参数校验失败")
    }

    _, err = l.svcCtx.UserBehaviorRpc.AddPraise(l.ctx, &rpcuserbehavior.AddPraiseReq{
        AppId:     req.AppId,
        TopicType: int32(req.TopicType),
        TopicId:   req.TopicId,
    })
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("点赞失败")
        return nil, errorx.NewDefaultError("点赞失败")
    }

    //调用点赞计数方法
    _, err = l.svcCtx.UserBehaviorRpc.AddPraiseNum(l.ctx, &rpcuserbehavior.AddPraiseNumReq{
        AppId:     req.AppId,
        TopicType: int32(req.TopicType),
        TopicId:   req.TopicId,
        Type:      cfgstatus.UserBehaviorOperationAddType,
    })
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("增加点赞计数失败")
    }

    return
}
