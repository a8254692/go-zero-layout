package handler

import (
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "minicode.com/sirius/go-back-server/config/cfgredis"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "minicode.com/sirius/go-back-server/crontab/app/model/producecount"
    "minicode.com/sirius/go-back-server/crontab/app/svc"
)

func TransferWorksDataFromMongo(svcCtx *svc.ServiceContext) {
    pageIndex := 0
    page := 10

    for {
        limit := pageIndex * page

        list, err := svcCtx.MgoWorksModel.FindList(int64(page), int64(limit))
        if err != nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("获取旧列表")
            continue
        }

        if list == nil {
            filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("list is nil")
            continue
        }

        if len(*list) <= 0 {
            break
        }

        for _, v := range *list {
            info := v
            if info.LikeNum <= 0 {
                continue
            }

            insertData := producecount.ProduceCount{
                AppId:     0,
                TopicId:   info.AuthorId,
                TopicType: cfgstatus.UserBehaviorWorkType,
                PraiseNum: int64(info.LikeNum),
            }
            _, err = svcCtx.ProduceCountModel.InsertUpdatePraiseNum(&insertData)
            if err != nil {
                filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("插入评论数据失败")
                continue
            }
        }

        svcCtx.Redis.Set("sir@TransferWorksDataFromMongo", pageIndex, 5*cfgredis.ExpirationDay)
        pageIndex++
    }

    return
}
