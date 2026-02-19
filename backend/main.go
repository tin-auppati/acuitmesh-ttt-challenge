package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	ConnectDB()

	defer DB.Close()

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status":  "server is running",
			"db":      "connected",
		})
	})

	api := r.Group("/api")
	{
		// --- ระบบ Auth ---
		api.POST("/register", RegisterHandler)
		api.POST("/login", LoginHandler)

		// --- ระบบเกม ---
		protected := api.Group("/games")
		protected.Use(AuthMiddleware())
		{
			protected.POST("", CreateGameHandler)
			protected.POST("/join", JoinGameHandler)
			protected.POST("/move", MakeMoveHandler)

			//ถ้าทัน spectator จะกลับมาแก้
			protected.GET("/:id", GetGameHandler)            // ดูสถานะเกม
			protected.GET("/:id/moves", GetGameMovesHandler) // ดูประวัติ
		}
	}

	r.Run(":8080") // listen and serve on
}
