package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/hiroyuki-takayama-RAIX/core"
)

// cannot use := because account is global
var accounts = map[float64]*core.Account{}

//

func main() {
	accounts[1001] = &core.Account{
		Customer: core.Customer{
			Name:    "John",
			Address: "Los Angeles, California",
			Phone:   "(213) 555 0147",
		},
		Number: 1001,
	}

	// set statment as handler function in '/statement'
	http.HandleFunc("/statement", statement)
	// log.fatal show you log with date_time
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

/*
api structure is as follows.
1. lines refering to query and its converted
2. lines refering to method you wanna use
*/

func statement(w http.ResponseWriter, req *http.Request) {
	// parse request
	numberqs := req.URL.Query().Get("number")

	if numberqs == "" {
		// http.ResponseWriter is io.Writer interface and Fprintf() write data into io.Writer interface.
		fmt.Fprintf(w, "Account number is missing!")
		return
	}

	// parse request
	if number, err := strconv.ParseFloat(numberqs, 64); err != nil {
		fmt.Fprintf(w, "Invalid account number!")
	} else {
		account, ok := accounts[number]
		if !ok {
			fmt.Fprintf(w, "Account with number %v can't be found!", number)
		} else {
			fmt.Fprintf(w, account.Statement())
		}
	}
}

func deposit(w http.ResponseWriter, req *http.Request) {
	// make reqests each query.
	numberqs := req.URL.Query().Get("number")
	amountqs := req.URL.Query().Get("amount")

	if numberqs == "" {
		fmt.Fprintf(w, "Account number is missing!")
		return
	}

	// below lines are error handlings of numberqs
	if number, err := strconv.ParseFloat(numberqs, 64); err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v is invalid account number!", number))
	} else if amount, err := strconv.ParseFloat(amountqs, 64); err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v is Invalid amount number!", amount))
	} else {
		account, ok := accounts[number]
		if !ok {
			fmt.Fprintf(w, "Account with number %v can't be found!", number)

			// below lines are error handling of amountqs
		} else {
			err := account.Deposit(amount)
			if err != nil {
				fmt.Fprintf(w, "%v", err)
			} else {
				fmt.Fprintf(w, account.Statement())
			}
		}
	}
}

func withdraw(w http.ResponseWriter, req *http.Request) {
	// make reqests each query.
	numberqs := req.URL.Query().Get("number")
	amountqs := req.URL.Query().Get("amount")

	if numberqs == "" {
		fmt.Fprintf(w, "Account number is missing!")
		return
	}

	// below lines are error handlings of numberqs
	if number, err := strconv.ParseFloat(numberqs, 64); err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v is invalid account number!", number))
	} else if amount, err := strconv.ParseFloat(amountqs, 64); err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v is Invalid amount number!", amount))
	} else {
		account, ok := accounts[number]
		if !ok {
			fmt.Fprintf(w, "Account with number %v can't be found!", number)

			// below lines are error handling of amountqs
		} else {
			err := account.Withdraw(amount)
			if err != nil {
				fmt.Fprintf(w, "%v", err)
			} else {
				fmt.Fprintf(w, account.Statement())
			}
		}
	}
}

func transfer(w http.ResponseWriter, req *http.Request) {
	// make reqests each query.
	senderNumberqs := req.URL.Query().Get("from")
	recieverNumberqs := req.URL.Query().Get("to")
	amountqs := req.URL.Query().Get("amount")

	if senderNumberqs == "" {
		fmt.Fprintf(w, "Account number is missing!")
		return
	}

	if recieverNumberqs == "" {
		fmt.Fprintf(w, "Account number is missing!")
		return
	}

	// below lines are error handlings of numberqs
	if senderNumber, err := strconv.ParseFloat(senderNumberqs, 64); err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v is invalid account number!", senderNumber))
	} else if recieverNumber, err := strconv.ParseFloat(recieverNumberqs, 64); err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v is Invalid account number!", recieverNumber))
	} else if amount, err := strconv.ParseFloat(amountqs, 64); err != nil {
		fmt.Fprintf(w, fmt.Sprintf("%v is Invalid amount number!", amount))
	} else {
		sender, senderOk := accounts[senderNumber]
		reciever, recieverOk := accounts[recieverNumber]
		if !senderOk {
			fmt.Fprintf(w, "Account of sender with number %v can't be found!", senderNumber)
		} else if !recieverOk {
			fmt.Fprintf(w, "Account of reciever with number %v can't be found!", recieverNumber)
			// below lines are error handling of amountqs
		} else {
			err := sender.Transfer(reciever, amount)
			if err != nil {
				fmt.Fprintf(w, "%v", err)
			} else {
				fmt.Fprintf(w, "sender : %v\nreviever : %v", sender.Statement(), reciever.Statement())
			}
		}
	}
}

func badRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Bad Request", http.StatusBadRequest)
}
