package logic

import (
    "context"
    "encoding/json"
    "errors"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "github.com/zeromicro/go-zero/core/logx"

    "minicode.com/sirius/go-back-server/service/userbehavior/model/userpraise"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"
    "minicode.com/sirius/go-back-server/utils/help"
)

type AddPraiseLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewAddPraiseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddPraiseLogic {
    return &AddPraiseLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 新增点赞详情
func (l *AddPraiseLogic) AddPraise(in *userBehaviorProto.AddPraiseReq) (*userBehaviorProto.AddPraiseResp, error) {
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

    _, err = l.svcCtx.UserPraiseModel.Insert(&userpraise.UserPraise{
        AppId:     in.AppId,
        TopicId:   in.TopicId,
        TopicType: int64(in.TopicType),
        Uin:       uin,
    })
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("新增点赞详情失败")
        return nil, err
    }

    return &userBehaviorProto.AddPraiseResp{}, nil
}
