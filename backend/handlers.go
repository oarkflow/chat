package main

import (
	"encoding/json"
	"net/http"
)

func (rs *RoomStore) AddPeerHandler(w http.ResponseWriter, r *http.Request) {
	req, err := ParseRequest[AddPeerRequest](r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response, err := rs.AddPeer(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeResponse(w, response)
}

func (rs *RoomStore) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	req, err := ParseRequest[CreateRoomRequest](r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response, err := rs.createRoom(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeResponse(w, response)
}

func (rs *RoomStore) GetAnswerHandler(w http.ResponseWriter, r *http.Request) {
	req, err := ParseRequest[GetAnswersRequest](r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response, err := rs.GetAnswers(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeResponse(w, response)
}

func (rs *RoomStore) GetPeerHandler(w http.ResponseWriter, r *http.Request) {
	req, err := ParseRequest[GetPeersRequest](r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	response, err := rs.GetPeers(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeResponse(w, response)
}

func (rs *RoomStore) GetAllRoomsHandler(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, rs.GetRooms())
}

func (rs *RoomStore) GetAllRoomsScaryHandler(w http.ResponseWriter, r *http.Request) {
	rooms := rs.GetAllRoomsScary()
	response := GetAllRoomsScaryResponse{
		Rooms: rooms,
	}
	writeResponse(w, response)
}

func (rs *RoomStore) SetAnswerHandler(w http.ResponseWriter, r *http.Request) {
	var req SetAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := rs.SetAnswer(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeResponse(w, map[string]any{"message": "Answer created"})
}
