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

func main() {
	// set statment as handler function in path
	http.HandleFunc("/statement", statement)
	http.HandleFunc("/withdraw", withdraw)
	http.HandleFunc("/deposit", deposit)
	http.HandleFunc("/transfer", transfer)
	http.HandleFunc("/teapot", teapot)
	http.HandleFunc("/createaccount", createAccount)
	http.HandleFunc("/deleteaccount", deleteAccount)

	fmt.Println("Server is running")

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
	if number, err := strconv.Atoi(numberqs); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid account number!", numberqs), http.StatusBadRequest)
	} else {
		s, err := core.NewNetBank().Statement(number)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
		} else {
			fmt.Fprintf(w, "%v", s)
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

	if number, err := strconv.Atoi(numberqs); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid account number!", numberqs), http.StatusBadRequest)
	} else if amount, err := strconv.ParseFloat(amountqs, 64); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid amount number!", amountqs), http.StatusBadRequest)
	} else {
		err := core.NewNetBank().Deposit(number, amount)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		} else {
			//when amount is less than zero, error is not nil.
			s, err := core.NewNetBank().Statement(number)
			if err != nil {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
			} else {
				fmt.Fprintf(w, "%v", s)
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
	if number, err := strconv.Atoi(numberqs); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid account number!", numberqs), http.StatusBadRequest)
	} else if amount, err := strconv.ParseFloat(amountqs, 64); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid amount number!", amountqs), http.StatusBadRequest)
	} else {
		err := core.NewNetBank().Withdraw(number, amount)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)

			// below lines are error handling of amountqs
			s, err := core.NewNetBank().Statement(number)
			if err != nil {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			} else {
				fmt.Fprintf(w, "%v", s)
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
	if senderNumber, err := strconv.Atoi(senderNumberqs); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid account number!", senderNumberqs), http.StatusBadRequest)
	} else if recieverNumber, err := strconv.Atoi(recieverNumberqs); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid account number!", recieverNumberqs), http.StatusBadRequest)
	} else if amount, err := strconv.ParseFloat(amountqs, 64); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid amount number!", amountqs), http.StatusBadRequest)
	} else {
		err := core.NewNetBank().Transfer(senderNumber, recieverNumber, amount)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
			// below lines are error handling of amountqs
		} else {
			sender, err := core.NewNetBank().Statement(senderNumber)
			if err != nil {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			} else {
				reciever, err := core.NewNetBank().Statement(recieverNumber)
				if err != nil {
					http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
				} else {
					fmt.Fprintf(w, "sender : %v\nreviever : %v", sender, reciever)
				}
			}
		}
	}
}

func teapot(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "418 : I'm a teapot.", http.StatusTeapot)
}

func createAccount(w http.ResponseWriter, req *http.Request) {
	// parse request
	numberqs := req.URL.Query().Get("number")
	name := req.URL.Query().Get("name")
	addr := req.URL.Query().Get("addr")
	phone := req.URL.Query().Get("phone")

	if numberqs == "" {
		// http.ResponseWriter is io.Writer interface and Fprintf() write data into io.Writer interface.
		http.Error(w, "Account number is missing!", http.StatusBadRequest)
		return
	}

	// parse request
	if number, err := strconv.Atoi(numberqs); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid account number!", numberqs), http.StatusBadRequest)
	} else {
		err := core.NewNetBank().CreateAccount(number, name, addr, phone)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
		} else {
			s, err := core.NewNetBank().Statement(number)
			if err != nil {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			} else {
				fmt.Fprintf(w, "%v", s)
			}
		}
	}
}

func deleteAccount(w http.ResponseWriter, req *http.Request) {
	// parse request
	numberqs := req.URL.Query().Get("number")

	if numberqs == "" {
		// http.ResponseWriter is io.Writer interface and Fprintf() write data into io.Writer interface.
		http.Error(w, "Account number is missing!", http.StatusBadRequest)
		return
	}

	// parse request
	if number, err := strconv.Atoi(numberqs); err != nil {
		http.Error(w, fmt.Sprintf("%v is invalid account number!", numberqs), http.StatusBadRequest)
	} else {
		err := core.NewNetBank().DeleteAccount(number)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
		} else {
			s, err := core.NewNetBank().Statement(number)
			if err != nil {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			} else {
				fmt.Fprintf(w, "%v", s)
			}
		}
	}
}
