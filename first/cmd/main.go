package main

import (
	"fmt"
	"go-rebbitMQ/first/pkg/syncmap"
	"log"

	"go-rebbitMQ/first/http"

	"github.com/gofiber/fiber/v3"
	amqp "github.com/rabbitmq/amqp091-go"
)

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

	sMap := syncmap.NewSyncMap()

	go http.Responser(ch, &output, sMap)

	router := fiber.New()
	http.NewRouting(
		router,
		http.NewHandler(
			ch,
			&input,
			sMap,
		),
	)

	log.Fatal(router.Listen(fmt.Sprintf(":%d", 8000)))
}
