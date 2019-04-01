package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

// ErrorJSON error struct to be used when error occured
type ErrorJSON struct {
	StatusCode int    `json:"status"`
	ErrMessage string `json:"message"`
}

var ch *amqp.Channel
var err error
var conn *amqp.Connection

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func startCompress(c *gin.Context) {
	var cfiles [10][]byte
	compressed := false

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		return
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return
	}

	chunks := Split(buf.Bytes())
	index := 0
	for i, v := range chunks {
		cfiles[i], err = Compress(v)
		if err != nil {
			err = ch.Publish("exchange_ping", "", false, false, amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte("Compression Error"),
			})
			failOnError(err, "Compression Error")
			break
		}
		
		percentage := (i+1) * 10
		err = ch.Publish("exchange_ping", "", false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(string(percentage) + "% Compressed"),
		})
		index = i
	}

	if index >= 9 {
		compressed = true
	}
	
	if compressed {

	} else {
		c.JSON(http.StatusUnprocessableEntity, "Failed to compress file")
		return
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
	r := gin.Default()
	r.POST("/compress", startCompress)
	r.Run("0.0.0.0:20606")
}
