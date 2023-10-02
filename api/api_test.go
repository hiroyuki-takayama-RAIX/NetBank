package api

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
		name: "Successfully Get all accounts.",
		uri:  "/accounts",
		code: http.StatusOK,
		body: `[{"name":"John","address":"Los Angeles, California","phone":"(213) 444 0147","id":1001,"balance":100},{"name":"Ide Non No","address":"Ta No Tsu","phone":"(0120) 117 117","id":3003,"balance":100}]`,
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
			router.GET("/accounts", GetAccounts)

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
		name: "Successfully Get an account.",
		uri:  "/accounts/1001",
		code: http.StatusOK,
		body: `{"name":"John","address":"Los Angeles, California","phone":"(213) 444 0147","id":1001,"balance":100}`,
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
			router.GET("/accounts/:id", GetAccount)
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

	fs := make([]*fixture, 8)
	fs[0] = &fixture{
		name:      "Successfully Create an account.",
		uri:       "/accounts",
		bodyParam: `{"name":"John","address":"Los Angeles, California","phone":"(213) 444 0147"}`,
		code:      http.StatusCreated,
		body:      `{"name":"John","address":"Los Angeles, California","phone":"(213) 444 0147","id":0,"balance":0}`,
	}
	fs[1] = &fixture{
		name:      "Invalied id number.",
		uri:       "/accounts",
		bodyParam: "{'name':'John','address':'Los Angeles, California','phone':'(213) 444 0147'}", // not json because using single quote.
		code:      http.StatusBadRequest,
		body:      `{"error":"Invalied request"}`,
	}
	fs[2] = &fixture{
		name:      "Empty name",
		uri:       "/accounts",
		bodyParam: `{"name":"","address":"Los Angeles, California","phone":"(213) 444 0147"}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"request has empty name"}`,
	}
	fs[3] = &fixture{
		name:      "Empty address",
		uri:       "/accounts",
		bodyParam: `{"name":"John","address":"","phone":"(213) 444 0147"}`,
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
	fs[5] = &fixture{
		name:      "Empty name (no name field)",
		uri:       "/accounts",
		bodyParam: `{"address":"Los Angeles, California","phone":"(213) 444 0147"}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"request has empty name"}`,
	}
	fs[6] = &fixture{
		name:      "Empty address 2 (no address field)",
		uri:       "/accounts",
		bodyParam: `{"name":"John","phone":"(213) 444 0147"}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"request has empty address"}`,
	}
	fs[7] = &fixture{
		name:      "Empty phone number 2 (no phone field)",
		uri:       "/accounts",
		bodyParam: `{"name":"John","address":"Los Angeles, California"}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"request has empty phone number"}`,
	}
	/*
		fs[8] = &fixture{
			name:      "Duplicate phone number",
			uri:       "/accounts",
			bodyParam: `{"name":"John","address":"Los Angeles, California","phone":"(213) 444 0147"}`,
			code:      http.StatusConflict,
			body:      `{"error":"there is posibblity of duplicate registration considering phone number"}`,
		}
	*/

	for _, f := range fs {
		t.Run(f.name, func(t *testing.T) {
			router := gin.Default()
			router.POST("/accounts", CreateAccount)

			// Create a new HTTP request with the JSON data
			bs := []byte(f.bodyParam)
			req, err := http.NewRequest("POST", f.uri, bytes.NewBuffer(bs))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			if assert.Equal(t, f.code, rr.Code) == true && f.code == 201 {
				// parse response body as Account instance
				var got core.Account
				gotJSON := rr.Body.Bytes()
				_ = json.Unmarshal(gotJSON, &got)

				// expected response body
				expected := &core.Account{
					Customer: core.Customer{
						Name:    "John",
						Address: "Los Angeles, California",
						Phone:   "(213) 444 0147",
					},
					Balance: 0,
				}

				// Check the response json except id because server assign random number as account id
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
			} else {
				assert.JSONEq(t, f.body, rr.Body.String())
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

	fs := make([]*fixture, 3)
	fs[0] = &fixture{
		name: "Successfully delete an account.",
		uri:  "/accounts/1001",
		code: http.StatusNoContent,
		body: "",
	}
	fs[1] = &fixture{
		name: "Invalied id number.",
		uri:  "/accounts/千百一",
		code: http.StatusBadRequest,
		body: `{"error":"got 千百一 as invalied id"}`,
	}
	fs[2] = &fixture{
		name: "Account not found",
		uri:  "/accounts/404",
		code: http.StatusNotFound,
		body: `{"error":"account(ID: 404) doesnt exist"}`,
	}

	for _, f := range fs {
		t.Run(f.name, func(t *testing.T) {
			req, err := http.NewRequest("DELETE", f.uri, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			router := gin.Default()
			router.DELETE("/accounts/:id", DeleteAccount)
			router.ServeHTTP(rr, req)
			assert.Equal(t, f.code, rr.Code)
			if f.code == 204 {
				assert.Equal(t, f.body, rr.Body.String())
			} else {
				assert.JSONEq(t, f.body, rr.Body.String())
			}
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	fs := make([]*fixture, 10)
	fs[0] = &fixture{
		name:      "Successfully update an account.",
		uri:       "/accounts/3003",
		bodyParam: `{"name":"John","address":"Los Angeles, California","phone":"(213) 444 0147"}`,
		code:      http.StatusCreated,
		body:      `{"name":"John","address":"Los Angeles, California","phone":"(213) 444 0147","id":3003,"balance":100}`,
	}
	fs[1] = &fixture{
		name:      "Invalied id number.",
		uri:       "/accounts/3003",
		bodyParam: "{'name':'John','address':'Los Angeles, California','phone':'(213) 444 0147'}", // not json because using single quote.
		code:      http.StatusBadRequest,
		body:      `{"error":"Invalied request"}`,
	}
	fs[2] = &fixture{
		name:      "Empty name",
		uri:       "/accounts/3003",
		bodyParam: `{"name":"","address":"Los Angeles, California","phone":"(213) 444 0147"}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"request has empty name"}`,
	}
	fs[3] = &fixture{
		name:      "Empty address",
		uri:       "/accounts/3003",
		bodyParam: `{"name":"John","address":"","phone":"(213) 444 0147"}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"request has empty address"}`,
	}
	fs[4] = &fixture{
		name:      "Empty phone number",
		uri:       "/accounts/3003",
		bodyParam: `{"name":"John","address":"Los Angeles, California","phone":""}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"request has empty phone number"}`,
	}
	fs[5] = &fixture{
		name:      "Empty name (no name field)",
		uri:       "/accounts/3003",
		bodyParam: `{"address":"Los Angeles, California","phone":"(213) 444 0147"}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"request has empty name"}`,
	}
	fs[6] = &fixture{
		name:      "Empty address 2 (no address field)",
		uri:       "/accounts/3003",
		bodyParam: `{"name":"John","phone":"(213) 444 0147"}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"request has empty address"}`,
	}
	fs[7] = &fixture{
		name:      "Empty phone number 2 (no phone field)",
		uri:       "/accounts/3003",
		bodyParam: `{"name":"John","address":"Los Angeles, California"}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"request has empty phone number"}`,
	}
	fs[8] = &fixture{
		name:      "Invalied id number.",
		uri:       "/accounts/千百一",
		bodyParam: `{"name":"John","address":"Los Angeles, California","phone":"(213) 444 0147"}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"got 千百一 as invalied id"}`,
	}
	fs[9] = &fixture{
		name:      "Account not found",
		uri:       "/accounts/404",
		bodyParam: `{"name":"John","address":"Los Angeles, California","phone":"(213) 444 0147"}`,
		code:      http.StatusNotFound,
		body:      `{"error":"account(ID: 404) doesnt exist"}`,
	}
	/*
		fs[10] = &fixture{
			name:      "Duplicate phone number",
			uri:       "/accounts",
			bodyParam: `{"name":"John","address":"Los Angeles, California","phone":"(213) 444 0147"}`,
			code:      http.StatusConflict,
			body:      `{"error":"there is posibblity of duplicate registration considering phone number"}`,
		}
	*/

	for _, f := range fs {
		t.Run(f.name, func(t *testing.T) {
			router := gin.Default()
			router.PUT("/accounts/:id", UpdateAccount)
			bs := []byte(f.bodyParam)
			req, err := http.NewRequest("PUT", f.uri, bytes.NewBuffer(bs))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, f.code, rr.Code)
			assert.JSONEq(t, f.body, rr.Body.String())
		})
	}
}

func TestGetBalance(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	fs := make([]*fixture, 3)
	fs[0] = &fixture{
		name: "Successfully Get a balance.",
		uri:  "/accounts/1001/balance",
		code: http.StatusOK,
		body: `{"id":1001,"balance":100}`,
	}
	fs[1] = &fixture{
		name: "Invalied id number.",
		uri:  "/accounts/千百一/balance",
		code: http.StatusBadRequest,
		body: `{"error":"got 千百一 as invalied id"}`,
	}
	fs[2] = &fixture{
		name: "Account not found.",
		uri:  "/accounts/404/balance",
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
			router.GET("/accounts/:id/balance", GetBalance)
			router.ServeHTTP(rr, req)
			assert.Equal(t, f.code, rr.Code)
			assert.JSONEq(t, f.body, rr.Body.String())
		})
	}
}

func TestDeposit(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	f := &fixture{
		name:      "Successfully deposit.",
		uri:       "/accounts/1001/balance",
		bodyParam: `{"class":"deposit","amount":20}`,
		code:      http.StatusOK,
		body:      `{"name":"John","address":"Los Angeles, California","phone":"(213) 444 0147","id":1001,"balance":120}`,
	}

	router := gin.Default()
	router.PATCH("/accounts/:id/balance", FinancialTransaction)
	bs := []byte(f.bodyParam)
	req, err := http.NewRequest("PATCH", f.uri, bytes.NewBuffer(bs))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, f.code, rr.Code)
	assert.JSONEq(t, f.body, rr.Body.String())
}

func TestWithdraw(t *testing.T) {
	fs := make([]*fixture, 2)
	fs[0] = &fixture{
		name:      "Successfully withdraw",
		uri:       "/accounts/1001/balance",
		bodyParam: `{"class":"withdraw","amount":20}`,
		code:      http.StatusOK,
		body:      `{"name":"John","address":"Los Angeles, California","phone":"(213) 444 0147","id":1001,"balance":80}`,
	}
	fs[1] = &fixture{
		name:      "Amount is grater than balance",
		uri:       "/accounts/1001/balance",
		bodyParam: `{"class":"withdraw","amount":120}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"amount is grater than the balance. your amount is 120, but the balance is 100"}`,
	}

	for _, f := range fs {
		t.Run(f.name, func(t *testing.T) {
			err := core.InsertTestData()
			if err != nil {
				t.Errorf("failed to insertTestData(): %v", err)
			}
			router := gin.Default()
			router.PATCH("/accounts/:id/balance", FinancialTransaction)
			bs := []byte(f.bodyParam)
			req, err := http.NewRequest("PATCH", f.uri, bytes.NewBuffer(bs))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, f.code, rr.Code)
			assert.JSONEq(t, f.body, rr.Body.String())
			core.DeleteTestData()
		})
	}
}

func TestTransfer(t *testing.T) {
	fs := make([]*fixture, 3)
	fs[0] = &fixture{
		name:      "Successfully transfer",
		uri:       "/accounts/1001/balance",
		bodyParam: `{"class":"transfer","amount":20,"from":1001,"to":3003}`,
		code:      http.StatusOK,
		body:      `[{"name":"John","address":"Los Angeles, California","phone":"(213) 444 0147","id":1001,"balance":80},{"name":"Ide Non No","address":"Ta No Tsu","phone":"(0120) 117 117","id":3003,"balance":120}]`,
	}
	fs[1] = &fixture{
		name:      "Amount is grater than balance",
		uri:       "/accounts/1001/balance",
		bodyParam: `{"class":"transfer","amount":120,"from":1001,"to":3003}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"amount is grater than the balance. sender's amount is 120, but the balance is 100"}`,
	}
	fs[2] = &fixture{
		name:      "Reciever's account not found",
		uri:       "/accounts/1001/balance",
		bodyParam: `{"class":"transfer","amount":20,"from":1001,"to":404}`,
		code:      http.StatusNotFound,
		body:      `{"error":"reciever's account(ID: 404) is not found: sql: no rows in result set"}`,
	}

	for _, f := range fs {
		t.Run(f.name, func(t *testing.T) {
			err := core.InsertTestData()
			if err != nil {
				t.Errorf("failed to insertTestData(): %v", err)
			}
			router := gin.Default()
			router.PATCH("/accounts/:id/balance", FinancialTransaction)
			bs := []byte(f.bodyParam)
			req, err := http.NewRequest("PATCH", f.uri, bytes.NewBuffer(bs))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, f.code, rr.Code)
			assert.JSONEq(t, f.body, rr.Body.String())
			core.DeleteTestData()
		})
	}
}

func TestFinancialTransaction(t *testing.T) {
	err := core.InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer core.DeleteTestData()

	fs := make([]*fixture, 5)
	fs[0] = &fixture{
		name:      "Amount is less than zero",
		uri:       "/accounts/1001/balance",
		bodyParam: `{"class":"test","amount":-20.0}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"amount is less than zero. your input is -20"}`,
	}
	fs[1] = &fixture{
		name:      "Invalied id number.",
		uri:       "/accounts/千百一/balance",
		bodyParam: `{"class":"test","amount":20}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"got 千百一 as invalied id"}`,
	}
	fs[2] = &fixture{
		name:      "Account not found",
		uri:       "/accounts/404/balance",
		bodyParam: `{"class":"test","amount":20}`,
		code:      http.StatusNotFound,
		body:      `{"error":"account(ID: 404) doesnt exist"}`,
	}
	fs[3] = &fixture{
		name:      "Invalied class",
		uri:       "/accounts/1001/balance",
		bodyParam: `{"class":"foreign exchange","amount":20}`,
		code:      http.StatusBadRequest,
		body:      `{"error":"you about to do foreign exchange, but its not defined."}`,
	}
	fs[4] = &fixture{
		name:      "Successfully trading",
		uri:       "/accounts/1001/balance",
		bodyParam: `{"class":"test","amount":20}`,
		code:      http.StatusOK,
		body:      `{"msg":"FinancialTransaction() is executed collectlly."}`,
	}

	for _, f := range fs {
		t.Run(f.name, func(t *testing.T) {
			router := gin.Default()
			router.PATCH("/accounts/:id/balance", FinancialTransaction)
			bs := []byte(f.bodyParam)
			req, err := http.NewRequest("PATCH", f.uri, bytes.NewBuffer(bs))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, f.code, rr.Code)
			assert.JSONEq(t, f.body, rr.Body.String())
		})
	}
}
