package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// เสกเลข 6 หลัก
func GenerateRoomCode() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func CreateGameHandler(c *gin.Context) {
	userIDContext, _ := c.Get("userID")
	playerID := userIDContext.(int)

	roomCode := GenerateRoomCode()

	var gameID int
	query := `
		INSERT INTO games (room_code, player1_id, current_turn_id, status, board) 
		VALUES ($1, $2, $2, 'WAITING', '---------') 
		RETURNING id`

	err := DB.QueryRow(query, roomCode, playerID).Scan(&gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create game", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Game created successfully",
		"room_code": roomCode,
		"status":    "WAITING",
	})
}

func JoinGameHandler(c *gin.Context) {
	var req struct {
		RoomCode string `json:"room_code" binding:"required,len=6"`
	}
	userIDContext, _ := c.Get("userID")
	playerID := userIDContext.(int)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room code format (must be 6 digits)"})
		return
	}
	//atomic ทั้งก้อน begin - commit
	tx, err := DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not start transaction"})
		return
	}

	// ถ้าเกิดอะไรขึ้นผิดพลาดให้ Rollback เสมอ
	defer tx.Rollback()

	// 1. SELECT ... FOR UPDATE เพื่อ Lock แถวเกมนั้นไว้ก่อน
	var gameID, p1ID int
	var p2ID *int
	var status string

	queryLock := `SELECT id, player1_id, player2_id, status FROM games WHERE room_code = $1 FOR UPDATE`
	err = tx.QueryRow(queryLock, req.RoomCode).Scan(&gameID, &p1ID, &p2ID, &status)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found. Check your code!"})
		return
	}

	// 2. validation ต่างๆ เช่น เช็คว่าเกมเต็มหรือยัง เช็คว่าเกมอยู่ในสถานะ WAITING หรือเปล่า เช็คว่า player ที่จะเข้ามาไม่ได้เป็น player1 อยู่แล้ว
	if p1ID == playerID {
		c.JSON(http.StatusConflict, gin.H{"error": "You are already the host of this room"})
		return
	}
	if p2ID != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Room is full"})
		return
	}
	if status != "WAITING" {
		c.JSON(http.StatusConflict, gin.H{"error": "Game is already in progress"})
		return
	}

	// 3. UPDATE เพื่อ set player2_id และเปลี่ยนสถานะเกมเป็น IN_PROGRESS
	queryUpdate := `UPDATE games SET player2_id = $1, status = 'IN_PROGRESS' WHERE id = $2`
	_, err = tx.Exec(queryUpdate, playerID, gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join game"})
		return
	}

	// 4. Commit Transaction เพื่อยืนยันข้อมูลและปล่อย Lock
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Joined game successfully",
		"room_code": req.RoomCode,
		"status":    "IN_PROGRESS",
	})
}

func MakeMoveHandler(c *gin.Context) {
	var req struct {
		RoomCode string `json:"room_code" binding:"required,len=6"`
		X        int    `json:"x" binding:"min=0,max=2"`
		Y        int    `json:"y" binding:"min=0,max=2"`
	}
	userIDContext, _ := c.Get("userID")
	playerID := userIDContext.(int)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//atomic ทั้งก้อน begin - commit
	tx, _ := DB.Begin()
	defer tx.Rollback()

	// 1. lock
	var gameID int
	var board, status string
	var p1ID, p2ID, turnID int
	query := `SELECT id, board, status, player1_id, player2_id, current_turn_id FROM games WHERE room_code = $1 FOR UPDATE`
	err := tx.QueryRow(query, req.RoomCode).Scan(&gameID, &board, &status, &p1ID, &p2ID, &turnID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	// 2. validation
	if status != "IN_PROGRESS" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game is not in progress"})
		return
	}
	if turnID != playerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your turn"})
		return
	}

	index := req.Y*3 + req.X
	if board[index] != '-' {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cell already occupied"})
		return
	}

	//3. update board
	char := "X"
	nextTurn := p2ID
	if playerID == p2ID {
		char = "O"
		nextTurn = p1ID
	}

	newBoard := board[:index] + char + board[index+1:]

	//4. check winner
	winnerSign := CheckWinner(newBoard)
	newStatus := "IN_PROGRESS"

	var winnerID *int

	if winnerSign != "" {
		if winnerSign == "DRAW" {
			newStatus = "DRAW"
		} else {
			newStatus = "FINISHED"
			winnerID = &playerID
		}
	}

	updateQuery := `UPDATE games SET board = $1, current_turn_id = $2, status = $3, winner_id = $4 WHERE id = $5`
	_, err = tx.Exec(updateQuery, newBoard, nextTurn, newStatus, winnerID, gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update game state"})
		return
	}

	_, err = tx.Exec(`INSERT INTO moves (game_id, player_id, x, y, move_order) 
			 VALUES ($1, $2, $3, $4, (SELECT count(*)+1 FROM moves WHERE game_id=$1))`,
		gameID, playerID, req.X, req.Y)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record move"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"board":  newBoard,
		"status": newStatus,
	})
}

// GetGameHandler - ดูสถานะเกมปัจจุบัน (ใช้สำหรับ Polling)
func GetGameHandler(c *gin.Context) {
	roomCode := c.Param("id")

	var game struct {
		ID            int    `json:"id"`
		RoomCode      string `json:"room_code"`
		Player1ID     int    `json:"player1_id"`
		Player2ID     *int   `json:"player2_id"`
		CurrentTurnID int    `json:"current_turn_id"`
		Board         string `json:"board"`
		Status        string `json:"status"`
		WinnerID      *int   `json:"winner_id"`
	}

	query := `SELECT id, room_code, player1_id, player2_id, current_turn_id, board, status, winner_id 
			  FROM games WHERE room_code = $1`

	row := DB.QueryRow(query, roomCode)
	err := row.Scan(&game.ID, &game.RoomCode, &game.Player1ID, &game.Player2ID, &game.CurrentTurnID, &game.Board, &game.Status, &game.WinnerID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	c.JSON(http.StatusOK, game)
}

// GetGameMovesHandler - ดูประวัติการเดินของเกม (ใช้สำหรับ Polling)
func GetGameMovesHandler(c *gin.Context) {
	roomCode := c.Param("id")

	query := `
		SELECT m.id, m.game_id, m.player_id, m.x, m.y, m.created_at 
		FROM moves m
		JOIN games g ON m.game_id = g.id
		WHERE g.room_code = $1 
		ORDER BY m.move_order ASC
	`
	rows, err := DB.Query(query, roomCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch moves"})
		return
	}

	defer rows.Close()

	var moves []Move
	for rows.Next() {
		var m Move
		if err := rows.Scan(&m.ID, &m.GameID, &m.PlayerID, &m.X, &m.Y, &m.CreatedAt); err != nil {
			continue
		}
		moves = append(moves, m)
	}
	c.JSON(http.StatusOK, gin.H{"moves": moves})
}
