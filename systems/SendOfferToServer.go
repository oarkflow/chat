package systems

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pion/webrtc/v4"
)

type AddPeerRequest struct {
	PeerID             string   `json:"peerId"`   // Unique identifier for the peer
	RoomName           string   `json:"roomName"` // Name of the room
	Password           string   `json:"password"` // Password for the room (optional)
	OfferSDP           string   `json:"offerSdp"`
	OfferIceCandidates []string `json:"offerIceCandidates"`
}

func SendOfferToServer(offerSDP webrtc.SessionDescription, pendingCandidates []*webrtc.ICECandidate, roomName, roomPassword, peerID string) (peerSecret string) {
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
	JsonResponse, statusCode, err := Request[map[string]string]("/add-peer", reqBody)
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
