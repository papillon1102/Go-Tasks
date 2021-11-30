package main

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/phuslu/log"
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
	URL string `json:"url"`
}

func ParseHandler(c *gin.Context) {
	var request Request
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	entries, err := GetFeedEntries(request.URL)
	if err != nil {
		log.Error().Err(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Err while parsing RSS feeds"})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func main() {
	router := gin.Default()
	router.POST("/parse", ParseHandler)
	router.Run(":5000")
}
