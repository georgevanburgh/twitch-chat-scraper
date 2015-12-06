package twitchchatscraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	TWITCH_API_CHAT_PROPERTIES_ENDPOINT = "https://api.twitch.tv/api/channels/%s/chat_properties"
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
