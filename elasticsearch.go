package twitchchatscraper

import (
	"fmt"
	"time"

	"github.com/sorcix/irc"
	"gopkg.in/olivere/elastic.v3"

	log "github.com/cihub/seelog"
)

type ElasticBroker struct {
	inputChannel <-chan *irc.Message
	elastiClient *elastic.Client
}

type TwitchMessage struct {
	Channel   string
	Message   string
	From      string
	Timestamp time.Time `json:"@timestamp"`
}

func (e *ElasticBroker) Connect() chan<- *irc.Message {
	inputChannel := make(chan *irc.Message, 10000)
	e.inputChannel = inputChannel
	var clientError error
	e.elastiClient, clientError = elastic.NewClient(elastic.SetURL("http://127.0.0.1:9200"), elastic.SetSniff(false)) // Make configurable
	if clientError != nil {
		log.Errorf("Error connecting to elasticsearch: %s", clientError.Error())
	}

	go e.listenForMessages()
	return inputChannel
}

func (e *ElasticBroker) listenForMessages() {
	bulkRequest := e.elastiClient.Bulk()
	for {
		message := <-e.inputChannel
		twitchMessage := TwitchMessage{Channel: message.Params[0], Message: message.Trailing, From: message.User, Timestamp: time.Now()}
		indexToInsertInto := fmt.Sprintf("twitch-%s", twitchMessage.Timestamp.Format("2006.01.02"))
		bulkRequest.Add(elastic.NewBulkIndexRequest().Index(indexToInsertInto).Type("chatmessage").Doc(twitchMessage))

		if bulkRequest.NumberOfActions() > 999 {
			log.Debugf("Applying %d bulk operations", bulkRequest.NumberOfActions())
			bulkRequest.Do()
		}
	}
}
