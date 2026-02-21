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

	var activeCount int
	checkQuery := `
		SELECT count(*) FROM games 
		WHERE (player1_id = $1 OR player2_id = $1) 
		AND status IN ('WAITING', 'IN_PROGRESS')`

	err := DB.QueryRow(checkQuery, playerID).Scan(&activeCount)

	if activeCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "You already have an active game session. Please finish or leave it first.",
		})
		return
	}

	roomCode := GenerateRoomCode()

	var gameID int
	query := `
		INSERT INTO games (room_code, player1_id, current_turn_id, status, board) 
		VALUES ($1, $2, $2, 'WAITING', '---------') 
		RETURNING id`

	err = DB.QueryRow(query, roomCode, playerID).Scan(&gameID)
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

	var activeCount int
	checkQuery := `
		SELECT count(*) FROM games 
		WHERE (player1_id = $1 OR player2_id = $1) 
		AND status IN ('WAITING', 'IN_PROGRESS')`

	err = DB.QueryRow(checkQuery, playerID).Scan(&activeCount)

	if activeCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "You already have an active game session. Please finish or leave it first.",
		})
		return
	}

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
		ID            int     `json:"id"`
		RoomCode      string  `json:"room_code"`
		Player1ID     int     `json:"player1_id"`
		Player2ID     *int    `json:"player2_id"`
		CurrentTurnID int     `json:"current_turn_id"`
		Board         string  `json:"board"`
		Status        string  `json:"status"`
		WinnerID      *int    `json:"winner_id"`
		NextRoomCode  *string `json:"next_room_code"`
		RematchP1     bool    `json:"rematch_p1"`
		RematchP2     bool    `json:"rematch_p2"`
	}

	query := `SELECT id, room_code, player1_id, player2_id, current_turn_id, board, status, winner_id, next_room_code, rematch_p1, rematch_p2 
			  FROM games WHERE room_code = $1`

	row := DB.QueryRow(query, roomCode)
	err := row.Scan(&game.ID, &game.RoomCode, &game.Player1ID, &game.Player2ID, &game.CurrentTurnID, &game.Board, &game.Status, &game.WinnerID, &game.NextRoomCode, &game.RematchP1, &game.RematchP2)
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

func CancelGameHandler(c *gin.Context) {
	roomCode := c.Param("id")
	userIDContext, _ := c.Get("userID")
	playerID := userIDContext.(int)

	//hard delete
	query := `DELETE FROM games WHERE room_code = $1 AND player1_id = $2 AND status = 'WAITING'`
	result, err := DB.Exec(query, roomCode, playerID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel game"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	//ลบไม่สำเร็จ
	if rowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot cancel this room. It may have already started or you are not the host."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Room destroyed successfully"})
}

func RematchHandler(c *gin.Context) {
	roomCode := c.Param("id")
	userIDContext, _ := c.Get("userID")
	playerID := userIDContext.(int)

	//Start Transaction
	tx, err := DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not start transaction"})
		return
	}

	defer tx.Rollback()

	var gameID, p1ID int
	var p2ID *int
	var status string
	var nextRoomCode *string
	var rematchP1, rematchP2 bool

	//ล็อคแถวไว้ป้องกัน rematch พร้อมกัน
	queryLock := `SELECT id, player1_id, player2_id, status, next_room_code, rematch_p1, rematch_p2 
	              FROM games WHERE room_code = $1 FOR UPDATE`
	err = tx.QueryRow(queryLock, roomCode).Scan(&gameID, &p1ID, &p2ID, &status, &nextRoomCode, &rematchP1, &rematchP2)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	//validate เกมจบมั้ย แล้วผู้เล่นห้องนี้กด rematch จริงมั้ย
	if status != "FINISHED" && status != "DRAW" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game is not finished yet"})
		return
	}
	isP1 := (playerID == p1ID)
	isP2 := (p2ID != nil && playerID == *p2ID)

	if !isP1 && !isP2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a player in this game"})
		return
	}

	// อัปเดตสถานะว่าคนนี้กด Rematch แล้ว
	if isP1 {
		rematchP1 = true
	} else if isP2 {
		rematchP2 = true
	}

	//บันทึกลง db
	updateRematch := `UPDATE games SET rematch_p1 = $1, rematch_p2 = $2 WHERE id = $3`
	_, err = tx.Exec(updateRematch, rematchP1, rematchP2, gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rematch status"})
		return
	}

	if rematchP1 && rematchP2 && nextRoomCode == nil {
		// ถ้าครบ 2 คนแล้ว ให้สร้างห้องใหม่เลย
		newRoomCode := GenerateRoomCode()

		// สลับฝั่ง P1 กับ P2
		newP1 := p1ID
		newP2 := p2ID
		if p2ID != nil {
			newP1 = *p2ID
			newP2 = &p1ID
		}

		var newGameID int
		insertQuery := `
			INSERT INTO games (room_code, player1_id, player2_id, current_turn_id, status, board) 
			VALUES ($1, $2, $3, $2, 'IN_PROGRESS', '---------') 
			RETURNING id`
		err = tx.QueryRow(insertQuery, newRoomCode, newP1, newP2).Scan(&newGameID)

		if err == nil {
			// อัปเดตห้องเก่า ให้ชี้เป้าไปห้องใหม่
			tx.Exec(`UPDATE games SET next_room_code = $1 WHERE id = $2`, newRoomCode, gameID)

			// Commit เลย
			tx.Commit()
			c.JSON(http.StatusOK, gin.H{
				"message":   "Both players agreed. Match started!",
				"status":    "2/2",
				"room_code": newRoomCode,
			})
			return
		}
	}
	//ถ้ามีแค่ 1 คน
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"message": "Waiting for opponent...",
		"status":  "1/2",
	})

}

func LeaveGameHandler(c *gin.Context) {
	roomCode := c.Param("id")
	userIDContext, _ := c.Get("userID")
	playerID := userIDContext.(int)

	var gameID, p1ID int
	var p2ID *int
	var status string

	err := DB.QueryRow(`SELECT id, player1_id, player2_id, status FROM games WHERE room_code = $1`, roomCode).Scan(&gameID, &p1ID, &p2ID, &status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	if status == "IN_PROGRESS" {
		// ถ้ากดออกกลางเกม = ยอมแพ้ (Surrender) ให้อีกฝั่งชนะทันที
		var winnerID int
		if playerID == p1ID && p2ID != nil {
			winnerID = *p2ID
		} else {
			winnerID = p1ID
		}
		DB.Exec(`UPDATE games SET status = 'ABANDONED', winner_id = $1 WHERE id = $2`, winnerID, gameID)

	} else if status == "FINISHED" || status == "DRAW" {
		// ถ้ากดออกตอนเกมจบแล้ว (ทิ้งหน้าจอ Rematch) -> เปลี่ยนเป็น ABANDONED
		DB.Exec(`UPDATE games SET status = 'ABANDONED' WHERE id = $1`, gameID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Left the arena"})
}


func GetMyActiveGameHandler(c *gin.Context) {
	userIDContext, _ := c.Get("userID")
	playerID := userIDContext.(int)

	var roomCode string
	query := `SELECT room_code FROM games WHERE (player1_id = $1 OR player2_id = $1) AND status IN ('WAITING', 'IN_PROGRESS') LIMIT 1`
	err := DB.QueryRow(query, playerID).Scan(&roomCode)

	if err != nil { // ถ้าหาไม่เจอ (ไม่มีห้องค้าง)
		c.JSON(http.StatusOK, gin.H{"has_active_game": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"has_active_game": true, "room_code": roomCode})
}