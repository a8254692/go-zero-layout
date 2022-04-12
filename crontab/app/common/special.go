package common

import (
	"context"
	"encoding/json"
	"fmt"
	red "github.com/go-redis/redis"
	"github.com/zeromicro/go-zero/core/logx"
	"minicode.com/sirius/go-back-server/config/cfgredis"
	"minicode.com/sirius/go-back-server/crontab/app/model/specialmgo"
	"minicode.com/sirius/go-back-server/crontab/app/svc"
	"time"
)

type SpecialCommon struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

//  专题
func NewSpecialCommon(ctx context.Context, svcCtx *svc.ServiceContext) *SpecialCommon {
	return &SpecialCommon{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (this *SpecialCommon) GetSpecialInfoById(id string) (special *specialmgo.Special, err error) {

	expire := 7 * 24 * time.Hour
	specialInfoKey := fmt.Sprintf(cfgredis.SpecialInfo, id)

	info, redisErr := this.svcCtx.Redis.Get(specialInfoKey).Result()
	if redisErr != nil && redisErr != red.Nil {
		return nil, redisErr
	}

	if info != "" {
		special = new(specialmgo.Special)
		err = json.Unmarshal([]byte(info), special)
		if err != nil {
			return
		}
		return
	}

	special, err = this.svcCtx.MgoSpecialModel.FindOne(this.ctx, id)
	if err != nil && err != specialmgo.ErrNotFound {
		return nil, err
	}

	if special != nil {

		infoJsonMarshal, err := json.Marshal(special)
		if err != nil {
		} else {
			this.svcCtx.Redis.Set(specialInfoKey, string(infoJsonMarshal), expire)
		}
	}

	return
}
