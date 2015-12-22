package twitchchatscraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/cihub/seelog"
)

const (
	TWITCH_API_CHAT_PROPERTIES_ENDPOINT = "https://api.twitch.tv/api/channels/%s/chat_properties"
	TWITCH_API_TOP_STREAM_ENDPOINT      = "https://api.twitch.tv/kraken/streams?limit=%d&offset=%d&broadcaster_language=en" // TODO: Make language configurable
)

// Locate the best chat server to connect to for a given twitch channel
type Locator struct {
}

// Locator constructor
func NewLocator() *Locator {
	locatorToReturn := Locator{}
	return &locatorToReturn
}

type chatPropertiesResponse struct {
	ID                              int      `json:"_id"`
	HideChatLinks                   bool     `json:"hide_chat_links"`
	DevChat                         bool     `json:"devchat"`
	Game                            string   `json:"game"`
	RequireVerifiedAccount          bool     `json:"require_verified_account"`
	SubsOnly                        bool     `json:"false"`
	EventChat                       bool     `json:"eventchat"`
	ChatServers                     []string `json:"chat_servers"`
	WebSocketServers                []string `json:"web_socket_servers"`
	WebSocketPct                    float32  `json:"web_socket_pct"`
	DarklaunchPct                   float32  `json:"darklaunch_pct"`
	BlockChatNotificationToken      string   `json:"block_chat_notification_token"`
	AvailableChatNotificationTokens string   `json:"available_chat_notification_tokens"`
	SceTitlePresetText1             string   `json:"sce_title_preset_text_1"`
	SceTitlePresetText2             string   `json:"sce_title_preset_text_2"`
	SceTitlePresetText3             string   `json:"sce_title_preset_text_3"`
	SceTitlePresetText4             string   `json:"sce_title_preset_text_4"`
	SceTitlePresetText5             string   `json:"sce_title_preset_text_5"`
}

type streamsResponse struct {
	Streams []stream `json:"streams"`
}

type stream struct {
	ID      int    `json:"_id"`
	Game    string `json:"game"`
	Viewers int    `json:"viewers"`
	Channel struct {
		Name string `json:"name"`
	} `json:"channel"`
}

func (l *Locator) GetIrcServerAddress(givenChannelName string) []string {
	url := fmt.Sprintf(TWITCH_API_CHAT_PROPERTIES_ENDPOINT, givenChannelName)

	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	chatDetails := chatPropertiesResponse{}
	json.Unmarshal([]byte(body), &chatDetails)

	return chatDetails.ChatServers
}

func (l *Locator) GetTopNChannels(givenLimit int) []string {
	offset := 0
	toReturn := make([]string, 0)

	for len(toReturn) < givenLimit {
		streamDetails := l.getChannelDetails(100, offset)

		for _, stream := range streamDetails.Streams {
			if len(toReturn) == givenLimit {
				break
			}
			toReturn = append(toReturn, stream.Channel.Name)
		}
		offset += 100
	}

	return toReturn
}

func (l *Locator) getChannelDetails(givenLimit int, givenOffset int) *streamsResponse {
	url := fmt.Sprintf(TWITCH_API_TOP_STREAM_ENDPOINT, givenLimit, givenOffset)
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("Error whilst calling Twitch: %s", err.Error())
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	streamDetails := streamsResponse{}
	json.Unmarshal([]byte(body), &streamDetails)
	return &streamDetails
}
