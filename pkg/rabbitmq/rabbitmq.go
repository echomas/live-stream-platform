package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"live-stream-platform/pkg/config"
)

var (
	Conn    *amqp.Connection
	Channel *amqp.Channel
	cfg     *config.RabbitMQConfig
)

func Init(config *config.RabbitMQConfig) error {
	cfg = config
	var err error

	Conn, err = amqp.Dial(cfg.URL)
	if err != nil {
		return fmt.Errorf("failed to connect rabbitmq: %w", err)
	}

	Channel, err = Conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// 声明交换机
	err = Channel.ExchangeDeclare(
		cfg.Exchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Println("RabbitMQ connected successfully")
	return nil
}

func Close() error {
	if Channel != nil {
		Channel.Close()
	}
	if Conn != nil {
		return Conn.Close()
	}
	return nil
}

func GetChannel() *amqp.Channel {
	return Channel
}

func DeclareQueue(name string) (amqp.Queue, error) {
	return Channel.QueueDeclare(
		fmt.Sprintf("%s_%s", cfg.Prefix, name), // name
		true,                                   // durable
		false,                                  // delete when unused
		false,                                  // exclusive
		false,                                  // no-wait
		nil,                                    // arguments
	)
}

func BindQueue(queueName, routingKey string) error {
	return Channel.QueueBind(
		fmt.Sprintf("%s_%s", cfg.Prefix, queueName), // queue name
		routingKey,   // routing key
		cfg.Exchange, // exchange
		false,
		nil,
	)
}

func Publish(routingKey string, body []byte) error {
	return Channel.Publish(
		cfg.Exchange, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
