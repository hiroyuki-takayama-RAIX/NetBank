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
	name      string
	uri       string
	bodyParam string
	code      int
	body      string
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
	// set beforeEach and afterEach functoins.
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	fs := make([]*fixture, 1)
	fs[0] = &fixture{
		name: "Successfully get all accounts.",
		uri:  "/accounts",
		code: http.StatusOK,
		body: `[{"name":"John","address":"Los Angeles, California","phone":"(213) 555 0147","id":1001,"balance":100},{"name":"Ide Non No","address":"Ta No Tsu","phone":"(0120) 117 117","id":3003,"balance":100}]`,
	}

	for _, f := range fs {
		t.Run(f.name, func(t *testing.T) {
			// Create a new HTTP request
			req, err := http.NewRequest("GET", f.uri, nil)
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
			assert.Equal(t, f.code, rr.Code)

			// Compare actual response and expected response
			// reflect.DeepEqual or general logical operator cannot compare got and expected...
			assert.JSONEq(t, f.body, rr.Body.String())
		})
	}
}

func TestGetAccount(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	fs := make([]*fixture, 3)
	fs[0] = &fixture{
		name: "Successfully get an account.",
		uri:  "/accounts/1001",
		code: http.StatusOK,
		body: `{"name":"John","address":"Los Angeles, California","phone":"(213) 555 0147","id":1001,"balance":100}`,
	}
	fs[1] = &fixture{
		name: "Invalied id number.",
		uri:  "/accounts/千百一",
		code: http.StatusBadRequest,
		body: `{"error":"got 千百一 as invalied id"}`,
	}
	fs[2] = &fixture{
		name: "Account not found.",
		uri:  "/accounts/404",
		code: http.StatusNotFound,
		body: `{"error":"account(ID: 404) doesnt exist"}`,
	}

	for _, f := range fs {
		t.Run(f.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", f.uri, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			router := gin.Default()
			router.GET("/accounts/:id", getAccount)
			router.ServeHTTP(rr, req)
			assert.Equal(t, f.code, rr.Code)
			assert.JSONEq(t, f.body, rr.Body.String())
		})
	}
}

func TestCreateAccount(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	fs := make([]*fixture, 5)
	fs[0] = &fixture{
		name:      "Successfully create an account.",
		uri:       "/accounts",
		bodyParam: `{"name":"John","address":"Los Angeles, California","phone":"(213) 555 0147"}`,
		code:      http.StatusCreated,
		body:      `{"name":"John","address":"Los Angeles, California","phone":"(213) 555 0147","id":0,"balance":0}`,
	}
	fs[1] = &fixture{
		name:      "Invalied id number.",
		uri:       "/accounts",
		bodyParam: `{"name":"John","address":"Los Angeles, California","phone":"(213) 555 0147","id":1001,"balance":100}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"Invalied request"}`,
	}
	fs[2] = &fixture{
		name:      "Empty name",
		uri:       "/accounts",
		bodyParam: `{"name":"","address":"Los Angeles, California","phone":"(213) 555 0147"}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"request has empty name"}`,
	}
	fs[3] = &fixture{
		name:      "Empty address",
		uri:       "/accounts",
		bodyParam: `{"name":"John","address":"","phone":"(213) 555 0147"}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"request has empty address"}`,
	}
	fs[4] = &fixture{
		name:      "Empty phone number",
		uri:       "/accounts",
		bodyParam: `{"name":"John","address":"Los Angeles, California","phone":""}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"request has empty phone number"}`,
	}

	for _, f := range fs {
		t.Run(f.name, func(t *testing.T) {
			router := gin.Default()
			router.POST("/accounts", createAccount)

			// Create a new HTTP request with the JSON data
			bs := []byte(f.bodyParam)
			req, err := http.NewRequest("POST", f.uri, bytes.NewBuffer(bs))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, f.code, rr.Code)

			// parse response body as Account instance
			var got core.Account
			gotJSON := rr.Body.Bytes()
			_ = json.Unmarshal(gotJSON, &got)

			// expected response body
			expected := &core.Account{
				Customer: core.Customer{
					Name:    "John",
					Address: "Los Angeles, California",
					Phone:   "(213) 555 0147",
				},
				Balance: 0,
			}

			// Check the response json except id because server assign random number to id
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
		})
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
