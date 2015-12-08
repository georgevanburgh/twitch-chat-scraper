package twitchchatscraper

import (
	"testing"
)

func TestLocator_GetChatServerAddress_ReturnsValidChatServerAddress(t *testing.T) {
	underTest := NewLocator()

	chatServers := underTest.GetIrcServerAddress("test_channel")
	if len(chatServers) == 0 {
		t.Fail()
	}
}

func TestLocator_GetTopTwitchStreams_ReturnsListOfCorrectLength(t *testing.T) {
	underTest := NewLocator()

	topStreams := underTest.GetTopNChannels(10)
	if len(topStreams) != 10 {
		t.Fail()
	}
}
