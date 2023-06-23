package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hiroyuki-takayama-RAIX/core"
)

type fixture struct {
	name         string
	request      string
	expectedCode int
	expectedBody string
}

// i wanna change this name better...
// this function get fixture and show test result.
func ListenAndServe(f fixture, h func(w http.ResponseWriter, req *http.Request), t *testing.T) {
	req, err := http.NewRequest("GET", f.request, nil)
	if err != nil {
		t.Fatal(err)
	}

	// rr records response from handler
	rr := httptest.NewRecorder()

	// cast statment() to http.HandlerFunc, which makes a Handler object whithout stating a struct.
	handler := http.HandlerFunc(h)

	// handler gets req as HTTP request and sets response on rr
	handler.ServeHTTP(rr, req)

	// check status code
	if status := rr.Code; status != f.expectedCode {
		t.Errorf("Status code mismatch: expected %v, got %v", f.expectedCode, status)
	}

	// compare my expection and the response
	if rr.Body.String() != f.expectedBody {
		t.Errorf("Response body mismatch: expected '%v', got '%v'", f.expectedBody, rr.Body.String())
	}
}

func TestStatement(t *testing.T) {
	// set fs as fixture to test some cases in same programm.
	/*
		fs := []fixture{}
		fs = append(fs, fixture{"successfully getting statement", 1001, fmt.Sprintf("%v - %s - %v", 1001, "John", 0)})
		fs = append(fs, fixture{"account with number cant be found", 404, fmt.Sprintf("Account with number %v can't be found!", 404)})
	*/

	// upper lines are correct, but make()'s second argument is useful to make clear number of test pattern.
	fs := make([]fixture, 4)
	fs[0] = fixture{
		name:         "Successfully getting statement",
		request:      fmt.Sprintf("/statement?number=%v", 1001),
		expectedCode: http.StatusOK,
		expectedBody: fmt.Sprintf("%v - %s - %v", 1001, "John", 100),
	}
	fs[1] = fixture{
		name:         "Account with the number cant be found",
		request:      fmt.Sprintf("/statement?number=%v", 404),
		expectedCode: http.StatusNotFound,
		expectedBody: fmt.Sprintf("Account with number %v can't be found!\n", 404),
	}
	fs[2] = fixture{
		name:         "Account number is missing",
		request:      fmt.Sprintf("/statement?n=%v", 1001),
		expectedCode: http.StatusBadRequest,
		expectedBody: "Account number is missing!\n",
	}
	fs[3] = fixture{
		name:         "Invalid account number!",
		request:      fmt.Sprintf("/statement?number=%v", "千一"),
		expectedCode: http.StatusBadRequest,
		expectedBody: fmt.Sprintf("%v is invalid account number!\n", "千一"),
	}

	for i := 0; i < len(fs); i++ {

		f := fs[i]

		// t.Run() works as a sub-test function
		t.Run(f.name, func(t *testing.T) {

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

			ListenAndServe(f, statement, t)

		})
	}
}

func TestDeposit(t *testing.T) {
	fs := make([]fixture, 5)
	fs[0] = fixture{
		name:         "Successfully deposit",
		request:      "/deposit?number=1001&amount=20",
		expectedCode: http.StatusOK,
		expectedBody: fmt.Sprintf("%v - %s - %v", 1001, "John", 20),
	}
	fs[1] = fixture{
		name:         "Account with number cant be found!",
		request:      "deposite?number=404&amount=20",
		expectedCode: http.StatusNotFound,
		expectedBody: fmt.Sprintf("Account with number %v can't be found!\n", 404),
	}
	fs[2] = fixture{
		name:         "Amount of deposit must be more than zero",
		request:      "deposite?number=1001&amount=-20",
		expectedCode: http.StatusBadRequest,
		expectedBody: "Amount must be more than zero!\n",
	}
	fs[3] = fixture{
		name:         "Invalid account number!",
		request:      "deposite?number=千一&amount=20",
		expectedCode: http.StatusBadRequest,
		expectedBody: fmt.Sprintf("%v is invalid account number!\n", "千一"),
	}
	fs[4] = fixture{
		name:         "Invalid amount number!",
		request:      "deposite?number=1001&amount=二十",
		expectedCode: http.StatusBadRequest,
		expectedBody: fmt.Sprintf("%v is invalid amount number!\n", "二十"),
	}

	for i := 0; i < len(fs); i++ {

		f := fs[i]

		t.Run(f.name, func(t *testing.T) {
			accounts[1001] = &core.Account{
				Customer: core.Customer{
					Name:    "John",
					Address: "Los Angeles, California",
					Phone:   "(213) 555 0147",
				},
				Number: 1001,
			}

			ListenAndServe(f, deposit, t)
		})
	}
}

func TestWithdraw(t *testing.T) {
	fs := make([]fixture, 6)
	fs[0] = fixture{
		name:         "Successfully withdraw",
		request:      "/withdraw?number=1001&amount=10",
		expectedCode: http.StatusOK,
		expectedBody: fmt.Sprintf("%v - %s - %v", 1001, "John", 10),
	}
	fs[1] = fixture{
		name:         "Account with number cant be found!",
		request:      "/withdraw?number=404&amount=20",
		expectedCode: http.StatusNotFound,
		expectedBody: fmt.Sprintf("Account with number %v can't be found!\n", 404),
	}
	fs[2] = fixture{
		name:         "Amount of withdraw must be more than zero!",
		request:      "/withdraw?number=1001&amount=-20",
		expectedCode: http.StatusBadRequest,
		expectedBody: "Amount must be more than zero!\n",
	}
	fs[3] = fixture{
		name:         "Invalid account number!",
		request:      "/withdraw?number=千一&amount=20",
		expectedCode: http.StatusBadRequest,
		expectedBody: "千一 is invalid account number!\n",
	}
	fs[4] = fixture{
		name:         "Invalid amount number!",
		request:      "/withdraw?number=1001&amount=二十",
		expectedCode: http.StatusBadRequest,
		expectedBody: "二十 is invalid amount number!\n",
	}
	fs[5] = fixture{
		name:         "Amount of withdraw must be more than deposit!",
		request:      "/withdraw?number=1001&amount=30",
		expectedCode: http.StatusBadRequest,
		expectedBody: "Amount of withdraw must be more than deposit!\n",
	}

	for i := 0; i < len(fs); i++ {

		f := fs[i]

		t.Run(f.name, func(t *testing.T) {
			accounts[1001] = &core.Account{
				Customer: core.Customer{
					Name:    "John",
					Address: "Los Angeles, California",
					Phone:   "(213) 555 0147",
				},
				Number:  1001,
				Balance: 20,
			}

			ListenAndServe(f, withdraw, t)
		})
	}
}

func TestTransfar(t *testing.T) {
	fs := make([]fixture, 8)
	fs[0] = fixture{
		name:         "Successfully transfer",
		request:      "/transtfer?from=1001&to=2002&amount=10",
		expectedCode: http.StatusOK,
		expectedBody: fmt.Sprintf("sender : %v\nreviever : %v", "1001 - John - 90", "2002 - C.J. - 110"),
	}
	fs[1] = fixture{
		name:         "Account with number cant be found",
		request:      "/transtfer?from=404&to=2002&amount=20",
		expectedCode: http.StatusNotFound,
		expectedBody: fmt.Sprintf("Account of sender with number %v can't be found!\n", 404),
	}
	fs[2] = fixture{
		name:         "Account with number cant be found",
		request:      "/transtfer?from=1001&to=404&amount=20",
		expectedCode: http.StatusNotFound,
		expectedBody: fmt.Sprintf("Account of reciever with number %v can't be found!\n", 404),
	}
	fs[3] = fixture{
		name:         "Amount must be more than zero!",
		request:      "/transfer?from=1001&to=2002&amount=-20",
		expectedCode: http.StatusBadRequest,
		expectedBody: "Amount must be more than zero!\n",
	}
	fs[4] = fixture{
		name:         "transfer is greater than deposit",
		request:      "/transfer?from=1001&to=2002&amount=200",
		expectedCode: http.StatusBadRequest,
		expectedBody: "transfer is greater than deposit!\n",
	}
	fs[5] = fixture{
		name:         "Invalid sender's account number!",
		request:      "/transfer?from=千一&to=2002&amount=100",
		expectedCode: http.StatusBadRequest,
		expectedBody: "千一 is invalid account number!\n",
	}
	fs[6] = fixture{
		name:         "Invalid reciever's account number!",
		request:      "/transfer?from=1001&to=二千二&amount=100",
		expectedCode: http.StatusBadRequest,
		expectedBody: "二千二 is invalid account number!\n",
	}
	fs[7] = fixture{
		name:         "Invalid amont number!",
		request:      "/transfer?from=1001&to=2002&amount=百",
		expectedCode: http.StatusBadRequest,
		expectedBody: "百 is invalid amount number!\n",
	}

	for i := 0; i < len(fs); i++ {

		f := fs[i]

		t.Run(f.name, func(t *testing.T) {
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

			ListenAndServe(f, transfer, t)
		})
	}
}

func TestTeapot(t *testing.T) {

	f := fixture{
		name:         "I'm a teapot.",
		request:      "/teapot",
		expectedCode: http.StatusTeapot,
		expectedBody: "418 : I'm a teapot.\n",
	}

	ListenAndServe(f, teapot, t)
}
