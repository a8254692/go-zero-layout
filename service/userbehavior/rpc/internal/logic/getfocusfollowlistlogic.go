package logic

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "github.com/zeromicro/go-zero/core/logx"

    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/common"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"
    "minicode.com/sirius/go-back-server/utils/help"
)

type GetFocusFollowListLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewGetFocusFollowListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFocusFollowListLogic {
    return &GetFocusFollowListLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 获取关注/粉丝列表
func (l *GetFocusFollowListLogic) GetFocusFollowList(in *userBehaviorProto.GetFocusFollowListReq) (*userBehaviorProto.GetFocusFollowListResp, error) {
    reqByte, _ := json.Marshal(in)
    reqStr := string(reqByte)

    uin, err := help.GetRpcUinFromCtx(l.ctx)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取uin失败")
        return nil, errors.New("获取uin失败")
    }

    if in.TargetUin == "" {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("校验目标uin失败")
        return nil, errors.New("校验目标uin失败")
    }

    limit, offset := help.GetPagingParam(in.PageIndex, in.PageSize)

    var list []*userBehaviorProto.GetFocusFollowListData
    var total int64

    status := cfgstatus.UserBehaviorCanNotFocus
    var isMyList bool
    if in.TargetUin == uin {
        isMyList = true
    }

    commonUserInfo := common.NewUserInfoCommon(l.ctx, l.svcCtx)

    switch in.Type {
    case cfgstatus.UserBehaviorFocusListFocusType:

        dbList, err := l.svcCtx.UserFocusModel.FindUinFocusList(in.TargetUin, limit, offset)
        if err != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取关注/粉丝列表失败")
            return nil, err
        }

        if len(*dbList) > 0 {
            for _, v := range *dbList {
                if isMyList {
                    status = int(v.Status)
                }

                userInfo, err := commonUserInfo.GetUserInfoById(v.FocusUin)
                if err != nil {
                    filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
                    l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取用户信息失败")
                }
                var nickName string
                var avatarId string
                if userInfo != nil {
                    nickName = userInfo.NickName
                    avatarId = fmt.Sprintf("%d", userInfo.AvatarId)
                }

                list = append(list, &userBehaviorProto.GetFocusFollowListData{
                    Id:          v.Id,
                    UIn:         v.FocusUin,
                    UserName:    nickName,
                    UserHeadUrl: avatarId,
                    Status:      int32(status),
                })
            }

            dbCount, err := l.svcCtx.UserFocusModel.FindUinFocusCount(uin)
            if err != nil {
                filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
                l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取关注/粉丝列表总数失败")
                return nil, err
            }

            total = dbCount
        }

    case cfgstatus.UserBehaviorFocusListFollowType:

        dbList, err := l.svcCtx.UserFocusModel.FindUinFollowList(in.TargetUin, limit, offset)
        if err != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取关注/粉丝列表失败")
            return nil, err
        }

        if len(*dbList) > 0 {
            for _, v := range *dbList {
                if isMyList {
                    status = int(v.Status)
                }

                userInfo, err := commonUserInfo.GetUserInfoById(v.Uin)
                if err != nil {
                    filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
                    l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取用户信息失败")
                }
                var nickName string
                var avatarId string
                if userInfo != nil {
                    nickName = userInfo.NickName
                    avatarId = fmt.Sprintf("%d", userInfo.AvatarId)
                }

                list = append(list, &userBehaviorProto.GetFocusFollowListData{
                    Id:          v.Id,
                    UIn:         v.Uin,
                    UserName:    nickName,
                    UserHeadUrl: avatarId,
                    Status:      int32(status),
                })
            }

            dbCount, err := l.svcCtx.UserFocusModel.FindUinFollowCount(in.TargetUin)
            if err != nil {
                filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
                l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取关注/粉丝列表总数失败")
                return nil, err
            }

            total = dbCount
        }
    }

    return &userBehaviorProto.GetFocusFollowListResp{
        List:  list,
        Total: total,
    }, nil
}
