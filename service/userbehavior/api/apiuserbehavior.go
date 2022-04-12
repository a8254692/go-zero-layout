package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"

	"minicode.com/sirius/go-back-server/service/userbehavior/api/internal/config"
	"minicode.com/sirius/go-back-server/service/userbehavior/api/internal/handler"
	"minicode.com/sirius/go-back-server/service/userbehavior/api/internal/svc"
	"minicode.com/sirius/go-back-server/utils/errorx"
)

var configFile = flag.String("f", "etc/apiuserbehavior.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	// 自定义错误
	httpx.SetErrorHandler(func(err error) (int, interface{}) {
		switch e := err.(type) {
		case *errorx.CodeError:
			return http.StatusOK, e.Data()
		default:
			logx.Errorf("SetErrorHandler Err:%s Stack:", err.Error(), string(debug.Stack()))
			initErr := errorx.NewDefaultError("系统错误")
			return http.StatusInternalServerError, initErr.(*errorx.CodeError).Data()
		}
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
