package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"

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
	// set fs as fixture to test some cases in same programm.
	/*
		fs := []fixture{}
		fs = append(fs, fixture{"successfully getting statement", 1001, fmt.Sprintf("%v - %s - %v", 1001, "John", 0)})
		fs = append(fs, fixture{"account with number cant be found", 404, fmt.Sprintf("Account with number %v can't be found!", 404)})
	*/

	// upper lines are correct, but make()'s second argument is useful to make clear number of test pattern.
	// *fixture means
	fs := make([]*fixture, 4)
	fs[0] = &fixture{
		name:         "Successfully getting statement",
		request:      fmt.Sprintf("/statement?number=%v", 1001),
		expectedCode: http.StatusOK,
		expectedBody: fmt.Sprintf("%v - %s - %v", 1001, "John", 100),
	}
	fs[1] = &fixture{
		name:         "Account with the number cant be found",
		request:      fmt.Sprintf("/statement?number=%v", 404),
		expectedCode: http.StatusNotFound,
		expectedBody: fmt.Sprintf("Account with number %v can't be found!\n", 404),
	}
	fs[2] = &fixture{
		name:         "Account number is missing",
		request:      fmt.Sprintf("/statement?n=%v", 1001),
		expectedCode: http.StatusBadRequest,
		expectedBody: "Account number is missing!\n",
	}
	fs[3] = &fixture{
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

			CommonTestLogic(f, statement, t)

		})
	}
}

func TestDeposit(t *testing.T) {
	fs := make([]*fixture, 5)
	fs[0] = &fixture{
		name:         "Successfully deposit",
		request:      "/deposit?number=1001&amount=20",
		expectedCode: http.StatusOK,
		expectedBody: fmt.Sprintf("%v - %s - %v", 1001, "John", 20),
	}
	fs[1] = &fixture{
		name:         "Account with number cant be found!",
		request:      "deposite?number=404&amount=20",
		expectedCode: http.StatusNotFound,
		expectedBody: fmt.Sprintf("Account with number %v can't be found!\n", 404),
	}
	fs[2] = &fixture{
		name:         "Amount of deposit must be more than zero",
		request:      "deposite?number=1001&amount=-20",
		expectedCode: http.StatusBadRequest,
		expectedBody: "Amount must be more than zero!\n",
	}
	fs[3] = &fixture{
		name:         "Invalid account number!",
		request:      "deposite?number=千一&amount=20",
		expectedCode: http.StatusBadRequest,
		expectedBody: fmt.Sprintf("%v is invalid account number!\n", "千一"),
	}
	fs[4] = &fixture{
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

			CommonTestLogic(f, deposit, t)
		})
	}
}

func TestWithdraw(t *testing.T) {
	fs := make([]*fixture, 6)
	fs[0] = &fixture{
		name:         "Successfully withdraw",
		request:      "/withdraw?number=1001&amount=10",
		expectedCode: http.StatusOK,
		expectedBody: fmt.Sprintf("%v - %s - %v", 1001, "John", 10),
	}
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
	fs[3] = &fixture{
		name:         "Invalid account number!",
		request:      "/withdraw?number=千一&amount=20",
		expectedCode: http.StatusBadRequest,
		expectedBody: "千一 is invalid account number!\n",
	}
	fs[4] = &fixture{
		name:         "Invalid amount number!",
		request:      "/withdraw?number=1001&amount=二十",
		expectedCode: http.StatusBadRequest,
		expectedBody: "二十 is invalid amount number!\n",
	}
	fs[5] = &fixture{
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

			CommonTestLogic(f, withdraw, t)
		})
	}
}

func TestTransfar(t *testing.T) {
	fs := make([]*fixture, 8)
	fs[0] = &fixture{
		name:         "Successfully transfer",
		request:      "/transtfer?from=1001&to=2002&amount=10",
		expectedCode: http.StatusOK,
		expectedBody: fmt.Sprintf("sender : %v\nreviever : %v", "1001 - John - 90", "2002 - C.J. - 110"),
	}
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
	fs[5] = &fixture{
		name:         "Invalid sender's account number!",
		request:      "/transfer?from=千一&to=2002&amount=100",
		expectedCode: http.StatusBadRequest,
		expectedBody: "千一 is invalid account number!\n",
	}
	fs[6] = &fixture{
		name:         "Invalid reciever's account number!",
		request:      "/transfer?from=1001&to=二千二&amount=100",
		expectedCode: http.StatusBadRequest,
		expectedBody: "二千二 is invalid account number!\n",
	}
	fs[7] = &fixture{
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

// tests to comfirm usages of database/sql
func TestDatabaseSql(t *testing.T) {
	type account struct {
		id      int
		balance float64
	}

	type customer struct {
		id       int
		username string
		addr     string
		phone    string
	}

	// try to connect db.
	db, err := sql.Open("pgx", "host=localhost port=1234 user=postgres database=netbank password=passw0rd sslmode=disable")
	if err != nil {
		t.Errorf("failed to connect db: %v", err)
	}

	t.Run("db.Ping()", func(t *testing.T) {
		// confirme the connection
		err = db.Ping()
		if err != nil {
			t.Errorf("app and db are not connected: %v", err)
		}
	})

	t.Run("db.QueryRowContext()", func(t *testing.T) {
		var (
			id      int
			balance float64
		)

		row := db.QueryRowContext(context.Background(), "SELECT * FROM account")
		if err != nil {
			t.Errorf("failed to extract rows from DB: %v", err)
		}

		err = row.Scan(&id, &balance)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}

		a := account{
			id:      id,
			balance: balance,
		}

		expected := account{
			id:      1001,
			balance: 0,
		}

		if reflect.DeepEqual(a, expected) == false {
			t.Errorf("rows mismatch! expected %v, got %v", expected, a)
		}
	})

	t.Run("tx.ExecContext() and db.QueryContext()", func(t *testing.T) {
		// function to check a query is successfully commited or not.
		getAllRows := func(q string) ([]customer, error) {
			rows, err := db.QueryContext(context.Background(), q)
			if err != nil {
				t.Errorf("query all customer: %v", err)
				return nil, err
			}
			defer rows.Close()

			var customers []customer
			for rows.Next() {
				var (
					id       int
					username string
					addr     string
					phone    string
				)

				if err := rows.Scan(&id, &username, &addr, &phone); err != nil {
					t.Errorf("scan the customer: %v", err)
					return nil, err
				}
				customers = append(customers, customer{id: id, username: username, addr: addr, phone: phone})

				if err = rows.Close(); err != nil {
					t.Errorf("rows close: %v", err)
					return nil, err
				}
				if err = rows.Err(); err != nil {
					t.Errorf("scan customer: %v", err)
					return nil, err
				}
			}
			return customers, nil
		}

		// rows detail to be used as queries and comparisones
		var (
			id    = 2002
			name  = "C.J."
			addr  = "Los Santos"
			phone = "(213) 555 0147"
		)

		// db.Begin() is necessary before quering create, update, and delete.
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("failed to db.Begin(): %v", err)
			return
		}
		// when any kinds of error happens in a transaction, call tx.Rollback() to cancel queries which you execute
		defer tx.Rollback()

		// check insert statement
		insertQuery := "INSERT INTO customer (id, username, addr, phone) VALUES ($1, $2, $3, $4);"
		_, err = tx.ExecContext(context.Background(), insertQuery, id, name, addr, phone)
		if err != nil {
			t.Errorf("failed to tx.ExecuteContext(): %v", err)
			return
		} else {
			// to reflect query on DB, call tx.Commit
			tx.Commit()
		}
		cs, _ := getAllRows(fmt.Sprintf("SELECT * FROM customer WHERE id=%v", id))
		expected := customer{
			id:       id,
			username: name,
			addr:     addr,
			phone:    phone,
		}
		if reflect.DeepEqual(cs[0], expected) == false {
			t.Errorf("failed to insert rows: %v\n got: %v\n expected: %v\n", err, cs[0], expected)
		}

		// once you call tx.Commit(), you must call db.Begin() and set tx.Rollback() again
		tx, err = db.Begin()
		if err != nil {
			t.Errorf("failed to db.Begin(): %v", err)
			return
		}
		defer tx.Rollback()

		// check update statement
		updateQuery := "update customer set phone='(080) 1457 9387' WHERE id=$1;"
		_, err = tx.ExecContext(context.Background(), updateQuery, id)
		if err != nil {
			t.Errorf("failed to tx.ExecuteContext(): %v", err)
			return
		} else {
			tx.Commit()
		}
		cs, _ = getAllRows(fmt.Sprintf("SELECT * FROM customer WHERE id=%v", id))
		expected = customer{
			id:       id,
			username: name,
			addr:     addr,
			phone:    "(080) 1457 9387",
		}
		if reflect.DeepEqual(cs[0], expected) == false {
			t.Errorf("failed to update rows: %v\n got: %v\n expected: %v\n", err, cs[0], expected)
		}

		tx, err = db.Begin()
		if err != nil {
			t.Errorf("failed to db.Begin(): %v", err)
			return
		}
		defer tx.Rollback()

		// check update statement
		deleteQuery := "delete FROM customer WHERE id=$1;"
		_, err = tx.ExecContext(context.Background(), deleteQuery, id)
		if err != nil {
			t.Errorf("failed to tx.ExecuteContext(): %v", err)
			return
		} else {
			tx.Commit()
		}
		cs, _ = getAllRows(fmt.Sprintf("SELECT * FROM customer WHERE id=%v", id))
		if len(cs) != 0 {
			t.Errorf("id 5432 in customer shuould not exist: %v", cs)
		}
	})
}
