package main

import (
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/config"
	"minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/server"
	"minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
	"minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"
)

var configFile = flag.String("f", "etc/rpcuserbehavior.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)
	srv := server.NewRpcUserBehaviorServer(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		userBehaviorProto.RegisterRpcUserBehaviorServer(grpcServer, srv)

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
