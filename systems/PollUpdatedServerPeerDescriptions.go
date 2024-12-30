package systems

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type ServerPeerDescription struct {
	Secret              string
	OfferICECandidates  []string `json:"offerIceCandidates"`  // ICE candidates from the peer
	AnswerICECandidates []string `json:"answerIceCandidates"` // ICE candidates from the room creator
	OfferSDP            string   `json:"offerSdp"`
	AnswerSDP           string   `json:"answerSdp"`
}

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
		peers, statusCode, err := Request[map[string]ServerPeerDescription]("/get-peers", jsonData)
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
