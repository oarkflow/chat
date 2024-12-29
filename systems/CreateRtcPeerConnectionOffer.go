package systems

import (
	"fmt"
	"log"
	"sync"

	"github.com/pion/webrtc/v4"
)

func CreateOfferRTCPeerConnection(config webrtc.Configuration) (offer webrtc.SessionDescription, pendingIceCandidates []*webrtc.ICECandidate, peerConnection *webrtc.PeerConnection, dataChannel *webrtc.DataChannel) {
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Println(err)
	}

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
	dataChannel, err = peerConnection.CreateDataChannel("chat", nil)
	if err != nil {
		log.Fatalf("Failed to create data channel: %v", err)
	}

	dataChannel.OnError(func(err error) { fmt.Println("there was an error with the datachannel ", err) })
	offer, err = peerConnection.CreateOffer(nil)
	if err != nil {
		log.Fatal(err)
	}

	// this will start gathering ice candidates and call OnIceCandidates
	if err = peerConnection.SetLocalDescription(offer); err != nil {
		log.Println(err)
	}
	// wait till all candidates are gathered
	fmt.Println("gathering ice candidates, please wait!")
	<-webrtc.GatheringCompletePromise(peerConnection)
	return offer, pendingIceCandidates, peerConnection, dataChannel
}
