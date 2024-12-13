package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/pion/webrtc/v4"
)

// Function to create a room
func createRoom(name string, password string) (hostSecret string, err error) {
	// Create the request payload
	reqBody := map[string]string{
		"name":     name,
		"password": password,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Send the HTTP POST request
	resp, err := http.Post(SignalingServerAddress+"/create-room", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	respBody := map[string]string{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return respBody["hostSecret"], nil
}

type Peer struct {
	Secret              string
	OfferICECandidates  []string `json:"offerIceCandidates"`  // ICE candidates from the peer
	AnswerICECandidates []string `json:"answerIceCandidates"` // ICE candidates from the room creator
	OfferSDP            string   `json:"offerSdp"`
	AnswerSDP           string   `json:"answerSdp"`
}

// poll for offer ice candidates
func pollForPeerConnection(hostSecret string, roomName string, peerChan chan map[string]Peer) {
	reqBody := map[string]string{
		"hostSecret": hostSecret,
		"roomName":   roomName,
	}
	jsonData, _ := json.Marshal(reqBody)
	for {
		time.Sleep(time.Second * 6)
		resp, err := http.Post(SignalingServerAddress+"/get-peers", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			panic(err)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			// fmt.Println("error in polling for peer connection ", string(body))
			continue
		}
		peers := map[string]Peer{}
		json.Unmarshal(body, &peers)
		peerChan <- peers
		fmt.Println("finished polling for peers")
		return
	}
}

func sendAnswertoServer(roomName, hostSecret, peerID string, sdp string, answerIceCandidates []*webrtc.ICECandidate) {
	if hostSecret == "" || roomName == "" || peerID == "" {
		log.Fatal("hostSecret, roomName, or peerID is empty")
	}
	fmt.Println("the answer candidates are ", answerIceCandidates)
	fmt.Printf("Sending to server: roomName=%s, hostSecret=%s, peerID=%s\n", roomName, hostSecret, peerID)

	var iceCandidates []string
	for _, c := range answerIceCandidates {
		iceCandidates = append(iceCandidates, c.ToJSON().Candidate)
	}

	reqBody := map[string]any{
		"hostSecret":          hostSecret,
		"roomName":            roomName,
		"peerId":              peerID,
		"answerSdp":           sdp,
		"answerIceCandidates": iceCandidates,
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(SignalingServerAddress+"/set-answer", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal("Failed to send answer SDP to server: ", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("Failed to send answer SDP. Server response: %s\n", string(body))
	} else {
		fmt.Println("Sent the answer SDP to the server successfully.")
	}

}

///////client

func AddPeer(peerID string, roomName string, roomPassword string, offerSDP string, offerIceCandidates []*webrtc.ICECandidate) string {
	var iceCandiates []string
	for _, c := range offerIceCandidates {
		iceCandiates = append(iceCandiates, c.ToJSON().Candidate)
	}
	reqBody := map[string]any{
		"peerId":             peerID,
		"roomName":           roomName,
		"password":           roomPassword,
		"offerSdp":           offerSDP,
		"offerIceCandidates": iceCandiates,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return ""
	}

	resp, err := http.Post(SignalingServerAddress+"/add-peer", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error adding peer: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error response: %s\n", body)
		return ""
	}

	body, _ := io.ReadAll(resp.Body)
	JsonResponse := map[string]string{}
	json.Unmarshal(body, &JsonResponse)
	return JsonResponse["peerSecret"]
}

type GetAnswersResponse struct {
	AnswerSDP           string   `json:"answerSdp"`
	AnswerIceCandidates []string `json:"answerIceCandidates"`
}

func pollForAnswer(roomName string, peedId string, peerSecret string) GetAnswersResponse {
	reqBody := map[string]string{
		"roomName":   roomName,
		"peerSecret": peerSecret,
		"peerId":     peedId,
	}
	jsonData, _ := json.Marshal(reqBody)
	for {
		time.Sleep(time.Second * 2)
		resp, err := http.Post(SignalingServerAddress+"/get-answer", "application/json", bytes.NewBuffer(jsonData))

		if err != nil {
			fmt.Printf("Error polling for answer Sdp: %v\n", err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		var JsonResp GetAnswersResponse
		if err := json.Unmarshal(body, &JsonResp); err != nil {
			fmt.Println("error in polling for answer SDP", string(body), err)
			continue
		}
		return JsonResp

	}
}

func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}
