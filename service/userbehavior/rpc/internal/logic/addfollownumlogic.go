package logic

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/zeromicro/go-zero/core/logx"
    "minicode.com/sirius/go-back-server/config/cfgredis"
    "minicode.com/sirius/go-back-server/utils/help"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"
)

type AddFollowNumLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewAddFollowNumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddFollowNumLogic {
    return &AddFollowNumLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 新增粉丝计数
func (l *AddFollowNumLogic) AddFollowNum(in *userBehaviorProto.AddFollowNumReq) (*userBehaviorProto.AddFollowNumResp, error) {
    reqByte, _ := json.Marshal(in)
    reqStr := string(reqByte)

    uin, err := help.GetRpcUinFromCtx(l.ctx)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取uin失败")
        return nil, errors.New("获取uin失败")
    }

    if uin == in.FocusUin {
        return nil, errors.New("参数校验失败")
    }

    err = l.svcCtx.UserCountModel.UpdateNumIncr(in.FocusUin, 2, int8(in.Type))
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("编辑粉丝数量失败")
        return nil, errors.New("编辑粉丝数量失败")
    }

    countNumShowKey := fmt.Sprintf(cfgredis.UserBehaviorCountNumShow, in.FocusUin)
    l.svcCtx.Redis.Del(countNumShowKey)

    //today := time.Now().Format("20060102")
    //userFollowNumKey := fmt.Sprintf(cfgredis.UserBehaviorFollowNum, today)
    //
    ////先查数据库中是否存在数据
    //existsRsV := l.svcCtx.Redis.HExists(userFollowNumKey, in.FocusUin).Val()
    //if !existsRsV {
    //    dbInfo, err := l.svcCtx.UserCountModel.FindOneByUin(in.FocusUin)
    //    if err != nil || dbInfo == nil {
    //        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取DB粉丝计数失败")
    //    } else {
    //        if dbInfo.FollowNum > 0 {
    //            err = l.svcCtx.Redis.HSet(userFollowNumKey, in.FocusUin, dbInfo.FollowNum).Err()
    //            if err != nil {
    //                filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //                l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("设置DB粉丝计数失败")
    //            }
    //        }
    //    }
    //}
    //
    //incr := 1
    //if in.Type == cfgstatus.UserBehaviorOperationReduceType {
    //    incr = -1
    //}
    //
    //err := l.svcCtx.Redis.HIncrBy(userFollowNumKey, in.FocusUin, int64(incr)).Err()
    //if err != nil {
    //    filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //    l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("新增粉丝计数失败")
    //    return nil, errors.New("新增粉丝计数失败")
    //}
    //
    //l.svcCtx.Redis.Expire(userFollowNumKey, help.GetTodayTimeRemaining())

    return &userBehaviorProto.AddFollowNumResp{}, nil
}
