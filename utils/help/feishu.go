package help

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	red "github.com/go-redis/redis"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"minicode.com/sirius/go-back-server/config/cfgredis"
)

const FeiShuResponseSuccessCode = 0

type FeiShuFeedbackSendMsgParams struct {
	Content       string
	MsgType       string
	ReceiveId     string
	Authorization string
}

type FeiShuAppAccessToken struct {
	Code              int32  `json:"code"`
	Expire            int32  `json:"expire"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
}

type FeiShuAppUserResp struct {
	Code int32             `json:"code"`
	Msg  string            `json:"msg"`
	Data FeiShuAppUserData `json:"data"`
}

type FeiShuAppUserData struct {
	HasMore bool                    `json:"has_more"`
	Item    []FeiShuAppUserDataItem `json:"items"`
}

// 部分数据
type FeiShuAppUserDataItem struct {
	Name    string `json:"name"`
	OpenId  string `json:"open_id"`
	UnionId string `json:"union_id"`
}

type FeiShuFeedbackSendMsgResp struct {
	Code int32       `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type GetTokenParams struct {
	AppId     string
	AppSecret string
	Redis     *red.Client
}

// 获取token
func GetFeiShuAppAccessToken(appId, appSecret string) (err error, token string) {

	if appId == "" || appSecret == "" {
		return errors.New("params illegal"), ""
	}

	feiShuTokenUrl := "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json; charset=utf-8"

	data := make(map[string]interface{})
	data["app_id"] = appId
	data["app_secret"] = appSecret

	body, err := json.Marshal(data)
	if err != nil {
		logx.Error("GetFeiShuAppAccessToken.json.Marshal.error ,err ", err)
		return
	}

	status, resData := OnPostHttp(feiShuTokenUrl, body, headers)

	if status != http.StatusOK {
		msg := fmt.Sprintf("GetFeiShuAppAccessToken.post.not zeor ,status %d", status)
		return errors.New(msg), token
	}

	res := new(FeiShuAppAccessToken)
	err = json.Unmarshal(resData, res)

	if err != nil {
		return err, token
	}

	if res.Code != FeiShuResponseSuccessCode {
		msg := fmt.Sprintf("GetFeiShuAppAccessToken.res.code not zero %d ", res.Code)
		return errors.New(msg), token
	}

	return nil, res.TenantAccessToken

}

// https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/contact-v3/user/list
func GetFeiShuAppUser(authorization string) (err error, data FeiShuAppUserData) {
	userUrl := "https://open.feishu.cn/open-apis/contact/v3/users"

	headers := make(map[string]string)
	headers["Authorization"] = authorization

	status, resData := OnGetHttp(userUrl, headers)

	if status != http.StatusOK {
		msg := fmt.Sprintf("GetFeiShuAppAccessToken.post.not zeor ,status %d", status)
		return errors.New(msg), data
	}
	fmt.Printf("GetFeiShuAppUser.resData %s \n", string(resData))
	res := new(FeiShuAppUserResp)
	err = json.Unmarshal(resData, res)

	if err != nil {
		return err, data
	}

	if res.Code != FeiShuResponseSuccessCode {
		msg := fmt.Sprintf("GetFeiShuAppUser.res.code not zero %d ", res.Code)
		return errors.New(msg), data
	}

	return nil, res.Data

}

// https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=open_id
// 发送飞书消息
func FeiShuFeedbackMsg(feiShuMsgUrl string, params FeiShuFeedbackSendMsgParams) (err error, ret *FeiShuFeedbackSendMsgResp) {
	fmt.Printf("url %s ,params %+v \n", feiShuMsgUrl, params)

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json; charset=utf-8"
	headers["Authorization"] = params.Authorization

	data := make(map[string]interface{})
	data["receive_id"] = params.ReceiveId
	data["content"] = params.Content
	data["msg_type"] = params.MsgType
	fmt.Printf("headers %+v ,data %+v \n", headers, data)

	body, err := json.Marshal(data)
	if err != nil {
		logx.Error("FeiShuFeedbackMsg.json.Marshal.error ,err ", err)
		return
	}

	status, resData := OnPostHttp(feiShuMsgUrl, body, headers)

	if status != http.StatusOK {
		msg := fmt.Sprintf("FeiShuFeedbackMsg.post.not zeor ,status %d", status)
		return errors.New(msg), nil
	}

	res := new(FeiShuFeedbackSendMsgResp)
	err = json.Unmarshal(resData, res)

	if err != nil {
		logx.Error("FeiShuFeedbackMsg.unmarshal.error ", err)
		return err, nil
	}

	return nil, res
}

// 获取token 对 token 获取token 的再一次封装 ，加了对 redis 的操作
func GetFSFeedbackAppToken(params GetTokenParams) (err error, token string) {

	appId := params.AppId
	appSecret := params.AppSecret

	if len(appId) == 0 || len(appSecret) == 0 {
		msg := fmt.Sprintf("getFSFeedbackAppToken.appId or appSecret illegal ,appId %s ,appSecret %s ", appId, appSecret)
		logx.Error(msg)
		return
	}

	tokenKey := cfgredis.FeiShuFeedbackTokenKey
	cmd := params.Redis.Get(tokenKey)
	token, err = cmd.Result()
	if err != nil && err != redis.Nil {
		logx.Error("getFSFeedbackAppToken.error ", cmd)
		return err, token
	}

	if len(token) == 0 {
		err, token = GetFeiShuAppAccessToken(appId, appSecret)
		params.Redis.Set(tokenKey, token, time.Hour)
	}

	return nil, token

}

// 获取Authorization
func GetAuthorization(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}

// 获取飞书反馈应用的可见用户 -- 对 GetFeiShuAppUser 的封装
func GetFSFeedbackAppUser(params GetTokenParams) (err error, item []FeiShuAppUserDataItem) {
	userKey := cfgredis.FeiShuFeedbackUserKey
	cmd := params.Redis.Get(userKey)
	res, err := cmd.Result()
	if err != nil && err != redis.Nil {
		logx.Error("GetFSFeedbackAppUser.get.error ", err)
		return err, nil
	}

	if len(res) != 0 {

		userInfoArr := make([]FeiShuAppUserDataItem, 0)
		unmarshalErr := json.Unmarshal([]byte(res), &userInfoArr)
		if unmarshalErr != nil {
			logx.Error("GetFSFeedbackAppUser.Unmarshal.error ", unmarshalErr)
			return unmarshalErr, nil
		}

		return nil, userInfoArr
	}

	return RefreshFSFeedbackAppUser(params)
}

// 刷新飞书反馈应用的可见用户
func RefreshFSFeedbackAppUser(params GetTokenParams) (err error, item []FeiShuAppUserDataItem) {
	userKey := cfgredis.FeiShuFeedbackUserKey
	err, token := GetFSFeedbackAppToken(params)
	if err != nil {
		logx.Error("RefreshFSFeedbackAppUser.GetFSFeedbackAppToken.error ", err)
		return
	}

	authorization := GetAuthorization(token)
	err, data := GetFeiShuAppUser(authorization)
	if err != nil {
		logx.Error("RefreshFSFeedbackAppUser.GetFeiShuAppUser.error ", err)
		return err, nil
	}

	if len(data.Item) == 0 {
		logx.Error("RefreshFSFeedbackAppUser.data.Item no data")
		return nil, nil
	}

	marshalData, marshalErr := json.Marshal(data.Item)
	if marshalErr != nil {
		logx.Error("RefreshFSFeedbackAppUser.json.Marshal", marshalErr)
		return marshalErr, nil
	}

	setCmd := params.Redis.Set(userKey, marshalData, time.Hour*24*60) // 存储2个月
	_, err = setCmd.Result()
	if err != nil {
		logx.Error("RefreshFSFeedbackAppUser.Set.error ", err)
		return err, nil
	}

	return nil, data.Item
}
