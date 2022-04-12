package benchmark

import (
    "context"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "testing"

    "github.com/zeromicro/go-zero/core/conf"
    "github.com/zeromicro/go-zero/zrpc"

    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/rpcuserbehavior"
    "minicode.com/sirius/go-back-server/test/config/apiuserbehavior"
    "minicode.com/sirius/go-back-server/utils/help"
)

var configFile = "../config/apiuserbehavior/apiuserbehavior.local.yaml"

func BenchmarkAddFocus(b *testing.B) {
    var c apiuserbehavior.Config
    conf.MustLoad(configFile, &c)

    UserBehaviorRpc := rpcuserbehavior.NewRpcUserBehavior(zrpc.MustNewClient(c.UserBehaviorRpc))
    ctx := context.Background()
    ctx, _ = help.SetUinToMetadataCtx(ctx, "1000004312")

    b.ResetTimer() // 重置计时器，忽略前面的准备时间

    for n := 0; n < b.N; n++ {
        randStr := help.GetRandstring(8)
        //增加关注关系
        _, _ = UserBehaviorRpc.AddFocus(ctx, &rpcuserbehavior.AddFocusReq{
            AppId:    0,
            FocusUin: randStr,
        })
    }
}

func BenchmarkAddFollowNum(b *testing.B) {
    var c apiuserbehavior.Config
    conf.MustLoad(configFile, &c)

    UserBehaviorRpc := rpcuserbehavior.NewRpcUserBehavior(zrpc.MustNewClient(c.UserBehaviorRpc))
    ctx := context.Background()
    ctx, _ = help.SetUinToMetadataCtx(ctx, "1000004312")

    b.ResetTimer() // 重置计时器，忽略前面的准备时间

    for n := 0; n < b.N; n++ {
        randStr := help.GetRandstring(8)

        _, _ = UserBehaviorRpc.AddFollowNum(ctx, &rpcuserbehavior.AddFollowNumReq{
            AppId:    0,
            FocusUin: randStr,
            Type:     cfgstatus.UserBehaviorOperationReduceType,
        })
    }
}

func BenchmarkAddPraise(b *testing.B) {
    var c apiuserbehavior.Config
    conf.MustLoad(configFile, &c)

    UserBehaviorRpc := rpcuserbehavior.NewRpcUserBehavior(zrpc.MustNewClient(c.UserBehaviorRpc))
    ctx := context.Background()
    ctx, _ = help.SetUinToMetadataCtx(ctx, "1000004312")

    b.ResetTimer() // 重置计时器，忽略前面的准备时间

    for n := 0; n < b.N; n++ {
        randStr := help.GetRandstring(8)

        _, _ = UserBehaviorRpc.AddPraise(ctx, &rpcuserbehavior.AddPraiseReq{
            AppId:     0,
            TopicType: cfgstatus.UserBehaviorWorkType,
            TopicId:   randStr,
        })
    }
}

func BenchmarkAddPraiseNum(b *testing.B) {
    var c apiuserbehavior.Config
    conf.MustLoad(configFile, &c)

    UserBehaviorRpc := rpcuserbehavior.NewRpcUserBehavior(zrpc.MustNewClient(c.UserBehaviorRpc))
    ctx := context.Background()
    ctx, _ = help.SetUinToMetadataCtx(ctx, "1000004312")

    b.ResetTimer() // 重置计时器，忽略前面的准备时间

    for n := 0; n < b.N; n++ {
        randStr := help.GetRandstring(8)

        //调用点赞计数方法
        _, _ = UserBehaviorRpc.AddPraiseNum(ctx, &rpcuserbehavior.AddPraiseNumReq{
            AppId:     0,
            TopicType: cfgstatus.UserBehaviorWorkType,
            TopicId:   randStr,
            Type:      cfgstatus.UserBehaviorOperationAddType,
        })
    }
}

func BenchmarkAddShareNum(b *testing.B) {
    var c apiuserbehavior.Config
    conf.MustLoad(configFile, &c)

    UserBehaviorRpc := rpcuserbehavior.NewRpcUserBehavior(zrpc.MustNewClient(c.UserBehaviorRpc))
    ctx := context.Background()
    ctx, _ = help.SetUinToMetadataCtx(ctx, "1000004312")

    b.ResetTimer() // 重置计时器，忽略前面的准备时间

    for n := 0; n < b.N; n++ {
        randStr := help.GetRandstring(8)

        _, _ = UserBehaviorRpc.AddShareNum(ctx, &rpcuserbehavior.AddShareNumReq{
            AppId:     0,
            TopicType: cfgstatus.UserBehaviorWorkType,
            TopicId:   randStr,
            Type:      cfgstatus.UserBehaviorOperationAddType,
        })
    }
}

func BenchmarkGetFocusFollowList(b *testing.B) {
    var c apiuserbehavior.Config
    conf.MustLoad(configFile, &c)

    UserBehaviorRpc := rpcuserbehavior.NewRpcUserBehavior(zrpc.MustNewClient(c.UserBehaviorRpc))
    ctx := context.Background()
    ctx, _ = help.SetUinToMetadataCtx(ctx, "1000004312")

    b.ResetTimer() // 重置计时器，忽略前面的准备时间

    for n := 0; n < b.N; n++ {

        _, _ = UserBehaviorRpc.GetFocusFollowList(ctx, &rpcuserbehavior.GetFocusFollowListReq{
            AppId:     0,
            PageIndex: 0,
            PageSize:  10,
            Type:      1,
            TargetUin: "1000004312",
        })
    }
}

func BenchmarkGetPraiseList(b *testing.B) {
    var c apiuserbehavior.Config
    conf.MustLoad(configFile, &c)

    UserBehaviorRpc := rpcuserbehavior.NewRpcUserBehavior(zrpc.MustNewClient(c.UserBehaviorRpc))
    ctx := context.Background()
    ctx, _ = help.SetUinToMetadataCtx(ctx, "1000004312")

    b.ResetTimer() // 重置计时器，忽略前面的准备时间

    for n := 0; n < b.N; n++ {

        _, _ = UserBehaviorRpc.GetPraiseList(ctx, &rpcuserbehavior.GetPraiseListReq{
            AppId:     0,
            PageIndex: 0,
            PageSize:  10,
            TargetUin: "1000004312",
        })
    }
}

func BenchmarkGetProduceCount(b *testing.B) {
    var c apiuserbehavior.Config
    conf.MustLoad(configFile, &c)

    UserBehaviorRpc := rpcuserbehavior.NewRpcUserBehavior(zrpc.MustNewClient(c.UserBehaviorRpc))
    ctx := context.Background()
    ctx, _ = help.SetUinToMetadataCtx(ctx, "1000004312")

    b.ResetTimer() // 重置计时器，忽略前面的准备时间

    for n := 0; n < b.N; n++ {
        randStr := help.GetRandstring(8)

        _, _ = UserBehaviorRpc.GetProduceCount(ctx, &rpcuserbehavior.GetProduceCountReq{
            AppId:     0,
            TopicType: cfgstatus.UserBehaviorWorkType,
            TopicId:   randStr,
        })
    }
}

func BenchmarkGetUserCount(b *testing.B) {
    var c apiuserbehavior.Config
    conf.MustLoad(configFile, &c)

    UserBehaviorRpc := rpcuserbehavior.NewRpcUserBehavior(zrpc.MustNewClient(c.UserBehaviorRpc))
    ctx := context.Background()
    ctx, _ = help.SetUinToMetadataCtx(ctx, "1000004312")

    b.ResetTimer() // 重置计时器，忽略前面的准备时间

    for n := 0; n < b.N; n++ {
        _, _ = UserBehaviorRpc.GetUserCount(ctx, &rpcuserbehavior.GetUserCountReq{
            AppId:     0,
            TargetUin: "1000004312",
        })
    }
}

func BenchmarkGetUserFocusStatus(b *testing.B) {
    var c apiuserbehavior.Config
    conf.MustLoad(configFile, &c)

    UserBehaviorRpc := rpcuserbehavior.NewRpcUserBehavior(zrpc.MustNewClient(c.UserBehaviorRpc))
    ctx := context.Background()
    ctx, _ = help.SetUinToMetadataCtx(ctx, "1000004312")

    b.ResetTimer() // 重置计时器，忽略前面的准备时间

    for n := 0; n < b.N; n++ {
        _, _ = UserBehaviorRpc.GetUserFocusStatus(ctx, &rpcuserbehavior.GetUserFocusStatusReq{
            AppId:    0,
            FocusUin: "1000004312",
        })
    }
}

func BenchmarkGetUserIsPraise(b *testing.B) {
    var c apiuserbehavior.Config
    conf.MustLoad(configFile, &c)

    UserBehaviorRpc := rpcuserbehavior.NewRpcUserBehavior(zrpc.MustNewClient(c.UserBehaviorRpc))
    ctx := context.Background()
    ctx, _ = help.SetUinToMetadataCtx(ctx, "1000004312")

    b.ResetTimer() // 重置计时器，忽略前面的准备时间

    for n := 0; n < b.N; n++ {
        randStr := help.GetRandstring(8)

        _, _ = UserBehaviorRpc.GetUserIsPraise(ctx, &rpcuserbehavior.GetUserIsPraiseReq{
            AppId:     0,
            TopicType: cfgstatus.UserBehaviorWorkType,
            TopicId:   randStr,
        })
    }
}
