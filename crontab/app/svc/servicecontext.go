package svc

import (
    "errors"
    "fmt"
    "github.com/sirupsen/logrus"
    "minicode.com/sirius/go-back-server/config/cfgtables"
    "minicode.com/sirius/go-back-server/crontab/app/model/specialmgo"
    "minicode.com/sirius/go-back-server/crontab/app/model/userfocus"
    "minicode.com/sirius/go-back-server/crontab/app/model/userfocusmsglog"
    "minicode.com/sirius/go-back-server/crontab/app/model/usermgo"
    "minicode.com/sirius/go-back-server/crontab/app/model/worksmgo"
    "minicode.com/sirius/go-back-server/utils/mylogrus"
    "minicode.com/sirius/go-back-server/utils/rmq/rabbitmq"

    red "github.com/go-redis/redis"
    "github.com/zeromicro/go-zero/core/stores/sqlx"

    "minicode.com/sirius/go-back-server/crontab/app/config"
    "minicode.com/sirius/go-back-server/crontab/app/model/coingoods"
    "minicode.com/sirius/go-back-server/crontab/app/model/comment"
    "minicode.com/sirius/go-back-server/crontab/app/model/producecount"
    "minicode.com/sirius/go-back-server/crontab/app/model/usercount"
)

type ServiceContext struct {
    Config config.Config

    Redis *red.Client

    RmqCommentConn              *rabbitmq.RabbitMQ
    CommentReviewResultRmqQConn *rabbitmq.RabbitMQ // 评论审核通知
    SendMessageRmqQConn         *rabbitmq.RabbitMQ // 发送消息通知
    RmqDataReportConn           *rabbitmq.RabbitMQ // 数据上报
    AddFocusRmqQConn            *rabbitmq.RabbitMQ

    CoinGoodsModel       coingoods.CoinGoodsModel
    CommentModel         comment.CommentModel
    UserCountModel       usercount.UserCountModel
    ProduceCountModel    producecount.ProduceCountModel
    UserFocusModel       userfocus.UserFocusModel
    UserFocusMsgLogModel userfocusmsglog.UserFocusMsgLogModel
    MgoAccountsModel     usermgo.UserModel

    MgoWorksModel   worksmgo.WorksModel
    MgoSpecialModel specialmgo.SpecialModel

    MyLogger *logrus.Logger
}

func NewServiceContext(c config.Config) *ServiceContext {
    mylogger, err := mylogrus.InitLogger(c.Logrus.Path, c.Logrus.ServiceName, c.Logrus.Level, c.Logrus.RotationTime, c.Logrus.KeepDays)
    if err != nil {
        fmt.Println(err.Error())
        panic(errors.New("init logrus log err"))
    }

    rmqCommentConn, err := rabbitmq.NewRabbitMQSimple(c.Rmq.User, c.Rmq.Pwd, c.Rmq.Host, c.Rmq.Port, c.CommentRmqMsgQ.QuName)
    if err != nil {
        fmt.Println(err.Error())
        panic(errors.New("rmqMsgConn is err"))
    }
    commentReviewResultRmqQConn, err := rabbitmq.NewRabbitMQSimple(c.Rmq.User, c.Rmq.Pwd, c.Rmq.Host, c.Rmq.Port, c.CommentReviewResultRmqQ.QuName)
    if err != nil {
        fmt.Println(err.Error())
        panic(errors.New("commentReviewResultRmqQConn is err"))
    }

    sendMessageRmqQConn, err := rabbitmq.NewRabbitMQSimple(c.Rmq.User, c.Rmq.Pwd, c.Rmq.Host, c.Rmq.Port, c.SendMessageRmqQ.QuName)
    if err != nil {
        fmt.Println(err.Error())
        panic(errors.New("sendMessageRmqQConn is err"))
    }

    rmqDataReportConn, err := rabbitmq.NewRabbitMQSimple(c.Rmq.User, c.Rmq.Pwd, c.Rmq.Host, c.Rmq.Port, c.RmqDataReportQ.QuName)
    if err != nil {
        fmt.Println(err.Error())
        panic(errors.New("rmqDataReportConn is err"))
    }

    addFocusRmqQConn, err := rabbitmq.NewRabbitMQSimple(c.Rmq.User, c.Rmq.Pwd, c.Rmq.Host, c.Rmq.Port, c.AddFocusRmqQ.QuName)
    if err != nil {
        fmt.Println(err.Error())
        panic(errors.New("sendMessageRmqQConn is err"))
    }

    sqlConn := sqlx.NewMysql(c.Mysql.Datasource)
    if sqlConn == nil {
        panic(errors.New("sqlConn is nil"))
    }

    rdb := red.NewClient(&red.Options{
        Addr:     c.RedisConn.Address,
        Password: c.RedisConn.Pwd, // no password set
        DB:       c.RedisConn.Db,  // use default DB
    })
    if rdb == nil {
        panic(errors.New("rdbConn is nil"))
    }

    mgoConnStr := c.Mongo.Datasource
    if mgoConnStr == "" {
        panic(errors.New("mgoConnStr is nil"))
    }

    return &ServiceContext{
        Config: c,

        Redis: rdb,

        RmqCommentConn:              rmqCommentConn,
        CommentReviewResultRmqQConn: commentReviewResultRmqQConn,
        SendMessageRmqQConn:         sendMessageRmqQConn,
        RmqDataReportConn:           rmqDataReportConn,
        AddFocusRmqQConn:            addFocusRmqQConn,

        //Mysql的链接
        CoinGoodsModel:       coingoods.NewCoinGoodsModel(sqlConn),
        CommentModel:         comment.NewCommentModel(sqlConn),
        UserCountModel:       usercount.NewUserCountModel(sqlConn),
        ProduceCountModel:    producecount.NewProduceCountModel(sqlConn),
        UserFocusModel:       userfocus.NewUserFocusModel(sqlConn),
        UserFocusMsgLogModel: userfocusmsglog.NewUserFocusMsgLogModel(sqlConn),

        //Mgo的链接
        MgoWorksModel:    worksmgo.NewWorksModel(mgoConnStr, cfgtables.WorksTable),
        MgoSpecialModel:  specialmgo.NewSpecialModel(mgoConnStr, cfgtables.SpecialTable),
        MgoAccountsModel: usermgo.NewUserModel(mgoConnStr, cfgtables.AccountsTable),

        MyLogger: mylogger,
    }
}
