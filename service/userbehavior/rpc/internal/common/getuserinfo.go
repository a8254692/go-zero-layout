package common

import (
    "context"
    "encoding/json"
    "fmt"
    red "github.com/go-redis/redis"
    "github.com/zeromicro/go-zero/core/logx"
    "minicode.com/sirius/go-back-server/config/cfgredis"
    "minicode.com/sirius/go-back-server/service/userbehavior/model/usermgo"
    "minicode.com/sirius/go-back-server/service/userbehavior/rpc/internal/svc"
    "minicode.com/sirius/go-back-server/utils/mylogrus"
)

type UserInfoCommon struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewUserInfoCommon(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoCommon {
    return &UserInfoCommon{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

func (c *UserInfoCommon) GetUserInfoById(id string) (userInfo *usermgo.User, err error) {
    userInfoKey := fmt.Sprintf(cfgredis.UserInfo, id)

    getCmd := c.svcCtx.Redis.Get(userInfoKey)
    info, redisErr := getCmd.Result()
    if redisErr != nil && redisErr != red.Nil {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        c.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(c.ctx, filed)).Error("获取账号信息失败redis")
        return nil, redisErr
    }

    if info != "" {
        userInfo = new(usermgo.User)
        err = json.Unmarshal([]byte(info), userInfo)
        if err != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            c.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(c.ctx, filed)).Error("账号信息序列化失败")
            return
        }
        if userInfo.AvatarId == 0 {
            userInfo.AvatarId = 1 // 默认头像
        }
        return
    }

    userInfo, err = c.svcCtx.MgoAccountsModel.FindOne(c.ctx, id)
    if err != nil && err != usermgo.ErrNotFound {
        filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
        c.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(c.ctx, filed)).Error("获取账号信息失败db")
        return nil, err
    }

    if userInfo != nil {
        if userInfo.AvatarId == 0 {
            userInfo.AvatarId = 1 // 默认头像
        }

        infoJsonMarshalStr, err := json.Marshal(userInfo)
        if err != nil {
            filed := map[string]interface{}{"sender": "USER-BEHAVIOR-RPC", "code": 0, "uin": "", "req": "", "resp": "", "track_data": ""}
            c.svcCtx.MyLogger.WithFields(mylogrus.GetCommonField(c.ctx, filed)).Error("用户信息json加密失败")
        } else {
            c.svcCtx.Redis.Set(userInfoKey, string(infoJsonMarshalStr), cfgredis.Expiration60M)
        }
    }

    return
}
