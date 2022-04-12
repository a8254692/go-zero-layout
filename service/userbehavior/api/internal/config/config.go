package config

import (
    "github.com/zeromicro/go-zero/rest"
    "github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
    rest.RestConf

    Logrus struct {
        Path         string
        ServiceName  string
        Level        uint32
        RotationTime uint32
        KeepDays     uint32
    }

    //用户行为Rpc
    UserBehaviorRpc zrpc.RpcClientConf
}
