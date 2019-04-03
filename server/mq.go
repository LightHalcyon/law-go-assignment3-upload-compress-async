package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"io/ioutil"
	"path/filepath"
	"fmt"
	"math/rand"
	"mime/multipart"
	// "time"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"github.com/gin-contrib/cors"
)

// ErrorJSON error struct to be used when error occured
type appError struct {
	Code	int    `json:"status"`
	Message	string `json:"message"`
}

var ch *amqp.Channel
var err error
var conn *amqp.Connection
var files map[string]string
var url, vhost, exchangeName, exchangeType string

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// func (errs *appError) Error() string {
// 	return errs.Message
// }

// // JSONAppErrorReporter Error middleware
// func JSONAppErrorReporter() gin.HandlerFunc {
//     return jsonAppErrorReporterT(gin.ErrorTypeAny)
// }

// func jsonAppErrorReporterT(errType gin.ErrorType) gin.HandlerFunc {
//     return func(c *gin.Context) {
//         c.Next()
//         detectedErrors := c.Errors.ByType(errType)

//         log.Println("Handle APP error")
//         if len(detectedErrors) > 0 {
// 			errs := detectedErrors[0].Err
// 			log.Println(detectedErrors)
//             var parsedError *appError
//             switch errs.(type) {
// 				case *appError:
// 					parsedError = errs.(*appError)
// 				default:
// 					parsedError = &appError{ 
// 						Code: http.StatusInternalServerError,
// 						Message: "Internal Server Error",
// 					}
//             }
//             // Put the error into response
//             // c.IndentedJSON(parsedError.Code, parsedError)
//             // c.Abort()
//             c.AbortWithStatusJSON(parsedError.Code, parsedError)
//             return
//         }

//     }
// }

// TokenGenerator generates token for as key for downloading
func TokenGenerator() string {
	b := make([]byte, 18)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func startCompress(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin","*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, X-Routing-Key, Host")

	// log.Println(url, vhost)

	// compressed := false
	// log.Println(c.Request.Header)

	routingKey := c.GetHeader("X-Routing-Key")

	// log.Println(c.Request.Body)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		// log.Println(err)
		c.JSON(http.StatusBadRequest, appError{
			Code:		http.StatusBadRequest,
			Message:	"File get error, did you upload a file?",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, appError{
			Code:		http.StatusInternalServerError,
			Message:	err.Error(),
		})
		return
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		c.JSON(http.StatusInternalServerError, appError{
			Code:		http.StatusInternalServerError,
			Message:	err.Error(),
		})
		return
	}

	chunks := Split(buf.Bytes())

	c.JSON(http.StatusOK, appError{
		Code:		http.StatusOK,
		Message:	"File Compressed",
	})

	go func(chunks [10][]byte, ch *amqp.Channel, routingKey string, fileHeader *multipart.FileHeader) {
		var cfiles [10][]byte
		var err1 error

		compression := true

		log.Println(routingKey)

		// time.Sleep(10 * time.Second)

		for i, v := range chunks {
			// log.Println(i)
			cfiles[i], err1 = Compress(v)
			if err1 != nil {
				err = ch.Publish(exchangeName, routingKey, false, false, amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte("Compression Error"),
				})
				failOnError(err1, "Compression Error")
				compression = false
				break
			}
			
			percentage := (i+1) * 10
			// log.Println(string(percentage))
			err2 := ch.Publish(exchangeName, routingKey, false, false, amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(strconv.Itoa(percentage) + "% Compressed"),
			})
			failOnError(err2, "Publish Error")
			// time.Sleep(1 * time.Second)
		}

		if compression {
			cfile := Combine(cfiles)
			sep := string(filepath.Separator)
			filename := "dl/" + filepath.Base(sep + "dl" + sep + fileHeader.Filename) + ".gz"
	
			err = ioutil.WriteFile(filename, cfile, 0644)
	
			key := TokenGenerator()

			// currURL := "http://1406568753.law.infralabs.cs.ui.ac.id:20608/download" + key
			currURL := "http://localhost:20608/download/" + key

			err2 := ch.Publish(exchangeName, routingKey, false, false, amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte("<a href=" + currURL + ">" + currURL + "</a>"),
			})
			failOnError(err2, "Publish Error")
			files[key] = filename
		}
	}(chunks, ch, routingKey, fileHeader)

	return	
}

func download(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin","*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, X-Routing-Key, Host")

	id := c.Param("id")

	if _, ok := files[id]; !ok {
		c.JSON(http.StatusNotFound, appError{
			Code:		http.StatusNotFound,
			Message:	"File not found, have you uploaded it here?",
		})
		return
	}

	targetFile := files[id]
	fileName := strings.Replace(targetFile, "dl/" , "", -1)

    c.Header("Content-Description", "File Transfer")
    c.Header("Content-Transfer-Encoding", "binary")
    c.Header("Content-Disposition", "attachment; filename=" + fileName)
    c.Header("Content-Type", "application/octet-stream")
	c.File(targetFile)
	return
}

func main() {
	url := "amqp://" + os.Getenv("UNAME") + ":" + os.Getenv("PW") + "@" + os.Getenv("URL") + ":" + os.Getenv("PORT") + "/"
	// url = "amqp://1406568753:167664@152.118.148.103:5672/"
	vhost := os.Getenv("VHOST")
	// vhost = "1406568753"
	exchangeName := os.Getenv("EXCNAME")
	// exchangeName = "1406568753"
	exchangeType = "direct"

	files = make(map[string]string)

	conn, err = amqp.Dial(url + vhost)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err = conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(exchangeName, exchangeType, false, false, false, false, nil)
	failOnError(err, "Failed to declare exchange")

	r := gin.Default()
	// r.Use(JSONAppErrorReporter())
	r.POST("/compress", startCompress)
	r.GET("/download/:id", download)
	conf := cors.DefaultConfig()
	conf.AllowOrigins = []string{"*"}
	conf.AddAllowHeaders("X-ROUTING-KEY")
	conf.AddAllowHeaders("Content-Type")
	conf.AddAllowHeaders("Access-Control-Allow-Origin")
	conf.AddAllowHeaders("Access-Control-Allow-Headers")
	conf.AddAllowHeaders("Access-Control-Allow-Methods")
	conf.AddAllowHeaders("Host")
	r.Use(cors.New(conf))
	// log.Println("Running in localhost:20606")
	r.Run("0.0.0.0:20608")
}
