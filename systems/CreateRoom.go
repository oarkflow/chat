package systems

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func CreateRoom(name, password string) (hostSecret string, err error) {
	reqBody := map[string]string{
		"name":     name,
		"password": password,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}
	resp, err := http.Post(SignalingServerAddress+"/create-room", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}
	respBody := map[string]string{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	return respBody["hostSecret"], nil
}
