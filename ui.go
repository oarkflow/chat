package chat

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
)

type Option int

const (
	Host Option = iota
	Join
)

func (c Option) String() string {
	return [...]string{"host", "join"}[c]
}

func Init() (opt Option, roomName, roomPassword, userName string) {
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[Option]().
				Title("Would you like to Host a chat-room or Join an existing chat-room?").
				Value(&opt).
				Options(
					huh.NewOption("Host new chatroom", Host),
					huh.NewOption("Join existing chatroom", Join),
				),
		),
	).Run()
	if err != nil {
		fmt.Println("Trouble in chat paradise:", err)
		os.Exit(1)
	}
	if opt == Host {
		err = huh.NewForm(
			huh.NewGroup(
				roomInput("Enter Room Name (default - general)", &roomName),
				passwordInput("Set Room Password", &roomPassword),
				usernameInput("Your Host Username", &userName),
			),
		).Run()
		if err != nil {
			fmt.Println("Error while setting up hosting:", err)
			os.Exit(1)
		}
	} else if opt == Join {
		err = huh.NewForm(
			huh.NewGroup(
				roomInput("Enter Room Name to Join (default - general)", &roomName),
				passwordInput("Enter Room Password", &roomPassword),
				usernameInput("Your Username", &userName),
			),
		).Run()
		if err != nil {
			fmt.Println("Error while joining a chatroom:", err)
			os.Exit(1)
		}
	}
	return
}

func roomInput(label string, value *string) *huh.Input {
	return huh.NewInput().Inline(true).Validate(NotEmpty("Room name cannot be empty")).Title(label).Value(value)
}

func passwordInput(label string, value *string) *huh.Input {
	return huh.NewInput().Inline(true).Title(label).Value(value)
}

func usernameInput(label string, value *string) *huh.Input {
	return huh.NewInput().Inline(true).Validate(NotEmpty("Username cannot be empty")).Title(label).Value(value)
}

func NotEmpty(msg string) func(string) error {
	return func(s string) error {
		s = strings.TrimSpace(s)
		if s == "" {
			return fmt.Errorf(msg)
		}
		return nil
	}
}

func AskForMessageInput() (message string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Type a message: ")
	message, _ = reader.ReadString('\n')
	return
}
