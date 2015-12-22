package twitchchatscraper

import (
	// "crypto/tls"
	"fmt"

	"github.com/sorcix/irc"

	log "github.com/cihub/seelog"
)

const (
	IRC_PASS_STRING = "PASS %s"
	IRC_USER_STRING = "NICK %s"
	IRC_JOIN_STRING = "JOIN #%s"
)

type Scraper struct {
	chatServers *[]string
	conn        *irc.Conn
	reader      *irc.Decoder
	writer      *irc.Encoder
	readChan    chan *irc.Message
	writeChan   chan *string
	clientChan  chan *string
}

func NewScraper() *Scraper {
	newScraper := Scraper{}
	return &newScraper
}

func (s *Scraper) Connect(givenChannelName string) (chan<- *string, <-chan *irc.Message) {
	log.Debugf("Connecting to Twitch chat for %s", givenChannelName)

	// Grab the list of chat servers for this channel
	locator := NewLocator()
	chatServers := locator.GetIrcServerAddress(givenChannelName)
	s.chatServers = &chatServers

	if len(chatServers) == 0 {
		log.Errorf("Error whilst connecting to %s, no chat servers available", givenChannelName)
	}

	log.Debugf("Trying to connect to %s.", chatServers[0])

	// Connect to the first chat server in the list
	// TODO: There should probably be some intelligence around selecting this
	var err error
	for server := 0; server < len(chatServers); server++ {
		s.conn, err = irc.Dial(chatServers[server])

		if err == nil {
			break
		}
		log.Errorf("An error occurred whilst connecting to %s, %s.", chatServers[server], err.Error())
	}
	if err != nil {
		log.Criticalf("All servers exhausted. Will not collect metrics for %s", givenChannelName)
		return nil, nil
	}

	log.Debug("Connection established.")

	// Create and return the IRC channels
	s.writer = &s.conn.Encoder
	s.reader = &s.conn.Decoder

	readChannel := make(chan *irc.Message, 100)
	writeChannel := make(chan *string, 10)
	clientChannel := make(chan *string, 10)
	s.readChan = readChannel
	s.writeChan = writeChannel
	s.clientChan = clientChannel

	go s.Read(readChannel)
	go s.Write(writeChannel)
	go s.listenForNewClients()

	// Authenticate with the server
	authString := fmt.Sprintf(IRC_PASS_STRING, Configuration.TwitchOAuthToken)
	nickString := fmt.Sprintf(IRC_USER_STRING, Configuration.TwitchUsername)
	s.writeChan <- &authString
	s.writeChan <- &nickString
	s.clientChan <- &givenChannelName

	return clientChannel, readChannel
}

func (s *Scraper) listenForNewClients() {
	for {
		channelToSubscribeTo := <-s.clientChan
		log.Debugf("Asked to subscribe to: %s", *channelToSubscribeTo)

		joinString := fmt.Sprintf(IRC_JOIN_STRING, *channelToSubscribeTo)
		s.writeChan <- &joinString
	}
}

func (s *Scraper) Read(givenChan chan<- *irc.Message) {
	pongString := "PONG tmi.twitch.tv"
	for {
		msg, err := s.reader.Decode()
		if msg == nil {
			log.Errorf("Nil/deformed message %s", msg)
		} else if msg.Command == "PING" {
			log.Debug("Replying to ping")
			s.writeChan <- &pongString
		} else if msg.Command != "PRIVMSG" {
			log.Debugf("Control message received: %s", msg)
		} else if err != nil {
			log.Errorf("Error received whilst reading message: %s", err.Error())
			break
		} else {
			// We only care about user messages
			if !msg.IsServer() {
				givenChan <- msg
			}
		}
	}
}

func (s *Scraper) Write(givenChan <-chan *string) {
	for {
		messageToSend := <-givenChan
		ircMessageToSend := irc.ParseMessage(*messageToSend)
		if err := s.writer.Encode(ircMessageToSend); err != nil {
			log.Errorf("Error sending message %s: %s", *messageToSend, err.Error())
			break
		}
	}
}
