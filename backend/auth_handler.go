package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("super_secret_tictactoe_key_2026")

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// jwt token generation
func GenerateToken(userID int) (string, error) {
	// กำหนดวันหมดอายุของ Token
	expirationTime := time.Now().Add(24 * time.Hour)

	// สร้าง Payload
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// สร้าง Token พร้อมระบุ Algorithm ที่ใช้เข้ารหัส (HS256)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// registerHandler - สมัครสมาชิกใหม่
func RegisterHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,alphanum,min=3,max=20" `
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// บันทึก User ลง Database
	var userID int
	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`
	err = DB.QueryRow(query, req.Username, hashedPassword).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Username already exists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully", "user_id": userID})
}

// loginHandler - เข้าสู่ระบบ
func LoginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,alphanum,min=3,max=20" `
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID int
	var storedHash string
	query := `SELECT id, password_hash FROM users WHERE username = $1`
	err := DB.QueryRow(query, req.Username).Scan(&userID, &storedHash)

	if err != nil || !CheckPasswordHash(req.Password, storedHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	tokenString, err := GenerateToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   tokenString,
		"user_id": userID,
	})
}
