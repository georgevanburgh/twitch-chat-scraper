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
	elasticChannel = elasticBroker.Connect()

	for _, channelName := range twitchchatscraper.NewLocator().GetTopNChannels(numberOfChannels) {
		wg.Add(1)
		if scraper == nil {
			scraper = twitchchatscraper.NewScraper()
			newClientChannel, newReaderChan := scraper.Connect(channelName)
			clientChannel = newClientChannel

			go printOutput(newReaderChan)
		} else {
			clientChannel <- &channelName
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
