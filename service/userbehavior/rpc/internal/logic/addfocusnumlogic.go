package logic

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/zeromicro/go-zero/core/logx"
    "minicode.com/sirius/go-back-server/config/cfgredis"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"
    "minicode.com/sirius/go-back-server/utils/help"
)

type AddFocusNumLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewAddFocusNumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddFocusNumLogic {
    return &AddFocusNumLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 新增关注计数
func (l *AddFocusNumLogic) AddFocusNum(in *userBehaviorProto.AddFocusNumReq) (*userBehaviorProto.AddFocusNumResp, error) {
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

    err = l.svcCtx.UserCountModel.UpdateNumIncr(uin, 1, int8(in.Type))
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("编辑关注数量失败")
        return nil, errors.New("编辑关注数量失败")
    }

    countNumShowKey := fmt.Sprintf(cfgredis.UserBehaviorCountNumShow, uin)
    l.svcCtx.Redis.Del(countNumShowKey)

    //today := time.Now().Format("20060102")
    //userFocusNumKey := fmt.Sprintf(cfgredis.UserBehaviorFocusNum, today)
    //
    ////先查数据库中是否存在数据
    //existsRsV := l.svcCtx.Redis.HExists(userFocusNumKey, uin).Val()
    //if !existsRsV {
    //    dbInfo, err := l.svcCtx.UserCountModel.FindOneByUin(uin)
    //    if err != nil || dbInfo == nil {
    //        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取DB关注计数失败")
    //    } else {
    //        if dbInfo.FocusNum > 0 {
    //            err = l.svcCtx.Redis.HSet(userFocusNumKey, uin, dbInfo.FocusNum).Err()
    //            if err != nil {
    //                filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //                l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("设置DB关注计数失败")
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
    //err = l.svcCtx.Redis.HIncrBy(userFocusNumKey, uin, int64(incr)).Err()
    //if err != nil {
    //    filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
    //    l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("新增关注计数失败")
    //    return nil, errors.New("新增关注计数失败")
    //}
    //
    //l.svcCtx.Redis.Expire(userFocusNumKey, help.GetTodayTimeRemaining())

    return &userBehaviorProto.AddFocusNumResp{}, nil
}
