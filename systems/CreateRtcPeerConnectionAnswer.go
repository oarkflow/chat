package systems

import (
	"log"
	"sync"

	"github.com/pion/webrtc/v4"
)

func CreateWebrtcConfiguration(StunServerAddress string) webrtc.Configuration {
	return webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:" + StunServerAddress}}}}
}

func CreateAnswerRTCPeerConnection(config webrtc.Configuration,
	peerDescription ServerPeerDescription) (answer webrtc.SessionDescription, pendingIceCandidates []*webrtc.ICECandidate, peerConnection *webrtc.PeerConnection) {
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Println(err)
	}
	if err != nil {
		log.Println("could not set remote description on host: ", err)
	}
	pendingIceCandidates = []*webrtc.ICECandidate{}
	var pendingCandidatesMux sync.Mutex
	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		pendingCandidatesMux.Lock()
		defer pendingCandidatesMux.Unlock()
		if c != nil {
			pendingIceCandidates = append(pendingIceCandidates, c)
		}
	})
	err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  peerDescription.OfferSDP,
	})
	for _, candidate := range peerDescription.OfferICECandidates {
		err := peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate})
		if err != nil {
			log.Println(err)
		}
	}
	answer, err = peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println("failed to create answer on host: ", answer)
	}
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		log.Println(err)
	}
	<-webrtc.GatheringCompletePromise(peerConnection)
	return answer, pendingIceCandidates, peerConnection
}
