package logic

import (
    "context"

    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"

    "github.com/zeromicro/go-zero/core/logx"
)

type CheckHealthLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewCheckHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckHealthLogic {
    return &CheckHealthLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

func (l *CheckHealthLogic) CheckHealth(in *userBehaviorProto.CheckHealthRequest) (*userBehaviorProto.CheckHealthReply, error) {
    return &userBehaviorProto.CheckHealthReply{
        Message: "Ok",
    }, nil
}
