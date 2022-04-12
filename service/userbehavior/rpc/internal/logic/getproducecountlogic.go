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

type GetProduceCountLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewGetProduceCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProduceCountLogic {
    return &GetProduceCountLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 获取作品点赞评论数量
func (l *GetProduceCountLogic) GetProduceCount(in *userBehaviorProto.GetProduceCountReq) (*userBehaviorProto.GetProduceCountResp, error) {
    reqByte, _ := json.Marshal(in)
    reqStr := string(reqByte)

    if in.TopicId == "" {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("校验主题ID失败")
        return nil, errors.New("校验主题ID失败")
    }

    var commentNum int64
    var praiseNum int64
    var shareNum int64

    appIdStr := fmt.Sprintf("%d", in.AppId)
    topicTypeStr := fmt.Sprintf("%d", in.TopicType)
    field := appIdStr + "|" + topicTypeStr + "|" + in.TopicId
    produceCountKey := fmt.Sprintf(cfgredis.UserBehaviorProduceCountShow, field)

    redisExistsCmd := l.svcCtx.Redis.Exists(produceCountKey)
    redisExistsV, err := redisExistsCmd.Result()
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("redisExists链接失败")
        return nil, errors.New("redis链接失败")
    }

    if redisExistsV <= 0 {
        dbInfo, err := l.svcCtx.ProduceCountModel.FindOneByParam(in.TopicType, in.TopicId)
        if err != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("db查询作品点赞评论数量失败")
            return nil, errors.New("db查询计数信息失败")
        }

        l.svcCtx.Redis.HSet(produceCountKey, cfgredis.UserBehaviorProduceCountShowFieldComment, dbInfo.CommentNum)
        l.svcCtx.Redis.HSet(produceCountKey, cfgredis.UserBehaviorProduceCountShowFieldPraise, dbInfo.PraiseNum)
        l.svcCtx.Redis.HSet(produceCountKey, cfgredis.UserBehaviorProduceCountShowFieldShare, dbInfo.ShareNum)
        l.svcCtx.Redis.Expire(produceCountKey, cfgredis.ExpirationTenM)

        commentNum = dbInfo.CommentNum
        praiseNum = dbInfo.PraiseNum
        shareNum = dbInfo.ShareNum
    } else {
        redisShowCountRs := l.svcCtx.Redis.HGetAll(produceCountKey)
        if redisShowCountRs.Err() != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("redis链接失败")
            return nil, errors.New("redis链接失败")
        }

        commentNum, _ = strconv.ParseInt(redisShowCountRs.Val()[cfgredis.UserBehaviorProduceCountShowFieldComment], 10, 64)
        praiseNum, _ = strconv.ParseInt(redisShowCountRs.Val()[cfgredis.UserBehaviorProduceCountShowFieldPraise], 10, 64)
        shareNum, _ = strconv.ParseInt(redisShowCountRs.Val()[cfgredis.UserBehaviorProduceCountShowFieldShare], 10, 64)
    }

    return &userBehaviorProto.GetProduceCountResp{
        CommentNum: commentNum,
        PraiseNum:  praiseNum,
        ShareNum:   shareNum,
    }, nil
}
