package consume

import (
    "encoding/json"
    "errors"
    "fmt"
    "minicode.com/sirius/go-back-server/crontab/app/common"
    "minicode.com/sirius/go-back-server/crontab/app/model/usercount"
    "minicode.com/sirius/go-back-server/crontab/app/model/userfocusmsglog"
    "strconv"

    red "github.com/go-redis/redis"
    "minicode.com/sirius/go-back-server/config/cfgmsg"
    "minicode.com/sirius/go-back-server/config/cfgredis"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "minicode.com/sirius/go-back-server/crontab/app/model/userfocus"
    "minicode.com/sirius/go-back-server/crontab/app/svc"
    "minicode.com/sirius/go-back-server/utils/message"
    "minicode.com/sirius/go-back-server/utils/mylogrus"
    "minicode.com/sirius/go-back-server/utils/tool"
)

//消息队列关注详情参数
type AddFocusReq struct {
    OpType   int64  `json:"opType"`
    AppId    int64  `json:"appId"`
    Uin      string `json:"uin"`
    FocusUin string `json:"focusUin"`
}

// 直接在main 函数中加入异常捕获，并不能阻止程序崩溃
func AddFocusFromRmq(svcCtx *svc.ServiceContext) {
    defer func() {
        if err := recover(); err != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("[panic] err: %v ,stack: %s \n", err, tool.GetCurrentGoroutineStack())
        }
    }()

    msg, err := svcCtx.AddFocusRmqQConn.ConsumeSimple()
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("AddFocusFromRmq.ConsumeSimple.error ", err)
        return
    }

    //启用协程处理
    go func() {
        for d := range msg {
            if len(d.Body) != 0 {
                focus := new(AddFocusReq)
                err = json.Unmarshal(d.Body, focus)
                if err != nil {
                    filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": string(d.Body), "resp": err.Error(), "track_data": ""}
                    svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("AddFocus.Unmarshal.error")
                    continue
                }

                if focus.OpType == cfgstatus.UserBehaviorRmqFocusType {
                    err = AddFocus(svcCtx, focus)
                    if err != nil {
                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": string(d.Body), "resp": err.Error(), "track_data": ""}
                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("AddFocus.error")
                        continue
                    }
                } else if focus.OpType == cfgstatus.UserBehaviorRmqCancelFocusType {
                    err = DelFocus(svcCtx, focus)
                    if err != nil {
                        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": string(d.Body), "resp": err.Error(), "track_data": ""}
                        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("DelFocus.error")
                        continue
                    }
                }
            }
        }
    }()
}

// 新增关注详情
func AddFocus(svcCtx *svc.ServiceContext, focus *AddFocusReq) error {
    uin := focus.Uin
    focusUin := focus.FocusUin
    appId := focus.AppId

    if uin == "" || focusUin == "" || uin == focusUin {
        return errors.New("参数校验失败")
    }

    //先查询是否是对方已关注
    myDbInfo, err := svcCtx.UserFocusModel.FindOneByUinFocusUin(uin, focusUin)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": fmt.Sprintf("%s|%s", uin, focusUin), "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("获取是已关注信息失败")
        return errors.New("获取是已关注信息失败")
    }
    if myDbInfo.Id > 0 {
        return errors.New("已关注对方")
    }

    status := cfgstatus.UserBehaviorOneFocus
    //先查询是否是对方已关注
    bothWayDbInfo, err := svcCtx.UserFocusModel.FindOneByUinFocusUin(focusUin, uin)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": fmt.Sprintf("%s|%s", focusUin, uin), "resp": err.Error(), "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("获取对方是否关注失败")
        return errors.New("获取对方是否关注失败")
    }

    if bothWayDbInfo != nil {
        if bothWayDbInfo.Status > cfgstatus.UserBehaviorCanNotFocus {
            status = cfgstatus.UserBehaviorMutuallyFocus
        }
    }

    focusIns := userfocus.UserFocus{
        AppId:    appId,
        Uin:      uin,
        FocusUin: focusUin,
        Status:   int64(status),
    }
    _, err = svcCtx.UserFocusModel.Insert(&focusIns)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": focusIns, "resp": err.Error(), "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("新增关注详情失败")
        return err
    }

    //双向关注则更新下对方的状态
    if bothWayDbInfo.Status == cfgstatus.UserBehaviorOneFocus {
        err = svcCtx.UserFocusModel.UpdateStatus(focusUin, uin, cfgstatus.UserBehaviorMutuallyFocus)
        if err != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": fmt.Sprintf("%s|%s", focusUin, uin), "resp": err.Error(), "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("双向关注则更新下对方的状态失败")
            return err
        }
    }

    //先查是否有数据
    dbInfo, err := svcCtx.UserCountModel.FindOneByUin(uin)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": uin, "resp": err.Error(), "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("查询粉丝数据是否存在失败")
        return err
    }
    if dbInfo.Id <= 0 {
        insertData := usercount.UserCount{
            Uin:      uin,
            FocusNum: 1,
        }
        _, err = svcCtx.UserCountModel.Insert(&insertData)
        if err != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": insertData, "resp": err.Error(), "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("插入关注数据失败")
            return err
        }
    } else {
        err = svcCtx.UserCountModel.UpdateNumIncr(uin, 1, cfgstatus.UserBehaviorOperationAddType)
        if err != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": uin, "resp": err.Error(), "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("增加关注数量失败")
            return errors.New("增加关注数量失败")
        }
    }

    countNumShowKey := fmt.Sprintf(cfgredis.UserBehaviorCountNumShow, uin)
    svcCtx.Redis.Del(countNumShowKey)

    //先查是否有数据
    dbFocusInfo, err := svcCtx.UserCountModel.FindOneByUin(focusUin)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": focusUin, "resp": err.Error(), "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("查询粉丝数据是否存在失败")
        return err
    }
    if dbFocusInfo.Id <= 0 {
        fInsertData := usercount.UserCount{
            Uin:       focusUin,
            FollowNum: 1,
        }
        _, err = svcCtx.UserCountModel.Insert(&fInsertData)
        if err != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": fInsertData, "resp": err.Error(), "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("插入粉丝数据失败")
            return err
        }
    } else {
        err = svcCtx.UserCountModel.UpdateNumIncr(focusUin, 2, cfgstatus.UserBehaviorOperationAddType)
        if err != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": focusUin, "resp": err.Error(), "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("增加粉丝数量失败")
            return errors.New("增加粉丝数量失败")
        }
    }

    countNumFUinShowKey := fmt.Sprintf(cfgredis.UserBehaviorCountNumShow, focusUin)
    svcCtx.Redis.Del(countNumFUinShowKey)

    isSendKey := fmt.Sprintf(cfgredis.UserBehaviorIsSendFocusMsg, uin, focusUin)
    redisCmd := svcCtx.Redis.Get(isSendKey)
    err = redisCmd.Err()
    if err != nil || err == red.Nil {
        var redisValNum int64
        if err == red.Nil {
            dbInfo, err := svcCtx.UserFocusMsgLogModel.FindOneByUinFocusUin(uin, focusUin)
            if err != nil {
                filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": fmt.Sprintf("%s|%s", uin, focusUin), "resp": err.Error(), "track_data": ""}
                svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("查询发送消息记录失败")
                return nil
            }

            if dbInfo.Id > 0 {
                redisValNum = 1
            }

            svcCtx.Redis.Set(isSendKey, redisValNum, cfgredis.ExpirationDay)
        } else {
            redisValNum, _ = strconv.ParseInt(redisCmd.Val(), 10, 64)
        }

        if redisValNum <= 0 {
            //发送关注消息
            sendDataFrom := message.SendMessageRmqQFrom{
                AuthorId: uin,
            }
            sendDataExtra := message.SendMessageRmqQExtra{}

            userInfoStc := common.NewUserInfoCommon(svcCtx)
            userInfo, err := userInfoStc.GetUserInfoById(uin)
            if err != nil {
                filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": uin, "resp": err.Error(), "track_data": ""}
                svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("查询用户信息失败")
            }
            var userName string
            if userInfo != nil {
                userName = userInfo.NickName
            }

            mobPushContent := fmt.Sprintf("%s 关注了你", userName)
            sendDataMobPush := message.SendMessageRmqQMobPush{
                Title:    "收获新粉丝",
                Content:  mobPushContent,
                NextType: cfgstatus.SendMessageNextTypeScheme,
                Url:      cfgstatus.SendMessageUrlMsgCenter,
            }
            sendData := message.SendMessageRmqQ{
                Type:    cfgstatus.SendMessageTypeFocus,
                Uin:     focusUin,
                Content: cfgmsg.FocusMsg,
                From:    sendDataFrom,
                Extra:   sendDataExtra,
                MobPush: sendDataMobPush,
            }
            err = message.SendMessageUserCenter(svcCtx.SendMessageRmqQConn, sendData)
            if err != nil {
                filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": sendData, "resp": err.Error(), "track_data": ""}
                svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("发送关注消息失败")
                return err
            }

            insData := userfocusmsglog.UserFocusMsgLog{
                AppId:    appId,
                Uin:      uin,
                FocusUin: focusUin,
            }
            _, err = svcCtx.UserFocusMsgLogModel.Insert(&insData)
            if err != nil {
                filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": insData, "resp": err.Error(), "track_data": ""}
                svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("插入关注消息发送记录失败")
                return nil
            }

            svcCtx.Redis.Set(isSendKey, 1, cfgredis.ExpirationDay)
        }
    }

    return nil
}

// 新增关注详情
func DelFocus(svcCtx *svc.ServiceContext, focus *AddFocusReq) error {
    uin := focus.Uin
    focusUin := focus.FocusUin
    //appId := focus.AppId

    if uin == "" || focusUin == "" || uin == focusUin {
        return errors.New("参数校验失败")
    }

    //先查询是否是对方已关注
    bothWayDbInfo, err := svcCtx.UserFocusModel.FindOneByUinFocusUin(focusUin, uin)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": fmt.Sprintf("%s|%s", focusUin, uin), "resp": err.Error(), "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("获取对方是否关注失败")
        return errors.New("获取对方是否关注失败")
    }

    //如果对方是双向关注则更新对方状态
    if bothWayDbInfo.Status == cfgstatus.UserBehaviorMutuallyFocus {
        err := svcCtx.UserFocusModel.UpdateStatus(focusUin, uin, cfgstatus.UserBehaviorOneFocus)
        if err != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": fmt.Sprintf("%s|%s", focusUin, uin), "resp": err.Error(), "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("更新对方状态失败")
            return err
        }
    }

    err = svcCtx.UserFocusModel.DeleteByUinFocusUin(uin, focusUin)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": fmt.Sprintf("%s|%s", uin, focusUin), "resp": err.Error(), "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("取消关注详情失败")
        return err
    }

    err = svcCtx.UserCountModel.UpdateNumIncr(uin, 1, cfgstatus.UserBehaviorOperationReduceType)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": uin, "resp": err.Error(), "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("减少关注数量失败")
        return errors.New("减少关注数量失败")
    }

    err = svcCtx.UserCountModel.UpdateNumIncr(focusUin, 2, cfgstatus.UserBehaviorOperationReduceType)
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": focusUin, "resp": err.Error(), "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("减少粉丝数量失败")
        return errors.New("减少粉丝数量失败")
    }

    countNumShowKey := fmt.Sprintf(cfgredis.UserBehaviorCountNumShow, uin)
    svcCtx.Redis.Del(countNumShowKey)

    countNumFUinShowKey := fmt.Sprintf(cfgredis.UserBehaviorCountNumShow, focusUin)
    svcCtx.Redis.Del(countNumFUinShowKey)

    return nil
}
