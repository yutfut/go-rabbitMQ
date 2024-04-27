package main

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
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
	if err != nil {
		fmt.Println(err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	response := &Message{}

	for d := range msgs {
		if err = json.Unmarshal(d.Body, response); err != nil {
			fmt.Println(err)
		}
		fmt.Printf("Received a message: %s\n", response.Body)
	}
}
