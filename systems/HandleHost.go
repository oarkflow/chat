package systems

import (
	"fmt"
	"log"
	"sync"

	"github.com/pion/webrtc/v4"

	"chat-app/utils"
)

type ConnectedPeer struct {
	peerId          string
	peerDescription ServerPeerDescription
	peerConnection  *webrtc.PeerConnection
	dataChannel     *webrtc.DataChannel
}

func UpdateHost(hostSecret, roomName string, WebrtcConfiguration webrtc.Configuration) {
	AllPeersDescription := make(chan map[string]ServerPeerDescription)
	ConnectedPeers := make(map[string]*ConnectedPeer)
	var ConnectedPeersMutex sync.Mutex
	go PollUpdatedServerPeerDescriptions(AllPeersDescription, hostSecret, roomName)
	for {
		serverPeers := <-AllPeersDescription
		for peerId, peer := range serverPeers {
			ConnectedPeersMutex.Lock()
			if _, exists := ConnectedPeers[peerId]; exists {
				ConnectedPeersMutex.Unlock()
				continue
			}
			ConnectedPeers[peerId] = &ConnectedPeer{}
			ConnectedPeersMutex.Unlock()
			go func() {
				answerSDP, pendingCandidates, peerConnection := CreateAnswerRTCPeerConnection(WebrtcConfiguration, peer)
				SendAnswerToServer(answerSDP, pendingCandidates, roomName, hostSecret, peerId)
				ConnectedPeersMutex.Lock()
				ConnectedPeers[peerId].peerConnection = peerConnection
				ConnectedPeersMutex.Unlock()
				peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
					ConnectedPeersMutex.Lock()
					ConnectedPeers[peerId].dataChannel = dc
					ConnectedPeersMutex.Unlock()
					dc.OnMessage(func(msg webrtc.DataChannelMessage) {
						ConnectedPeersMutex.Lock()
						defer ConnectedPeersMutex.Unlock()
						for _, connectedPeer := range ConnectedPeers {
							if connectedPeer.dataChannel != nil {
								err := connectedPeer.dataChannel.SendText(string(msg.Data))
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
	offer, pendingCandidates, peerConnection, dataChannel := CreateOfferRTCPeerConnection(WebrtcConfiguration)
	peerSecret := SendOfferToServer(offer, pendingCandidates, roomName, roomPassword, username)
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		MessageHistory += fmt.Sprintln(string(msg.Data))
		utils.ClearScreen()
		fmt.Println(MessageHistory)
		fmt.Print("Type a message: ")
	})
	answerSdp, answerIceCandidates := PollServerAnswer(roomName, peerSecret, username)
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
