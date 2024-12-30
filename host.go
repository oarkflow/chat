package chat

import (
	"fmt"
	"log"
	"sync"

	"github.com/pion/webrtc/v4"

	"github.com/oarkflow/chat/utils"
)

func UpdateHost(hostSecret, roomName string, WebrtcConfiguration webrtc.Configuration) {
	AllPeersDescription := make(chan map[string]ServerPeerDescription)
	ConnectedPeers := make(map[string]*ConnectedPeer)
	var m sync.Mutex
	go PollPeer(AllPeersDescription, hostSecret, roomName)
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
				answerSDP, pendingCandidates, peerConnection := CreateAnswer(WebrtcConfiguration, peer)
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

func ConnectToHost(roomName, roomPassword, username string, WebrtcConfiguration webrtc.Configuration) {
	MessageHistory := ""
	offer, pendingCandidates, peerConnection, dataChannel := CreateOffer(WebrtcConfiguration)
	peerSecret := SendOffer(offer, pendingCandidates, roomName, roomPassword, username)
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		MessageHistory += fmt.Sprintln(string(msg.Data))
		utils.ClearScreen()
		fmt.Println(MessageHistory)
		fmt.Print("Type a message: ")
	})
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
	for {
		fmt.Println(MessageHistory)
		message := AskForMessageInput()
		if err := dataChannel.SendText(username + ": " + message); err != nil {
			fmt.Println(err)
		}
	}
}
