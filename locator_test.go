package twitchchatscraper

import (
	"testing"
)

func TestLocator_GetChatServerAddress_ReturnsValidChatServerAddress(t *testing.T) {
	underTest := Locator{}

	chatServers := underTest.GetIrcServerAddress("test_channel")
	if len(chatServers) == 0 {
		t.Fail()
	}
}
