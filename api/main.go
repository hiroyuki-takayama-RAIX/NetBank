package main

import (
	// "_" in import statement means blank import

	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hiroyuki-takayama-RAIX/core"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	router := gin.Default()
	router.GET("/accounts", getAccounts)
	router.GET("/accounts/:id", getAccount)
	router.POST("/accounts", createAccount)
	router.DELETE("/accounts/:id", deleteAccount)
	router.PUT("/accountts/:id", updateAccount)

	router.Run("localhost:8080")
}

func getAccounts(c *gin.Context) {
	nb, _ := core.NewNetBank()
	/*
		if err != nil {
			http.Error(w, "initialize error!", http.StatusBadRequest)
		}
	*/
	defer nb.Close()

	accounts, _ := nb.GetAccounts()
	c.IndentedJSON(http.StatusOK, accounts)
}

func getAccount(c *gin.Context) {
	nb, _ := core.NewNetBank()
	/*
		if err != nil {
			http.Error(w, "initialize error!", http.StatusBadRequest)
		}
	*/
	defer nb.Close()

	param := c.Param("id")
	id, _ := strconv.Atoi(param)
	/*
		if err != nil {
			errors.Errorf("got %v as invailed id: %v", param, err)
		}
	*/

	account, _ := nb.GetAccount(id)
	c.IndentedJSON(http.StatusOK, account)
}

func createAccount(c *gin.Context) {
	nb, _ := core.NewNetBank()
	defer nb.Close()

	// postaccounts adds an account from JSON received in the request body.
	var customer core.Customer

	_ = c.BindJSON(&customer)

	id, _ := nb.CreateAccount(&customer)

	account, _ := nb.GetAccount(id)

	c.IndentedJSON(http.StatusCreated, account)
}

func deleteAccount(c *gin.Context) {
	nb, _ := core.NewNetBank()
	defer nb.Close()

	param := c.Param("id")
	id, _ := strconv.Atoi(param)

	_ = nb.DeleteAccount(id)

	response := fmt.Sprintf("successfully delete account(ID: %v)", id)
	c.IndentedJSON(http.StatusOK, response)
}

func updateAccount(c *gin.Context) {
	nb, _ := core.NewNetBank()
	defer nb.Close()

	param := c.Param("id")
	id, _ := strconv.Atoi(param)

	var customer core.Customer
	_ = c.BindJSON(&customer)

	fmt.Println(id)

	_ = nb.UpdateAccount(id, &customer)

	account, _ := nb.GetAccount(id)
	customer = account.Customer

	c.IndentedJSON(http.StatusOK, customer)
}
