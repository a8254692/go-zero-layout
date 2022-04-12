package config

type Config struct {
    Logrus struct {
        Path         string
        ServiceName  string
        Level        uint32
        RotationTime uint32
        KeepDays     uint32
    }

    Mysql struct {
        Datasource string
    }

    Mongo struct {
        Datasource string
    }

    RedisConn struct {
        Address string
        Pwd     string
        Db      int
    }

    Rmq struct {
        User string
        Pwd  string
        Host string
        Port int32
    }

    CommentRmqMsgQ struct {
        QuName string
        RtKey  string
        ExName string
        ExType string
    }

    CommentReviewResultRmqQ struct {
        QuName string
        RtKey  string
        ExName string
        ExType string
    }

    SendMessageRmqQ struct {
        QuName string
        RtKey  string
        ExName string
        ExType string
    }

    CommentReview struct {
        Url  string
        Cmd  string
        Env  int
        From int
        Type int
    }

    RmqDataReportQ struct {
        QuName string
        RtKey  string
        ExName string
        ExType string
    }

    CoinGoodsIncrSection int

    AddFocusRmqQ struct {
        QuName string
        RtKey  string
        ExName string
        ExType string
    }
}
