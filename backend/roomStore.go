package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID         string
	HostSecret string
	Password   string
	IsPrivate  bool
	MaxPeers   int
	Peers      map[string]*Peer
	CreatedAt  int64
	LastActive int64
}

type Peer struct {
	Secret              string
	OfferICECandidates  []string `json:"offerIceCandidates"`
	AnswerICECandidates []string `json:"answerIceCandidates"`
	OfferSDP            string   `json:"offerSdp"`
	AnswerSDP           string   `json:"answerSdp"`
}

type RoomSummary struct {
	ID             string `json:"id"`
	NumPeers       int    `json:"numPeers"`
	MaxPeers       int    `json:"maxPeers"`
	AvailableSlots int    `json:"availableSlots"`
	IsPrivate      bool   `json:"isPrivate"`
}

type RoomStore struct {
	rooms map[string]*Room
	mutex sync.RWMutex
}

func (rs *RoomStore) CreateRoom(id, hostSecret, password string, isPrivate bool, maxPeers int) (*Room, error) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	if room, exists := rs.rooms[id]; exists {
		return room, nil
	}
	room := &Room{
		ID:         id,
		HostSecret: hostSecret,
		Password:   password,
		IsPrivate:  isPrivate,
		Peers:      make(map[string]*Peer),
		MaxPeers:   maxPeers,
		CreatedAt:  time.Now().Unix(),
		LastActive: time.Now().Unix(),
	}
	rs.rooms[id] = room
	return room, nil
}
func (rs *RoomStore) GetAllRoomsScary() []Room {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	var rooms []Room
	for _, room := range rs.rooms {
		rooms = append(rooms, *room)
	}
	return rooms
}
func (rs *RoomStore) GetAllRooms() []RoomSummary {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	var summaries []RoomSummary
	for _, room := range rs.rooms {
		if room.IsPrivate {
			continue
		}
		summaries = append(summaries, RoomSummary{
			ID:        room.ID,
			NumPeers:  len(room.Peers),
			IsPrivate: room.IsPrivate,
			MaxPeers:  room.MaxPeers,
		})
	}
	return summaries
}

func (rs *RoomStore) RemovePeerFromRoom(roomID, peerID string) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	room, exists := rs.rooms[roomID]
	if !exists {
		return errors.New("room not found")
	}
	if _, exists := room.Peers[peerID]; !exists {
		return errors.New("peer not found in the room")
	}
	delete(room.Peers, peerID)
	room.LastActive = time.Now().Unix()
	return nil
}

func (rs *RoomStore) GetRoom(name string) (*Room, error) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	room, exists := rs.rooms[name]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}
	return room, nil
}

func (rs *RoomStore) UpdateRoom(roomID, hostSecret, password string, isPrivate bool) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	room, exists := rs.rooms[roomID]
	if !exists {
		return errors.New("room not found")
	}
	if room.HostSecret != hostSecret {
		return errors.New("invalid host secret")
	}
	room.Password = password
	room.IsPrivate = isPrivate
	room.LastActive = time.Now().Unix()
	return nil
}

func (rs *RoomStore) DeleteRoom(roomID, hostSecret string) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	room, exists := rs.rooms[roomID]
	if !exists {
		return errors.New("room not found")
	}
	if room.HostSecret != hostSecret {
		return errors.New("invalid host secret")
	}
	delete(rs.rooms, roomID)
	return nil
}
func (rs *RoomStore) CleanupInactiveRooms(timeout int64) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	now := time.Now().Unix()
	for id, room := range rs.rooms {
		if now-room.LastActive > timeout {
			delete(rs.rooms, id)
		}
	}
}

func NewRoomStore() *RoomStore {
	return &RoomStore{
		rooms: make(map[string]*Room),
	}
}

func (rs *RoomStore) AddPeer(req AddPeerRequest) (*AddPeerResponse, error) {
	if req.PeerID == "" || req.RoomName == "" || req.OfferSDP == "" || len(req.OfferIceCandidates) == 0 {
		return nil, fmt.Errorf("peer ID, SDP, offerIceCandidates[], and room name are required")
	}
	room, err := rs.GetRoom(req.RoomName)
	if err != nil {
		return nil, fmt.Errorf("room not found")
	}
	if room.Password != "" && room.Password != req.Password {
		return nil, fmt.Errorf("invalid room password")
	}
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	if len(room.Peers) >= room.MaxPeers {
		return nil, fmt.Errorf("room is full")
	}
	if _, exists := room.Peers[req.PeerID]; exists {
		delete(room.Peers, req.PeerID)
	}
	newPeer := &Peer{
		Secret:              uuid.New().String(),
		AnswerICECandidates: make([]string, 0),
		OfferICECandidates:  req.OfferIceCandidates,
		OfferSDP:            req.OfferSDP,
	}
	room.Peers[req.PeerID] = newPeer
	room.LastActive = time.Now().Unix()
	return &AddPeerResponse{Secret: newPeer.Secret}, nil
}

func (rs *RoomStore) createRoom(req CreateRoomRequest) (*CreateRoomResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("room name is required")
	}
	if req.MaxPeers <= 0 {
		req.MaxPeers = 10
	}
	hostSecret := uuid.New().String()
	_, err := rs.CreateRoom(req.Name, hostSecret, req.Password, req.IsPrivate, req.MaxPeers)
	if err != nil {
		return nil, err
	}
	return &CreateRoomResponse{
		RoomID:     req.Name,
		HostSecret: hostSecret,
	}, nil
}

func (rs *RoomStore) GetAnswers(req GetAnswersRequest) (*GetAnswersResponse, error) {
	if req.PeerId == "" || req.PeerSecret == "" || req.RoomName == "" {
		return nil, fmt.Errorf("roomName, peerId and peerSecret are required")
	}
	room, err := rs.GetRoom(req.RoomName)
	if err != nil {
		return nil, err
	}
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	peer, ok := room.Peers[req.PeerId]
	if !ok {
		return nil, fmt.Errorf("this peer is not in this room")
	}
	if peer.AnswerSDP == "" || len(peer.AnswerICECandidates) == 0 {
		return nil, fmt.Errorf("waiting on answer")
	}
	return &GetAnswersResponse{
		AnswerSDP:           peer.AnswerSDP,
		AnswerIceCandidates: peer.AnswerICECandidates,
	}, nil
}

func (rs *RoomStore) GetPeers(req GetPeersRequest) (map[string]*Peer, error) {
	if req.HostSecret == "" || req.RoomName == "" {
		return nil, fmt.Errorf("hostSecret and roomName are required")
	}
	room, err := rs.GetRoom(req.RoomName)
	if err != nil {
		return nil, err
	}
	if len(room.Peers) == 0 {
		return nil, fmt.Errorf("no peers yet")
	}
	return room.Peers, nil
}

func (rs *RoomStore) GetRooms() GetAllRoomsResponse {
	rooms := rs.GetAllRooms()
	for i := range rooms {
		rooms[i].AvailableSlots = rooms[i].MaxPeers - rooms[i].NumPeers
	}
	return GetAllRoomsResponse{
		Rooms: rooms,
	}
}

func (rs *RoomStore) SetAnswer(req SetAnswerRequest) error {
	if req.PeerId == "" || req.HostSecret == "" || req.RoomName == "" || req.AnswerSDP == "" || len(req.AnswerIceCandidates) == 0 {
		return fmt.Errorf("answerSdp, roomName, peerId, answerIceCandidates, and hostSecret are required")
	}
	room, err := rs.GetRoom(req.RoomName)
	if err != nil {
		return err
	}
	if room.HostSecret != req.HostSecret {
		return err
	}
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	peer, ok := room.Peers[req.PeerId]
	if !ok {
		return fmt.Errorf("this peer is not in this room")
	}
	peer.AnswerSDP = req.AnswerSDP
	peer.AnswerICECandidates = req.AnswerIceCandidates
	return nil
}
