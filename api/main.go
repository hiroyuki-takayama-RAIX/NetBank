package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	// "_" in import statement means blank import
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/hiroyuki-takayama-RAIX/core"
)

// cannot use := because account is global
var accounts = map[float64]*core.Account{}

func main() {
	accounts[1001] = &core.Account{
		Customer: core.Customer{
			Name:    "John",
			Address: "Los Angeles, California",
			Phone:   "(213) 555 0147",
		},
		Number: 1001,
	}

	// set statment as handler function in path
	http.HandleFunc("/statement", statement)
	http.HandleFunc("/withdraw", withdraw)
	http.HandleFunc("/deposit", deposit)
	http.HandleFunc("/transfer", transfer)
	http.HandleFunc("/teapot", teapot)

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
		http.Error(w, "Account number is missing!", http.StatusBadRequest)
		return
	}

	// parse request
	if number, err := strconv.ParseFloat(numberqs, 64); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid account number!", numberqs), http.StatusBadRequest)
	} else {
		account, ok := accounts[number]
		if !ok {
			http.Error(w, fmt.Sprintf("Account with number %v can't be found!", number), http.StatusNotFound)
		} else {
			fmt.Fprintf(w, "%v", account.Statement())
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

	if number, err := strconv.ParseFloat(numberqs, 64); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid account number!", numberqs), http.StatusBadRequest)
	} else if amount, err := strconv.ParseFloat(amountqs, 64); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid amount number!", amountqs), http.StatusBadRequest)
	} else {
		account, ok := accounts[number]
		if !ok {
			http.Error(w, fmt.Sprintf("Account with number %v can't be found!", number), http.StatusNotFound)

		} else {
			//when amount is less than zero, error is not nil.
			err := account.Deposit(amount)
			if err != nil {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			} else {
				fmt.Fprintf(w, "%v", account.Statement())
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
		http.Error(w, fmt.Sprintf("%v is invalid account number!", numberqs), http.StatusBadRequest)
	} else if amount, err := strconv.ParseFloat(amountqs, 64); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid amount number!", amountqs), http.StatusBadRequest)
	} else {
		account, ok := accounts[number]
		if !ok {
			http.Error(w, fmt.Sprintf("Account with number %v can't be found!", number), http.StatusNotFound)

			// below lines are error handling of amountqs
		} else {
			err := account.Withdraw(amount)
			if err != nil {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			} else {
				fmt.Fprintf(w, "%v", account.Statement())
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
		http.Error(w, fmt.Sprintf("%v is invalid account number!", senderNumberqs), http.StatusBadRequest)
	} else if recieverNumber, err := strconv.ParseFloat(recieverNumberqs, 64); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid account number!", recieverNumberqs), http.StatusBadRequest)
	} else if amount, err := strconv.ParseFloat(amountqs, 64); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid amount number!", amountqs), http.StatusBadRequest)
	} else {
		sender, senderOk := accounts[senderNumber]
		reciever, recieverOk := accounts[recieverNumber]
		if !senderOk {
			http.Error(w, fmt.Sprintf("Account of sender with number %v can't be found!", senderNumber), http.StatusNotFound)
		} else if !recieverOk {
			http.Error(w, fmt.Sprintf("Account of reciever with number %v can't be found!", recieverNumber), http.StatusNotFound)
			// below lines are error handling of amountqs
		} else {
			err := sender.Transfer(reciever, amount)
			if err != nil {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			} else {
				fmt.Fprintf(w, "sender : %v\nreviever : %v", sender.Statement(), reciever.Statement())
			}
		}
	}
}

func teapot(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "418 : I'm a teapot.", http.StatusTeapot)
}
