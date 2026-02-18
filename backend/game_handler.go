package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateGameHandler(c *gin.Context) {
	var req struct {
		PlayerID int `json:"player1_id" binding:"required"`
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
