package systems

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pion/webrtc/v4"

	"chat-app/utils"
)

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
