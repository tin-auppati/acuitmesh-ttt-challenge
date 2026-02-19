package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	ConnectDB()

	defer DB.Close()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status":  "server is running",
			"db":      "connected",
		})
	})

	api := r.Group("/api")
	{
		api.GET("/games", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"games": []string{"Room 1", "Room 2"},
			})
			fmt.Println("GET /api/games")
		})
		api.POST("/games", CreateGameHandler)
		api.POST("/games/join", JoinGameHandler)
		api.POST("/games/move", MakeMoveHandler)

		api.GET("/games/:id", GetGameHandler)       // ดูสถานะเกม
		api.GET("/games/:id/moves", GetGameMovesHandler) // ดูประวัติ
	}

	r.Run(":8080") // listen and serve on
}
