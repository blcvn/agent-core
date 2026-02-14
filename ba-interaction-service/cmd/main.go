package main

import (
	"log"

	"github.com/blcvn/backend/services/ba-interaction-service/internal/ws"
	"github.com/gin-gonic/gin"
)

func main() {
	hub := ws.NewHub()
	go hub.Run()

	r := gin.Default()
	r.GET("/ws", func(c *gin.Context) {
		ws.ServeWs(hub, c.Writer, c.Request)
	})

	log.Println("Interaction Service running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
