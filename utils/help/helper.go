package help

import (
    "context"
    "database/sql"
    "errors"
    "math/rand"
    "strings"
    "time"

    clock "github.com/AlpacaLabs/go-timestamp"
    "github.com/golang/protobuf/ptypes/timestamp"
    "github.com/zeromicro/go-zero/core/logx"
    "google.golang.org/grpc/metadata"

    "minicode.com/sirius/go-back-server/config/cfglogs"
)

func GetUinFromCtx(ctx context.Context) (uin string, err error) {
    uinInter := ctx.Value("Uin")
    if uinInter == nil {
        logx.WithContext(ctx).Errorf(cfglogs.LogPrefix, "API", "utils", "GetUinFromCtx", "ctx uin interface is nil", "", "uinInter != nil")
        err = errors.New("ctx uin interface is nil")
        return
    }

    uin = uinInter.(string)
    if uin == "" {
        logx.WithContext(ctx).Errorf(cfglogs.LogPrefix, "API", "utils", "GetUinFromCtx", "uin is nil", "", "uin == nil")
        err = errors.New("uin is nil")
        return
    }

    return
}

func SetUinToMetadataCtx(ctx context.Context, uin string) (rsCtx context.Context, err error) {
    if uin == "" {
        logx.WithContext(ctx).Errorf(cfglogs.LogPrefix, "API", "utils", "SetUinToMetadataCtx", "uin is nil", "", "uin is nil")
        err = errors.New("uin is nil")
        return
    }

    md := metadata.Pairs("rUin", uin)
    rsCtx = metadata.NewOutgoingContext(ctx, md)

    return
}

// GetRpcUinFromCtx 从metadata中获取Uin (需要上层ctx透传)
func GetRpcUinFromCtx(ctx context.Context) (uin string, err error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        logx.WithContext(ctx).Errorf(cfglogs.LogPrefix, "RPC", "utils", "GetUinFromCtx", "metadata fromIncomingContext is false", "", "metadata.FromIncomingContext is false")
        err = errors.New("metadata fromIncomingContext is false")
        return
    }

    uinArr := md.Get("rUin")
    if len(uinArr) < 0 {
        logx.WithContext(ctx).Errorf(cfglogs.LogPrefix, "RPC", "utils", "GetUinFromCtx", "uinArr len < 0", "", "uinArr len < 0")
        err = errors.New("uinArr is nil")
        return
    }

    uin = uinArr[0]
    if uin == "" {
        logx.WithContext(ctx).Errorf(cfglogs.LogPrefix, "RPC", "utils", "GetUinFromCtx", "uin is nil", "", "uin is nil")
        err = errors.New("uin is nil")
        return
    }

    return
}

func GetTodayTimeRemaining() time.Duration {
    todayLast := time.Now().Format("2006-01-02") + " 23:59:59"

    todayLastTime, _ := time.ParseInLocation("2006-01-02 15:04:05", todayLast, time.Local)

    remainSecond := time.Duration(todayLastTime.Unix()-time.Now().Local().Unix()) * time.Second

    return remainSecond
}

func GetPagingParam(pageIndex int64, pageSize int64) (limit int64, offset int64) {
    if pageSize <= 0 {
        limit = 10
    } else {
        limit = pageSize
    }

    if pageIndex <= 1 {
        offset = 0
    } else {
        offset = (pageIndex - 1) * pageSize
    }

    return
}

func TimestampToSqlNullTime(pbTime *timestamp.Timestamp) (rs sql.NullTime, err error) {
    if pbTime == nil {
        err = errors.New("helper timestampToSqlNullTime pbTime is nil")
        return rs, err
    }

    pbCreatedAt := clock.TimestampToTime(pbTime)
    rs = sql.NullTime{
        Time:  pbCreatedAt,
        Valid: true,
    }

    return
}

func StringToSqlNullSting(str string) (rs sql.NullString, err error) {
    if str == "" {
        err = errors.New("helper stringToSqlNullSting str is nil")
        return rs, err
    }

    rs = sql.NullString{
        String: str,
        Valid:  true,
    }

    return
}

func Int64ToSqlNullInt64(int int64) (rs sql.NullInt64, err error) {
    if int <= 0 {
        err = errors.New("helper int64ToSqlNullInt64 int <= 0 ")
        return rs, err
    }

    rs = sql.NullInt64{
        Int64: int,
        Valid: true,
    }

    return
}

func TimeToSqlNullTime(time time.Time) (rs sql.NullTime, err error) {
    rs = sql.NullTime{
        Time:  time,
        Valid: true,
    }

    return
}

func GetRandstring(length int) string {
    if length < 1 {
        return ""
    }
    char := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    charArr := strings.Split(char, "")
    charlen := len(charArr)
    ran := rand.New(rand.NewSource(time.Now().Unix()))

    var rchar string = ""
    for i := 1; i <= length; i++ {
        rchar = rchar + charArr[ran.Intn(charlen)]
    }
    return rchar
}
