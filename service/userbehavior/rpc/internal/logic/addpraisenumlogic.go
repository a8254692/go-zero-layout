package logic

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "minicode.com/sirius/go-back-server/utils/mylogrus"
    "time"

    "github.com/zeromicro/go-zero/core/logx"

    "minicode.com/sirius/go-back-server/config/cfgredis"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"
    "minicode.com/sirius/go-back-server/utils/help"
)

type AddPraiseNumLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewAddPraiseNumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddPraiseNumLogic {
    return &AddPraiseNumLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 新增点赞计数
func (l *AddPraiseNumLogic) AddPraiseNum(in *userBehaviorProto.AddPraiseNumReq) (*userBehaviorProto.AddPraiseNumResp, error) {
    reqByte, _ := json.Marshal(in)
    reqStr := string(reqByte)

    if in.TopicId == "" {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("校验TopicId失败")
        return nil, errors.New("校验TopicId失败")
    }

    appIdStr := fmt.Sprintf("%d", in.AppId)
    topicTypeStr := fmt.Sprintf("%d", in.TopicType)

    today := time.Now().Format("20060102")
    field := appIdStr + "|" + topicTypeStr + "|" + in.TopicId
    userPraiseNumKey := fmt.Sprintf(cfgredis.UserBehaviorPraiseNum, today)

    //先查数据库中是否存在数据
    existsRsV := l.svcCtx.Redis.HExists(userPraiseNumKey, field).Val()
    if !existsRsV {
        dbInfo, err := l.svcCtx.ProduceCountModel.FindOneByParam(int64(in.TopicType), in.TopicId)
        if err != nil {

            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取DB点赞计数失败")
        } else {
            if dbInfo.PraiseNum > 0 {
                err = l.svcCtx.Redis.HSet(userPraiseNumKey, field, dbInfo.PraiseNum).Err()
                if err != nil {
                    filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
                    l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("设置DB点赞计数失败")
                }
            }
        }
    }

    incr := 1
    if in.Type == cfgstatus.UserBehaviorOperationReduceType {
        incr = -1
    }

    err := l.svcCtx.Redis.HIncrBy(userPraiseNumKey, field, int64(incr)).Err()
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("新增点赞计数失败")
        return nil, errors.New("新增点赞计数失败")
    }

    l.svcCtx.Redis.Expire(userPraiseNumKey, help.GetTodayTimeRemaining())

    return &userBehaviorProto.AddPraiseNumResp{}, nil
}
