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

func SendOfferToServer(offerSDP webrtc.SessionDescription, pendingCandidates []*webrtc.ICECandidate, roomName, roomPassword, peerID string) (peerSecret string) {
	var iceCandidates []string
	for _, c := range pendingCandidates {
		iceCandidates = append(iceCandidates, c.ToJSON().Candidate)
	}
	type AddPeerRequest struct {
		PeerID            string   `json:"peerId"`   // Unique identifier for the peer
		RoomName          string   `json:"roomName"` // Name of the room
		Password          string   `json:"password"` // Password for the room (optional)
		OfferSDP          string   `json:"offerSdp"`
		OfferIceCandiates []string `json:"offerIceCandidates"`
	}
	reqBody := AddPeerRequest{
		PeerID:            peerID,
		RoomName:          roomName,
		Password:          roomPassword,
		OfferSDP:          offerSDP.SDP,
		OfferIceCandiates: iceCandidates,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatal("failed to marshal json offer body", err)
	}
	resp, err := http.Post(SignalingServerAddress+"/add-peer", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal("failed to send offer to server.", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Server responded with error after sending offer: %s \n", string(body))
	} else {
		fmt.Println("Sent the offer to the server successfully.")
	}
	body, _ := io.ReadAll(resp.Body)
	JsonResponse := map[string]string{}
	json.Unmarshal(body, &JsonResponse)
	return JsonResponse["peerSecret"]
}
