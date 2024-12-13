package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var (
	SignalingServerAddress = "https://71776aac073ddf.lhr.life"
	StunServerAddress      = "stun1.l.google.com:19302"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("enter the server address")
	address, _ := reader.ReadString('\n')
	SignalingServerAddress = "https://" + strings.TrimSpace(address)

	fmt.Println("Select an option:")
	fmt.Println("1. Host a Room")
	fmt.Println("2. Join a Room")
	option, _ := reader.ReadString('\n')
	option = strings.TrimSpace(option)

	switch option {
	case "1":
		fmt.Print("Enter room name: ")
		roomName, _ := reader.ReadString('\n')
		roomName = strings.TrimSpace(roomName)

		fmt.Print("Enter room password: ")
		roomPassword, _ := reader.ReadString('\n')
		roomPassword = strings.TrimSpace(roomPassword)

		HostRoomAndEstablishConnection(roomName, roomPassword)

	case "2":
		fmt.Print("Enter room name: ")
		roomName, _ := reader.ReadString('\n')
		roomName = strings.TrimSpace(roomName)

		fmt.Print("Enter room password: ")
		roomPassword, _ := reader.ReadString('\n')
		roomPassword = strings.TrimSpace(roomPassword)

		fmt.Print("Enter your peer ID: ")
		peerID, _ := reader.ReadString('\n')
		peerID = strings.TrimSpace(peerID)

		connectToRoom(roomName, roomPassword, peerID)

	default:
		fmt.Println("Invalid option. Exiting.")
	}
}
