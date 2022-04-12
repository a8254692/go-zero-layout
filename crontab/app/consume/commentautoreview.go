package consume

import (
    "crypto/md5"
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/go-redis/redis"
    "minicode.com/sirius/go-back-server/config/cfgredis"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    modelComment "minicode.com/sirius/go-back-server/crontab/app/model/comment"
    "minicode.com/sirius/go-back-server/crontab/app/svc"
    utils "minicode.com/sirius/go-back-server/utils/help"
    "minicode.com/sirius/go-back-server/utils/mylogrus"
    "minicode.com/sirius/go-back-server/utils/tool"
    "net/http"
    "net/url"
    "time"
)

// 使用 time.Time 会有时区问题，数据库和 redis 存储的时间格式会有问题，导致更新的时候会有异常
type Comment struct {
    Id        int64          `db:"id"`
    AppId     int64          `db:"app_id"`     // app
    TopicId   string         `db:"topic_id"`   // 主题id
    TopicType int64          `db:"topic_type"` // 主题类型
    Content   sql.NullString `db:"content"`    // 评论内容
    Uin       string         `db:"uin"`        // 评论用户id
    CreatedAt time.Time      `db:"created_at"`
    CreatedTs int64          `db:"created_ts"` // 创建时间戳
}

type CommentReviewRes struct {
    Ret int64  `json:"ret"`
    Msg string `json:"msg"`
}

const CommentReviewSuccessRet = 0

const (
    ReviewStatusFailed  = 1 // 审核未通过
    ReviewStatusSuccess = 2 // 审核通过
)

// 直接在main 函数中加入异常捕获，并不能阻止程序崩溃
// 评论自动审核
func CommentAutoReview(svcCtx *svc.ServiceContext) {

    defer func() {
        if err := recover(); err != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("[panic] err: %v ,stack: %s \n", err, tool.GetCurrentGoroutineStack())
        }
    }()

    msg, err := svcCtx.RmqCommentConn.ConsumeSimple()
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("CommentAutoReview.ConsumeSimple.error ", err)
        return
    }

    //启用协程处理
    go func() {
        for d := range msg {
            if len(d.Body) != 0 {
                err = commentAutoReview(svcCtx, d.Body)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentAutoReview.error %s", err.Error())
                }
            }
        }
    }()
}

func commentAutoReview(svcCtx *svc.ServiceContext, data []byte) error {
    fmt.Printf("commentAutoReview.data %s \n", data)
    comment := new(Comment)
    err := json.Unmarshal(data, comment)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.Unmarshal.error:%s", err.Error())
        return err
    }

    fmt.Printf("comment %+v ,ts %d \n", comment, comment.CreatedAt.Unix())
    count, err := updateUserUnReviewCommentCount(svcCtx, comment.Uin, comment.AppId, comment.TopicType, comment.TopicId, 0)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.updateUserUnReviewCommentCount.error:%s", err.Error())
        return err
    }

    if count > 0 {
        // 更新用户未审核评论数据
        userUnReviewCommentKey := fmt.Sprintf(cfgredis.UserBehaviorUserUnReviewComment, comment.Uin, comment.AppId, comment.TopicType, comment.TopicId)
        fmt.Printf("commentAutoReview.userUnReviewCommentKey %s \n", userUnReviewCommentKey)
        zCount, err := svcCtx.Redis.ZCard(userUnReviewCommentKey).Result()

        if err != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.ZCard.error:%s", err.Error())
            return err
        }
        if zCount > 0 {
            content := sql.NullString{
                String: comment.Content.String,
                Valid:  true,
            }
            rInfo := cfgstatus.CommentRInfo{
                Id:        comment.Id,
                AppId:     comment.AppId,
                TopicId:   comment.TopicId,
                TopicType: comment.TopicType,
                Content:   content,
                Uin:       comment.Uin,
                CreatedTs: comment.CreatedAt.Unix(),
            }

            fmt.Printf("commentAutoReview.rInfo %+v \n", rInfo)
            bVal, err := json.Marshal(rInfo)
            if err != nil {
                return err
            }
            member := string(bVal)
            z := redis.Z{
                Score:  float64(rInfo.CreatedTs),
                Member: member,
            }
            fmt.Printf("commentAutoReview.bVal %s \n", bVal)
            c, err := svcCtx.Redis.ZAdd(userUnReviewCommentKey, z).Result()
            fmt.Printf("ZAdd.c %d ,err %v \n", c, err)
        }
    }

    reviewConf := svcCtx.Config.CommentReview

    params := url.Values{}
    Url, err := url.Parse(reviewConf.Url)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.url.Parse.err %s", err.Error())
        return err
    }
    now := time.Now()
    nowTs := now.Unix()
    params.Set("cmd", reviewConf.Cmd)
    params.Set("type", fmt.Sprintf("%d", reviewConf.Type))
    params.Set("time", fmt.Sprintf("%d", nowTs))
    params.Set("env", fmt.Sprintf("%d", reviewConf.Env))
    params.Set("uin", comment.Uin)
    params.Set("token", generateReviewToken(nowTs, comment.Uin))
    params.Set("from", fmt.Sprintf("%d", reviewConf.From))
    params.Set("key", comment.Content.String)
    //如果参数中有中文参数,这个方法会进行URLEncode
    Url.RawQuery = params.Encode()
    urlPath := Url.String()

    httpStatus, resp := utils.OnGetHttp(urlPath, nil)

    if httpStatus != http.StatusOK {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.OnGetHttp.res:%s,httpStatus %d ", resp, httpStatus)
        return errors.New("CommentReview.OnGetHttp.error")
    }

    res := new(CommentReviewRes)
    err = json.Unmarshal(resp, res)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.json.Unmarshal.error %s", err.Error())
        return err
    }

    reviewStatus := int64(0)
    reviewMark := ""
    if res.Ret == CommentReviewSuccessRet {
        reviewStatus = ReviewStatusSuccess
    } else {
        reviewStatus = ReviewStatusFailed
        reviewMark = res.Msg

        if reviewMark == "" {
            // 审核不通过后，接口不会返回msg
            reviewMark = cfgstatus.CommentReviewErrMsg[int(res.Ret)]
        }
    }

    commentParams := &modelComment.Comment{
        Id:               comment.Id,
        AutoReviewTime:   sql.NullTime{Time: now},
        AutoReviewStatus: reviewStatus,
    }
    err = updateComment(svcCtx, commentParams)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("CommentReview.updateComment.error %s", err.Error())
        return err
    }

    return nil
}

// 生成自动审核接口调用 token
func generateReviewToken(timeStamp int64, uin string) string {
    str := fmt.Sprintf("%s#heihei#%d#check#", uin, timeStamp)
    data := []byte(str)
    has := md5.Sum(data)

    return fmt.Sprintf("%x", has) // 将 []byte 转成 16 进制

}

// 修改表数据
func updateComment(svcCtx *svc.ServiceContext, comment *modelComment.Comment) (err error) {

    now := time.Now()
    params := &modelComment.Comment{
        Id:               comment.Id,
        AutoReviewTime:   comment.AutoReviewTime,
        AutoReviewStatus: comment.AutoReviewStatus,
        UpdatedAt:        now,
    }

    return svcCtx.CommentModel.Update(params)
}

/**
* 更新用户未被审核的评论数
* operateType 0 递增 1 递减
 */
func updateUserUnReviewCommentCount(svcCtx *svc.ServiceContext, uin string, appId int64, topicType int64, topicId string, operateType int) (int64, error) {
    countKey := fmt.Sprintf(cfgredis.UserBehaviorUserUnReviewCommentCount, uin, appId, topicType, topicId)
    count, err := svcCtx.Redis.Get(countKey).Int64()
    if err != nil && err != redis.Nil {
        return count, err
    }

    if err == redis.Nil {
        return 0, nil
    }

    var intCmd *redis.IntCmd
    if operateType == 0 {
        intCmd = svcCtx.Redis.Incr(countKey)
    } else {
        if count == 0 {
            // 一般情况下不会执行到这里，加个判断，以防数据出错
            return 0, err
        }
        intCmd = svcCtx.Redis.Decr(countKey)
    }

    count, err = intCmd.Result()
    return count, err
}
