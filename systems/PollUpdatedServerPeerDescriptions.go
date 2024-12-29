package systems

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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

func PollUpdatedServerPeerDescriptions(AllPeerDescriptionsChan chan map[string]ServerPeerDescription, SignalingServerAddress string, hostSecret, roomName string) {
	reqBody := map[string]string{
		"hostSecret": hostSecret,
		"roomName":   roomName,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("Failed to marshal request body: %v", err)
	}

	// Poll the server continuously
	for {
		resp, err := http.Post(SignalingServerAddress+"/get-peers", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Failed to make HTTP request: %v", err)
			time.Sleep(6 * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			time.Sleep(6 * time.Second)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read response body after polling for answer: %v", err)
			time.Sleep(6 * time.Second)
			continue
		}

		peers := map[string]ServerPeerDescription{}
		if err := json.Unmarshal(body, &peers); err != nil {
			log.Printf("Failed to unmarshal response body: %v", err)
			time.Sleep(6 * time.Second)
			continue
		}

		// Send the updated peer descriptions to the channel
		AllPeerDescriptionsChan <- peers

		// Wait before polling again
		time.Sleep(6 * time.Second)
	}
}
