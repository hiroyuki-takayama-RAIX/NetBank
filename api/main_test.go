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

	// Create a new HTTP request to the "/accounts" endpoint
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
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
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

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	got := rr.Body.String()
	expected := `{"name":"John","address":"Los Angeles, California","phone":"(213) 555 0147","id":1001,"balance":100}`

	assert.JSONEq(t, got, expected)
}

func TestCreateAccount(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	// Create a new Gin router
	router := gin.Default()
	router.POST("/accounts", createAccount)

	// Create a sample account to send in the request
	c := core.Customer{
		Name:    "C.J.",
		Address: "Los Santos",
		Phone:   "(080) 1457 9387",
	}

	// Convert the newaccount struct to JSON
	j, err := json.Marshal(c)
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
	var got core.Account
	gotJSON := rr.Body.Bytes()
	_ = json.Unmarshal(gotJSON, &got)
	expected := &core.Account{
		Customer: c,
		Number:   0,
		Balance:  0,
	}

	// Check that the response matches the newaccount we sent
	nameBool := got.Name == expected.Name
	addressBool := got.Address == expected.Address
	phoneBool := got.Phone == expected.Phone
	balanceBool := got.Balance == expected.Balance
	result := nameBool && addressBool && phoneBool && balanceBool
	if result == false {
		t.Errorf("except id in Account, handler returned wrong response body: \n[got]\n%v \n[want]\n%v", got, expected)
	}

	idBool := (got.Number != 1001) && (got.Number != 3003)
	if idBool == false {
		t.Errorf("api registerd duplicate id: %v", got.Number)
	}
}

func TestDeleteAccount(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	// Create a new Gin router
	router := gin.Default()
	router.DELETE("/accounts/:id", deleteAccount)

	// Create a new HTTP request with the JSON data
	req, err := http.NewRequest("DELETE", "/accounts/1001", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Serve the request and record the response
	router.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `"successfully delete account(ID: 1001)"`
	got := rr.Body.String()
	if got != expected {
		t.Errorf("handler returned wrong response body: \n[got]\n%v \n[want]\n%v", got, expected)
	}
}

func TestUpdateAccount(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	// Create a sample account to send in the request
	c := &core.Customer{
		Name:    "johnson",
		Address: "Libercity",
		Phone:   "(080) 4075 8704",
	}

	// Convert the newaccount struct to JSON
	j, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new Gin router
	router := gin.Default()
	router.PUT("/accounts/:id", updateAccount)

	// Create a new HTTP request with the JSON data
	req, err := http.NewRequest("PUT", "/accounts/1001", bytes.NewBuffer(j))
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Serve the request and record the response
	router.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"name":"johnson","address":"Libercity","phone":"(080) 4075 8704"}`
	got := rr.Body.String()
	assert.JSONEq(t, got, expected)
}
