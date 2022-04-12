package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
    zrpc.RpcServerConf

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

    RedisConn struct {
        Address string
        Pwd     string
        Db      int
    }

    Mongo struct {
        Datasource string
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

    SendMessageRmqQ struct {
        QuName string
        RtKey  string
        ExName string
        ExType string
    }

    AddFocusRmqQ struct {
        QuName string
        RtKey  string
        ExName string
        ExType string
    }
}
