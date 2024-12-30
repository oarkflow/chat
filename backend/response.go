package main

import (
	"encoding/json"
	"net/http"
)

type AddPeerResponse struct {
	Secret string `json:"peerSecret"`
}

type CreateRoomResponse struct {
	RoomID     string `json:"roomId"`
	HostSecret string `json:"hostSecret"`
}

type GetAnswersResponse struct {
	AnswerSDP           string   `json:"answerSdp"`
	AnswerIceCandidates []string `json:"answerIceCandidates"`
}

type GetAllRoomsResponse struct {
	Rooms []RoomSummary `json:"rooms"`
}

type GetAllRoomsScaryResponse struct {
	Rooms []Room `json:"rooms"`
}

func writeResponse(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if response != nil {
		json.NewEncoder(w).Encode(response)
	}
}
