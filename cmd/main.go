package main

import (
	"log"

	"github.com/oarkflow/chat"
	"github.com/oarkflow/chat/utils"
)

var (
	StunServerAddress = "stun1.l.google.com:19302"
)

func main() {
	AppMode := chat.DisplayModeOptions()
	WebrtcConfiguration := utils.GetConfig(StunServerAddress)
	switch AppMode {
	case chat.ModeHost:
		roomName, roomPassword := chat.DisplayRoomConfigOptions()
		hostSecret, err := chat.CreateRoom(roomName, roomPassword)
		if err != nil {
			log.Fatal(err)
		}
		go chat.UpdateHost(hostSecret, roomName, WebrtcConfiguration)
		username := chat.AskForUsernameInput()
		chat.ConnectToHost(roomName, roomPassword, username, WebrtcConfiguration)
	case chat.ModeJoin:
		roomName, roomPassword, username := chat.DisplayRoomJoinOptions()
		chat.ConnectToHost(roomName, roomPassword, username, WebrtcConfiguration)
	}
}
