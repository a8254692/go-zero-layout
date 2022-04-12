package main

import (
    "flag"
    "fmt"
    "minicode.com/sirius/go-back-server/crontab/app/consume"
    "minicode.com/sirius/go-back-server/crontab/app/handler"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "minicode.com/sirius/go-back-server/crontab/app/config"
    "minicode.com/sirius/go-back-server/crontab/app/svc"

    "github.com/robfig/cron"
    "github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/admin.yaml", "the config file")

func main() {
    flag.Parse()

    var cof config.Config
    conf.MustLoad(*configFile, &cof)

    ctx := svc.NewServiceContext(cof)

    myChan := make(chan int64)

    c := cron.New()
    defer c.Stop()

    initCron(c, ctx)

    startConsume(ctx)

    fmt.Printf("Starting app server ...\n")

    c.Start()
    <-myChan
}

func startConsume(ctx *svc.ServiceContext) {
    consume.CommentAutoReview(ctx)
    consume.CommentReviewResult(ctx)
    consume.AddFocusFromRmq(ctx)
}

func initCron(c *cron.Cron, ctx *svc.ServiceContext) {
    // ps: 详解请查看根目录readme
    m5Spec := "0 */5 * * * ?"
    err := c.AddFunc(m5Spec, func() {
        fmt.Println("=================translateIng=================")
    })
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        ctx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("监控检查调度器注册失败:%s", err)
        return
    }

    spec := "0 */1 * * * ?" // 每分钟执行
    err = c.AddFunc(spec, func() {
        //用户行为数据落地
        handler.UserBehavior(ctx)
    })
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        ctx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("监控检查调度器注册失败:%s", err)
        return
    }

    hSpec := "0 0 */1 * * ?" // 每小时执行
    err = c.AddFunc(hSpec, func() {
        _ = handler.Guidance(ctx)
    })
    if err != nil {
        filed := map[string]interface{}{"sender": "CRONTAB-APP", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        ctx.MyLogger.WithFields(mylogrus.GetCommonFieldNoTrace(filed)).Errorf("监控检查调度器注册失败:%s", err)
        return
    }
}
