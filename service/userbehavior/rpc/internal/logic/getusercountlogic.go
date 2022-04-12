package logic

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "minicode.com/sirius/go-back-server/utils/mylogrus"
    "strconv"

    "github.com/zeromicro/go-zero/core/logx"

    "minicode.com/sirius/go-back-server/config/cfgredis"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"
)

type GetUserCountLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewGetUserCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserCountLogic {
    return &GetUserCountLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 获取粉丝关注数量
func (l *GetUserCountLogic) GetUserCount(in *userBehaviorProto.GetUserCountReq) (*userBehaviorProto.GetUserCountResp, error) {
    reqByte, _ := json.Marshal(in)
    reqStr := string(reqByte)

    uin := in.TargetUin
    if uin == "" {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("目标UIN校验失败")
        return nil, errors.New("目标UIN校验失败")
    }

    var followNum int64
    var focusNum int64

    userFollowNumKey := fmt.Sprintf(cfgredis.UserBehaviorCountNumShow, uin)

    redisExistsCmd := l.svcCtx.Redis.Exists(userFollowNumKey)
    redisExistsV, err := redisExistsCmd.Result()
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("redisExists链接失败")
        return nil, errors.New("redis链接失败")
    }

    if redisExistsV <= 0 {
        dbInfo, err := l.svcCtx.UserCountModel.FindOneByUin(uin)
        if err != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("db查询计数信息失败")
            return nil, errors.New("db查询计数信息失败")
        }

        l.svcCtx.Redis.HSet(userFollowNumKey, cfgredis.UserBehaviorCountNumShowFieldFollow, dbInfo.FollowNum)
        l.svcCtx.Redis.HSet(userFollowNumKey, cfgredis.UserBehaviorCountNumShowFieldFocus, dbInfo.FocusNum)
        l.svcCtx.Redis.Expire(userFollowNumKey, cfgredis.ExpirationTenM)

        followNum = dbInfo.FollowNum
        focusNum = dbInfo.FocusNum
    } else {
        redisShowCountCmd := l.svcCtx.Redis.HGetAll(userFollowNumKey)
        redisV, err := redisShowCountCmd.Result()
        if err != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("redis链接失败")
            return nil, errors.New("redis链接失败")
        }

        followNum, _ = strconv.ParseInt(redisV[cfgredis.UserBehaviorCountNumShowFieldFollow], 10, 64)
        focusNum, _ = strconv.ParseInt(redisV[cfgredis.UserBehaviorCountNumShowFieldFocus], 10, 64)
    }

    return &userBehaviorProto.GetUserCountResp{
        FollowNum: followNum,
        FocusNum:  focusNum,
    }, nil
}
