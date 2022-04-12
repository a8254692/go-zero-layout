package middleware

import (
    "bytes"
    "context"
    "crypto/md5"
    "encoding/hex"
    "encoding/json"
    "errors"
    "io/ioutil"
    "net/http"
    "reflect"
    "strconv"

    "github.com/zeromicro/go-zero/core/logx"
    "github.com/zeromicro/go-zero/core/service"
    "github.com/zeromicro/go-zero/rest/httpx"

    "minicode.com/sirius/go-back-server/config/cfginit"
    "minicode.com/sirius/go-back-server/utils/errorx"
    "minicode.com/sirius/go-back-server/utils/help"
)

type authInfoGeneral struct {
    Uin  string      `json:"uin"`
    Time interface{} `json:"time"`
    S2t  interface{} `json:"s2t"`
    Sign string      `json:"sign"`
}

type authInfo struct {
    Uin  string `json:"uin"`
    Time int64  `json:"time"`
    S2t  int64  `json:"s2t"`
    Sign string `json:"sign"`
}

func GetCheckAuthFun(mod string, next http.HandlerFunc) func(w http.ResponseWriter, r *http.Request) {
    var authKey string
    switch mod {
    case service.DevMode, service.TestMode, service.PreMode:
        authKey = cfginit.AuthKeyTest
    case service.ProMode:
        authKey = cfginit.AuthKeyPro
    default:
        authKey = cfginit.AuthKeyPro
    }

    return func(w http.ResponseWriter, r *http.Request) {
        var uin string
        var timeInt64 int64
        var s2tInt64 int64
        var sign string

        //先获取head中是否存在Authorization
        hAuthorization := r.Header.Get("Authorization")
        if hAuthorization != "" {
            hAuthorizationByte := []byte(hAuthorization)

            sign = string(hAuthorizationByte[len(hAuthorizationByte)-32:])
            s2tStr := string(hAuthorizationByte[len(hAuthorizationByte)-43 : len(hAuthorizationByte)-33])
            s2tInt64, _ = strconv.ParseInt(s2tStr, 10, 64)
            timeStr := string(hAuthorizationByte[len(hAuthorizationByte)-54 : len(hAuthorizationByte)-44])
            timeInt64, _ = strconv.ParseInt(timeStr, 10, 64)
            uin = string(hAuthorizationByte[:len(hAuthorizationByte)-55])
        } else {
            body, err := ioutil.ReadAll(r.Body)
            if err != nil {
                logx.Errorf("鉴权读取BODY参数失败")
                httpx.Error(w, errorx.NewDefaultError("鉴权读取参数失败"))
                return
            }

            var ag authInfoGeneral
            err = json.Unmarshal(body, &ag)
            if err != nil {
                logx.Errorf("鉴权参数BODY解码失败", err.Error())
                httpx.Error(w, errorx.NewDefaultError("鉴权参数解码失败"))
                return
            }

            if ag.S2t == nil || ag.Time == nil || ag.Uin == "" || ag.Sign == "" {
                logx.Errorf("鉴权参数校验无效")
                httpx.Error(w, errorx.NewSystemUserAuthParamError("鉴权参数校验无效"))
                return
            }

            s2tInt64, err = typeInterfaceToInt64(ag.S2t)
            if err != nil {
                logx.Errorf("鉴权S2t转换失败", err.Error())
                httpx.Error(w, errorx.NewSystemUserAuthChangeTypeError("鉴权类型校验失败"))
                return
            }

            timeInt64, err = typeInterfaceToInt64(ag.Time)
            if err != nil {
                logx.Errorf("鉴权time转换失败", err.Error())
                httpx.Error(w, errorx.NewSystemUserAuthChangeTypeError("鉴权类型校验失败"))
                return
            }
            uin = ag.Uin
            sign = ag.Sign

            //body回写
            r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
        }

        if timeInt64 <= 0 || s2tInt64 <= 0 || uin == "" || sign == "" {
            logx.Errorf("鉴权转换参数校验无效")
            httpx.Error(w, errorx.NewSystemUserAuthParamError("鉴权转换参数校验无效"))
            return
        }

        a := authInfo{
            Uin:  uin,
            Time: timeInt64,
            S2t:  s2tInt64,
            Sign: sign,
        }

        var hash1 = md5.Sum([]byte(a.Uin + authKey + strconv.Itoa(int(a.S2t))))
        var hash2 = md5.Sum([]byte(strconv.Itoa(int(a.Time)) + hex.EncodeToString(hash1[:]) + a.Uin))
        if hex.EncodeToString(hash2[:]) != a.Sign {
            logx.Errorf("鉴权失败")
            httpx.Error(w, errorx.NewSystemUserAuthError("鉴权失败"))
            return
        }

        ctx := context.WithValue(r.Context(), "Uin", a.Uin)
        ctx, _ = help.SetUinToMetadataCtx(ctx, a.Uin)

        // Passthrough to next handler if need
        next(w, r.WithContext(ctx))
    }
}

func typeInterfaceToInt64(in interface{}) (val int64, err error) {
    if in == nil {
        err = errors.New("鉴权类型转换参数为空")
        logx.Errorf("鉴权类型转换参数为空")
    }

    switch reflect.TypeOf(in).Kind() {
    case reflect.Float64:
        val = int64(in.(float64))
    case reflect.String:
        val, err = strconv.ParseInt(in.(string), 10, 64)
        if err != nil {
            logx.Errorf("鉴权类型转换失败", err.Error())
            return
        }
    case reflect.Int64:
        val = in.(int64)
    case reflect.Int32:
        val = int64(in.(int32))
    default:
        err = errors.New("鉴权类型不合法")
        logx.Errorf("鉴权类型不合法")
        return
    }

    return
}
