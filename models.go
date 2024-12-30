package chat

import (
	"github.com/pion/webrtc/v4"
)

var (
	SignalingServerAddress = "http://localhost:8080"
)

func getUrl(uri string) string {
	return SignalingServerAddress + uri
}

type ServerPeerDescription struct {
	Secret              string
	OfferICECandidates  []string `json:"offerIceCandidates"`
	AnswerICECandidates []string `json:"answerIceCandidates"`
	OfferSDP            string   `json:"offerSdp"`
	AnswerSDP           string   `json:"answerSdp"`
}

type ConnectedPeer struct {
	PeerId          string
	PeerDescription ServerPeerDescription
	PeerConnection  *webrtc.PeerConnection
	DataChannel     *webrtc.DataChannel
}

type AnswerResponse struct {
	AnswerSDP           string   `json:"answerSdp"`
	AnswerIceCandidates []string `json:"answerIceCandidates"`
}

type AddPeerRequest struct {
	PeerID             string   `json:"peerId"`
	RoomName           string   `json:"roomName"`
	Password           string   `json:"password"`
	OfferSDP           string   `json:"offerSdp"`
	OfferIceCandidates []string `json:"offerIceCandidates"`
}
