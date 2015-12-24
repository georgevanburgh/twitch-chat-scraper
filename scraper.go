package twitchchatscraper

import (
	// "crypto/tls"
	"fmt"
	"time"

	"github.com/sorcix/irc"

	log "github.com/cihub/seelog"
)

const (
	IRC_PASS_STRING              = "PASS %s"
	IRC_USER_STRING              = "NICK %s"
	IRC_JOIN_STRING              = "JOIN #%s"
	TIME_TO_WAIT_FOR_CONNECTION  = time.Second * 5
	TIME_BETWEEN_CHANNEL_SCRAPES = time.Minute * 20
	CHANNELS_TO_GET_PER_SCRAPE   = 1000
)

type Scraper struct {
	chatServers  *[]string
	conn         *irc.Conn
	reader       *irc.Decoder
	writer       *irc.Encoder
	readChan     chan *irc.Message
	writeChan    chan *string
	clientChan   chan *string
	SubscribedTo map[string]bool
	connected    bool
}

func NewScraper() *Scraper {
	newScraper := Scraper{}
	newScraper.SubscribedTo = make(map[string]bool, 0)
	return &newScraper
}

func (s *Scraper) Connect() (chan<- *string, <-chan *irc.Message) {
	log.Debug("Connecting to Twitch IRC")

	// List of twitch chat servers (should definitely not be hardcoded)
	chatServers := [...]string{
		"192.16.64.174:6667",
		"192.16.64.175:6667",
		"192.16.64.176:6667",
		"192.16.64.177:6667",
		"192.16.64.178:6667",
		"192.16.64.179:6667",
		"192.16.64.205:6667",
		"192.16.64.206:6667",
		"192.16.64.207:6667",
		"192.16.64.208:6667",
		"192.16.64.209:6667",
		"192.16.64.210:6667",
		"192.16.64.211:6667"}

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
		log.Criticalf("All servers exhausted. Could not connect to IRC")
		panic("All servers exhausted. Could not connect to IRC")
	}

	log.Debug("Connection established.")

	// Create the IRC communication channels
	s.writer = &s.conn.Encoder
	s.reader = &s.conn.Decoder

	readChannel := make(chan *irc.Message, 100)
	writeChannel := make(chan *string, 10)
	clientChannel := make(chan *string, 10000)
	s.readChan = readChannel
	s.writeChan = writeChannel
	s.clientChan = clientChannel

	go s.Read(readChannel)
	go s.Write(writeChannel)

	// Authenticate with the server
	authString := fmt.Sprintf(IRC_PASS_STRING, Configuration.TwitchOAuthToken)
	nickString := fmt.Sprintf(IRC_USER_STRING, Configuration.TwitchUsername)
	s.writeChan <- &authString
	s.writeChan <- &nickString

	// Timeout after a certain amount of time waiting for confirmation
	timer := time.NewTimer(TIME_TO_WAIT_FOR_CONNECTION)
	connectedNotification := make(chan bool, 1)
	go func() {
		for !s.connected {
		}
		connectedNotification <- true
	}()

	select {
	case <-timer.C:
		log.Critical("Timeout whilst authenticating with Twitch chat")
		panic("Timeout whilst authenticating with Twitch chat")
	case <-connectedNotification:
		log.Debug("Successfully authenticated with Twitch chat")
	}

	// Assuming we've successfully authenticated - we can start subscribing to
	// chats, and return the channels for use
	go s.listenForNewClients()
	return clientChannel, readChannel
}

func (s *Scraper) listenForNewClients() {
	for {
		channelToSubscribeTo := <-s.clientChan
		log.Debugf("Asked to subscribe to: %s", *channelToSubscribeTo)

		if !s.SubscribedTo[*channelToSubscribeTo] {
			s.SubscribedTo[*channelToSubscribeTo] = true
			joinString := fmt.Sprintf(IRC_JOIN_STRING, *channelToSubscribeTo)
			s.writeChan <- &joinString
			time.Sleep(time.Second * 2) // We don't want to get rate limited
		} else { // No need to worry about rate limit if we're already subscribed
			log.Debugf("We're already subscribed to %s", channelToSubscribeTo)
		}

	}
}

func (s *Scraper) Read(givenChan chan<- *irc.Message) {
	pongString := "PONG tmi.twitch.tv"
	for {
		msg, err := s.reader.Decode()
		if msg == nil {
			log.Errorf("Nil/deformed message %s", msg)
			break
		} else if msg.Command == "PING" {
			log.Debug("Replying to ping")
			s.writeChan <- &pongString
		} else if msg.Command == "001" {
			log.Debugf("IRC Connection message received")
			s.connected = true
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

func (s *Scraper) StartMessages() {
	log.Debug("Starting messages")
	go s.refreshChannels()

	ticker := time.NewTicker(TIME_BETWEEN_CHANNEL_SCRAPES)
	go func() {
		for {
			<-ticker.C
			s.refreshChannels()
			log.Debug("Grabbed new channels")
			log.Debugf("Now subscribed to %d channels", len(s.SubscribedTo))
		}
	}()
}

func (s *Scraper) refreshChannels() {
	locator := NewLocator()
	topChannels := locator.GetTopNChannels(CHANNELS_TO_GET_PER_SCRAPE)
	for i := 0; i < len(topChannels); i++ {
		s.clientChan <- &topChannels[i]
	}
}
