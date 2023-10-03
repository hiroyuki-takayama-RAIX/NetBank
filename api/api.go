package api

import (
	// "_" in import means blank import

	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hiroyuki-takayama-RAIX/core"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func GetAccounts(c *gin.Context) {
	nb, err := core.NewNetBank()
	if err != nil {
		// Handle initialization error and send an error response
		msg := fmt.Sprintf("failed to initialize netbank instance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	defer nb.Close()

	minBalanceStr := c.DefaultQuery("min-balance", "0")
	maxBalanceStr := c.DefaultQuery("max-balance", "2147483647")

	// Convert query parameters to float64
	minBalance, err := strconv.ParseFloat(minBalanceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'minBalance' parameter"})
		return
	}

	maxBalance, err := strconv.ParseFloat(maxBalanceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'maxBalance' parameter"})
		return
	}

	accounts, err := nb.GetAccounts(minBalance, maxBalance)
	if err != nil {
		// Handle the error returned by nb.GetAccounts() and send an error response
		// the frist arguement is actual status code, the second one is expected status code
		msg := fmt.Sprintf("failed to Get accounts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
	} else {
		// Send a successful response with the accounts data
		c.IndentedJSON(http.StatusOK, accounts)
	}
}

func GetAccount(c *gin.Context) {
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

func CreateAccount(c *gin.Context) {
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
		account, err := nb.CreateAccount(&customer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create a new account"})
		} else {
			c.IndentedJSON(http.StatusCreated, account)
		}
	}
}

func DeleteAccount(c *gin.Context) {
	nb, err := core.NewNetBank()
	if err != nil {
		msg := fmt.Sprintf("failed to initialize netbank instance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	defer nb.Close()

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

func UpdateAccount(c *gin.Context) {
	nb, err := core.NewNetBank()
	if err != nil {
		msg := fmt.Sprintf("failed to initialize netbank instance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	defer nb.Close()

	param := c.Param("id")
	id, err := strconv.Atoi(param)
	if err != nil {
		msg := fmt.Sprintf("got %v as invalied id", param)
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	} else {
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

const (
	TEST     = "test"
	DEPOSIT  = "deposit"
	WITHDRAW = "withdraw"
	TRANSFER = "transfer"
)

func FinancialTransaction(c *gin.Context) {
	nb, err := core.NewNetBank()
	if err != nil {
		msg := fmt.Sprintf("failed to initialize netbank instance: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	defer nb.Close()

	param := c.Param("id")
	id, err := strconv.Atoi(param)
	if err != nil {
		msg := fmt.Sprintf("got %v as invalied id", param)
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	} else {
		_, err := nb.GetAccount(id)
		if err != nil {
			msg := fmt.Sprintf("account(ID: %v) doesnt exist", id)
			c.JSON(http.StatusNotFound, gin.H{"error": msg})
		} else {
			var t core.Trade
			err = c.BindJSON(&t)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalied request"})
			} else if t.Amount <= 0 {
				msg := fmt.Sprintf("amount is less than zero. your input is %v", t.Amount)
				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			} else {
				/*
					accounts, err := nb.Execute(ft)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					} else {
						c.JSON(http.StatusOK, account)
					}
				*/
				switch t.Class {
				case DEPOSIT:
					account, err := nb.Deposit(id, t.Amount)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					} else {
						c.JSON(http.StatusOK, account)
					}
				case WITHDRAW:
					account, err := nb.Withdraw(id, t.Amount)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					} else {
						c.JSON(http.StatusOK, account)
					}
				case TRANSFER:
					accounts, err := nb.Transfer(id, t.To, t.Amount)
					if err != nil {
						if errors.Is(err, sql.ErrNoRows) {
							c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
						} else {
							c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
						}
					} else {
						c.JSON(http.StatusOK, accounts)
					}
				case TEST:
					c.JSON(http.StatusOK, gin.H{"msg": "FinancialTransaction() is executed collectlly."})
				default:
					msg := fmt.Sprintf("you about to do %v, but its not defined.", t.Class)
					c.JSON(http.StatusBadRequest, gin.H{"error": msg})
				}
			}
		}
	}
}

func GetBalance(c *gin.Context) {
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
		// coreパッケージにGetBalance()を作成するのではなく、GetAccount()を流用して必要な情報を抽出する。
		account, err := nb.GetAccount(id)
		if err != nil {
			msg := fmt.Sprintf("account(ID: %v) doesnt exist", id)
			c.JSON(http.StatusNotFound, gin.H{"error": msg})
		} else {
			r := make(map[string]any)
			r["balance"] = account.Balance
			r["id"] = account.Number
			c.IndentedJSON(http.StatusOK, r)
		}
	}
}
