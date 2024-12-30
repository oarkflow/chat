package systems

import (
	"fmt"
	"net/http"
)

func CreateRoom(name, password string) (hostSecret string, err error) {
	reqBody := map[string]string{
		"name":     name,
		"password": password,
	}
	respBody, statusCode, err := Request[map[string]string]("/create-room", reqBody)
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("server returned status %d", statusCode)
	}
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	return respBody["hostSecret"], nil
}
