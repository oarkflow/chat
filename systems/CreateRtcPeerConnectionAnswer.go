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

// create peer connection
// gather ice candidates
// wait till all of the candidates are gathered
// return answer sdp and candidates so you can send them to the server
func CreateAnswerRTCPeerConnection(config webrtc.Configuration,
	peerDescription ServerPeerDescription) (answer webrtc.SessionDescription, pendingIceCandidates []*webrtc.ICECandidate, peerConnection *webrtc.PeerConnection) {

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Println(err)
	}

	// add the peer's offer sdp to this peer connection

	if err != nil {
		log.Println("could not set remote description on host: ", err)
	}
	//gather answer ice candidates in here
	pendingIceCandidates = []*webrtc.ICECandidate{}
	var pendingCandidatesMux sync.Mutex
	//add gathered ice candidates to pendingIceCandidates[]
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
	// add offer ice candidates given by the peer
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
	// this will start gathering ice candidates and call OnIceCandidates
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		log.Println(err)
	}
	// wait till all candidates are gathered
	<-webrtc.GatheringCompletePromise(peerConnection)
	return answer, pendingIceCandidates, peerConnection
}
