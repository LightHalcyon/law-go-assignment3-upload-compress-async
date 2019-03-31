package main

import (
	"log"
	"os"

    "github.com/streadway/amqp"
)

var ch *amqp.Channel
var err error
var conn *amqp.Connection

func failOnError(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %s", msg, err)
    }
}

func init() {
	url := os.Getenv("URL")
	vhost := os.Getenv("VHOST")
	exchangeName := os.Getenv("EXCNAME")
	exchangeType := os.Getenv("EXCTYPE")

	conn, err = amqp.Dial(url + vhost)
    failOnError(err, "Failed to connect to RabbitMQ")
    defer conn.Close()

    ch, err = conn.Channel()
    failOnError(err, "Failed to open a channel")
    defer ch.Close()

    err = ch.ExchangeDeclare(exchangeName, exchangeType, false, false, false, false, nil)
    failOnError(err, "Failed to declare exchange")
}

func main() {

	body := "Test"
	
	err = ch.Publish("exchange_ping", "", false, false, amqp.Publishing{
		ContentType:	"text/plain",
		Body:			[]byte(body),
	})
	failOnError(err, "Failed to publish message")
	log.Print("sent")
}