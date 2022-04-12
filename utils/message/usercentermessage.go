package message

import (
    "encoding/json"
    "errors"
    "github.com/zeromicro/go-zero/core/logx"
    "minicode.com/sirius/go-back-server/utils/rmq/rabbitmq"
)

// 发送到用户中心的消息

// 发送消息rmq
type SendMessageRmqQ struct {
    Type    int                    `json:"type"`    // 21关注 22评论 23评论未通过 24评论审核通过
    Uin     string                 `json:"uin"`     // 消息接受人
    Content string                 `json:"content"` // 内容
    From    SendMessageRmqQFrom    `json:"from"`
    Extra   SendMessageRmqQExtra   `json:"extra"`
    MobPush SendMessageRmqQMobPush `json:"mobPush"`
}

type SendMessageRmqQFrom struct {
    AuthorId string `json:"authorId"` // 消息发送人
}

type SendMessageRmqQExtra struct {
    TopicType int64  `json:"topicType"` // 1作品 2专题
    TopicId   string `json:"topicId"`   // 1作品 2专题
}

// 3.6 手机消息推送配置 作品被评论 & 被关注
//mobPush: {
//title,  // 标题
//content,  // 内容
//nextType, // 0 打开首页 1 link跳转 2 scheme 跳转 (默认为0)
//url // 应用内跳转协议
//}
type SendMessageRmqQMobPush struct {
    Title    string `json:"title"`    // 标题
    Content  string `json:"content"`  // 内容
    NextType int32  `json:"nextType"` // 0 打开首页 1 link跳转 2 scheme 跳转 (默认为0)
    Url      string `json:"url"`      // 应用内跳转协议
}

// 发送消息到用户中心
func SendMessageUserCenter(messageRmq *rabbitmq.RabbitMQ, message SendMessageRmqQ) (err error) {

    if messageRmq == nil {
        return errors.New("messageRmq pointer is nil ")
    }

    msg, err := json.Marshal(message)
    if err != nil {
        return
    }
    logx.Infof("SendMessageUserCenter.msg %s ", string(msg))
    return messageRmq.PublishSimple(string(msg))
}
