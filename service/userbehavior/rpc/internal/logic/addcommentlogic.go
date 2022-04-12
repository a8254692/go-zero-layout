package logic

import (
    "context"
    "database/sql"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "minicode.com/sirius/go-back-server/service/userbehavior/model/comment"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"
    "minicode.com/sirius/go-back-server/utils/errorx"
    "minicode.com/sirius/go-back-server/utils/help"
    "minicode.com/sirius/go-back-server/utils/mylogrus"
    "time"
    "unicode/utf8"

    "encoding/json"
    "github.com/zeromicro/go-zero/core/logx"
)

type AddCommentLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewAddCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCommentLogic {
    return &AddCommentLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

//  添加评论
func (l *AddCommentLogic) AddComment(in *userBehaviorProto.AddCommentReq) (*userBehaviorProto.AddCommentResp, error) {
    reqByte, _ := json.Marshal(in)
    reqStr := string(reqByte)

    uin, err := help.GetRpcUinFromCtx(l.ctx)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": uin, "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取uin失败")
        return nil, errorx.NewSystemGetUinFromContextError()
    }

    if in.AppId < 0 || len(in.Content) == 0 || len(in.TopicId) == 0 || in.TopicType <= 0 {
        return nil, errorx.NewParamUserBehaviorRpcError("评论参数错误")
    }

    // 判断字符串字数
    if utf8.RuneCountInString(in.Content) > cfgstatus.CommentContentMaxLength {
        return nil, errorx.NewParamUserBehaviorRpcError("评论内容超过最大长度限制")
    }
    now := time.Now()
    params := &comment.Comment{
        AppId:     in.AppId,
        TopicId:   in.TopicId,
        TopicType: in.TopicType,
        Content:   sql.NullString{String: in.Content},
        Uin:       uin,
        CreatedAt: now,
    }
    result, err := l.svcCtx.CommentModel.Insert(params)

    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": uin, "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("添加评论")
        return nil, errorx.NewCodeErrorRPC(cfgstatus.UserBehaviorDbInsert, "添加评论失败")
    }

    id, err := result.LastInsertId()
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": uin, "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取插入id失败")
        return nil, errorx.NewCodeErrorRPC(cfgstatus.UserBehaviorDbGet, "获取插入id失败")
    }
    // 往消息队列塞数据，然后在crontab 里去调用迷你世界评论自动审核接口
    // 这里之所以用 crontab 是因为，自动审核可能会挂掉或者超时不返回( 对方服务接口 ),如果用groutine + 超时去做，那么超时的审核还得去触发重试机制
    // 与其这样还不如直接往消息队列中丢，然后在 crontab 中去处理，如果超时消息堆积就不确认就可以

    params.Id = id
    msg, err := json.Marshal(params)
    err = l.svcCtx.RmqCommentConn.PublishSimple(string(msg))
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": uin, "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("添加评论rmq发送消息失败")
    }

    return &userBehaviorProto.AddCommentResp{}, nil
}
