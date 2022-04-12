package rabbitmq

import (
    "errors"
    "fmt"

    "github.com/streadway/amqp"
    "github.com/zeromicro/go-zero/core/logx"
)

type RabbitMQ struct {
    //连接
    conn *amqp.Connection
    //管道
    channel *amqp.Channel
    //队列名称
    QueueName string
    //交换机
    Exchange string
    //key Simple模式 几乎用不到
    Key string
    //连接信息
    MqUrl string
}

// 申请队列参数
type QueueDeclareParams struct {
    QueueName string // queue 名
    Durable,AutoDelete, Exclusive, NoWait bool
    Args amqp.Table
}

// 消费者参数
type ConsumeSimpleParams struct {
    QueueName, Consumer string
    AutoAck, Exclusive, NoLocal, NoWait bool
    Args amqp.Table
    QueueDeclare QueueDeclareParams
}


// newRabbitMQ 创建RabbitMQ结构体实例
func newRabbitMQ(user string, pwd string, host string, port int32, queueName string, exchange string, key string) (*RabbitMQ, error) {
    var err error

    rabbitUrl := fmt.Sprintf("amqp://%s:%s@%s:%d/sirius", user, pwd, host, port)
    rabbitmq := &RabbitMQ{QueueName: queueName, Exchange: exchange, Key: key, MqUrl: rabbitUrl}
    //创建rabbitmq连接
    rabbitmq.conn, err = amqp.Dial(rabbitmq.MqUrl)
    if err != nil {
        logx.Errorf("创建连接错误")
        return nil, err
    }

    rabbitmq.channel, err = rabbitmq.conn.Channel()
    if err != nil {
        logx.Errorf("获取channel失败")
        return nil, err
    }

    return rabbitmq, nil
}

// Destroy 断开channel和connection
func (r *RabbitMQ) Destroy() error {
    // 先关闭管道,再关闭链接
    err := r.channel.Close()
    if err != nil {
        logx.Errorf("rmq关闭channel失败")
        return errors.New("rmq关闭channel失败")
    }
    err = r.conn.Close()
    if err != nil {
        logx.Errorf("rmq链接关闭失败")
        return errors.New("rmq链接关闭失败")
    }

    return nil
}

// NewRabbitMQSimple  创建简单模式下RabbitMQ实例
func NewRabbitMQSimple(user string, pwd string, host string, port int32, queueName string) (*RabbitMQ, error) {
    rmq, err := newRabbitMQ(user, pwd, host, port, queueName, "", "")
    return rmq, err
}

// NewRabbitMQPubSub  创建订阅模式rabbitmq实例
func NewRabbitMQPubSub(user string, pwd string, host string, port int32, exchangeName string) (*RabbitMQ, error) {
    rmq, err := newRabbitMQ(user, pwd, host, port, "", exchangeName, "")
    return rmq, err
}

// NewRabbitMQTopic  创建话题模式RabbitMQ实例
func NewRabbitMQTopic(user string, pwd string, host string, port int32, exchange string, routingKey string) (*RabbitMQ, error) {
    rmq, err := newRabbitMQ(user, pwd, host, port, "", exchange, routingKey)
    return rmq, err
}

// NewRabbitMQRouting  创建路由模式RabbitMQ实例
func NewRabbitMQRouting(user string, pwd string, host string, port int32, exchange string, routingKey string) (*RabbitMQ, error) {
    rmq, err := newRabbitMQ(user, pwd, host, port, "", exchange, routingKey)
    return rmq, err
}

// PublishSimple 简单模式下生产者
func (r *RabbitMQ) PublishSimple(message string) error {
    //1、申请队列，如果队列存在就跳过，不存在创建
    //优点：保证队列存在，消息能发送到队列中
    _, err := r.channel.QueueDeclare(
        //队列名称
        r.QueueName,
        //是否持久化
        true,
        //是否为自动删除 当最后一个消费者断开连接之后，是否把消息从队列中删除
        false,
        //是否具有排他性 true表示自己可见 其他用户不能访问
        false,
        //是否阻塞 true表示要等待服务器的响应
        false,
        //额外数学系
        nil,
    )
    if err != nil {
        logx.Errorf("rmq申请队列失败", err)
        return err
    }

    //2.发送消息到队列中
    err = r.channel.Publish(
        //默认的Exchange交换机是default,类型是direct直接类型
        r.Exchange,
        //要赋值的队列名称
        r.QueueName,
        //如果为true，根据exchange类型和routkey规则，如果无法找到符合条件的队列那么会把发送的消息返回给发送者
        false,
        //如果为true,当exchange发送消息到队列后发现队列上没有绑定消费者，则会把消息还给发送者
        false,
        //消息
        amqp.Publishing{
            //类型
            ContentType: "text/plain",
            //消息
            Body: []byte(message),
        })

    if err != nil {
        logx.Errorf("rmq发送消息失败", err)
        return err
    }

    return nil
}

//简单模式Step:2、简单模式下消费者
func (r *RabbitMQ) ConsumeSimple() (<-chan amqp.Delivery, error) {
    //1、申请队列，如果队列存在就跳过，不存在创建
    //优点：保证队列存在，消息能发送到队列中
    _, err := r.channel.QueueDeclare(
        //队列名称
        r.QueueName,
        //是否持久化
        true,
        //是否为自动删除 当最后一个消费者断开连接之后，是否把消息从队列中删除
        false,
        //是否具有排他性
        false,
        //是否阻塞
        false,
        //额外数学系
        nil,
    )
    if err != nil {
        return nil, err
    }
    //接收消息
    return r.channel.Consume(
        r.QueueName,
        //用来区分多个消费者
        "",
        //是否自动应答
        true,
        //是否具有排他性
        false,
        //如果设置为true,表示不能同一个connection中发送的消息传递给这个connection中的消费者
        false,
        //队列是否阻塞
        false,
        nil,
    )
}

// ConsumeSimple() 中有些参数是写死的，不支持灵活配置，但是有些地方已经在用了，不好去改，新的用这个
func (r *RabbitMQ) ConsumeSimpleNew(consume ConsumeSimpleParams) (<-chan amqp.Delivery, error) {
    //1、申请队列，如果队列存在就跳过，不存在创建
    //优点：保证队列存在，消息能发送到队列中
    _, err := r.channel.QueueDeclare(
        //队列名称
        consume.QueueDeclare.QueueName,
        //是否持久化
        consume.QueueDeclare.Durable,
        //是否为自动删除 当最后一个消费者断开连接之后，是否把消息从队列中删除
        consume.QueueDeclare.AutoDelete,
        //是否具有排他性
        consume.QueueDeclare.Exclusive,
        //是否阻塞
        consume.QueueDeclare.NoWait,
        //额外数学系
        consume.QueueDeclare.Args,
    )
    if err != nil {
        return nil, err
    }
    //接收消息
    return r.channel.Consume(
        consume.QueueName,
        //用来区分多个消费者
        consume.Consumer,
        //是否自动应答
        consume.AutoAck,
        //是否具有排他性
        consume.Exclusive,
        //如果设置为true,表示不能同一个connection中发送的消息传递给这个connection中的消费者
        consume.NoLocal,
        //队列是否阻塞
        consume.NoWait,
        consume.Args,
    )
}

/* TODO 预留代码先别删，后续有需要再整理

//简单模式Step:2、简单模式下消费者
func (r *RabbitMQ) ConsumeSimple() {
    //1、申请队列，如果队列存在就跳过，不存在创建
    //优点：保证队列存在，消息能发送到队列中
    _, err := r.channel.QueueDeclare(
        //队列名称
        r.QueueName,
        //是否持久化
        false,
        //是否为自动删除 当最后一个消费者断开连接之后，是否把消息从队列中删除
        false,
        //是否具有排他性
        false,
        //是否阻塞
        false,
        //额外数学系
        nil,
    )
    if err != nil {
        fmt.Println(err)
    }
    //接收消息
    msgs, err := r.channel.Consume(
        r.QueueName,
        //用来区分多个消费者
        "",
        //是否自动应答
        true,
        //是否具有排他性
        false,
        //如果设置为true,表示不能同一个connection中发送的消息传递给这个connection中的消费者
        false,
        //队列是否阻塞
        false,
        nil,
    )
    if err != nil {
        fmt.Println(err)
    }
    forever := make(chan bool)

    //启用协程处理
    go func() {
        for d := range msgs {
            //实现我们要处理的逻辑函数
            log.Printf("Received a message:%s", d.Body)
            //fmt.Println(d.Body)
        }
    }()

    log.Printf("【*】warting for messages, To exit press CCTRAL+C")
    <-forever
}


//订阅模式生成
func (r *RabbitMQ) PublishPub(message string) {
    //尝试创建交换机，不存在创建
    err := r.channel.ExchangeDeclare(
        //交换机名称
        r.Exchange,
        //交换机类型 广播类型
        "fanout",
        //是否持久化
        true,
        //是否字段删除
        false,
        //true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
        false,
        //是否阻塞 true表示要等待服务器的响应
        false,
        nil,
    )
    r.failOnErr(err, "failed to declare an excha"+"nge")

    //2 发送消息
    err = r.channel.Publish(
        r.Exchange,
        "",
        false,
        false,
        amqp.Publishing{
            //类型
            ContentType: "text/plain",
            //消息
            Body: []byte(message),
        })
}

//订阅模式消费端代码
func (r *RabbitMQ) RecieveSub() {
    //尝试创建交换机，不存在创建
    err := r.channel.ExchangeDeclare(
        //交换机名称
        r.Exchange,
        //交换机类型 广播类型
        "fanout",
        //是否持久化
        true,
        //是否字段删除
        false,
        //true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
        false,
        //是否阻塞 true表示要等待服务器的响应
        false,
        nil,
    )
    r.failOnErr(err, "failed to declare an excha"+"nge")
    //2试探性创建队列，创建队列
    q, err := r.channel.QueueDeclare(
        "", //随机生产队列名称
        false,
        false,
        true,
        false,
        nil,
    )
    r.failOnErr(err, "Failed to declare a queue")
    //绑定队列到exchange中
    err = r.channel.QueueBind(
        q.Name,
        //在pub/sub模式下，这里的key要为空
        "",
        r.Exchange,
        false,
        nil,
    )
    //消费消息
    message, err := r.channel.Consume(
        q.Name,
        "",
        true,
        false,
        false,
        false,
        nil,
    )
    forever := make(chan bool)
    go func() {
        for d := range message {
            log.Printf("Received a message:%s,", d.Body)
        }
    }()
    fmt.Println("退出请按 Ctrl+C")
    <-forever
}

//话题模式发送信息
func (r *RabbitMQ) PublishTopic(message string) {
    //尝试创建交换机，不存在创建
    err := r.channel.ExchangeDeclare(
        //交换机名称
        r.Exchange,
        //交换机类型 话题模式
        "topic",
        //是否持久化
        true,
        //是否字段删除
        false,
        //true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
        false,
        //是否阻塞 true表示要等待服务器的响应
        false,
        nil,
    )
    r.failOnErr(err, "topic failed to declare an excha"+"nge")
    //2发送信息
    err = r.channel.Publish(
        r.Exchange,
        //要设置
        r.Key,
        false,
        false,
        amqp.Publishing{
            //类型
            ContentType: "text/plain",
            //消息
            Body: []byte(message),
        })
}

//话题模式接收信息
//要注意key
//其中* 用于匹配一个单词，#用于匹配多个单词（可以是零个）
//匹配 表示匹配imooc.* 表示匹配imooc.hello,但是imooc.hello.one需要用imooc.#才能匹配到
func (r *RabbitMQ) RecieveTopic() {
    //尝试创建交换机，不存在创建
    err := r.channel.ExchangeDeclare(
        //交换机名称
        r.Exchange,
        //交换机类型 话题模式
        "topic",
        //是否持久化
        true,
        //是否字段删除
        false,
        //true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
        false,
        //是否阻塞 true表示要等待服务器的响应
        false,
        nil,
    )
    r.failOnErr(err, "failed to declare an excha"+"nge")
    //2试探性创建队列，创建队列
    q, err := r.channel.QueueDeclare(
        "", //随机生产队列名称
        false,
        false,
        true,
        false,
        nil,
    )
    r.failOnErr(err, "Failed to declare a queue")
    //绑定队列到exchange中
    err = r.channel.QueueBind(
        q.Name,
        //在pub/sub模式下，这里的key要为空
        r.Key,
        r.Exchange,
        false,
        nil,
    )
    //消费消息
    message, err := r.channel.Consume(
        q.Name,
        "",
        true,
        false,
        false,
        false,
        nil,
    )
    forever := make(chan bool)
    go func() {
        for d := range message {
            log.Printf("Received a message:%s,", d.Body)
        }
    }()
    fmt.Println("退出请按 Ctrl+C")
    <-forever
}

//路由模式发送信息
func (r *RabbitMQ) PublishRouting(message string) {
    //尝试创建交换机，不存在创建
    err := r.channel.ExchangeDeclare(
        //交换机名称
        r.Exchange,
        //交换机类型 广播类型
        "direct",
        //是否持久化
        true,
        //是否字段删除
        false,
        //true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
        false,
        //是否阻塞 true表示要等待服务器的响应
        false,
        nil,
    )
    r.failOnErr(err, "failed to declare an excha"+"nge")
    //发送信息
    err = r.channel.Publish(
        r.Exchange,
        //要设置
        r.Key,
        false,
        false,
        amqp.Publishing{
            //类型
            ContentType: "text/plain",
            //消息
            Body: []byte(message),
        })
}

//路由模式接收信息
func (r *RabbitMQ) RecieveRouting() {
    //尝试创建交换机，不存在创建
    err := r.channel.ExchangeDeclare(
        //交换机名称
        r.Exchange,
        //交换机类型 广播类型
        "direct",
        //是否持久化
        true,
        //是否字段删除
        false,
        //true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
        false,
        //是否阻塞 true表示要等待服务器的响应
        false,
        nil,
    )
    r.failOnErr(err, "failed to declare an excha"+"nge")
    //2试探性创建队列，创建队列
    q, err := r.channel.QueueDeclare(
        "", //随机生产队列名称
        false,
        false,
        true,
        false,
        nil,
    )
    r.failOnErr(err, "Failed to declare a queue")
    //绑定队列到exchange中
    err = r.channel.QueueBind(
        q.Name,
        //在pub/sub模式下，这里的key要为空
        r.Key,
        r.Exchange,
        false,
        nil,
    )
    //消费消息
    message, err := r.channel.Consume(
        q.Name,
        "",
        true,
        false,
        false,
        false,
        nil,
    )
    forever := make(chan bool)
    go func() {
        for d := range message {
            log.Printf("Received a message:%s,", d.Body)
        }
    }()
    fmt.Println("退出请按 Ctrl+C")
    <-forever
}
*/
