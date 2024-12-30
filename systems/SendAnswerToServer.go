package systems

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pion/webrtc/v4"

	"chat-app/utils"
)

func SendAnswerToServer(answerSDP webrtc.SessionDescription, pendingCandidates []*webrtc.ICECandidate, roomName, hostSecret, peerID string) {
	var iceCandidates []string
	for _, c := range pendingCandidates {
		iceCandidates = append(iceCandidates, c.ToJSON().Candidate)
	}
	reqBody := map[string]any{
		"hostSecret":          hostSecret,
		"roomName":            roomName,
		"peerId":              peerID,
		"answerSdp":           answerSDP.SDP,
		"answerIceCandidates": iceCandidates,
	}
	_, statusCode, err := utils.Request[map[string]any](getUrl("/set-answer"), reqBody)
	if err != nil {
		log.Fatal("Failed to send answer to server: ", err)
	}
	if statusCode != http.StatusOK {
		log.Fatal("Failed to send answer")
	} else {
		fmt.Println("Sent the answer to the server successfully.")
	}
}
