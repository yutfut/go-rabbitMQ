package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"time"

	"go-rebbitMQ/first/pkg/syncmap"

	"github.com/gofiber/fiber/v3"
	amqp "github.com/rabbitmq/amqp091-go"
)

type HTTPInterface interface {
	Get(ctx fiber.Ctx) error
}

func NewRouting(r *fiber.App, a HTTPInterface) {
	r.Post("/v1/ping", a.Get)
}

type Handler struct {
	producer     *amqp.Channel
	queue        *amqp.Queue
	keyChanelMap syncmap.SyncMapInterface
}

func NewHandler(
	producer *amqp.Channel,
	queue *amqp.Queue,
	sMap syncmap.SyncMapInterface,
) HTTPInterface {
	return &Handler{
		producer:     producer,
		queue:        queue,
		keyChanelMap: sMap,
	}
}

type Ping struct {
	UUID string `json:"UUID"`
	Ping string `json:"ping"`
}

type Pong struct {
	UUID string `json:"UUID"`
	Ping string `json:"ping"`
	Pong string `json:"pong"`
}

func (a *Handler) Get(ctx fiber.Ctx) error {
	request := &Ping{}

	if err := json.Unmarshal(ctx.Body(), request); err != nil {
		return ctx.SendStatus(http.StatusBadRequest)
	}

	uuidRequest := uuid.NewString()

	requestRMQ, err := json.Marshal(
		&Ping{
			UUID: uuidRequest,
			Ping: request.Ping,
		},
	)

	if err != nil {
		return ctx.SendStatus(http.StatusInternalServerError)
	}

	ctxRMQ, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	if err = a.producer.PublishWithContext(
		ctxRMQ,
		"",
		a.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        requestRMQ,
		},
	); err != nil {
		fmt.Println(err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	outputChanel := make(chan []byte)
	if err = a.keyChanelMap.Set(uuidRequest, outputChanel); err != nil {
		fmt.Println(err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	responseData := <-outputChanel

	if _, err = ctx.Write(responseData); err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return ctx.SendStatus(fiber.StatusOK)
}

func Responser(
	consumer *amqp.Channel,
	queue *amqp.Queue,
	keyChanelMap syncmap.SyncMapInterface,
) {
	defer func() {
		if err := consumer.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	messages, err := consumer.Consume(
		queue.Name,
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

	response := &Pong{}

	for d := range messages {
		if err = json.Unmarshal(d.Body, response); err != nil {
			fmt.Println(err)
			continue
		}

		outputChanel, err := keyChanelMap.Get(response.UUID)
		if err != nil {
			fmt.Println(err)
			continue
		}

		outputChanel <- d.Body
		if err = keyChanelMap.Del(response.UUID); err != nil {
			fmt.Println(err)
		}
	}
}
