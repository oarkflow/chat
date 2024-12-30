package systems

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/pion/webrtc/v4"
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
	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(SignalingServerAddress+"/set-answer", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal("Failed to send answer to server: ", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to send answer. Server response: %s\n", string(body))
	} else {
		fmt.Println("Sent the answer to the server successfully.")
	}
}
