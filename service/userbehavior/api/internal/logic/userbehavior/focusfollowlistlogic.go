package userbehavior

import (
    "context"
    "encoding/json"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "github.com/zeromicro/go-zero/core/logx"

    "minicode.com/sirius/go-back-server/service/userbehavior/api/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/api/internal/types"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/rpcuserbehavior"
    "minicode.com/sirius/go-back-server/utils/errorx"
)

type FocusFollowListLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewFocusFollowListLogic(ctx context.Context, svcCtx *svc.ServiceContext) FocusFollowListLogic {
    return FocusFollowListLogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *FocusFollowListLogic) FocusFollowList(req types.FocusFollowListReq) (resp *types.FocusFollowListResp, err error) {
    reqByte, _ := json.Marshal(req)
    reqStr := string(reqByte)

    if req.TargetUin == "" {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("参数TargetUin无效")
        return nil, errorx.NewDefaultError("参数校验失败")
    }

    list, err := l.svcCtx.UserBehaviorRpc.GetFocusFollowList(l.ctx, &rpcuserbehavior.GetFocusFollowListReq{
        AppId:     req.AppId,
        PageIndex: req.PageIndex,
        PageSize:  req.PageSize,
        Type:      int32(req.Type),
        TargetUin: req.TargetUin,
    })
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取关注/粉丝列表失败")
        return nil, errorx.NewDefaultError("获取关注/粉丝列表失败")
    }

    var rsList []types.FocusFollowListData
    if len(list.List) > 0 {
        for _, v := range list.List {
            rsList = append(rsList, types.FocusFollowListData{
                Id:          v.Id,
                UIn:         v.UIn,
                UserName:    v.UserName,
                UserHeadUrl: v.UserHeadUrl,
                Status:      v.Status,
            })
        }
    }

    return &types.FocusFollowListResp{
        List:  rsList,
        Total: list.Total,
    }, nil
}
