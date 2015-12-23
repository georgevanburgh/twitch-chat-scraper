package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/FireEater64/twitch-chat-scraper"
	"github.com/sorcix/irc"

	log "github.com/cihub/seelog"
)

var wg sync.WaitGroup
var configurationFile string
var elasticSearchHost string
var numberOfChannels int
var elasticChannel chan<- *irc.Message
var scraper *twitchchatscraper.Scraper
var clientChannel chan<- *string

func main() {
	defer log.Flush()
	// Load configuration from file
	initializeLogging()
	parseCommandLineFlags()
	parseConfigurationFile()

	wg = sync.WaitGroup{}

	elasticBroker := twitchchatscraper.ElasticBroker{}
	elasticChannel = elasticBroker.Connect(elasticSearchHost)

	locator := twitchchatscraper.NewLocator()
	topChannels := locator.GetTopNChannels(numberOfChannels)
	for i := 0; i < len(topChannels); i++ {
		wg.Add(1)
		if scraper == nil {
			scraper = twitchchatscraper.NewScraper()
			newClientChannel, newReaderChan := scraper.Connect(topChannels[i])
			clientChannel = newClientChannel

			go printOutput(newReaderChan)
		} else {
			clientChannel <- &topChannels[i]
		}

	}
	wg.Wait()
}

func initializeLogging() {
	logger, err := log.LoggerFromConfigAsFile("logconfig.xml")

	if err != nil {
		log.Criticalf("An error occurred whilst initializing logging\n", err.Error())
		panic(err)
	}

	log.ReplaceLogger(logger)
}

func parseCommandLineFlags() {
	flag.StringVar(&configurationFile, "config", "config.json", "The location of the config.json file")
	flag.IntVar(&numberOfChannels, "channels", 1000, "The number of top channels to subscribe to")
	flag.StringVar(&elasticSearchHost, "eshost", "http://127.0.0.1:9200", "The address of the ElasticSearch server to send stats to")

	flag.Parse()
}

func parseConfigurationFile() {
	configurationFileContents, readError := ioutil.ReadFile(configurationFile)
	if readError != nil {
		panic(fmt.Sprintf("Could not read configuration file: %s", readError.Error()))
	}

	config := twitchchatscraper.Config{}
	parseError := json.Unmarshal(configurationFileContents, &config)
	if parseError != nil {
		panic(fmt.Sprintf("Could not parse configuration file: %s", parseError.Error()))
	}
	twitchchatscraper.SetConfig(&config)
}

func printOutput(givenChannel <-chan *irc.Message) {
	defer wg.Done()
	for {
		message, more := <-givenChannel
		if more {
			// log.Debugf("%s: %s: %s", message.Params[0], message.User, message.Trailing)
			elasticChannel <- message
		} else {
			break
		}
	}
}
