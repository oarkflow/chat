package systems

import (
	"bufio"
	"fmt"
	"os"
)

func AskForMessageInput() (message string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Type a message: ")
	message, _ = reader.ReadString('\n')

	return message
}
