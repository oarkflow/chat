package main

import (
	"encoding/json"
	"io"
)

type AddPeerRequest struct {
	PeerID             string   `json:"peerId"`
	RoomName           string   `json:"roomName"`
	Password           string   `json:"password"`
	OfferSDP           string   `json:"offerSdp"`
	OfferIceCandidates []string `json:"offerIceCandidates"`
}

type CreateRoomRequest struct {
	Name      string `json:"name"`
	Password  string `json:"password,omitempty"`
	MaxPeers  int    `json:"max_peers"`
	IsPrivate bool   `json:"isPrivate"`
}

type GetAnswersRequest struct {
	RoomName   string `json:"roomName"`
	PeerId     string `json:"peerId"`
	PeerSecret string `json:"peerSecret"`
}

type GetPeersRequest struct {
	RoomName   string `json:"roomName"`
	HostSecret string `json:"hostSecret"`
}

type SetAnswerRequest struct {
	RoomName            string   `json:"roomName"`
	PeerId              string   `json:"peerId"`
	HostSecret          string   `json:"hostSecret"`
	AnswerSDP           string   `json:"answerSdp"`
	AnswerIceCandidates []string `json:"answerIceCandidates"`
}

func ParseRequest[T any](body io.Reader) (T, error) {
	var req T
	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return req, err
	}
	return req, nil
}
