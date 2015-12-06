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
)

var wg sync.WaitGroup
var configurationFile string

func main() {
	// Load configuration from file
	parseCommandLineFlags()
	parseConfigurationFile()

	scraper := twitchchatscraper.NewScraper()
	writerChan, readerChan := scraper.Connect("capcomfighters")

	wg = sync.WaitGroup{}

	wg.Add(2)
	go printOutput(readerChan)
	go readInput(writerChan)
	wg.Wait()
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
	fmt.Println(config.TwitchUsername)
	twitchchatscraper.SetConfig(&config)
}

func printOutput(givenChannel <-chan *irc.Message) {
	defer wg.Done()
	for {
		message, more := <-givenChannel
		if more {
			fmt.Printf("%s: %s: %s\n", message.Params[0], message.User, message.Trailing)
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
