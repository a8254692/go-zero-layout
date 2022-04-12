package logic

import (
    "context"
    "encoding/json"
    "errors"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"
    "minicode.com/sirius/go-back-server/utils/help"

    "github.com/zeromicro/go-zero/core/logx"
)

type GetUserIsPraiseLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewGetUserIsPraiseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserIsPraiseLogic {
    return &GetUserIsPraiseLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 获取用户是否点赞
func (l *GetUserIsPraiseLogic) GetUserIsPraise(in *userBehaviorProto.GetUserIsPraiseReq) (*userBehaviorProto.GetUserIsPraiseResp, error) {
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

    dbInfo, err := l.svcCtx.UserPraiseModel.FindOneByParam(uin, in.TopicType, in.TopicId)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取用户是否点赞失败")
        return nil, err
    }

    isRs := false
    if dbInfo.Id > 0 {
        isRs = true
    }

    return &userBehaviorProto.GetUserIsPraiseResp{
        IsPraise: isRs,
    }, nil
}
