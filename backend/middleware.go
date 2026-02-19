// backend/middleware.go
package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware - ตรวจสอบ JWT Token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		//ตรวจสอบรูปแบบ ต้องเป็น "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		//ดึง Key จาก Environment
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "super_secret_tictactoe_key_2026"
		}

		//แกะ Token และตรวจสอบความถูกต้อง
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// userID ที่แกะได้ไปฝากไว้ใน Context
		c.Set("userID", claims.UserID)
		c.Next() // อนุญาตให้ผ่านเข้าสู่ API เกมได้
	}
}
