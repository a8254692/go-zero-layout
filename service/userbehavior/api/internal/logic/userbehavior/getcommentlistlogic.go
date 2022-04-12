package userbehavior

import (
    "context"
    "encoding/json"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/rpcuserbehavior"
    "minicode.com/sirius/go-back-server/utils/errorx"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "minicode.com/sirius/go-back-server/service/userbehavior/api/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/api/internal/types"

    "github.com/zeromicro/go-zero/core/logx"
)

type GetCommentListLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewGetCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) GetCommentListLogic {
    return GetCommentListLogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *GetCommentListLogic) GetCommentList(req types.GetCommentListReq) (resp *types.GetCommentListResp, err error) {
    reqByte, _ := json.Marshal(req)
    reqStr := string(reqByte)

    resp = new(types.GetCommentListResp)
    topicType := req.TopicType
    if _, ok := cfgstatus.UserBehaviorTypeMap[int(topicType)]; !ok {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("topicType非法")
        return resp, errorx.NewParamUserBehaviorApiError("参数非法")
    }

    var page, pageSize int32

    page, pageSize = req.Page, req.PageSize
    if req.Page < 0 || req.PageSize < 0 {
        page, pageSize = 1, 10
    }

    appId := req.AppId
    if appId < 0 {
        appId = 0
    }

    params := &rpcuserbehavior.GetCommentListReq{
        AppId:     req.AppId,
        TopicId:   req.TopicId,
        TopicType: req.TopicType,
        Page:      page,
        PageSize:  pageSize,
        Sort:      req.Sort,
    }
    commentListResp, err := l.svcCtx.UserBehaviorRpc.GetCommentList(l.ctx, params)

    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-API", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取评论数据失败")
        return resp, errorx.Parse(err)
    }

    resp.Total = commentListResp.Total
    resp.List = make([]types.CommentList, 0)
    for _, v := range commentListResp.List {
        resp.List = append(resp.List, types.CommentList{
            Id:          v.Id,
            Uin:         v.Uin,
            NickName:    v.NickName,
            UserAvatar:  v.UserAvatar,
            Content:     v.Content,
            CreatedAt:   v.CreatedAt,
            PraiseCount: v.PraiseCount,
            IsAuthor:    v.IsAuthor,
            IsPraise:    v.IsPraise,
        })
    }
    return resp, nil
}
