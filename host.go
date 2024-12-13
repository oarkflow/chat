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

func HostRoomAndEstablishConnection(roomName, roomPassword string) {
	hostSecret, err := createRoom(roomName, roomPassword)
	if err != nil {
		log.Fatalf("Failed to create room: %v", err)
	}
	fmt.Printf("Room '%s' created successfully. Host secret: %s\n", roomName, hostSecret)

	peerChan := make(chan map[string]Peer, 1)
	go pollForPeerConnection(hostSecret, roomName, peerChan)

	var candidatesMux sync.Mutex
	pendingCandidates := make([]*webrtc.ICECandidate, 0)

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:" + StunServerAddress},
			},
		},
	}

	peers := <-peerChan
	fmt.Println("mhm. yep, stopped polling alright", " peers are", peers)
	for peerID, peer := range peers {
		fmt.Printf("Connecting to peer: %s\n", peerID)

		peerConnection, err := webrtc.NewPeerConnection(config)
		if err != nil {
			log.Printf("Failed to create PeerConnection: %v", err)
			continue
		}
		peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
			candidatesMux.Lock()
			defer candidatesMux.Unlock()
			if c != nil {
				pendingCandidates = append(pendingCandidates, c)
			}
		})

		peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
			fmt.Printf("Connection state with peer %s: %s\n", peerID, state.String())
		})

		err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  peer.OfferSDP,
		})
		if err != nil {
			log.Printf("Failed to set remote description: %v", err)
			continue
		}
		for _, candidate := range peer.OfferICECandidates {
			err := peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate})
			if err != nil {
				log.Printf("Failed to add ICE candidate for peer: %v", err)
			}
		}
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			log.Printf("Failed to create answer: %v", err)
			continue
		}

		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			log.Printf("Failed to set local description: %v", err)
			continue
		}

		peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
			dc.OnMessage(func(msg webrtc.DataChannelMessage) {
				fmt.Println(string(msg.Data))
			})
			dc.OnOpen(func() {
				reader := bufio.NewReader(os.Stdin)
				clearScreen()
				fmt.Println("data channel has been opened")
				for {
					input, _ := reader.ReadString('\n')
					dc.SendText("HOST: " + input)
				}
			})
		})
		fmt.Println("gathering started")
		<-webrtc.GatheringCompletePromise(peerConnection)
		fmt.Println("gathering complete")
		sendAnswertoServer(roomName, hostSecret, peerID, answer.SDP, pendingCandidates)
		fmt.Printf("Answer sent to peer: %s\n", peerID)

	}
	for {
		time.Sleep(time.Millisecond * 16)
	}
}
