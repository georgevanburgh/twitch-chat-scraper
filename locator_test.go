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
	numberOfChannelsToRequest := 10

	topStreams := underTest.GetTopNChannels(numberOfChannelsToRequest)
	if len(topStreams) != numberOfChannelsToRequest {
		t.Fatalf("Expected %d channels, got %d", numberOfChannelsToRequest,
			len(topStreams))
		t.Fail()
	}
}

func TestLocator_GetTopTwitchStreams_ReturnsBigListOfCorrectLength(t *testing.T) {
	underTest := NewLocator()
	numberOfChannelsToRequest := 200

	topStreams := underTest.GetTopNChannels(numberOfChannelsToRequest)
	if len(topStreams) != numberOfChannelsToRequest {
		t.Fatalf("Expected %d channels, got %d", numberOfChannelsToRequest,
			len(topStreams))
		t.Fail()
	}
}
