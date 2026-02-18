package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status": "server is running",
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
	}

	r.Run(":8080") // listen and serve on
}
