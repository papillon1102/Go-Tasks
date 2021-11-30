package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/phuslu/log"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/mgo.v2/bson"
)

type Entry struct {
	Link struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Thumbnail struct {
		URL string `xml:"url,attr"`
	} `xml:"thumbnail"`
	Title string `xml:"title"`
}

type Feed struct {
	Entries []Entry `xml:"entry"`
}

func GetFeedEntries(url string) ([]Entry, error) {

	// Create new http.Client & make request
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Simulation of request send from browser (NOTE)
	// Find out list of valid User-Agents
	// https://developers.whatismybrowser.com/useragents/explore/.
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36(KHTML, like Gecko) Chrome/70.0.3538.110 Safari/537.36")

	// Get return from request
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	byteValue, _ := ioutil.ReadAll(res.Body)
	feed := Feed{}
	xml.Unmarshal(byteValue, &feed)

	return feed.Entries, nil
}

type Request struct {
	URL string
}

func main() {

	// Connect to mongo
	ctx := context.Background()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = mongoClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal().Err(err)
	} else {
		log.Info().Msg("Connected to MongoDB")
	}

	// Connect to rabbitmq
	amqpConnection, err := amqp.Dial(os.Getenv("RABBITMQ_URI"))
	if err != nil {
		log.Fatal().Err(err)
	}
	defer amqpConnection.Close()

	channelAmqp, err := amqpConnection.Channel()
	if err != nil {
		log.Debug().Err(err).Msg("Err from connecting to rabbitmq")
	} else {
		log.Info().Msg("Connected to RabbitMQ")
	}

	defer channelAmqp.Close()

	forever := make(chan bool)
	msgs, err := channelAmqp.Consume(
		os.Getenv("RABBITMQ_QUEUE"),
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	go func() {
		for msg := range msgs {
			log.Info().Msgf("Received a message: %s\n", msg.Body)

			var request Request
			json.Unmarshal(msg.Body, &request)
			log.Debug().Msgf("RSS-URL : %s\n", request.URL)

			entries, _ := GetFeedEntries(request.URL)

			collection := mongoClient.Database(os.Getenv("MONGO_DATABASE")).Collection("tasks")
			for _, e := range entries {
				collection.InsertOne(ctx, bson.M{
					"title":     e.Title,
					"thumbnail": e.Thumbnail.URL,
					"url":       e.Link.Href,
				})
			}

		}
	}()

	// <-forever : locking "main go-routine" to avoid quitting instantly
	log.Printf("Waiting for message. To exit press CTRL+C")
	<-forever
}
