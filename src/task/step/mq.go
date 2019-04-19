package step

import (
    "log"

    "github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %s", msg, err)
    }
}

// 只能在安装 rabbitmq 的服务器上操作
func MQSend(protocol, host, user, pwd, topic string, data []byte) {
    mq_connct := protocol+"://"+user+":"+pwd+"@"+host+"/"
    conn, err := amqp.Dial(mq_connct)
    //conn, err := amqp.Dial("amqp://root:123456@47.97.182.182:32222/")
    failOnError(err, "Failed to connect to RabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    failOnError(err, "Failed to open a channel")
    defer ch.Close()

    q, err := ch.QueueDeclare(
        //"hello", // name
        topic, // name
        false,   // durable
        false,   // delete when unused
        false,   // exclusive
        false,   // no-wait
        nil,     // arguments
    )
    failOnError(err, "Failed to declare a queue")

    //body := "Hello World!"
    err = ch.Publish(
        "",     // exchange
        q.Name, // routing key
        false,  // mandatory
        false,  // immediate
        amqp.Publishing{
            ContentType: "text/plain",
            Body:        data,
        })
    log.Printf(" [x] Sent %s", string(data))
    failOnError(err, "Failed to publish a message")
}
