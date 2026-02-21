// backend/scripts/concurrency_test.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const baseURL = "http://localhost:8080/api"

type APIResponse struct {
	Token   string  `json:"token"`
	GameID  float64 `json:"game_id"`
	ID      float64 `json:"id"`
	Error   string  `json:"error"`
	Message string  `json:"message"`
}

// ฟังก์ชันช่วยยิง HTTP Request และแปลงผลลัพธ์เป็น JSON
func sendRequest(method, url string, payload map[string]interface{}, token string) (map[string]interface{}, error) {
	var reqBody []byte
	if payload != nil {
		reqBody, _ = json.Marshal(payload)
	}

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	return result, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	testUser := fmt.Sprintf("testuser%d", rand.Intn(100000))
	testPass := "password123"

	fmt.Printf("[1/4] Registering new user: %s...\n", testUser)
	_, err := sendRequest("POST", baseURL+"/register", map[string]interface{}{
		"username": testUser,
		"password": testPass,
	}, "")
	if err != nil {
		fmt.Println("Register Failed:", err)
		return
	}

	fmt.Printf("[2/4] Logging in...\n")
	loginResp, err := sendRequest("POST", baseURL+"/login", map[string]interface{}{
		"username": testUser,
		"password": testPass,
	}, "")
	if err != nil {
		fmt.Println("Login Failed:", err)
		return
	}

	token, ok := loginResp["token"].(string)
	if !ok {
		fmt.Println("Cannot extract token from response:", loginResp)
		return
	}

	fmt.Printf("[3/4] Creating a new game...\n")
	gameResp, err := sendRequest("POST", baseURL+"/games", nil, token)
	if err != nil {
		fmt.Println("Create Game Failed:", err)
		return
	}

	// เช็คคีย์ "game_id" ให้ตรงกับที่ API Return ออกมา
	roomCode, ok := gameResp["room_code"].(string)
	if !ok {
		fmt.Println("Cannot extract room_code from response:", gameResp)
		return
	}

	fmt.Printf("Game Created! Room Code: %s\n\n", roomCode)

	fmt.Printf("[3.5/4] Creating Player 2 to join the game...\n")
    testUser2 := fmt.Sprintf("player2%d", rand.Intn(100000)) 
    
    // Register Player 2
    _, err = sendRequest("POST", baseURL+"/register", map[string]interface{}{
        "username": testUser2,
        "password": testPass,
    }, "")
    if err != nil {
        fmt.Println("Player 2 Register Failed:", err)
        return
    }

    // Login Player 2
    loginResp2, err := sendRequest("POST", baseURL+"/login", map[string]interface{}{
        "username": testUser2,
        "password": testPass,
    }, "")
    if err != nil {
        fmt.Println("Player 2 Login Failed:", err)
        return
    }
    
    // ดึง Token ของ Player 2
    token2, ok := loginResp2["token"].(string)
    if !ok {
        fmt.Println("Cannot extract token for Player 2")
        return
    }

    // Player 2 Join Room ด้วย roomCode
    _, err = sendRequest("POST", baseURL+"/games/join", map[string]interface{}{
        "room_code": roomCode, 
    }, token2)

    if err != nil {
        fmt.Println("Player 2 Join Failed:", err)
        return
    }
    fmt.Println("Player 2 joined successfully! Game is now IN_PROGRESS.")

	fmt.Printf("[4/4] FIRING CONCURRENT REQUESTS (Race Condition Test)\n")
	requestsCount := 20
	var wg sync.WaitGroup
	var successCount int32
	var failCount int32

	movePayload, _ := json.Marshal(map[string]interface{}{
		"room_code": roomCode,
		"x":         0,
		"y":         0,
	})

	startSignal := make(chan struct{})

	for i := 0; i < requestsCount; i++ {
		wg.Add(1)
		go func(reqID int) {
			defer wg.Done()
			<-startSignal

			req, _ := http.NewRequest("POST", baseURL+"/games/move", bytes.NewBuffer(movePayload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)

			if err != nil {
				atomic.AddInt32(&failCount, 1)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				atomic.AddInt32(&successCount, 1)
				fmt.Printf("[Req %02d] Success! Move accepted.\n", reqID)
			} else {
				atomic.AddInt32(&failCount, 1)
				fmt.Printf("[Req %02d] Rejected (Status %d)\n", reqID, resp.StatusCode)
			}
		}(i)
	}
	time.Sleep(1 * time.Second)
	fmt.Println("GO! Firing 20 requests at the exact same time...")
	close(startSignal)
	wg.Wait()

	fmt.Println("\n --- TEST RESULTS ---")
	fmt.Printf("Success (Expected: 1): %d\n", successCount)
	fmt.Printf("Rejected (Expected: 19): %d\n", failCount)

	if successCount == 1 {
		fmt.Println("RESULT: PASS! Your database properly locked the row.")
	} else {
		fmt.Println("RESULT: FAIL! Race condition detected.")
	}
}
