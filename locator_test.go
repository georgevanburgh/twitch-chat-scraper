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

	if !arrayContainsUniqueValues(topStreams) {
		t.Fatalf("Returned channels should be unique")
		t.Fail()
	}

}

func TestLocator_GetTopTwitchStreams_ReturnsBigListOfCorrectLength(t *testing.T) {
	underTest := NewLocator()
	numberOfChannelsToRequest := 201

	topStreams := underTest.GetTopNChannels(numberOfChannelsToRequest)
	if len(topStreams) != numberOfChannelsToRequest {
		t.Fatalf("Expected %d channels, got %d", numberOfChannelsToRequest,
			len(topStreams))
		t.Fail()
	}

	if !arrayContainsUniqueValues(topStreams) {
		t.Fatalf("Returned channels should be unique")
		t.Fail()
	}
}

func arrayContainsUniqueValues(givenArray []string) bool {
	valueMap := make(map[string]bool, 0)
	for _, value := range givenArray {
		if valueMap[value] {
			return false
		}
		valueMap[value] = true
	}
	return true
}
