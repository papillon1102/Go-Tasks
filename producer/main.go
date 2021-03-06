package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/phuslu/log"
	"github.com/streadway/amqp"
)

// MONGO_URI="mongodb://admin:password@192.168.99.100:27017/test?authSourc=admin&readPreference=primary&appname=MongoDB%20Compass&ssl=false" MONGO_DATABASE=test-rss  RABBITMQ_URI="amqp://user:password@192.168.99.100:5672/" RABBITMQ_QUEUE=rss_urls go run main.go
var channelAmqp *amqp.Channel

func init() {

	// Connection-string will be provided via "RABBITMQ_URI" (NOTE)
	amqpConnection, err := amqp.Dial(os.Getenv("RABBITMQ_URI"))
	if err != nil {
		log.Error().Err(err)
	}

	channelAmqp, _ = amqpConnection.Channel()
}

type Request struct {
	URL string `json:"url"`
}

func ParserHandler(c *gin.Context) {

	var request Request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	data, _ := json.Marshal(request)
	err := channelAmqp.Publish(
		"",
		os.Getenv("RABBITMQ_QUEUE"),
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(data),
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("Err from publish message")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Err while publishing rabbitmq"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Publish Success"})
}

func main() {
	r := gin.Default()
	r.POST("/parse", ParserHandler)
	r.Run(":5001")
}
