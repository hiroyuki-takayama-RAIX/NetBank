package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hiroyuki-takayama-RAIX/api"
)

func main() {
	env := os.Getenv("Env")
	if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.GET("/accounts", api.GetAccounts)
	router.GET("/accounts/:id", api.GetAccount)
	router.POST("/accounts", api.CreateAccount)
	router.DELETE("/accounts/:id", api.DeleteAccount)
	router.PUT("/accounts/:id", api.UpdateAccount)
	router.GET("/accounts/:id/balance", api.GetBalance)
	router.PATCH("/accounts/:id/balance", api.FinancialTransaction)

	if env == "prod" {
		router.Run("0.0.0.0:80")
	} else {
		router.Run("localhost:8080")
	}
}
