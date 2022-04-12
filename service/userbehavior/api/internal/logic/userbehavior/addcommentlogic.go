package userbehavior

import (
    "context"
    "encoding/json"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "minicode.com/sirius/go-back-server/service/userbehavior/api/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/api/internal/types"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"
    "minicode.com/sirius/go-back-server/utils/errorx"
    "minicode.com/sirius/go-back-server/utils/mylogrus"
    "unicode/utf8"

    "github.com/zeromicro/go-zero/core/logx"
)

type AddCommentLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewAddCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) AddCommentLogic {
    return AddCommentLogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *AddCommentLogic) AddComment(req types.AddCommentReq) (resp *types.AddCommentResp, err error) {
    reqByte, _ := json.Marshal(req)
    reqStr := string(reqByte)

    if req.AppId < 0 || len(req.Content) == 0 || len(req.TopicId) == 0 || req.TopicType <= 0 {
        return nil, errorx.NewParamUserBehaviorApiError("评论参数错误")
    }

    if _, ok := cfgstatus.UserBehaviorCanCommentTopicType[int(req.TopicType)]; !ok {
        return nil, errorx.NewParamUserBehaviorApiError("topicType 不合法！")
    }

    // 判断字符串字数
    if utf8.RuneCountInString(req.Content) > cfgstatus.CommentContentMaxLength {
        return nil, errorx.NewParamUserBehaviorApiError("评论内容超过最大长度限制")
    }

    addParams := &userBehaviorProto.AddCommentReq{
        AppId:     req.AppId,
        TopicId:   req.TopicId,
        TopicType: req.TopicType,
        Content:   req.Content,
    }
    _, err = l.svcCtx.UserBehaviorRpc.AddComment(l.ctx, addParams)

    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("添加评论失败")
        return nil, errorx.Parse(err)
    }

    return
}
