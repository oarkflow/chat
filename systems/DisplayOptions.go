package systems

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	ModeHost = iota
	ModeJoin
)

func DisplayModeOptions(SignalingServerAddress string) (Mode int) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("enter the signaling server address. \npress enter to use default [%s]: ", SignalingServerAddress)
	address, _ := reader.ReadString('\n')
	if address := strings.TrimSpace(address); address != "" {
		SignalingServerAddress = address
	}

	fmt.Println("Select an option:")
	fmt.Println("1. Host a Room")
	fmt.Println("2. Join a Room")
	option, _ := reader.ReadString('\n')
	option = strings.TrimSpace(option)
	switch option {
	case "1":
		return ModeHost
	case "2":
		return ModeJoin
	default:
		log.Fatal("Invalid option, select 1 or 2")
	}
	return 0
}

func DisplayRoomConfigOptions() (roomName, roomPassword string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter room name: ")
	roomName, _ = reader.ReadString('\n')

	fmt.Print("Enter room password (press enter for no password): ")
	roomPassword, _ = reader.ReadString('\n')
	if roomName := strings.TrimSpace(roomName); roomName != "" {
		return roomName, roomPassword
	} else {
		log.Fatal("Room name cannot be empty")
	}
	return "", ""
}

func DisplayRoomJoinOptions() (roomName, roomPassword, username string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter room name: ")
	roomName, _ = reader.ReadString('\n')

	fmt.Print("Enter room password (press enter for no password): ")
	roomPassword, _ = reader.ReadString('\n')

	fmt.Print("Enter your username: ")
	username, _ = reader.ReadString('\n')
	username = strings.Trim(username, "\n")
	if cleanName := strings.TrimSpace(username); cleanName != "" {
		username = cleanName
	} else {
		log.Fatal("username cannot be empty")
	}

	if roomName := strings.TrimSpace(roomName); roomName != "" {
		return roomName, roomPassword, username
	} else {
		log.Fatal("Room name cannot be empty")
	}
	return "", "", ""
}
