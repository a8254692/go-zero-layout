package svc

import (
    "errors"
    "fmt"
    "minicode.com/sirius/go-back-server/service/userbehavior/model/specialmgo"
    "minicode.com/sirius/go-back-server/service/userbehavior/model/worksmgo"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    red "github.com/go-redis/redis"
    "github.com/sirupsen/logrus"
    "github.com/zeromicro/go-zero/core/stores/sqlx"

    "minicode.com/sirius/go-back-server/config/cfgtables"
    "minicode.com/sirius/go-back-server/service/userbehavior/model/comment"
    "minicode.com/sirius/go-back-server/service/userbehavior/model/producecount"
    "minicode.com/sirius/go-back-server/service/userbehavior/model/usercount"
    "minicode.com/sirius/go-back-server/service/userbehavior/model/userfocus"
    "minicode.com/sirius/go-back-server/service/userbehavior/model/usermgo"
    "minicode.com/sirius/go-back-server/service/userbehavior/model/userpraise"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/config"
    "minicode.com/sirius/go-back-server/utils/rmq/rabbitmq"
)

type ServiceContext struct {
    Config config.Config

    Redis *red.Client

    UserCountModel    usercount.UserCountModel
    UserFocusModel    userfocus.UserFocusModel
    UserPraiseModel   userpraise.UserPraiseModel
    CommentModel      comment.CommentModel
    ProduceCountModel producecount.ProduceCountModel
    MgoAccountsModel  usermgo.UserModel
    MgoWorksModel     worksmgo.WorksModel
    MgoSpecialModel   specialmgo.SpecialModel

    RmqCommentConn      *rabbitmq.RabbitMQ
    SendMessageRmqQConn *rabbitmq.RabbitMQ // 发送消息通知
    AddFocusRmqQConn    *rabbitmq.RabbitMQ

    MyLogger *logrus.Logger
}

func NewServiceContext(c config.Config) *ServiceContext {
    mylogger, err := mylogrus.InitLogger(c.Logrus.Path, c.Logrus.ServiceName, c.Logrus.Level, c.Logrus.RotationTime, c.Logrus.KeepDays)
    if err != nil {
        fmt.Println(err.Error())
        panic(errors.New("init logrus log err"))
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

    rmqCommentConn, err := rabbitmq.NewRabbitMQSimple(c.Rmq.User, c.Rmq.Pwd, c.Rmq.Host, c.Rmq.Port, c.CommentRmqMsgQ.QuName)
    if err != nil {
        fmt.Println(err.Error())
        panic(errors.New("rmqMsgConn is err"))
    }

    sendMessageRmqQConn, err := rabbitmq.NewRabbitMQSimple(c.Rmq.User, c.Rmq.Pwd, c.Rmq.Host, c.Rmq.Port, c.SendMessageRmqQ.QuName)
    if err != nil {
        fmt.Println(err.Error())
        panic(errors.New("sendMessageRmqQConn is err"))
    }

    addFocusRmqQConn, err := rabbitmq.NewRabbitMQSimple(c.Rmq.User, c.Rmq.Pwd, c.Rmq.Host, c.Rmq.Port, c.AddFocusRmqQ.QuName)
    if err != nil {
        fmt.Println(err.Error())
        panic(errors.New("sendMessageRmqQConn is err"))
    }

    return &ServiceContext{
        Config: c,

        Redis: rdb,

        UserCountModel:    usercount.NewUserCountModel(sqlConn),
        UserFocusModel:    userfocus.NewUserFocusModel(sqlConn),
        UserPraiseModel:   userpraise.NewUserPraiseModel(sqlConn),
        CommentModel:      comment.NewCommentModel(sqlConn),
        ProduceCountModel: producecount.NewProduceCountModel(sqlConn),

        //Mgo的链接
        MgoWorksModel:    worksmgo.NewWorksModel(mgoConnStr, cfgtables.WorksTable),
        MgoAccountsModel: usermgo.NewUserModel(mgoConnStr, cfgtables.AccountsTable),
        MgoSpecialModel:  specialmgo.NewSpecialModel(mgoConnStr, cfgtables.SpecialTable),

        RmqCommentConn:      rmqCommentConn,
        SendMessageRmqQConn: sendMessageRmqQConn,
        AddFocusRmqQConn:    addFocusRmqQConn,

        MyLogger: mylogger,
    }
}
