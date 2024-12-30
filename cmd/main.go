package main

import (
	"log"

	"github.com/oarkflow/chat"
)

var ()

func main() {
	mode, roomName, roomPassword, userName := chat.Init()
	switch mode {
	case chat.Host:
		hostSecret, err := chat.CreateRoom(roomName, roomPassword)
		if err != nil {
			log.Fatal(err)
		}
		go chat.UpdateHost(hostSecret, roomName)
	}
	chat.ConnectToHost(roomName, roomPassword, userName)
}
