package logic

import (
    "context"
    "encoding/json"
    "errors"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"

    "github.com/zeromicro/go-zero/core/logx"
)

type GetProduceCountsLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewGetProduceCountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProduceCountsLogic {
    return &GetProduceCountsLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 批量获取作品点赞评论数量
func (l *GetProduceCountsLogic) GetProduceCounts(in *userBehaviorProto.GetProduceCountsReq) (*userBehaviorProto.GetProduceCountsResp, error) {
    reqByte, _ := json.Marshal(in)
    reqStr := string(reqByte)

    if len(in.TopicId) <= 0 {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("校验主题ID失败")
        return nil, errors.New("校验主题ID失败")
    }

    dbList, err := l.svcCtx.ProduceCountModel.FindListByParam(in.TopicType, in.TopicId)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("db批量查询作品点赞评论数量失败")
        return nil, errors.New("db批量查询作品点赞评论数量失败")
    }

    var rsList []*userBehaviorProto.GetProduceCountsData
    if len(*dbList) > 0 {
        for _, v := range *dbList {
            info := userBehaviorProto.GetProduceCountsData{
                CommentNum: v.CommentNum,
                PraiseNum:  v.PraiseNum,
                ShareNum:   v.ShareNum,
                TopicId:    v.TopicId,
                TopicType:  v.TopicType,
            }

            rsList = append(rsList, &info)
        }
    }

    return &userBehaviorProto.GetProduceCountsResp{
        List: rsList,
    }, nil
}
