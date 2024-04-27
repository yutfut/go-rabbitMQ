package main

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type Message struct {
	Body string `json:"body"`
}

func main() {
	conn, err := amqp.Dial("amqp://user:pass@localhost:5672/")
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello",
		false,
		false,
		false,
		false,
		nil,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	marshal, err := json.Marshal(Message{Body: "Hello World1"})
	if err != nil {
		fmt.Println(err)
	}

	if err = ch.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        marshal,
		},
	); err != nil {
		fmt.Println(err)
	}
}
