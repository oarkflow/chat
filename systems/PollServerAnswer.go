package systems

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func PollServerAnswer(SignalingServerAddress, roomName, peerSecret, peerId string) (answerSdp string, answerIceCandidates []string) {
	reqBody := map[string]string{
		"roomName":   roomName,
		"peerSecret": peerSecret,
		"peerId":     peerId,
	}
	jsonData, _ := json.Marshal(reqBody)
	for {
		time.Sleep(time.Second * 2)
		resp, err := http.Post(SignalingServerAddress+"/get-answer", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error polling for answer Sdp: %v\n", err)
			continue
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			fmt.Println(string(body))
			continue
		}
		var JsonResp struct {
			AnswerSDP           string   `json:"answerSdp"`
			AnswerIceCandidates []string `json:"answerIceCandidates"`
		}
		if err := json.Unmarshal(body, &JsonResp); err != nil {
			fmt.Println("error in polling for answer SDP", string(body), err)
			continue
		}
		return JsonResp.AnswerSDP, JsonResp.AnswerIceCandidates
	}
}
