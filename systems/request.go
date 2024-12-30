package systems

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	SignalingServerAddress = "http://localhost:8080"
)

func Request[T any](uri string, data any) (T, int, error) {
	uri = SignalingServerAddress + uri
	var t T
	var buf *bytes.Buffer
	switch data := data.(type) {
	case []byte:
		buf = bytes.NewBuffer(data)
	default:
		jsonData, err := json.Marshal(data)
		if err != nil {
			return t, http.StatusBadRequest, fmt.Errorf("failed to marshal request body: %w", err)
		}
		buf = bytes.NewBuffer(jsonData)
	}
	resp, err := http.Post(uri, "application/json", buf)
	if err != nil {
		return t, http.StatusBadRequest, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return t, resp.StatusCode, nil
	}
	err = json.NewDecoder(resp.Body).Decode(&t)
	if err != nil {
		return t, http.StatusBadRequest, fmt.Errorf("failed to decode response: %w", err)
	}
	return t, http.StatusOK, nil
}
