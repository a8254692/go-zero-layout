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

type GetPraiseListLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewGetPraiseListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPraiseListLogic {
    return &GetPraiseListLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 获取点赞列表
func (l *GetPraiseListLogic) GetPraiseList(in *userBehaviorProto.GetPraiseListReq) (*userBehaviorProto.GetPraiseListResp, error) {
    reqByte, _ := json.Marshal(in)
    reqStr := string(reqByte)

    uin := in.TargetUin
    if uin == "" {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("目标UIN校验失败")
        return nil, errors.New("目标UIN校验失败")
    }

    limit, offset := help.GetPagingParam(in.PageIndex, in.PageSize)
    dbList, err := l.svcCtx.UserPraiseModel.FindUinList(uin, limit, offset)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取点赞列表失败")
        return nil, err
    }

    var list []*userBehaviorProto.GetPraiseListData
    var total int64

    if len(*dbList) > 0 {
        for _, v := range *dbList {
            list = append(list, &userBehaviorProto.GetPraiseListData{
                TopicType: v.TopicType,
                TopicId:   v.TopicId,
            })
        }

        dbCount, err := l.svcCtx.UserPraiseModel.FindUinCount(uin)
        if err != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取点赞列表总数失败")
            return nil, err
        }

        total = dbCount
    }

    //filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Info("=========这是获取点赞列表的返回值和Total========", list)

    return &userBehaviorProto.GetPraiseListResp{
        List:  list,
        Total: total,
    }, nil
}
