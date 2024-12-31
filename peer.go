package chat

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/pion/webrtc/v4"
)

func UpdateHost(hostSecret, roomName string) {
	AllPeersDescription := make(chan map[string]ServerPeerDescription)
	ConnectedPeers := make(map[string]*ConnectedPeer)
	var m sync.Mutex
	go PollPeers(AllPeersDescription, hostSecret, roomName)
	for {
		serverPeers := <-AllPeersDescription
		for peerId, peer := range serverPeers {
			peer := peer
			peerId := peerId
			m.Lock()
			if _, exists := ConnectedPeers[peerId]; exists {
				m.Unlock()
				continue
			}
			connPeer := &ConnectedPeer{}
			ConnectedPeers[peerId] = connPeer
			m.Unlock()
			go func() {
				answerSDP, pendingCandidates, peerConnection := CreateAnswer(peer)
				SendAnswer(answerSDP, pendingCandidates, roomName, hostSecret, peerId)
				m.Lock()
				connPeer.PeerConnection = peerConnection
				m.Unlock()
				peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
					m.Lock()
					connPeer.DataChannel = dc
					m.Unlock()
					dc.OnMessage(func(msg webrtc.DataChannelMessage) {
						m.Lock()
						defer m.Unlock()
						for _, connectedPeer := range ConnectedPeers {
							if connectedPeer.DataChannel != nil {
								err := connectedPeer.DataChannel.SendText(string(msg.Data))
								if err != nil {
									log.Printf("Failed to forward message to peer: %v", err)
								}
							}
						}
					})
				})
			}()
		}
	}
}

func ConnectToHost(roomName, roomPassword, username string) {
	offer, pendingCandidates, peerConnection, dataChannel := CreateOffer()
	peerSecret := SendOffer(offer, pendingCandidates, roomName, roomPassword, username)
	answerSdp, answerIceCandidates := PollAnswer(roomName, peerSecret, username)
	for _, candidate := range answerIceCandidates {
		peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate})
	}
	err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answerSdp,
	})
	if err != nil {
		log.Fatal("Could not set remote description on client: ", err)
	}
	view := InitView(username, func(msg string) {
		if err := dataChannel.SendText(username + ": " + msg); err != nil {
			fmt.Println(err)
		}
	})
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		message := string(msg.Data)
		username := username + ":"
		if !strings.Contains(message, username) {
			view.messages = append(view.messages, string(msg.Data))
		}
		view.viewport.SetContent(strings.Join(view.messages, "\n"))
		view.textarea.Reset()
		view.viewport.GotoBottom()
	})
	RenderView(view)
}

func CreateAnswer(peerDescription ServerPeerDescription) (answer webrtc.SessionDescription, pendingIceCandidates []*webrtc.ICECandidate, peerConnection *webrtc.PeerConnection) {
	peerConnection, err := webrtc.NewPeerConnection(webRTCConfig)
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

func CreateOffer() (offer webrtc.SessionDescription, pendingIceCandidates []*webrtc.ICECandidate, peerConnection *webrtc.PeerConnection, dataChannel *webrtc.DataChannel) {
	peerConnection, err := webrtc.NewPeerConnection(webRTCConfig)
	if err != nil {
		log.Println(err)
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
	dataChannel, err = peerConnection.CreateDataChannel("chat", nil)
	if err != nil {
		log.Fatalf("Failed to create data channel: %v", err)
	}
	dataChannel.OnError(func(err error) { fmt.Println("there was an error with the datachannel ", err) })
	offer, err = peerConnection.CreateOffer(nil)
	if err != nil {
		log.Fatal(err)
	}
	if err = peerConnection.SetLocalDescription(offer); err != nil {
		log.Println(err)
	}
	fmt.Println("Gathering peers information, please wait!")
	<-webrtc.GatheringCompletePromise(peerConnection)
	return offer, pendingIceCandidates, peerConnection, dataChannel
}
