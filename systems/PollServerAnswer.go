package systems

import (
	"fmt"
	"net/http"
	"time"
)

type AnswerResponse struct {
	AnswerSDP           string   `json:"answerSdp"`
	AnswerIceCandidates []string `json:"answerIceCandidates"`
}

func PollServerAnswer(roomName, peerSecret, peerId string) (answerSdp string, answerIceCandidates []string) {
	reqBody := map[string]string{
		"roomName":   roomName,
		"peerSecret": peerSecret,
		"peerId":     peerId,
	}
	for {
		time.Sleep(time.Second * 2)
		JsonResp, statusCode, err := Request[AnswerResponse]("/get-answer", reqBody)
		if statusCode != http.StatusOK {
			continue
		}
		if err != nil {
			fmt.Println("error in polling for answer SDP", err)
			continue
		}
		return JsonResp.AnswerSDP, JsonResp.AnswerIceCandidates
	}
}
