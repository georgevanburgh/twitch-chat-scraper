package twitchchatscraper

import (
	// "crypto/tls"
	"fmt"
	"github.com/sorcix/irc"
	"net"
	"time"

	log "github.com/cihub/seelog"
)

const (
	IRC_PASS_STRING = "PASS %s"
	IRC_USER_STRING = "NICK %s"
	IRC_JOIN_STRING = "JOIN #%s"
)

type Scraper struct {
	chatServers *[]string
	conn        net.Conn
	reader      *irc.Decoder
	writer      *irc.Encoder
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

	log.Debugf("Trying to connect to %s.", chatServers[0])

	// Connect to the first chat server in the list
	// TODO: There should probably be some intelligence around selecting this
	var err error
	s.conn, err = net.Dial("tcp", chatServers[0])

	if err != nil {
		log.Errorf("An error occurred whilst connecting to %s, %s", chatServers[0], err.Error())
		return nil, nil
	}

	log.Debug("Connection established.")

	// Create and return the IRC channels
	s.writer = irc.NewEncoder(s.conn)
	s.reader = irc.NewDecoder(s.conn)

	readChannel := make(chan *irc.Message, 100)
	writeChannel := make(chan *string, 10)

	go s.Read(readChannel)
	go s.Write(writeChannel)

	// Authenticate with the server
	authString := fmt.Sprintf(IRC_PASS_STRING, Configuration.TwitchOAuthToken)
	nickString := fmt.Sprintf(IRC_USER_STRING, Configuration.TwitchUsername)
	joinString := fmt.Sprintf(IRC_JOIN_STRING, givenChannelName)
	writeChannel <- &authString
	writeChannel <- &nickString
	writeChannel <- &joinString

	return writeChannel, readChannel
}

func (s *Scraper) Read(givenChan chan<- *irc.Message) {
	for {
		s.conn.SetDeadline(time.Now().Add(300 * time.Second))
		msg, err := s.reader.Decode()
		if err != nil {
			log.Errorf("Error received whilst reading message: %s", err.Error())
			close(givenChan)
			s.conn.Close()
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
