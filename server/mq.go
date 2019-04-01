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
type appError struct {
	Code	int    `json:"status"`
	Message	string `json:"message"`
}

var ch *amqp.Channel
var err error
var conn *amqp.Connection

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func (errs *appError) Error() string {
	return errs.Message
}

// JSONAppErrorReporter Error middleware
func JSONAppErrorReporter() gin.HandlerFunc {
    return jsonAppErrorReporterT(gin.ErrorTypeAny)
}

func jsonAppErrorReporterT(errType gin.ErrorType) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        detectedErrors := c.Errors.ByType(errType)

        log.Println("Handle APP error")
        if len(detectedErrors) > 0 {
            errs := detectedErrors[0].Err
            var parsedError *appError
            switch errs.(type) {
				case *appError:
					parsedError = errs.(*appError)
				default:
					parsedError = &appError{ 
						Code: http.StatusInternalServerError,
						Message: "Internal Server Error",
					}
            }
            // Put the error into response
            c.IndentedJSON(parsedError.Code, parsedError)
            // c.Abort()
            c.AbortWithStatusJSON(parsedError.Code, parsedError)
            return
        }

    }
}

func startCompress(c *gin.Context) {
	var cfiles [10][]byte
	compressed := false

	routingKey := c.GetHeader("X-ROUTING-KEY")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.Error(err)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.Error(err)
		return
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		c.Error(err)
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
		err = ch.Publish("exchange_ping", routingKey, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(string(percentage) + "% Compressed"),
		})
		if err != nil {
			c.Error(err)
			break
		}
		index = i
	}

	if index >= 9 {
		compressed = true
	}
	
	if compressed {
		c.JSON(http.StatusOK, appError{
			Code:		http.StatusOK,
			Message:	"File Compressed",
		})
		return
	} else {
		c.JSON(http.StatusUnprocessableEntity, appError{
			Code:		http.StatusUnprocessableEntity,
			Message:	"Failed to compress file",
		})
		return
	}
}

func init() {
	// url := os.Getenv("URL")
	url := "amqp://0806444524:0806444524@152.118.148.103:5672/"
	// vhost := os.Getenv("VHOST")
	vhost := "%2f0806444524"
	// exchangeName := os.Getenv("EXCNAME")
	exchangeName := "1406568753"
	// exchangeType := os.Getenv("EXCTYPE")
	exchangeType := "direct"

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
	r.Use(JSONAppErrorReporter())
	r.POST("/compress", startCompress)
	r.Run("0.0.0.0:20606")
}
