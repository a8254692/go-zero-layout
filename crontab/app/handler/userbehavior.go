package handler

import (
    "errors"
    "fmt"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "strconv"
    "strings"
    "time"

    "minicode.com/sirius/go-back-server/config/cfgredis"
    "minicode.com/sirius/go-back-server/crontab/app/model/producecount"
    "minicode.com/sirius/go-back-server/crontab/app/svc"
    "minicode.com/sirius/go-back-server/utils/mylogrus"
)

func UserBehavior(svcCtx *svc.ServiceContext) {
    err := userBehaviorUpdateCommentNum(svcCtx)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("落地评论数失败")
    }
    err = userBehaviorUpdatePraiseNum(svcCtx)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("落地点赞数失败")
    }
    err = userBehaviorUpdateShareNum(svcCtx)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("落地分享数失败")
    }

    //2022年3月16日14:13:22方案改了不需要了
    //err = userBehaviorUpdateFocusNum(svcCtx)
    //if err != nil {
    //    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
    //    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("落地点赞数失败")
    //}
    //err = userBehaviorUpdateFollowNum(svcCtx)
    //if err != nil {
    //    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
    //    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("落地粉丝数失败")
    //}

    return
}

func userBehaviorUpdateCommentNum(svcCtx *svc.ServiceContext) (err error) {
    today := time.Now().Format("20060102")
    sep := "|"

    userCommentNumKey := fmt.Sprintf(cfgredis.UserBehaviorCommentNum, today)
    all := svcCtx.Redis.HGetAll(userCommentNumKey)
    if all.Err() != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("获取redis中评论数据失败")
        return errors.New("获取redis中评论数据失败")
    }

    if all != nil {
        if len(all.Val()) > 0 {
            for k, v := range all.Val() {
                params := strings.Split(k, sep)
                if len(params) != 3 {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("all的key拆分校验失败")
                    continue
                }

                if v == "" {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("all的val为0")
                    continue
                }

                appId, err := strconv.ParseInt(params[0], 10, 64)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("appId转为int64失败")
                    continue
                }

                topicType, err := strconv.ParseInt(params[1], 10, 64)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("topicType转为int64失败")
                    continue
                }

                topicId := params[2]

                changeNum, err := strconv.ParseInt(v, 10, 64)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("changeNum转为int64失败")
                    continue
                }

                //先查是否有数据
                dbInfo, err := svcCtx.ProduceCountModel.FindOneByParam(topicType, topicId)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("查询评论数据是否存在失败")
                    continue
                }

                if changeNum == 0 {
                    continue
                }

                if dbInfo.Id <= 0 {
                    insertData := producecount.ProduceCount{
                        AppId:      appId,
                        TopicId:    topicId,
                        TopicType:  topicType,
                        CommentNum: changeNum,
                    }
                    _, err = svcCtx.ProduceCountModel.Insert(&insertData)
                    if err != nil {
                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("插入评论数据失败")
                        continue
                    }
                } else {
                    updateData := producecount.ProduceCount{
                        Id:         dbInfo.Id,
                        CommentNum: changeNum,
                    }
                    err = svcCtx.ProduceCountModel.UpdateNum(&updateData)
                    if err != nil {
                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("更新评论数据失败")
                        continue
                    }
                }

                appIdStr := fmt.Sprintf("%d", appId)
                topicTypeStr := fmt.Sprintf("%d", topicType)
                field := appIdStr + "|" + topicTypeStr + "|" + topicId
                countNumShowKey := fmt.Sprintf(cfgredis.UserBehaviorProduceCountShow, field)
                svcCtx.Redis.Del(countNumShowKey)
            }
        }
    }
    return
}

func userBehaviorUpdatePraiseNum(svcCtx *svc.ServiceContext) (err error) {
    today := time.Now().Format("20060102")
    sep := "|"

    userPraiseNumKey := fmt.Sprintf(cfgredis.UserBehaviorPraiseNum, today)
    all := svcCtx.Redis.HGetAll(userPraiseNumKey)
    if all.Err() != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("获取redis中点赞数据失败")
        return errors.New("获取redis中点赞数据失败")
    }

    if all != nil {
        if len(all.Val()) > 0 {
            for k, v := range all.Val() {
                params := strings.Split(k, sep)
                if len(params) != 3 {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("all的key拆分校验失败")
                    continue
                }

                if v == "" {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("all的val为0")
                    continue
                }

                appId, err := strconv.ParseInt(params[0], 10, 64)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("appId转为int64失败")
                    continue
                }

                topicType, err := strconv.ParseInt(params[1], 10, 64)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("topicType转为int64失败")
                    continue
                }

                topicId := params[2]

                changeNum, err := strconv.ParseInt(v, 10, 64)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("changeNum转为int64失败")
                    continue
                }

                //先查是否有数据
                dbInfo, err := svcCtx.ProduceCountModel.FindOneByParam(topicType, topicId)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("查询点赞数据是否存在失败")
                    continue
                }

                if changeNum == 0 {
                    continue
                }

                if dbInfo.Id <= 0 {
                    insertData := producecount.ProduceCount{
                        AppId:     appId,
                        TopicId:   topicId,
                        TopicType: topicType,
                        PraiseNum: changeNum,
                    }
                    _, err = svcCtx.ProduceCountModel.Insert(&insertData)
                    if err != nil {
                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("插入点赞数据失败")
                        continue
                    }
                } else {
                    updateData := producecount.ProduceCount{
                        Id:        dbInfo.Id,
                        PraiseNum: changeNum,
                    }
                    err = svcCtx.ProduceCountModel.UpdateNum(&updateData)
                    if err != nil {
                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("插入点赞数据失败")
                        continue
                    }

                }

                // 更新评论点赞数
                if topicType == cfgstatus.UserBehaviorCommentType {
                    id ,err := strconv.ParseInt(topicId,10,64)
                    if err != nil {
                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("strconv.ParseInt.err:" + err.Error())
                        continue
                    }
                    // 更新点赞数
                    rows, err := svcCtx.CommentModel.UpdatePraiseNum(id,changeNum)
                    if err != nil {
                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("UpdatePraiseNum更新评论点赞数失败")
                        continue
                    }

                    if rows == 0 {
                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("更新评论点赞数失败")
                        continue
                    }
                }

                appIdStr := fmt.Sprintf("%d", appId)
                topicTypeStr := fmt.Sprintf("%d", topicType)
                field := appIdStr + "|" + topicTypeStr + "|" + topicId
                countNumShowKey := fmt.Sprintf(cfgredis.UserBehaviorProduceCountShow, field)
                svcCtx.Redis.Del(countNumShowKey)
            }
        }
    }
    return
}

func userBehaviorUpdateShareNum(svcCtx *svc.ServiceContext) (err error) {
    today := time.Now().Format("20060102")
    sep := "|"

    userShareNumKey := fmt.Sprintf(cfgredis.UserBehaviorShareNum, today)
    all := svcCtx.Redis.HGetAll(userShareNumKey)
    if all.Err() != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("获取redis中分享数据失败")
        return errors.New("获取redis中分享数据失败")
    }

    if all != nil {
        if len(all.Val()) > 0 {
            for k, v := range all.Val() {
                params := strings.Split(k, sep)
                if len(params) != 3 {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("all的key拆分校验失败")
                    continue
                }

                if v == "" {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("all的val为0")
                    continue
                }

                appId, err := strconv.ParseInt(params[0], 10, 64)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("appId转为int64失败")
                    continue
                }

                topicType, err := strconv.ParseInt(params[1], 10, 64)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("topicType转为int64失败")
                    continue
                }

                topicId := params[2]

                changeNum, err := strconv.ParseInt(v, 10, 64)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("changeNum转为int64失败")
                    continue
                }

                //先查是否有数据
                dbInfo, err := svcCtx.ProduceCountModel.FindOneByParam(topicType, topicId)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("查询分享数据是否存在失败")
                    continue
                }

                if changeNum == 0 {
                    continue
                }

                if dbInfo.Id <= 0 {
                    insertData := producecount.ProduceCount{
                        AppId:     appId,
                        TopicId:   topicId,
                        TopicType: topicType,
                        ShareNum:  changeNum,
                    }
                    _, err = svcCtx.ProduceCountModel.Insert(&insertData)
                    if err != nil {
                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("插入分享数据失败")
                        continue
                    }
                } else {
                    updateData := producecount.ProduceCount{
                        Id:       dbInfo.Id,
                        ShareNum: changeNum,
                    }
                    err = svcCtx.ProduceCountModel.UpdateNum(&updateData)
                    if err != nil {
                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("插入分享数据失败")
                        continue
                    }

                }

                appIdStr := fmt.Sprintf("%d", appId)
                topicTypeStr := fmt.Sprintf("%d", topicType)
                field := appIdStr + "|" + topicTypeStr + "|" + topicId
                countNumShowKey := fmt.Sprintf(cfgredis.UserBehaviorProduceCountShow, field)
                svcCtx.Redis.Del(countNumShowKey)
            }
        }
    }
    return
}

//func userBehaviorUpdateFocusNum(svcCtx *svc.ServiceContext) (err error) {
//    today := time.Now().Format("20060102")
//
//    userFocusNumKey := fmt.Sprintf(cfgredis.UserBehaviorFocusNum, today)
//    all := svcCtx.Redis.HGetAll(userFocusNumKey)
//    if all.Err() != nil {
//        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
//        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("获取redis中关注数据失败")
//        return errors.New("获取redis中关注数据失败")
//    }
//
//    if all != nil {
//        if len(all.Val()) > 0 {
//            for k, v := range all.Val() {
//                if k == "" {
//                    continue
//                }
//
//                uin := k
//
//                changeNum, err := strconv.ParseInt(v, 10, 64)
//                if err != nil {
//                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
//                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("changeNum转为int64失败")
//                    continue
//                }
//
//                if changeNum == 0 {
//                    continue
//                }
//
//                //先查是否有数据
//                dbInfo, err := svcCtx.UserCountModel.FindOneByUin(uin)
//                if err != nil {
//                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
//                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("查询关注数据是否存在失败")
//                    continue
//                }
//
//                if dbInfo.Id <= 0 {
//                    insertData := usercount.UserCount{
//                        Uin:      uin,
//                        FocusNum: changeNum,
//                    }
//                    _, err = svcCtx.UserCountModel.Insert(&insertData)
//                    if err != nil {
//                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
//                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("插入关注数据失败")
//                        continue
//                    }
//                } else {
//                    updateData := usercount.UserCount{
//                        Id:       dbInfo.Id,
//                        FocusNum: changeNum,
//                    }
//                    err = svcCtx.UserCountModel.UpdateNum(&updateData)
//                    if err != nil {
//                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
//                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("插入关注数据失败")
//                        continue
//                    }
//                }
//
//                countNumShowKey := fmt.Sprintf(cfgredis.UserBehaviorCountNumShow, uin)
//                svcCtx.Redis.Del(countNumShowKey)
//            }
//        }
//    }
//
//    return
//}
//
//func userBehaviorUpdateFollowNum(svcCtx *svc.ServiceContext) (err error) {
//    today := time.Now().Format("20060102")
//
//    userFollowNumKey := fmt.Sprintf(cfgredis.UserBehaviorFollowNum, today)
//    all := svcCtx.Redis.HGetAll(userFollowNumKey)
//    if all.Err() != nil {
//        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
//        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf(cfglogs.LogPrefix, "CRONTAB", "handler", "userBehaviorUpdateFollowNum", "获取redis中粉丝数据失败", "", all.Err())
//        return errors.New("获取redis中粉丝数据失败")
//    }
//
//    if all != nil {
//        if len(all.Val()) > 0 {
//            for k, v := range all.Val() {
//                if k == "" {
//                    continue
//                }
//
//                uin := k
//
//                changeNum, err := strconv.ParseInt(v, 10, 64)
//                if err != nil {
//                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
//                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf(cfglogs.LogPrefix, "CRONTAB", "handler", "userBehaviorUpdateFollowNum", "changeNum转为int64失败", v, err.Error())
//                    continue
//                }
//
//                //先查是否有数据
//                dbInfo, err := svcCtx.UserCountModel.FindOneByUin(uin)
//                if err != nil {
//                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
//                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf(cfglogs.LogPrefix, "CRONTAB", "handler", "userBehaviorUpdateFollowNum", "查询粉丝数据是否存在失败", "", err.Error())
//                    continue
//                }
//
//                if changeNum == 0 {
//                    continue
//                }
//
//                if dbInfo.Id <= 0 {
//                    insertData := usercount.UserCount{
//                        Uin:       uin,
//                        FollowNum: changeNum,
//                    }
//                    _, err = svcCtx.UserCountModel.Insert(&insertData)
//                    if err != nil {
//                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
//                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf(cfglogs.LogPrefix, "CRONTAB", "handler", "userBehaviorUpdateFollowNum", "插入粉丝数据失败", insertData, err.Error())
//                        continue
//                    }
//                } else {
//                    updateData := usercount.UserCount{
//                        Id:        dbInfo.Id,
//                        FollowNum: changeNum,
//                    }
//                    err = svcCtx.UserCountModel.UpdateNum(&updateData)
//                    if err != nil {
//                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
//                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf(cfglogs.LogPrefix, "CRONTAB", "handler", "userBehaviorUpdateFollowNum", "插入粉丝数据失败", updateData, err.Error())
//                        continue
//                    }
//                }
//
//                countNumShowKey := fmt.Sprintf(cfgredis.UserBehaviorCountNumShow, uin)
//                svcCtx.Redis.Del(countNumShowKey)
//            }
//        }
//    }
//
//    return
//}
