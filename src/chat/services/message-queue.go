package services

import (
	"log"

	"github.com/streadway/amqp"
)

func NewRabbitMQ() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer channel.Close()

	return channel
}

func PublishToQueue(ch *amqp.Channel, key string, body []byte, headers amqp.Table) error {
	err := ch.Publish(
		"",    // exchange
		key,   // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
			Headers:     headers,
		})

	return err
}

func ConsumeFromQueue(ch *amqp.Channel, key string) (<-chan amqp.Delivery, error) {
	msgs, err := ch.Consume(
		key,   // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	return msgs, err
}

func DeclareQueue(ch *amqp.Channel, name string) (amqp.Queue, error) {
	queue, err := ch.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return queue, err
}
