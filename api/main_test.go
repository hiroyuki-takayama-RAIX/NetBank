package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hiroyuki-takayama-RAIX/core"
)

type statementFixture struct {
	testName         string
	number           float64
	expectedResponse string
}

func TestStatement(t *testing.T) {

	// set fs to check test of two case in same programm.
	/*
		fs := []statementFixture{}
		fs = append(fs, statementFixture{"successfully getting statement", 1001, fmt.Sprintf("%v - %s - %v", 1001, "John", 0)})
		fs = append(fs, statementFixture{"account with number cant be found", 404, fmt.Sprintf("Account with number %v can't be found!", 404)})
	*/

	// upper lines are correct, but make()'s second argument is useful to make clear number of test pattern,
	fs := make([]statementFixture, 2)
	fs[0] = statementFixture{"successfully getting statement", 1001, fmt.Sprintf("%v - %s - %v", 1001, "John", 100)}
	fs[1] = statementFixture{"account with number cant be found", 404, fmt.Sprintf("Account with number %v can't be found!", 404)}

	for i := 0; i < len(fs); i++ {

		f := fs[i]
		name := fmt.Sprintf("test : %v", f.testName)

		// t.Run() works as a sub-test function
		t.Run(name, func(t *testing.T) {

			// make an account and insert into account (global variable in main.go)
			accounts[1001] = &core.Account{
				Customer: core.Customer{
					Name:    "John",
					Address: "Los Angeles, California",
					Phone:   "(213) 555 0147",
				},
				Number:  1001,
				Balance: 100,
			}

			// make a request
			request := fmt.Sprintf("/statement?number=%v", f.number)
			req, err := http.NewRequest("GET", request, nil)
			if err != nil {
				t.Fatal(err)
			}

			// rr records response from handler
			rr := httptest.NewRecorder()

			// cast statment() to http.HandlerFunc, which makes a Handler object whithout stating a struct.
			handler := http.HandlerFunc(statement)

			// handler gets req as HTTP request and sets response on rr
			handler.ServeHTTP(rr, req)

			// check status code
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Status code mismatch: expected %v, got %v", http.StatusOK, status)
			}

			// compare my expection and the response
			if rr.Body.String() != f.expectedResponse {
				t.Errorf("Response body mismatch: expected '%v', got '%v'", f.expectedResponse, rr.Body.String())
			}
		})
	}
}

type depositFixture struct {
	name             string
	number           float64
	amount           float64
	expectedResponse string
}

func TestDeposite(t *testing.T) {

	fs := make([]depositFixture, 3)
	fs[0] = depositFixture{"successfully doing deposit", 1001, 20, fmt.Sprintf("%v - %s - %v", 1001, "John", 20)}
	fs[1] = depositFixture{"account with number cant be found", 404, 20, fmt.Sprintf("Account with number %v can't be found!", 404)}
	fs[2] = depositFixture{"deposite is less than zero", 1001, -20, "An amount is less than zero!"}

	for i := 0; i < len(fs); i++ {

		f := fs[i]
		name := fmt.Sprintf("test %v", f.name)

		t.Run(name, func(t *testing.T) {
			accounts[1001] = &core.Account{
				Customer: core.Customer{
					Name:    "John",
					Address: "Los Angeles, California",
					Phone:   "(213) 555 0147",
				},
				Number: 1001,
			}

			// "&" conbines two or more queries
			request := fmt.Sprintf("/deposit?number=%v&amount=%f", f.number, f.amount)
			req, err := http.NewRequest("GET", request, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(deposit)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Status code mismatch: expected %v, got %v", http.StatusOK, status)
			}

			if rr.Body.String() != f.expectedResponse {
				t.Errorf("Response body mismatch: expected '%v', got '%v'", f.expectedResponse, rr.Body.String())
			}
		})
	}
}

type withdrawFixture struct {
	name             string
	number           float64
	amount           float64
	expectedResponse string
}

func TestWithdraw(t *testing.T) {

	fs := make([]withdrawFixture, 3)
	fs[0] = withdrawFixture{"successfully doing withdraw", 1001, 10, fmt.Sprintf("%v - %s - %v", 1001, "John", 10)}
	fs[1] = withdrawFixture{"account with number cant be found", 404, 20, fmt.Sprintf("Account with number %v can't be found!", 404)}
	fs[2] = withdrawFixture{"withdraw is less than zero", 1001, -20, "An amount is less than zero!"}

	for i := 0; i < len(fs); i++ {

		f := fs[i]
		name := fmt.Sprintf("test %v", f.name)

		t.Run(name, func(t *testing.T) {
			accounts[1001] = &core.Account{
				Customer: core.Customer{
					Name:    "John",
					Address: "Los Angeles, California",
					Phone:   "(213) 555 0147",
				},
				Number:  1001,
				Balance: 20,
			}

			request := fmt.Sprintf("/withdraw?number=%v&amount=%f", f.number, f.amount)
			req, err := http.NewRequest("GET", request, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(withdraw)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Status code mismatch: expected %v, got %v", http.StatusOK, status)
			}

			if rr.Body.String() != f.expectedResponse {
				t.Errorf("Response body mismatch: expected '%v', got '%v'", f.expectedResponse, rr.Body.String())
			}
		})
	}
}

type transferFixture struct {
	name             string
	senderNumber     float64
	recieverNumber   float64
	amount           float64
	expectedResponse string
}

func TestTransfar(t *testing.T) {

	fs := make([]transferFixture, 5)
	fs[0] = transferFixture{"successfully doing transfer", 1001, 2002, 10, fmt.Sprintf("sender : %v\nreviever : %v", "1001 - John - 90", "2002 - C.J. - 110")}
	fs[1] = transferFixture{"account with number cant be found", 404, 2002, 20, fmt.Sprintf("Account of sender with number %v can't be found!", 404)}
	fs[2] = transferFixture{"account with number cant be found", 1001, 404, 20, fmt.Sprintf("Account of reciever with number %v can't be found!", 404)}
	fs[3] = transferFixture{"transfer is less than zero", 1001, 2002, -20, "An amount is less than zero!"}
	fs[4] = transferFixture{"transfer is greater than the present deposit", 1001, 2002, 1000, "transfer is greater than deposit!"}

	for i := 0; i < len(fs); i++ {

		f := fs[i]
		name := fmt.Sprintf("test %v", f.name)

		t.Run(name, func(t *testing.T) {
			accounts[1001] = &core.Account{
				Customer: core.Customer{
					Name:    "John",
					Address: "Los Angeles, California",
					Phone:   "(213) 555 0147",
				},
				Number:  1001,
				Balance: 100,
			}
			accounts[2002] = &core.Account{
				Customer: core.Customer{
					Name:    "C.J.",
					Address: "Los Santos, San And Leas",
					Phone:   "(080) 1457 9387",
				},
				Number:  2002,
				Balance: 100,
			}

			request := fmt.Sprintf("/withdraw?from=%v&to=%v&amount=%f", f.senderNumber, f.recieverNumber, f.amount)
			req, err := http.NewRequest("GET", request, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(transfer)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Status code mismatch: expected %v, got %v", http.StatusOK, status)
			}

			if rr.Body.String() != f.expectedResponse {
				t.Errorf("Response body mismatch: expected '%v', got '%v'", f.expectedResponse, rr.Body.String())
			}
		})
	}
}

func TestErrorCodes(t *testing.T) {

	t.Run("404 Not Found", func(t *testing.T) {
		queryParams := make(url.Values)
		queryParams.Set("from", "sender")
		queryParams.Set("to", "receiver")
		queryParams.Set("amount", "invalid")

		// make api having invaild characters
		//request := "/" + queryParams.Encode()
		req, err := http.NewRequest("POST", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(badRequest)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Status code mismatch: expected %v, got %v", http.StatusBadRequest, status)
		}
	})

}
