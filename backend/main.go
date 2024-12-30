package main

import (
	"fmt"
	"log"
	"net/http"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	roomStore := NewRoomStore()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /get-rooms", roomStore.GetAllRoomsHandler)
	mux.HandleFunc("GET /get-rooms-scary", roomStore.GetAllRoomsScaryHandler)
	mux.HandleFunc("POST /get-peers", roomStore.GetPeerHandler)
	mux.HandleFunc("POST /get-answer", roomStore.GetAnswerHandler)
	mux.HandleFunc("POST /create-room", roomStore.CreateRoomHandler)
	mux.HandleFunc("POST /set-answer", roomStore.SetAnswerHandler)
	mux.HandleFunc("POST /add-peer", roomStore.AddPeerHandler)

	fmt.Println("starting server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(mux)))
}
