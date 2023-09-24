package main

import (
	"github.com/gin-gonic/gin"
	"github.com/hiroyuki-takayama-RAIX/api"
)

func main() {
	router := gin.Default()
	router.GET("/accounts", api.GetAccounts)
	router.GET("/accounts/:id", api.GetAccount)
	router.POST("/accounts", api.CreateAccount)
	router.DELETE("/accounts/:id", api.DeleteAccount)
	router.PUT("/accounts/:id", api.UpdateAccount)
	router.GET("/accounts/:id/balance", api.GetBalance)

	router.Run("localhost:8080")
}
