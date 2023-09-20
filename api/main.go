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
	nb, err := core.NewNetBank()
	if err != nil {
		// Handle initialization error and send an error response
		msg := fmt.Sprintf("failed to initialize netbank instance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	defer nb.Close()

	accounts, err := nb.GetAccounts()
	if err != nil {
		// Handle the error returned by nb.GetAccounts() and send an error response
		// the frist arguement is actual status code, the second one is expected status code
		msg := fmt.Sprintf("failed to get accounts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
	} else {
		// Send a successful response with the accounts data
		c.IndentedJSON(http.StatusOK, accounts)
	}
}

func getAccount(c *gin.Context) {
	nb, err := core.NewNetBank()
	if err != nil {
		msg := fmt.Sprintf("failed to initialize netbank instance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	defer nb.Close()

	// parse and validate a path parameter
	param := c.Param("id")
	id, err := strconv.Atoi(param)
	if err != nil {
		msg := fmt.Sprintf("got %v as invalied id", param)
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	} else {
		account, err := nb.GetAccount(id)
		if err != nil {
			msg := fmt.Sprintf("account(ID: %v) doesnt exist", id)
			c.JSON(http.StatusNotFound, gin.H{"error": msg})
		} else {
			c.IndentedJSON(http.StatusOK, account)
		}
	}
}

func createAccount(c *gin.Context) {
	nb, err := core.NewNetBank()
	if err != nil {
		msg := fmt.Sprintf("failed to initialize netbank instance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	defer nb.Close()

	// mapping request body into empty customer variable
	var customer core.Customer
	err = c.BindJSON(&customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalied request"})
	} else if customer.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request has empty name"})
	} else if customer.Address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request has empty address"})
	} else if customer.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request has empty phone number"})
	} else {
		id, err := nb.CreateAccount(&customer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create a new account"})
		} else {
			account, err := nb.GetAccount(id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get the new account information"})
			} else {
				c.IndentedJSON(http.StatusCreated, account)
			}
		}
	}
}

func deleteAccount(c *gin.Context) {
	nb, err := core.NewNetBank()
	if err != nil {
		msg := fmt.Sprintf("failed to initialize netbank instance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	defer nb.Close()

	// parse and validate a path parameter
	param := c.Param("id")
	id, err := strconv.Atoi(param)
	if err != nil {
		msg := fmt.Sprintf("got %v as invalied id", param)
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	} else {
		err := nb.DeleteAccount(id)
		if err != nil {
			msg := fmt.Sprintf("account(ID: %v) doesnt exist", id)
			c.JSON(http.StatusNotFound, gin.H{"error": msg})
		} else {
			c.Status(http.StatusNoContent)
		}
	}
}

func updateAccount(c *gin.Context) {
	nb, err := core.NewNetBank()
	if err != nil {
		msg := fmt.Sprintf("failed to initialize netbank instance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	defer nb.Close()

	// parse and validate a path parameter
	param := c.Param("id")
	id, err := strconv.Atoi(param)
	if err != nil {
		msg := fmt.Sprintf("got %v as invalied id", param)
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	} else {
		var customer core.Customer
		err = c.BindJSON(&customer)
		fmt.Println(customer)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalied request"})
		} else if customer.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "request has empty name"})
		} else if customer.Address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "request has empty address"})
		} else if customer.Phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "request has empty phone number"})
		} else {
			account, err := nb.UpdateAccount(id, &customer)
			if err != nil {
				msg := fmt.Sprintf("account(ID: %v) doesnt exist", id)
				c.JSON(http.StatusNotFound, gin.H{"error": msg})
			} else {
				c.IndentedJSON(http.StatusCreated, account)
			}
		}
	}
}
