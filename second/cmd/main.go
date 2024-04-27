package main

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type Ping struct {
	UUID string `json:"UUID"`
	Ping string `json:"ping"`
}

type Pong struct {
	UUID string `json:"UUID"`
	Ping string `json:"ping"`
	Pong string `json:"pong"`
}

func main() {
	conn, err := amqp.Dial("amqp://user:pass@localhost:5672/")
	if err != nil {
		fmt.Println(err)
	}
	defer func(conn *amqp.Connection) {
		if err = conn.Close(); err != nil {
			fmt.Println(err)
		}
	}(conn)

	ch, err := conn.Channel()
	defer func(ch *amqp.Channel) {
		if err = ch.Close(); err != nil {
			fmt.Println(err)
		}
	}(ch)

	input, err := ch.QueueDeclare(
		"input",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	output, err := ch.QueueDeclare(
		"output",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	messages, err := ch.Consume(
		input.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	for d := range messages {
		request := &Ping{}
		if err = json.Unmarshal(d.Body, request); err != nil {
			fmt.Println(err)
			continue
		}

		response := &Pong{
			UUID: request.UUID,
			Ping: request.Ping,
			Pong: request.Ping,
		}

		responseByte, err := json.Marshal(response)
		if err != nil {
			fmt.Println(err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err = ch.PublishWithContext(
			ctx,
			"",
			output.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        responseByte,
			},
		); err != nil {
			fmt.Println(err)
			continue
		}
	}
}
