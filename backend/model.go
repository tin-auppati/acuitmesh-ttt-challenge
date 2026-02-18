// backend/models.go
package main

import (
	"time"
)

// User - แทนตาราง users
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // ไม่ส่ง password กลับไปหน้าบ้าน
	CreatedAt    time.Time `json:"created_at"`
}

// Game - แทนตาราง games
type Game struct {
	ID            int       `json:"id"`
	Player1ID     *int      `json:"player1_id"` // ใช้ pointer เพราะอาจจะเป็น null ได้ (เผื่อไว้)
	Player2ID     *int      `json:"player2_id"`
	CurrentTurnID *int      `json:"current_turn_id"`
	Board         string    `json:"board"`   // "---------"
	Status        string    `json:"status"`  // WAITING, IN_PROGRESS, FINISHED
	WinnerID      *int      `json:"winner_id"`
	CreatedAt     time.Time `json:"created_at"`
}

// Move - แทนตาราง moves
type Move struct {
	ID        int       `json:"id"`
	GameID    int       `json:"game_id"`
	PlayerID  int       `json:"player_id"`
	X         int       `json:"x"`
	Y         int       `json:"y"`
	CreatedAt time.Time `json:"created_at"`
}