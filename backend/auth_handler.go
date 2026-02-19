package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// registerHandler - สมัครสมาชิกใหม่
func RegisterHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// บันทึก User ลง Database
	var userID int
	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`
	err := DB.QueryRow(query, req.Username, req.Password).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Username already exists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully", "user_id": userID})
}

// loginHandler - เข้าสู่ระบบ
func LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID int
	var passwordHash string
	query := `SELECT id, password_hash FROM users WHERE username = $1`
	err := DB.QueryRow(query, req.Username).Scan(&userID, &passwordHash)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user_id": userID,
	})
}
