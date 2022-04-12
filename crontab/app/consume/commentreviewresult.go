package consume

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/go-redis/redis"
    "minicode.com/sirius/go-back-server/config/cfgmsg"
    "minicode.com/sirius/go-back-server/config/cfgredis"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "minicode.com/sirius/go-back-server/crontab/app/common"
    "minicode.com/sirius/go-back-server/crontab/app/svc"
    "minicode.com/sirius/go-back-server/utils/help"
    utilMessage "minicode.com/sirius/go-back-server/utils/message"
    "minicode.com/sirius/go-back-server/utils/mylogrus"
    "minicode.com/sirius/go-back-server/utils/tool"
    "time"
)

// 心得：从数据库读取到的时间格式数据，不经过转换，直接写缓存会有时区问题
type CommentReviewResultMsg struct {
    Id           int64 `json:"id"`
    ReviewStatus int64 `json:"reviewStatus"`
}

const (
    CommentReviewStatusBan = 1 // 评论审核不通过 -- 封禁
    CommentReviewStatusOk  = 2 // 评论审核通过
)

type CommentRInfo struct {
    Id        int64          `db:"id"`
    AppId     int64          `db:"app_id"`     // app
    TopicId   string         `db:"topic_id"`   // 主题id
    TopicType int64          `db:"topic_type"` // 主题类型
    Content   sql.NullString `db:"content"`    // 评论内容
    Uin       string         `db:"uin"`        // 评论用户id
    CreatedAt time.Time      `db:"created_at"`
}

type ExchangeDataReport struct {
    Event      string     `json:"event"`
    Uin        string     `json:"uin"`
    Methed     string     `json:"methed"`
    Properties Properties `json:"properties"`
}

type Properties struct {
    WorkId       string `json:"work_id"`       // 作品id
    WorkAuthor   string `json:"work_author"`   // 作品作者
    WorkCritics  string `json:"work_critics"`  // 评论者uin
    Contents     string `json:"contents"`      // 内容
    AuditResults string `json:"audit_results"` // 审核结果 "0" 通过审核 "1" 未通过审核
}

// 评论人工审核结果
func CommentReviewResult(svcCtx *svc.ServiceContext) {

    defer func() {
        if err := recover(); err != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("[panic] err: %v ,stack: %s \n", err, tool.GetCurrentGoroutineStack())
        }
    }()

    msg, err := svcCtx.CommentReviewResultRmqQConn.ConsumeSimple()
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("CommentReviewResult.ConsumeSimple.error ", err)
        return
    }

    //启用协程处理
    go func() {
        for d := range msg {
            if len(d.Body) != 0 {
                filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Infof("CommentReviewResult.receive.body %s ", d.Body)
                err = commentReviewResult(svcCtx, d.Body)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReviewResult.error %s", err.Error())
                }
            }
        }
    }()
}

func commentReviewResult(svcCtx *svc.ServiceContext, data []byte) error {
    reviewResult := new(CommentReviewResultMsg)
    err := json.Unmarshal(data, reviewResult)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.Unmarshal.error:%s", err.Error())
        return err
    }

    if reviewResult.Id <= 0 {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("commentReviewResult.reviewResult.Id lessthan 0")
        return nil
    }

    commentInfo, err := svcCtx.CommentModel.FindOne(reviewResult.Id)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.FindOne.error:%s", err.Error())
        return err
    }

    fmt.Printf("createdAt:db %v ,local %v \n", commentInfo.CreatedAt, commentInfo.CreatedAt.Local())
    createdTs := commentInfo.CreatedAt.Unix()
    content := sql.NullString{
        String: commentInfo.Content.String,
        Valid:  true,
    }
    rInfo := &cfgstatus.CommentRInfo{
        Id:        commentInfo.Id,
        AppId:     commentInfo.AppId,
        TopicId:   commentInfo.TopicId,
        TopicType: commentInfo.TopicType,
        Content:   content,
        Uin:       commentInfo.Uin,
        CreatedTs: createdTs,
    }

    bVal, err := json.Marshal(rInfo)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.rInfo.Marshal.error %s ", err.Error())
        return err
    }
    member := string(bVal)

    now := time.Now()

    hotKey := fmt.Sprintf(cfgredis.UserBehaviorCommentHot, commentInfo.AppId, commentInfo.TopicType, commentInfo.TopicId, now.Hour())
    latestKey := fmt.Sprintf(cfgredis.UserBehaviorCommentLatest, commentInfo.AppId, commentInfo.TopicType, commentInfo.TopicId)

    hotKeyExists := svcCtx.Redis.Exists(hotKey).Val()
    latestKeyExists := svcCtx.Redis.Exists(latestKey).Val()

    messageContent := ""
    auditResult := "0"
    messageType := cfgstatus.SendMessageTypeCommentReviewFailed
    // 审核不通过
    if reviewResult.ReviewStatus == CommentReviewStatusBan {
        reason := cfgstatus.CommentBanReasonMap[int(commentInfo.SensitiveEntry)]
        messageContent = fmt.Sprintf(cfgmsg.CommentReviewFailedMsg, commentInfo.Content.String, reason)

        // 存在才删除
        if hotKeyExists > 0 {
            cmd := svcCtx.Redis.ZRem(hotKey, member)
            _, err = cmd.Result()
            if err != nil {
                filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.ZRem.error:%s,key %s ", hotKey, err.Error())
                return err
            }
        }

        if latestKeyExists > 0 {
            cmd := svcCtx.Redis.ZRem(latestKey, member)
            _, err = cmd.Result()
            if err != nil {
                filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.ZRem.error:%s,key %s ", latestKey, err.Error())
                return err
            }
        }

    } else if reviewResult.ReviewStatus == CommentReviewStatusOk {
        // 审核通过
        auditResult = "1"
        messageType = cfgstatus.SendMessageTypeCommentReviewOk
        messageContent = fmt.Sprintf(cfgmsg.CommentReviewOkMsg, commentInfo.Content.String)

        if hotKeyExists > 0 {
            // 有则添加，否则缓存有效期会变成 -1
            cmd := svcCtx.Redis.ZAdd(hotKey, redis.Z{
                Score:  0,
                Member: member,
            })
            _, err = cmd.Result()
            if err != nil {
                filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.ZAdd.error:%s,key %s ", hotKey, err.Error())
                return err
            }
        }

        if latestKeyExists > 0 {
            cmd := svcCtx.Redis.ZAdd(latestKey, redis.Z{
                Score:  float64(createdTs),
                Member: member,
            })
            _, err := cmd.Result()
            if err != nil {
                filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.ZAdd.error:%s,key %s ,", latestKey, err.Error())
                return err
            }
        }

        countKey := fmt.Sprintf(cfgredis.UserBehaviorCommentCount, commentInfo.AppId, commentInfo.TopicType, commentInfo.TopicId)
        _, err = svcCtx.Redis.Get(countKey).Int64()
        if err != nil && err != redis.Nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.Get.error:%s,key %s ", countKey, err.Error())
            return err
        }

        if err != redis.Nil {
            // 数量存在则更新否则更新有可能会出现数据不准确的情况
            svcCtx.Redis.Incr(countKey)
        }
    }

    // 不管审核通不通过，更新用户未审核评论的数据
    _, err = updateUserUnReviewCommentCount(svcCtx, commentInfo.Uin, commentInfo.AppId, commentInfo.TopicType, commentInfo.TopicId, 1)
    if err != nil {
        return err
    }

    userUnReviewCommentKey := fmt.Sprintf(cfgredis.UserBehaviorUserUnReviewComment, commentInfo.Uin, commentInfo.AppId, commentInfo.TopicType, commentInfo.TopicId)
    // 直接删除改key zrem 会有问题，目前发现问题是，写入缓存的 时间戳和 写入数据库的时间不一致，数据库时间慢一秒
    svcCtx.Redis.Del(userUnReviewCommentKey)

    //   userUnReviewCommentKey := fmt.Sprintf(cfgredis.UserBehaviorUserUnReviewComment,commentInfo.Uin,commentInfo.AppId,commentInfo.TopicType,commentInfo.TopicId)
    //   // 直接删除改key zrem 会有问题，目前发现问题是，写入缓存的 时间戳和 写入数据库的时间不一致，数据库时间慢一秒
    //   svcCtx.Redis.Del(userUnReviewCommentKey)
    //   //svcCtx.Redis.ZRem(userUnReviewCommentKey,member)

    err = commentAddNum(svcCtx, commentInfo.AppId, commentInfo.TopicType, commentInfo.TopicId)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("commentReviewResult.commentAddNum.error:%s ", err.Error())
    }

    // 发送消息
    authorId := ""
    workTitle := "" // 作品名称
    ctx := context.Background()
    if commentInfo.TopicType == cfgstatus.UserBehaviorWorkType {
        // 作品
        workInfo, err := common.NewWorksCommon(ctx, svcCtx).GetWorkById(commentInfo.TopicId)

        if err != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("commentReviewResult.GetWorkById.error:%s,topicId %s ", err.Error(), commentInfo.TopicId)
            return err
        }
        authorId = workInfo.AuthorId
        workTitle = workInfo.Title

        if authorId == "" {
            return errors.New("作品作者id 为空")
        }
    } else if commentInfo.TopicType == cfgstatus.UserBehaviorProjectType {
        // 作品
        specialInfo, err := common.NewSpecialCommon(ctx, svcCtx).GetSpecialInfoById(commentInfo.TopicId)

        if err != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("commentReviewResult.GetSpecialInfoById.error:%s,topicId %s ", err.Error(), commentInfo.TopicId)
            return err
        }
        authorId = specialInfo.AuthorId

        if authorId == "" {
            return errors.New("专题作者id 为空")
        }
    }

    // 数据上报到神策
    properties := Properties{
        WorkId:       commentInfo.TopicId,
        WorkAuthor:   authorId,
        WorkCritics:  commentInfo.Uin,
        Contents:     commentInfo.Content.String,
        AuditResults: auditResult,
    }

    err = DataReport(svcCtx, commentInfo.Uin, properties)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("commentReviewResult.DataReport.err %s ", err.Error())
    }

    extra := utilMessage.SendMessageRmqQExtra{
        TopicType: commentInfo.TopicType,
        TopicId:   commentInfo.TopicId,
    }

    userInfoStc := common.NewUserInfoCommon(svcCtx)
    userInfo, err := userInfoStc.GetUserInfoById(commentInfo.Uin)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": commentInfo.Uin, "resp": err.Error(), "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("查询用户信息失败")
    }

    var nickName string
    if userInfo != nil {
        nickName = userInfo.NickName
    }

    // {评论者昵称} 评论了你的作品 《{你的作品}》:"{评论内容}"
    mobPushContentTemplate := ` %s 评论了你的作品《%s》:"%s" `
    mobPushContent := fmt.Sprintf(mobPushContentTemplate, nickName,workTitle,commentInfo.Content.String)
    sendDataMobPush := utilMessage.SendMessageRmqQMobPush{
        Title:    "作品被评论了",
        Content:  mobPushContent,
        NextType: cfgstatus.SendMessageNextTypeScheme,
        Url:      cfgstatus.SendMessageUrlMsgCenter,
    }

    // 发送消息
    message := utilMessage.SendMessageRmqQ{
        Type:    messageType,
        Uin:     commentInfo.Uin,
        Content: messageContent,
        Extra:   extra,
        MobPush:sendDataMobPush,
    }

    sendMsgErr := utilMessage.SendMessageUserCenter(svcCtx.SendMessageRmqQConn, message)
    if sendMsgErr != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("commentReviewResult.SendMessageUserCenter.error:%s  ", sendMsgErr.Error())
        return sendMsgErr
    }

    if reviewResult.ReviewStatus == CommentReviewStatusOk && authorId != commentInfo.Uin &&
        commentInfo.TopicType == cfgstatus.UserBehaviorWorkType {
        // 作品被评论且评论审核通过 需要给作者发消息
        message.Type = cfgstatus.SendMessageTypeComment
        message.Uin = authorId // 消息发给作者
        message.Content = fmt.Sprintf(cfgmsg.CommentMsg, commentInfo.Content.String)
        from := utilMessage.SendMessageRmqQFrom{AuthorId: commentInfo.Uin} // 消息来自于
        message.From = from

        sendMsgErr = utilMessage.SendMessageUserCenter(svcCtx.SendMessageRmqQConn, message)
        if sendMsgErr != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("commentReviewResult.SendMessageUserCenter.error:%s  ", sendMsgErr.Error())
            return sendMsgErr
        }
    }

    return nil
}

// 评论统计计数
func commentAddNum(svcCtx *svc.ServiceContext, appId, topicType int64, topicId string) error {

    today := time.Now().Format("20060102")
    field := fmt.Sprintf("%d|%d|%s", appId, topicType, topicId)
    userCommentNumKey := fmt.Sprintf(cfgredis.UserBehaviorCommentNum, today)

    existsRsV := svcCtx.Redis.HExists(userCommentNumKey, field).Val()

    if !existsRsV {
        //先查数据库中是否存在数据
        dbInfo, err := svcCtx.ProduceCountModel.FindOneByParam(topicType, topicId)
        if err != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("commentAddNum.FindOneByParam.error:%s  ", err.Error())
        } else {
            if dbInfo.CommentNum > 0 {
                err = svcCtx.Redis.HSet(userCommentNumKey, field, dbInfo.CommentNum).Err()
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("commentAddNum.HSet.error:%s  ", err.Error())
                }
            }
        }
    }

    err := svcCtx.Redis.HIncrBy(userCommentNumKey, field, 1).Err()
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("commentAddNum.HIncrBy.error:%s  ", err.Error())
        return err
    }

    svcCtx.Redis.Expire(userCommentNumKey, help.GetTodayTimeRemaining())
    return nil
}

// 神策数据上报 ( 需在 queue  中再处理 )
func DataReport(svcCtx *svc.ServiceContext, uin string, properties Properties) (err error) {

    msgStr, err := json.Marshal(ExchangeDataReport{
        Event:      "AuditComments",
        Uin:        uin,
        Methed:     "track",
        Properties: properties,
    })
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("DataReport.Marshal.error:%s  ", err.Error())
        return err
    }

    err = svcCtx.RmqDataReportConn.PublishSimple(string(msgStr))
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("DataReport.RmqDataReportConn.PublishSimple.error:%s  ", err.Error())
        return err
    }
    fmt.Printf("DataReport.msg %s \n", string(msgStr))
    return nil
}
