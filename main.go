package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"chat-app/systems"
)

var (
	StunServerAddress = "stun1.l.google.com:19302"
)

func main() {
	AppMode := systems.DisplayModeOptions()
	var WebrtcConfiguration = systems.CreateWebrtcConfiguration(StunServerAddress)
	switch AppMode {
	case systems.ModeHost:
		roomName, roomPassword := systems.DisplayRoomConfigOptions()
		hostSecret, err := systems.CreateRoom(roomName, roomPassword)
		if err != nil {
			log.Fatal(err)
		}
		go systems.UpdateHost(hostSecret, roomName, WebrtcConfiguration)
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your username: ")
		username, _ := reader.ReadString('\n')
		username = strings.Trim(username, "\n")
		systems.ConnectToHost(roomName, roomPassword, username, WebrtcConfiguration)
	case systems.ModeJoin:
		roomName, roomPassword, username := systems.DisplayRoomJoinOptions()
		systems.ConnectToHost(roomName, roomPassword, username, WebrtcConfiguration)
	}
}
