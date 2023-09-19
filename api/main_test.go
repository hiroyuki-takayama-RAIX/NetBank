package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hiroyuki-takayama-RAIX/core"
)

type fixture struct {
	name         string
	request      string
	expectedCode int
	expectedBody string
}

func TestMain(m *testing.M) {
	err := core.ConnectTestDB()
	if err != nil {
		msg := fmt.Sprintf("failed to connect test db: %v", err)
		panic(msg)
	}

	code := m.Run()

	err = core.DisconnectTestDB()
	if err != nil {
		msg := fmt.Sprintf("failed to disconnect test db: %v", err)
		panic(msg)
	}

	os.Exit(code)
}

func TestGetAccounts(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	// Create a new HTTP request to the "/albums" endpoint
	req, err := http.NewRequest("GET", "/accounts", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Create a new Gin router and handler function
	router := gin.Default()
	router.GET("/accounts", getAccounts)

	// Serve the request and record the response
	router.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: \n[got]\n%v \n[want]\n%v", status, http.StatusOK)
	}

	got := rr.Body.String()
	expected := `[{"name":"John","address":"Los Angeles, California","phone":"(213) 555 0147","id":1001,"balance":100},{"name":"Ide Non No","address":"Ta No Tsu","phone":"(0120) 117 117","id":3003,"balance":100}]`

	assert.JSONEq(t, got, expected)

	// reflect.DeepEqual or general logical operator cannot compare got and expected...
}

func TestGetAccount(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	req, err := http.NewRequest("GET", "/accounts/1001", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	router := gin.Default()
	router.GET("/accounts/:id", getAccount)

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: \n[got]\n%v \n[want]\n%v", status, http.StatusOK)
	}

	got := rr.Body.String()
	expected := `{"name":"John","address":"Los Angeles, California","phone":"(213) 555 0147","id":1001,"balance":100}`

	assert.JSONEq(t, got, expected)
}

func TestPostAlbums(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	// Create a new Gin router
	router := gin.Default()
	router.POST("/albums", createAccount)

	// Create a sample album to send in the request
	account := core.Account{
		Customer: core.Customer{
			Name:    "C.J.",
			Address: "Los Santos",
			Phone:   "(080) 1457 9387",
		},
		Number:  4,
		Balance: 100,
	}

	// Convert the newAlbum struct to JSON
	j, err := json.Marshal(account)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new HTTP request with the JSON data
	req, err := http.NewRequest("POST", "/accounts", bytes.NewBuffer(j))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Serve the request and record the response
	router.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Check the response body (JSON content)
	got := rr.Body.String()
	j, _ = json.Marshal(account)
	expected := string(j)

	// Check that the response matches the newAlbum we sent
	assert.JSONEq(t, got, expected)
}

// i wanna change this name better...
// this function get fixture and show test result.
/*
func CommonTestLogic(f *fixture, h func(w http.ResponseWriter, req *http.Request), t *testing.T) {
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
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()
	// set fs as fixture to test some cases in same programm.
	/*
		fs := []fixture{}
		fs = append(fs, fixture{"successfully getting statement", 1001, fmt.Sprintf("%v - %s - %v", 1001, "John", 0)})
		fs = append(fs, fixture{"account with number cant be found", 404, fmt.Sprintf("Account with number %v can't be found!", 404)})
*/

// upper lines are correct, but make()'s second argument is useful to make clear number of test pattern.
// *fixture means
/*
	fs := make([]*fixture, 3)
	fs[0] = &fixture{
		name:         "Successfully getting statement",
		request:      fmt.Sprintf("/statement?number=%v", 1001),
		expectedCode: http.StatusOK,
		expectedBody: fmt.Sprintf("%v - %s - %v", 1001, "John", 100),
	}
	/* necessary to write more detail fixtures!
	fs[1] = &fixture{
		name:         "Account with the number cant be found",
		request:      fmt.Sprintf("/statement?number=%v", 404),
		expectedCode: http.StatusNotFound,
		expectedBody: fmt.Sprintf("Account with number %v can't be found!\n", 404),
	}
*/
/*
	fs[1] = &fixture{
		name:         "Account number is missing",
		request:      fmt.Sprintf("/statement?n=%v", 1001),
		expectedCode: http.StatusBadRequest,
		expectedBody: "Account number is missing!\n",
	}
	fs[2] = &fixture{
		name:         "Invalid account number!",
		request:      fmt.Sprintf("/statement?number=%v", "千一"),
		expectedCode: http.StatusBadRequest,
		expectedBody: fmt.Sprintf("%v is invalid account number!\n", "千一"),
	}

	for i := 0; i < len(fs); i++ {

		f := fs[i]

		// t.Run() works as a sub-test function
		t.Run(f.name, func(t *testing.T) {
			CommonTestLogic(f, statement, t)
		})
	}
}

func TestDeposit(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	fs := make([]*fixture, 4)
	fs[0] = &fixture{
		name:         "Successfully deposit",
		request:      "/deposit?number=1001&amount=20",
		expectedCode: http.StatusOK,
		expectedBody: fmt.Sprintf("%v - %s - %v", 1001, "John", 120),
	}
	/* necessary to write more detail fixtures!
	fs[1] = &fixture{
		name:         "Account with number cant be found!",
		request:      "deposite?number=404&amount=20",
		expectedCode: http.StatusNotFound,
		expectedBody: fmt.Sprintf("Account with number %v can't be found!\n", 404),
	}
*/
/*
	fs[1] = &fixture{
		name:         "Amount of deposit must be more than zero",
		request:      "deposite?number=1001&amount=-20",
		expectedCode: http.StatusBadRequest,
		expectedBody: "deposit of account_1001 is less than 0. you was going to deposit -20$\n",
	}
	fs[2] = &fixture{
		name:         "Invalid account number!",
		request:      "deposite?number=千一&amount=20",
		expectedCode: http.StatusBadRequest,
		expectedBody: fmt.Sprintf("%v is invalid account number!\n", "千一"),
	}
	fs[3] = &fixture{
		name:         "Invalid amount number!",
		request:      "deposite?number=1001&amount=二十",
		expectedCode: http.StatusBadRequest,
		expectedBody: fmt.Sprintf("%v is invalid amount number!\n", "二十"),
	}

	for i := 0; i < len(fs); i++ {

		f := fs[i]

		t.Run(f.name, func(t *testing.T) {
			CommonTestLogic(f, deposit, t)
		})
	}
}

func TestWithdraw(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	fs := make([]*fixture, 3)
	fs[0] = &fixture{
		name:         "Successfully withdraw",
		request:      "/withdraw?number=1001&amount=10",
		expectedCode: http.StatusOK,
		expectedBody: fmt.Sprintf("%v - %s - %v", 1001, "John", 90),
	}
	/* necessary to more detail fixtures!
	fs[1] = &fixture{
		name:         "Account with number cant be found!",
		request:      "/withdraw?number=404&amount=20",
		expectedCode: http.StatusNotFound,
		expectedBody: fmt.Sprintf("Account with number %v can't be found!\n", 404),
	}
	fs[2] = &fixture{
		name:         "Amount of withdraw must be more than zero!",
		request:      "/withdraw?number=1001&amount=-20",
		expectedCode: http.StatusBadRequest,
		expectedBody: "Amount must be more than zero!\n",
	}
*/
/*
	fs[1] = &fixture{
		name:         "Invalid account number!",
		request:      "/withdraw?number=千一&amount=20",
		expectedCode: http.StatusBadRequest,
		expectedBody: "千一 is invalid account number!\n",
	}
	fs[2] = &fixture{
		name:         "Invalid amount number!",
		request:      "/withdraw?number=1001&amount=二十",
		expectedCode: http.StatusBadRequest,
		expectedBody: "二十 is invalid amount number!\n",
	}
	/*
		fs[5] = &fixture{
			name:         "Amount of withdraw must be more than deposit!",
			request:      "/withdraw?number=1001&amount=30",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Amount of withdraw must be more than deposit!\n",
		}
*/
/*

	for i := 0; i < len(fs); i++ {
		f := fs[i]
		t.Run(f.name, func(t *testing.T) {
			CommonTestLogic(f, withdraw, t)
		})
	}
}

func TestTransfar(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	fs := make([]*fixture, 4)
	fs[0] = &fixture{
		name:         "Successfully transfer",
		request:      "/transtfer?from=1001&to=3003&amount=10",
		expectedCode: http.StatusOK,
		expectedBody: fmt.Sprintf("sender : %v\nreviever : %v", "1001 - John - 90", "3003 - Ide Non No - 110"),
	}
	/* necessary to write more detail fixture!
	fs[1] = &fixture{
		name:         "Account with number cant be found",
		request:      "/transtfer?from=404&to=2002&amount=20",
		expectedCode: http.StatusNotFound,
		expectedBody: fmt.Sprintf("Account of sender with number %v can't be found!\n", 404),
	}
	fs[2] = &fixture{
		name:         "Account with number cant be found",
		request:      "/transtfer?from=1001&to=404&amount=20",
		expectedCode: http.StatusNotFound,
		expectedBody: fmt.Sprintf("Account of reciever with number %v can't be found!\n", 404),
	}
	fs[3] = &fixture{
		name:         "Amount must be more than zero!",
		request:      "/transfer?from=1001&to=2002&amount=-20",
		expectedCode: http.StatusBadRequest,
		expectedBody: "Amount must be more than zero!\n",
	}
	fs[4] = &fixture{
		name:         "transfer is greater than deposit",
		request:      "/transfer?from=1001&to=2002&amount=200",
		expectedCode: http.StatusBadRequest,
		expectedBody: "transfer is greater than deposit!\n",
	}
*/
/*
	fs[1] = &fixture{
		name:         "Invalid sender's account number!",
		request:      "/transfer?from=千一&to=2002&amount=100",
		expectedCode: http.StatusBadRequest,
		expectedBody: "千一 is invalid account number!\n",
	}
	fs[2] = &fixture{
		name:         "Invalid reciever's account number!",
		request:      "/transfer?from=1001&to=二千二&amount=100",
		expectedCode: http.StatusBadRequest,
		expectedBody: "二千二 is invalid account number!\n",
	}
	fs[3] = &fixture{
		name:         "Invalid amont number!",
		request:      "/transfer?from=1001&to=2002&amount=百",
		expectedCode: http.StatusBadRequest,
		expectedBody: "百 is invalid amount number!\n",
	}

	for i := 0; i < len(fs); i++ {
		f := fs[i]
		t.Run(f.name, func(t *testing.T) {
			CommonTestLogic(f, transfer, t)
		})
	}
}

func TestTeapot(t *testing.T) {

	f := &fixture{
		name:         "I'm a teapot.",
		request:      "/teapot",
		expectedCode: http.StatusTeapot,
		expectedBody: "418 : I'm a teapot.\n",
	}

	CommonTestLogic(f, teapot, t)
}

func TestCreateAccount(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	fs := make([]*fixture, 2)
	fs[0] = &fixture{
		name:         "Successfully create a new account",
		request:      "/createaccount?number=2002&name=C.J.&addr=Los Santos&phone=(080) 1457 9387",
		expectedCode: http.StatusOK,
		expectedBody: "2002 - C.J. - 0",
	}
	fs[1] = &fixture{
		name:         "Invalid sender's account number!",
		request:      "/createaccount?number=二千二",
		expectedCode: http.StatusBadRequest,
		expectedBody: "二千二 is invalid account number!\n",
	}

	for i := 0; i < len(fs); i++ {
		f := fs[i]
		t.Run(f.name, func(t *testing.T) {
			CommonTestLogic(f, createAccount, t)
		})
	}
}

func TestDeleteAccount(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	fs := make([]*fixture, 2)
	fs[0] = &fixture{
		name:         "Successfully delete an account",
		request:      "/deleteaccount?number=1001",
		expectedCode: http.StatusBadRequest,
		expectedBody: "sql: no rows in result set\n",
	}
	fs[1] = &fixture{
		name:         "Invalid sender's account number!",
		request:      "/deleteaccount?number=二千二",
		expectedCode: http.StatusBadRequest,
		expectedBody: "二千二 is invalid account number!\n",
	}

	for i := 0; i < len(fs); i++ {
		f := fs[i]
		t.Run(f.name, func(t *testing.T) {
			CommonTestLogic(f, deleteAccount, t)
		})
	}
}
*/
