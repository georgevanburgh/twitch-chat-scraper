package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/FireEater64/twitch-chat-scraper"
	"github.com/sorcix/irc"
	"io/ioutil"
	"os"
	"sync"

	log "github.com/cihub/seelog"
)

var wg sync.WaitGroup
var configurationFile string
var elasticChannel chan<- *irc.Message

func main() {
	defer log.Flush()
	// Load configuration from file
	initializeLogging()
	parseCommandLineFlags()
	parseConfigurationFile()

	wg = sync.WaitGroup{}
	// wg.Add(1)

	elasticBroker := twitchchatscraper.ElasticBroker{}
	elasticChannel = elasticBroker.Connect()

	for _, channel := range twitchchatscraper.NewLocator().GetTopNChannels(1000) {
		wg.Add(3)
		go func(givenChannel string) {
			defer wg.Done()
			scraper := twitchchatscraper.NewScraper()
			writerChan, readerChan := scraper.Connect(givenChannel)

			go printOutput(readerChan)
			go readInput(writerChan)
		}(channel)
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

func readInput(givenChannel chan<- *string) {
	defer wg.Done()
	consoleReader := bufio.NewReader(os.Stdin)
	for {
		consoleInput, _ := consoleReader.ReadString('\n')
		givenChannel <- &consoleInput
	}
}
