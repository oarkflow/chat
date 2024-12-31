package chat

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/pion/webrtc/v4"
)

type PeerMessage struct {
	peerId string
	peer   ServerPeerDescription
}

func UpdateHost(hostSecret, roomName string) {
	allPeersDescription := make(chan map[string]ServerPeerDescription)
	peerUpdates := make(chan PeerMessage)
	connectedPeers := make(map[string]*ConnectedPeer)
	go func() {
		for update := range peerUpdates {
			if _, exists := connectedPeers[update.peerId]; exists {
				continue
			}
			connPeer := &ConnectedPeer{}
			connectedPeers[update.peerId] = connPeer
			go func(peerId string, peer ServerPeerDescription, connPeer *ConnectedPeer) {
				answerSDP, pendingCandidates, peerConnection := CreateAnswer(peer)
				SendAnswer(answerSDP, pendingCandidates, roomName, hostSecret, peerId)
				connPeer.PeerConnection = peerConnection
				peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
					connPeer.DataChannel = dc
					dc.OnMessage(func(msg webrtc.DataChannelMessage) {
						for _, connectedPeer := range connectedPeers {
							if connectedPeer.DataChannel != nil {
								err := connectedPeer.DataChannel.SendText(string(msg.Data))
								if err != nil {
									log.Printf("Failed to forward message to peer: %v", err)
								}
							}
						}
					})
				})
			}(update.peerId, update.peer, connPeer)
		}
	}()
	go PollPeers(allPeersDescription, hostSecret, roomName)
	for {
		serverPeers := <-allPeersDescription
		for peerId, peer := range serverPeers {
			peerUpdates <- PeerMessage{peerId: peerId, peer: peer}
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

func CreateOffer() (
	offer webrtc.SessionDescription,
	pendingIceCandidates []*webrtc.ICECandidate,
	peerConnection *webrtc.PeerConnection,
	dataChannel *webrtc.DataChannel,
) {
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
