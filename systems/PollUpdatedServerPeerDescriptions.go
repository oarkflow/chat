package systems

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"chat-app/utils"
)

func PollUpdatedServerPeerDescriptions(AllPeerDescriptionsChan chan map[string]ServerPeerDescription, hostSecret, roomName string) {
	reqBody := map[string]string{
		"hostSecret": hostSecret,
		"roomName":   roomName,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("Failed to marshal request body: %v", err)
	}
	for {
		peers, statusCode, err := utils.Request[map[string]ServerPeerDescription](getUrl("/get-peers"), jsonData)
		if err != nil {
			log.Printf("Failed to make HTTP request: %v", err)
			time.Sleep(6 * time.Second)
			continue
		}
		if statusCode != http.StatusOK {
			time.Sleep(6 * time.Second)
			continue
		}
		AllPeerDescriptionsChan <- peers
		time.Sleep(6 * time.Second)
	}
}
