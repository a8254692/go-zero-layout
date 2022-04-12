package handler

import (
    "errors"
    "fmt"
    "math/rand"
    "minicode.com/sirius/go-back-server/utils/mylogrus"
    "time"

    "minicode.com/sirius/go-back-server/config/cfgredis"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "minicode.com/sirius/go-back-server/crontab/app/svc"
)

func Guidance(svcCtx *svc.ServiceContext) (err error) {
    max := svcCtx.Config.CoinGoodsIncrSection
    if max <= 0 {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("获取区间最大值失败")
        return errors.New("获取区间最大值失败")
    }

    all, err := svcCtx.CoinGoodsModel.FindAll()
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("获取商品列表失败")
        return err
    }

    rand.Seed(time.Now().UnixNano())

    for _, item := range *all {
        if item.Id > 0 {
            x := rand.Intn(max) + 1 //生成1-N随机整数

            err := svcCtx.CoinGoodsModel.UpdateShowExchangeNum(item.Id, int64(x))
            if err != nil {
                filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
                svcCtx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Error("更新展示库存失败")
            }
        }
    }

    goodsListEntityKey := fmt.Sprintf(cfgredis.IntegrateGoodsList, cfgstatus.CoinGoodsTypeEntity)
    svcCtx.Redis.Del(goodsListEntityKey)
    goodsListVirtualKey := fmt.Sprintf(cfgredis.IntegrateGoodsList, cfgstatus.CoinGoodsTypeVirtual)
    svcCtx.Redis.Del(goodsListVirtualKey)

    return
}
