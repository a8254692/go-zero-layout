package logic

import (
    "context"
    "encoding/json"
    "errors"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "github.com/zeromicro/go-zero/core/logx"

    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"
    "minicode.com/sirius/go-back-server/utils/help"
)

type DelPraiseLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewDelPraiseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DelPraiseLogic {
    return &DelPraiseLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 取消点赞详情
func (l *DelPraiseLogic) DelPraise(in *userBehaviorProto.DelPraiseReq) (*userBehaviorProto.DelPraiseResp, error) {
    reqByte, _ := json.Marshal(in)
    reqStr := string(reqByte)

    uin, err := help.GetRpcUinFromCtx(l.ctx)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取uin失败")
        return nil, errors.New("获取uin失败")
    }

    if in.TopicId == "" {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("校验主题ID失败")
        return nil, errors.New("校验主题ID失败")
    }

    err = l.svcCtx.UserPraiseModel.DeleteByParam(uin, int64(in.TopicType), in.TopicId)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("取消点赞详情失败")
        return nil, err
    }

    return &userBehaviorProto.DelPraiseResp{}, nil
}
