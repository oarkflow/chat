package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pion/webrtc/v4"

	"github.com/oarkflow/chat/utils"
)

func CreateRoom(name, password string) (hostSecret string, err error) {
	reqBody := map[string]string{
		"name":     name,
		"password": password,
	}
	respBody, statusCode, err := utils.Request[map[string]string](getUrl("/create-room"), reqBody)
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("server returned status %d", statusCode)
	}
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	return respBody["hostSecret"], nil
}

func PollAnswer(roomName, peerSecret, peerId string) (answerSdp string, answerIceCandidates []string) {
	reqBody := map[string]string{
		"roomName":   roomName,
		"peerSecret": peerSecret,
		"peerId":     peerId,
	}
	for {
		time.Sleep(time.Second * 2)
		JsonResp, statusCode, err := utils.Request[AnswerResponse](getUrl("/get-answer"), reqBody)
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

func PollPeers(AllPeerDescriptionsChan chan map[string]ServerPeerDescription, hostSecret, roomName string) {
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

func SendAnswer(answerSDP webrtc.SessionDescription, pendingCandidates []*webrtc.ICECandidate, roomName, hostSecret, peerID string) {
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

func SendOffer(offerSDP webrtc.SessionDescription, pendingCandidates []*webrtc.ICECandidate, roomName, roomPassword, peerID string) (peerSecret string) {
	var iceCandidates []string
	for _, c := range pendingCandidates {
		iceCandidates = append(iceCandidates, c.ToJSON().Candidate)
	}
	reqBody := AddPeerRequest{
		PeerID:             peerID,
		RoomName:           roomName,
		Password:           roomPassword,
		OfferSDP:           offerSDP.SDP,
		OfferIceCandidates: iceCandidates,
	}
	JsonResponse, statusCode, err := utils.Request[map[string]string](getUrl("/add-peer"), reqBody)
	if err != nil {
		log.Fatal("failed to marshal json offer body", err)
	}
	if statusCode != http.StatusOK {
		log.Fatalf("Server responded with error after sending offer \n")
	} else {
		fmt.Println("Sent the offer to the server successfully.")
	}
	return JsonResponse["peerSecret"]
}
