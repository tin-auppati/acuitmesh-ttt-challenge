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
