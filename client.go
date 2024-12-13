package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pion/webrtc/v4"
)

func connectToRoom(roomName string, password string, peerID string) {
	var candidatesMux sync.Mutex
	pendingCandidates := make([]*webrtc.ICECandidate, 0)

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:" + StunServerAddress},
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Fatalf("Failed to create peer connection: %v", err)
	}
	defer peerConnection.Close()

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		candidatesMux.Lock()
		defer candidatesMux.Unlock()
		if c != nil {
			fmt.Println("Generated ICE candidate:", c.ToJSON().Candidate)
			pendingCandidates = append(pendingCandidates, c)
		}
	})

	dataChannel, err := peerConnection.CreateDataChannel("chat", nil)
	if err != nil {
		log.Fatalf("Failed to create data channel: %v", err)
	}

	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("Connection state with peer %s: %s\n", peerID, state.String())
	})
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Println(string(msg.Data))
	})
	dataChannel.OnOpen(func() { fmt.Println("datachannel opened, start chatting") })
	dataChannel.OnError(func(err error) { fmt.Println("there was an error with the datachannel ", err) })
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		log.Fatalf("Failed to create offer: %v", err)
	}

	if err := peerConnection.SetLocalDescription(offer); err != nil {
		log.Fatalf("Failed to set local description: %v", err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	<-gatherComplete
	fmt.Println("gathered ice candidates")

	peerSecret := AddPeer(peerID, roomName, password, offer.SDP, pendingCandidates)
	if peerSecret == "" {
		log.Fatal("Failed to add peer to the signaling server.")
	}

	answerResponse := pollForAnswer(roomName, peerID, peerSecret)
	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answerResponse.AnswerSDP,
	}); err != nil {
		log.Fatalf("Failed to set remote description: %v", err)
	}
	fmt.Println("answer reponse is here ", answerResponse)
	fmt.Println("ice candidates are ", answerResponse.AnswerIceCandidates)
	for _, candidate := range answerResponse.AnswerIceCandidates {
		if err := peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate}); err != nil {
			log.Printf("Failed to add ICE candidate: %v", err)
		}
	}
	// clearScreen()

	for {
		time.Sleep(time.Second)

		reader := bufio.NewReader(os.Stdin)

		input, _ := reader.ReadString('\n')
		dataChannel.SendText(peerID + ": " + input)

	}
}
