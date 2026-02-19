package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateGameHandler(c *gin.Context) {
	var req struct {
		PlayerID int `json:"player_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var gameID int
	query := `
		INSERT INTO games (player1_id, current_turn_id, status, board) 
		VALUES ($1, $1, 'WAITING', '---------') 
		RETURNING id`

	err := DB.QueryRow(query, req.PlayerID).Scan(&gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create game", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Game created successfully",
		"game_id": gameID,
		"status":  "WAITING",
	})
}


func JoinGameHandler(c *gin.Context) {
	var req struct {
		GameID   int `json:"game_id" binding:"required"`
		PlayerID int `json:"player_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//atomic ทั้งก้อน begin - commit
	tx, err := DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Could not start transaction"})
		return
	}

	// ถ้าเกิดอะไรขึ้นผิดพลาดให้ Rollback เสมอ
	defer tx.Rollback()

	// 1. SELECT ... FOR UPDATE เพื่อ Lock แถวเกมนั้นไว้ก่อน
	var p1ID int
	var p2ID *int
	var status string

	queryLock := `SELECT player1_id, player2_id, status FROM games WHERE id = $1 FOR UPDATE`
	err = tx.QueryRow(queryLock, req.GameID).Scan(&p1ID, &p2ID, &status)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}
	
	// 2. validation ต่างๆ เช่น เช็คว่าเกมเต็มหรือยัง เช็คว่าเกมอยู่ในสถานะ WAITING หรือเปล่า เช็คว่า player ที่จะเข้ามาไม่ได้เป็น player1 อยู่แล้ว
	if p1ID == req.PlayerID {
		c.JSON(http.StatusConflict, gin.H{"error": "You are already Player 1"})
		return
	}
	if p2ID != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Game is already full"})
		return
	}
	if status != "WAITING" {
		c.JSON(http.StatusConflict, gin.H{"error": "Game is not in waiting status Can't join"})
		return
	}

	// 3. UPDATE เพื่อ set player2_id และเปลี่ยนสถานะเกมเป็น IN_PROGRESS 
	queryUpdate := `UPDATE games SET player2_id = $1, status = 'IN_PROGRESS' WHERE id = $2` //atomic
	_, err = tx.Exec(queryUpdate, req.PlayerID, req.GameID)
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
		"message": "Joined game successfully",
		"status":  "IN_PROGRESS",
	})
}


func MakeMoveHandler(c *gin.Context){
	var req struct {
		GameID   int `json:"game_id" binding:"required"`
		PlayerID int `json:"player_id" binding:"required"`
		X		int `json:"x"`
		Y		int `json:"y"`

	}

	if err := c.ShouldBindJSON(&req) ; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//atomic ทั้งก้อน begin - commit
	tx, _ := DB.Begin()
	defer tx.Rollback()

	// 1. lock
	var board, statue string
	var p1ID, p2ID, turnID int
	query := `SELECT board, status, player1_id, player2_id, current_turn_id FROM games WHERE id = $1 FOR UPDATE`
	err := tx.QueryRow(query, req.GameID).Scan(&board, &statue, &p1ID, &p2ID, &turnID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	// 2. validation
	if statue != "IN_PROGRESS" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game is not in progress"})
		return
	}
	if turnID != req.PlayerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your turn"})
		return
	}

	index := req.X*3 + req.Y
	if board[index] != '-' {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cell already occupied"})
		return
	}


	//3. update board
	char := "X"
	nextTurn := p2ID
	if req.PlayerID == p2ID {
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
			winnerID = &req.PlayerID
		}
	}

	updateQuery := `UPDATE games SET board = $1, current_turn_id = $2, status = $3, winner_id = $4 WHERE id = $5`
	_, err = tx.Exec(updateQuery, newBoard, nextTurn, newStatus, winnerID, req.GameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update game state"})
		return
	}

	_, err = tx.Exec(`INSERT INTO moves (game_id, player_id, x, y, move_order) 
             VALUES ($1, $2, $3, $4, (SELECT count(*)+1 FROM moves WHERE game_id=$1))`,
		req.GameID, req.PlayerID, req.X, req.Y)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record move"})
		return
	}
	
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"board":   newBoard,
		"status":  newStatus,
	})
}