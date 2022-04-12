package logic

import (
    "context"
    "encoding/json"
    "errors"
    "minicode.com/sirius/go-back-server/config/cfgstatus"
    "minicode.com/sirius/go-back-server/utils/help"
    "minicode.com/sirius/go-back-server/utils/mylogrus"

    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/userBehaviorProto"

    "github.com/zeromicro/go-zero/core/logx"
)

type GetUserFocusStatusLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewGetUserFocusStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserFocusStatusLogic {
    return &GetUserFocusStatusLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 获取用户关注状态
func (l *GetUserFocusStatusLogic) GetUserFocusStatus(in *userBehaviorProto.GetUserFocusStatusReq) (*userBehaviorProto.GetUserFocusStatusResp, error) {
    reqByte, _ := json.Marshal(in)
    reqStr := string(reqByte)

    uin, err := help.GetRpcUinFromCtx(l.ctx)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取uin失败")
        return nil, errors.New("获取uin失败")
    }

    if in.FocusUin == "" {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("校验FocusUin失败")
        return nil, errors.New("校验FocusUin失败")
    }

    var rsStatus int32
    //先获取我与ta的关注状态
    myFocusDbInfo, err := l.svcCtx.UserFocusModel.FindOneByUinFocusUin(uin, in.FocusUin)
    if err != nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
        l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取我的关注状态失败")
        return nil, errors.New("获取我的关注状态失败")
    }

    if myFocusDbInfo.Status == cfgstatus.UserBehaviorMutuallyFocus {
        rsStatus = cfgstatus.UserBehaviorMutuallyFocus
    } else if myFocusDbInfo.Status == cfgstatus.UserBehaviorOneFocus {
        rsStatus = cfgstatus.UserBehaviorOnlyMyFocus
    } else {
        rsStatus = cfgstatus.UserBehaviorNoFocus

        //获取ta与我的关注状态
        heFocusDbInfo, err := l.svcCtx.UserFocusModel.FindOneByUinFocusUin(in.FocusUin, uin)
        if err != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": reqStr, "resp": "", "track_data": ""}
            l.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(l.ctx, filed)).Error("获取ta的关注状态失败")
            return nil, errors.New("获取ta的关注状态失败")
        }
        if heFocusDbInfo.Status == cfgstatus.UserBehaviorOneFocus {
            rsStatus = cfgstatus.UserBehaviorOnlyHeFocus
        }
    }

    return &userBehaviorProto.GetUserFocusStatusResp{
        Status: rsStatus,
    }, nil
}
