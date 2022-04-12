package logic

import (
    "context"
    "fmt"
    "github.com/go-redis/redis"
    "minicode.com/sirius/go-back-server/config/cfgredis"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    modelComment "minicode.com/sirius/go-back-server/service/userbehavior/model/comment"
    "minicode.com/sirius/go-back-server/service/userbehavior/model/usermgo"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/common"
    "minicode.com/sirius/go-back-server/utils/errorx"
    "minicode.com/sirius/go-back-server/utils/mylogrus"
    "time"

    "encoding/json"
    "github.com/zeromicro/go-zero/core/logx"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"
    "minicode.com/sirius/go-back-server/utils/help"
)

type GetCommentListLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

const (
    CommentLatestSort = 0
    CommentHotSort    = 1
)

var AvatarsIdMap = map[int]string{
    1: "https://minioss.miniaixue.com/images/coolmini.png",
    2: "https://minioss.miniaixue.com/images/misra.png",
    3: "https://minioss.miniaixue.com/images/tumeimei.png",
}

const defaultAvatr = "https://minioss.miniaixue.com/images/coolmini.png"

func NewGetCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentListLogic {
    return &GetCommentListLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

//  获取评论列表
func (l *GetCommentListLogic) GetCommentList(in *userBehaviorProto.GetCommentListReq) (*userBehaviorProto.GetCommentListResp, error) {
    reqByte, _ := json.Marshal(in)
    reqStr := string(reqByte)

    if in.AppId < 0 || len(in.TopicId) == 0 || in.TopicType <= 0 {
        return nil, errorx.NewParamUserBehaviorRpcError("获取评论参数错误")
    }

    uin, err := help.GetRpcUinFromCtx(l.ctx)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取uin失败")
        return nil, errorx.NewSystemGetUinFromContextError()
    }

    page, pageSize := in.Page, in.PageSize
    if page <= 0 || pageSize <= 0 {
        page, pageSize = 1, 10
    }

    if in.PageSize > 50 {
        pageSize = 50
    }

    resp := &userBehaviorProto.GetCommentListResp{}

    count, data, err := l.getCommentList(uin, in.AppId, in.TopicType, in.TopicId, pageSize, page, in.Sort)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取评论列表失败")
        return resp, err
    }
    resp.List = data
    resp.Total = count

    return resp, nil
}

// 获取评论列表数据
func (l *GetCommentListLogic) getCommentList(uin string, appId int64, topicType int64, topicId string, pageSize, page, sort int32) (int64, []*userBehaviorProto.CommentList, error) {
    now := time.Now()
    duration := 7 * 24 * time.Hour
    key := fmt.Sprintf(cfgredis.UserBehaviorCommentLatest, appId, topicType, topicId)
    if sort == CommentHotSort {
        hour := now.Hour()
        duration = time.Hour
        key = fmt.Sprintf(cfgredis.UserBehaviorCommentHot, appId, topicType, topicId, hour)
    }

    var userComment *[]modelComment.Comment
    var userCommentErr error

    // 获取用户未被审核的评论数
    userCommentCount, err := l.GetUserUnReviewCommentCount(uin, appId, topicType, topicId)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取用户未经审核评论数量失败")
        return 0, nil, err
    }

    if userCommentCount > 0 && page == 1 {
        // 只在第一页展示用户自己的评论
        userComment, userCommentErr = l.getUserUnReviewCommentList(uin, appId, topicType, topicId)
        if userCommentErr != nil {
            return 0, nil, userCommentErr
        }
    }

    reviewCount, err := l.GetCommentCount(appId, topicType, topicId)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取评论数量失败")
        return 0, nil, err
    }

    count := userCommentCount + reviewCount // 展示给用户的总数 = 该用户未经审核数 + 该topic_id 已过审核评论数
    if userComment == nil && count == 0 {
        return count, nil, nil
    }

    start := (page - 1) * pageSize
    stop := (page * pageSize) - 1 // redis zrange stop 值

    if int64(start) > count {
        // 获取数据超过数据库总数直接返回
        return count, nil, nil
    }

    cmd := l.svcCtx.Redis.ZRevRange(key, int64(start), int64(stop))
    comment, err := cmd.Result()
    if err != nil && err != redis.Nil {
        return count, nil, errorx.NewGetCacheDataUserBehaviorRpcError("获取评论列表出错")
    }

    isUpdateCache := false
    commentList := new([]modelComment.Comment)
    if len(comment) == 0 {
        if sort == CommentHotSort {
            commentList, err = l.svcCtx.CommentModel.FindCommentHot(appId, topicType, topicId, pageSize, start)
        } else {
            commentList, err = l.svcCtx.CommentModel.FindCommentLatest(appId, topicType, topicId, pageSize, start)
        }

        if err != nil && err != modelComment.ErrNotFound {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取评论列表失败")
            return count, nil, err
        }
        isUpdateCache = true
    } else {
        list := make([]modelComment.Comment, 0)
        for _, v := range comment {
            //info := modelComment.Comment{}
            info := cfgstatus.CommentRInfo{}
            err = json.Unmarshal([]byte(v), &info)
            if err != nil {
                filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("comment.redis.unmarshal")
                continue
            }
            ret := modelComment.Comment{
                Id:        info.Id,
                AppId:     info.AppId,
                TopicId:   info.TopicId,
                TopicType: info.TopicType,
                Content:   info.Content,
                Uin:       info.Uin,
                CreatedAt: time.Unix(info.CreatedTs, 0),
            }
            list = append(list, ret)
        }

        commentList = &list
    }

    if userComment == nil && commentList == nil {
        return count, nil, nil
    }

    // 获取缓存数量
    cardCmd := l.svcCtx.Redis.ZCard(key)
    redisLength, zCardErr := cardCmd.Result()

    if zCardErr != nil && zCardErr != redis.Nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("Redis.ZCard.error")
    }

    // 预存储 ,直接一次性拉取后10页数据到缓存中
    totalLen := (page + 10) * pageSize
    if redisLength > 0 && redisLength < reviewCount && redisLength < int64(totalLen) {
        offset := page * pageSize
        size := 10 * pageSize
        go l.preloadCommentList(appId, topicType, topicId, size, offset, sort, key, duration)
    }

    // 获取作品作者信息
    authorId, err := l.GetAuthorId(topicType, topicId)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": "", "msg": err.Error()}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取作品作者id失败")
    }

    commonUserInfo := common.NewUserInfoCommon(l.ctx, l.svcCtx)
    uinMap := make(map[string]*usermgo.User)
    list := make([]*userBehaviorProto.CommentList, 0)

    if userComment != nil {
        for _, item := range *userComment {
            if _, ok := uinMap[item.Uin]; !ok {
                userInfo, err := commonUserInfo.GetUserInfoById(item.Uin)
                if err != nil {
                    filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取用户信息失败")
                }
                uinMap[item.Uin] = userInfo
            }
            createdStr := item.CreatedAt.Format("2006-01-02")

            userinfo := uinMap[item.Uin]
            var userAvatar, nickName string
            if userinfo != nil {
                nickName = userinfo.NickName

                if _, ok := AvatarsIdMap[userinfo.AvatarId]; ok {
                    userAvatar = AvatarsIdMap[userinfo.AvatarId]
                } else {
                    userAvatar = defaultAvatr
                }
            }

            dYear, dMonth, dDay := item.CreatedAt.Date()
            cYear, cMonth, cDay := now.Date()
            if dYear == cYear && dMonth == cMonth && dDay == cDay {
                createdStr = fmt.Sprintf("今天 %d:%d", now.Hour(), now.Minute())
            }

            ret := userBehaviorProto.CommentList{
                Id:          item.Id,
                Content:     item.Content.String,
                Uin:         item.Uin,
                NickName:    nickName,
                UserAvatar:  userAvatar,
                CreatedAt:   createdStr,
                PraiseCount: 0,
                IsAuthor:    authorId == item.Uin,
                IsPraise:    false,
            }
            list = append(list, &ret)
        }
    }

    if commentList == nil {
        return count, list, nil
    }

    members := make([]redis.Z, 0)

    for _, val := range *commentList {
        d := val.CreatedAt
        if isUpdateCache {
            score := d.Unix()
            if sort == CommentHotSort {
                score = val.PraiseCount
            }

            member := cfgstatus.CommentRInfo{
                Id:        val.Id,
                AppId:     val.AppId,
                TopicId:   val.TopicId,
                TopicType: val.TopicType,
                Content:   val.Content,
                Uin:       val.Uin,
                CreatedTs: val.CreatedAt.Unix(),
            }

            valByte, err := json.Marshal(member)
            if err == nil {

                members = append(members, redis.Z{
                    Score:  float64(score),
                    Member: string(valByte),
                })

            } else {
                filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("json.Marshal")
            }
        }

        if val.Uin != "" {
            if _, ok := uinMap[val.Uin]; !ok {
                userInfo, err := commonUserInfo.GetUserInfoById(val.Uin)
                if err != nil {
                    filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取用户信息失败")
                }
                uinMap[val.Uin] = userInfo
            }
        }

        createdStr := d.Format("2006-01-02")

        userinfo := uinMap[val.Uin]
        var userAvatar, nickName string
        if userinfo != nil {
            nickName = userinfo.NickName

            if _, ok := AvatarsIdMap[userinfo.AvatarId]; ok {
                userAvatar = AvatarsIdMap[userinfo.AvatarId]
            } else {
                userAvatar = defaultAvatr
            }
        }

        ret := userBehaviorProto.CommentList{
            Id:         val.Id,
            Content:    val.Content.String,
            Uin:        val.Uin,
            NickName:   nickName,
            UserAvatar: userAvatar,
        }

        if authorId == val.Uin {
            ret.IsAuthor = true
        }

        dYear, dMonth, dDay := d.Date()
        cYear, cMonth, cDay := now.Date()
        if dYear == cYear && dMonth == cMonth && dDay == cDay {
            createdStr = fmt.Sprintf("今天 %d:%d", now.Hour(), now.Minute())
        }
        ret.CreatedAt = createdStr

        idStr := fmt.Sprintf("%d", ret.Id)
        isPraise, IsPraiseErr := l.GetCommentIsPraise(uin, appId, cfgstatus.UserBehaviorCommentType, idStr)
        if IsPraiseErr != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取是否点赞失败")
        }
        ret.IsPraise = isPraise
        praiseCount, praiseErr := l.GetCommentPraiseCount(uin, appId, cfgstatus.UserBehaviorCommentType, idStr)
        if praiseErr != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取点赞数失败")
        }

        ret.PraiseCount = praiseCount

        list = append(list, &ret)
    }

    if len(members) > 0 {
        l.svcCtx.Redis.ZAdd(key, members...)
        l.svcCtx.Redis.Expire(key, duration)
    }

    return count, list, nil

}

/**
*   获取用户未审核评论
 */
func (l *GetCommentListLogic) getUserUnReviewCommentList(uin string, appId int64, topicType int64, topicId string) (*[]modelComment.Comment, error) {
    key := fmt.Sprintf(cfgredis.UserBehaviorUserUnReviewComment, uin, appId, topicType, topicId)
    // 单个用户对单个作品未审核的评论不会多
    getCmd := l.svcCtx.Redis.ZRange(key, 0, -1)
    comment, err := getCmd.Result()

    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取用户未审核评论列表出错")
        return nil, errorx.NewGetCacheDataUserBehaviorRpcError("获取用户未审核评论列表出错")
    }

    if len(comment) > 0 {
        list := make([]modelComment.Comment, 0)
        for _, v := range comment {
            ret := new(cfgstatus.CommentRInfo)
            err = json.Unmarshal([]byte(v), ret)
            if err != nil {
                filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("用户未审核数据反序列化失败")
                return nil, errorx.NewUserBehaviorCommentRpcUnMarshal("用户未审核数据反序列化失败")
            }

            info := modelComment.Comment{
                Id:        ret.Id,
                AppId:     ret.AppId,
                TopicId:   ret.TopicId,
                TopicType: ret.TopicType,
                Content:   ret.Content,
                Uin:       ret.Uin,
                CreatedAt: time.Unix(ret.CreatedTs, 0),
            }
            list = append(list, info)

        }
        return &list, nil
    }
    userComment, err := l.svcCtx.CommentModel.FindCommentByUin(uin, appId, topicType, topicId)
    if err != nil && err != modelComment.ErrNotFound {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取用户未审核评论列表出错")
        return nil, errorx.NewGetDbDataUserBehaviorRpcError("获取用户未审核评论列表出错")
    }

    if userComment != nil {
        members := make([]redis.Z, 0)
        for _, v := range *userComment {
            bInfo := cfgstatus.CommentRInfo{
                Id:        v.Id,
                AppId:     v.AppId,
                TopicId:   v.TopicId,
                TopicType: v.TopicType,
                Content:   v.Content,
                Uin:       v.Uin,
                CreatedTs: v.CreatedAt.Unix(),
            }

            member, err := json.Marshal(bInfo)
            if err != nil {
                filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("用户未审核数据序列化失败")
                return nil, errorx.NewUserBehaviorCommentRpcMarshal("用户未审核数据序列化失败")
            }

            members = append(members, redis.Z{
                Score:  float64(v.CreatedAt.Unix()),
                Member: string(member),
            })
        }

        if len(members) > 0 {
            l.svcCtx.Redis.ZAdd(key, members...)
            expire := 3 * 24 * time.Hour
            l.svcCtx.Redis.Expire(key, expire)
        }
    }

    return userComment, nil
}

// 数据预加载到缓存中
func (l *GetCommentListLogic) preloadCommentList(appId int64, topicType int64, topicId string, pageSize, offset, sort int32, key string, expire time.Duration) (err error) {
    fmt.Printf("preloadCommentList.预加载 key %s ,page %d ", key, offset)
    commentList := new([]modelComment.Comment)

    if sort == CommentHotSort {
        commentList, err = l.svcCtx.CommentModel.FindCommentHot(appId, topicType, topicId, pageSize, offset)
    } else {
        commentList, err = l.svcCtx.CommentModel.FindCommentLatest(appId, topicType, topicId, pageSize, offset)
    }

    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("preloadCommentList")
        return
    }
    score := int64(0)
    members := make([]redis.Z, 0)
    for _, val := range *commentList {
        if sort == CommentHotSort {
            score = val.PraiseCount
        } else {
            score = val.CreatedAt.Unix()
        }

        bInfo := cfgstatus.CommentRInfo{
            Id:        val.Id,
            AppId:     val.AppId,
            TopicId:   val.TopicId,
            TopicType: val.TopicType,
            Content:   val.Content,
            Uin:       val.Uin,
            CreatedTs: val.CreatedAt.Unix(),
        }

        valByte, err := json.Marshal(bInfo)
        if err != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("json.Marshal")
            continue
        }
        members = append(members, redis.Z{
            Score:  float64(score),
            Member: string(valByte),
        })

    }

    if len(members) > 0 {
        l.svcCtx.Redis.ZAdd(key, members...)
        l.svcCtx.Redis.Expire(key,expire)
    }

    return
}

// 获取评论数
func (l *GetCommentListLogic) GetCommentCount(appId int64, topicType int64, topicId string) (int64, error) {
    key := fmt.Sprintf(cfgredis.UserBehaviorCommentCount, appId, topicType, topicId)
    count, err := l.svcCtx.Redis.Get(key).Int64()
    if err != nil && err != redis.Nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取评论数失败Redis.Get")
        return 0, errorx.NewGetCacheDataUserBehaviorRpcError("获取评论数失败")
    }

    if err == redis.Nil {
        count, err = l.svcCtx.CommentModel.CommentCount(appId, topicType, topicId)
        if err != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取评论数失败")
            return count, errorx.NewGetDbDataUserBehaviorRpcError("获取评论数失败")
        }

        l.svcCtx.Redis.Set(key, count, time.Hour*24)
    }

    return count, nil
}

/**
*  获取用户未审核的评论数
 */
func (l *GetCommentListLogic) GetUserUnReviewCommentCount(uin string, appId int64, topicType int64, topicId string) (int64, error) {
    key := fmt.Sprintf(cfgredis.UserBehaviorUserUnReviewCommentCount, uin, appId, topicType, topicId)

    count, err := l.svcCtx.Redis.Get(key).Int64()
    if err != nil && err != redis.Nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取用户未审核评论数出错")
        return 0, errorx.NewGetCacheDataUserBehaviorRpcError("获取用户未审核评论数出错")
    }

    if err == redis.Nil {
        count, err = l.svcCtx.CommentModel.UserUnReviewCommentCount(uin, appId, topicType, topicId)
        if err != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取用户未审核评论数出错")
            return count, errorx.NewGetDbDataUserBehaviorRpcError("获取用户未审核评论数出错")
        }

        l.svcCtx.Redis.Set(key, count, time.Hour*24*3)
    }

    return count, nil
}

// 获取是否点赞
func (l *GetCommentListLogic) GetCommentIsPraise(uin string, AppId int64, TopicType int64, TopicId string) (bool, error) {

    praiseLogic := NewGetUserIsPraiseLogic(l.ctx, l.svcCtx)
    praiseParam := &userBehaviorProto.GetUserIsPraiseReq{
        AppId:     AppId,
        TopicType: TopicType,
        TopicId:   TopicId,
    }

    isPraiseResp, isPraiseErr := praiseLogic.GetUserIsPraise(praiseParam)
    if isPraiseErr != nil || isPraiseResp == nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取用户是否点赞失败")
        return false, errorx.NewGetDataUserBehaviorRpcError("获取点赞失败")
    }

    return isPraiseResp.IsPraise, nil
}

// 获取点赞数
func (l *GetCommentListLogic) GetCommentPraiseCount(uin string, AppId int64, TopicType int64, TopicId string) (int64, error) {

    count := int64(0)
    countLogic := NewGetProduceCountLogic(l.ctx, l.svcCtx)
    countParam := &userBehaviorProto.GetProduceCountReq{
        AppId:     AppId,
        TopicType: TopicType,
        TopicId:   TopicId,
    }

    countResp, countErr := countLogic.GetProduceCount(countParam)
    if countErr != nil || countResp == nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取点赞数失败")
        return count, errorx.NewGetDataUserBehaviorRpcError("获取点赞数失败")
    }

    count = countResp.PraiseNum
    return count, nil
}

func (l *GetCommentListLogic) GetAuthorId(topicType int64, topicId string) (authorId string, err error) {

    if topicType == cfgstatus.UserBehaviorWorkType {
        // 作品
        workInfo, workErr := common.NewWorksCommon(l.ctx, l.svcCtx).GetWorkById(topicId)

        if workErr != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("commentReviewResult.GetWorkById.error:%s,topicId %s ", workErr.Error(), topicId)
            return
        }
        authorId = workInfo.AuthorId
    } else if topicType == cfgstatus.UserBehaviorProjectType {
        // 作品
        specialInfo, specialErr := common.NewSpecialCommon(l.ctx, l.svcCtx).GetSpecialInfoById(topicId)

        if specialErr != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("commentReviewResult.GetSpecialInfoById.error:%s,topicId %s ", specialErr.Error(), topicId)
            return
        }
        authorId = specialInfo.AuthorId
    }
    return
}