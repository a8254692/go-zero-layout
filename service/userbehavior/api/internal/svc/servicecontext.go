package svc

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"

	"minicode.com/sirius/go-back-server/service/userbehavior/api/internal/config"
	"minicode.com/sirius/go-back-server/service/userbehavior/api/internal/middleware"
	"minicode.com/sirius/go-back-server/service/userbehavior/rpc/rpcuserbehavior"
	"minicode.com/sirius/go-back-server/utils/mylogrus"
)

type ServiceContext struct {
	Config    config.Config
	CheckAuth rest.Middleware

	UserBehaviorRpc rpcuserbehavior.RpcUserBehavior

	MyLogger *logrus.Logger
}

func NewServiceContext(c config.Config) *ServiceContext {

	mylogger, err := mylogrus.InitLogger(c.Logrus.Path, c.Logrus.ServiceName, c.Logrus.Level, c.Logrus.RotationTime, c.Logrus.KeepDays)
	if err != nil {
		fmt.Println(err.Error())
		panic(errors.New("init logrus log err"))
	}

	return &ServiceContext{
		Config:    c,
		CheckAuth: middleware.NewCheckAuthMiddleware(c).Handle,

		UserBehaviorRpc: rpcuserbehavior.NewRpcUserBehavior(zrpc.MustNewClient(c.UserBehaviorRpc)),

		MyLogger: mylogger,
	}
}
