package common

import (
	"context"
	"encoding/json"
	"fmt"
	red "github.com/go-redis/redis"
	"github.com/zeromicro/go-zero/core/logx"
	"minicode.com/sirius/go-back-server/config/cfgredis"
	"minicode.com/sirius/go-back-server/crontab/app/model/worksmgo"
	"minicode.com/sirius/go-back-server/crontab/app/svc"
	"time"
)

type WorksCommon struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

//  作品
func NewWorksCommon(ctx context.Context, svcCtx *svc.ServiceContext) *WorksCommon {
	return &WorksCommon{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (this *WorksCommon) GetWorkById(id string) (work *worksmgo.Works, err error) {

	expire := 7 * 24 * time.Hour
	workInfoKey := fmt.Sprintf(cfgredis.WorksInfo, id)

	info, redisErr := this.svcCtx.Redis.Get(workInfoKey).Result()
	if redisErr != nil && redisErr != red.Nil {
		return nil, redisErr
	}

	if info != "" {
		work = new(worksmgo.Works)
		err = json.Unmarshal([]byte(info), work)
		if err != nil {
			return
		}
		return
	}

	work, err = this.svcCtx.MgoWorksModel.FindOne(this.ctx, id)
	if err != nil && err != worksmgo.ErrNotFound {
		return nil, err
	}

	if work != nil {

		infoJsonMarshal, err := json.Marshal(work)
		if err != nil {
		} else {
			this.svcCtx.Redis.Set(workInfoKey, string(infoJsonMarshal), expire)
		}
	}

	return
}
