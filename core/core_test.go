package core

// without import bank.go, you can use objects ans functions because core_test.go and bank.go are in the same module.
import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
	"gotest.tools/v3/assert"
)

func TestMain(m *testing.M) {
	err := ConnectTestDB()
	if err != nil {
		msg := fmt.Sprintf("failed to connect test db: %v", err)
		panic(msg)
	}

	code := m.Run()

	err = DisconnectTestDB()
	if err != nil {
		msg := fmt.Sprintf("failed to disconnect test db: %v", err)
		panic(msg)
	}

	os.Exit(code)
}

func TestBegin(t *testing.T) {
	err := InsertTestData()
	if err != nil {
		t.Errorf("failed to insert test data: %v", err)
	}
	defer DeleteTestData()

	tx, err := tnb.Begin()
	if err != nil {
		t.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	q := `
	UPDATE account 
	SET balance=200 
	WHERE id=1001;
	`
	_, err = tx.ExecContext(context.Background(), q)
	if err != nil {
		t.Errorf("failed to query: %v", err)
	} else {
		tx.Commit()
	}

	s, err := tnb.Statement(1001)
	if err != nil {
		t.Errorf("cannnot generate the statement of account_%v: %v", 1001, err)
	}

	expected := "1001 - John - 200"

	if reflect.DeepEqual(s, expected) == false {
		t.Errorf("rows mismatch! expected %v, got %v", expected, s)
	}
}

func TestNewNetBank(t *testing.T) {
	got, err := NewNetBank()
	if err != nil {
		t.Errorf("failed to genetare a new netBank instance: %v", err)
	}
	defer got.Close()

	err = got.Ping()
	if err != nil {
		t.Errorf("connection between netbank and db is not build: %v", err)
	}

	err = got.Close()
	if err != nil {
		t.Errorf("failed to close the connection: %v", err)
	}

	err = got.Ping()
	if err == nil {
		t.Errorf("failed to close the connection: %v", err)
	}
}

func TestGetAccounts(t *testing.T) {
	err := InsertTestData()
	if err != nil {
		t.Errorf("failed to insert test data: %v", err)
	}
	defer DeleteTestData()

	got, err := tnb.GetAccounts(0, 2147483647)
	if err != nil {
		t.Errorf("failed to get all accounts: %v", err)
	}

	expected := make([]*Account, 2)
	expected[0] = &Account{
		Customer: Customer{
			Name:    "John",
			Address: "Los Angeles, California",
			Phone:   "(213) 444 0147",
		},
		Number:  1001,
		Balance: 100,
	}
	expected[1] = &Account{
		Customer: Customer{
			Name:    "Ide Non No",
			Address: "Ta No Tsu",
			Phone:   "(0120) 117 117",
		},
		Number:  3003,
		Balance: 100,
	}

	if len(got) != len(expected) {
		t.Errorf("there is an unexpected test data in the db: %v", got)
	}

	for i := 0; i < len(got); i++ {
		if !reflect.DeepEqual(*got[i], *expected[i]) {
			t.Errorf("got unexpected data:\nexpected: %v\ngot: %v", expected, got)
		}
	}
}

func TestGetAccount(t *testing.T) {
	err := InsertTestData()
	if err != nil {
		t.Errorf("failed to insert test data: %v", err)
	}
	defer DeleteTestData()

	got, err := tnb.GetAccount(1001)
	if err != nil {
		t.Errorf("failed to get all accounts: %v", err)
	}

	expected := &Account{
		Customer: Customer{
			Name:    "John",
			Address: "Los Angeles, California",
			Phone:   "(213) 444 0147",
		},
		Number:  1001,
		Balance: 100,
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got unexpected data:\nexpected: %v\ngot: %v", expected, got)
	}
}

func TestCreateAccount(t *testing.T) {
	err := InsertTestData()
	if err != nil {
		t.Errorf("failed to insert test data: %v", err)
	}
	defer DeleteTestData()

	name := "C.J."
	addr := "Los Santos"
	phone := "(080) 1457 9387"

	c := &Customer{
		Name:    name,
		Address: addr,
		Phone:   phone,
	}

	got, err := tnb.CreateAccount(c)
	if err != nil {
		t.Errorf("failed to create a new account_%v: %v", got.Number, err)
	}

	expected := &Account{
		Customer: *c,
		Number:   got.Number,
		Balance:  0,
	}

	if !reflect.DeepEqual(*got, *expected) {
		t.Errorf("got unexpected value:\nexpected %v\ngot %v", expected, got)
	}
}

func TestDeleteAccount(t *testing.T) {
	err := InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer DeleteTestData()

	id := 1001

	err = tnb.DeleteAccount(id)
	if err != nil {
		t.Errorf("failed to delete a new account_%v: %v", id, err)
	}

	_, err = tnb.Statement(id)
	if err == nil {
		t.Errorf("failed to delete account_%v", id)
	}

	id = 404

	err = tnb.DeleteAccount(id)
	if err == nil {
		t.Errorf("account_%v shall not exist: %v", id, err)
	}
}

func TestUpdateAccount(t *testing.T) {
	err := InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer DeleteTestData()

	id := 1001

	c := &Customer{
		Name:    "johnson",
		Address: "Libercity",
		Phone:   "(080) 4075 8704",
	}

	got, err := tnb.UpdateAccount(id, c)
	if err != nil {
		t.Errorf("failed to update account_%v info: %v", id, err)
	}

	expected := &Account{
		Customer: *c,
		Number:   id,
		Balance:  100,
	}

	if !reflect.DeepEqual(*got, *expected) {
		t.Errorf("got unexpected value:\nexpected %v\ngot %v", expected, got)
	}
}

func TestDeposit(t *testing.T) {
	err := InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer DeleteTestData()

	id := 1001
	money := float64(100)

	got, err := tnb.Deposit(id, money)
	if err != nil {
		t.Errorf("failed to deposit on account_%v: %v", id, err)
	}

	expected := &Account{
		Customer: Customer{
			Name:    "John",
			Address: "Los Angeles, California",
			Phone:   "(213) 444 0147",
		},
		Number:  1001,
		Balance: 200,
	}
	assert.DeepEqual(t, expected, got)
}

func TestDeposit_Exception(t *testing.T) {}

func TestWithdraw(t *testing.T) {
	type fixture struct {
		name     string
		id       int
		money    float64
		expected *Account
		err      error
	}

	fs := make([]*fixture, 2)

	fs[0] = &fixture{
		name:  "Successfully Withdraw",
		id:    1001,
		money: 100,
		expected: &Account{
			Customer: Customer{
				Name:    "John",
				Address: "Los Angeles, California",
				Phone:   "(213) 444 0147",
			},
			Number:  1001,
			Balance: 0,
		},
		err: nil,
	}
	fs[1] = &fixture{
		name:     "Amount is grater than balance",
		id:       1001,
		money:    120,
		expected: nil,
		err:      errors.New("amount is grater than the balance. your amount is 120, but the balance is 100"),
	}

	for _, f := range fs {
		t.Run(f.name, func(t *testing.T) {
			err := InsertTestData()
			if err != nil {
				t.Errorf("failed to setup talbes: %v", err)
			}

			got, err := tnb.Withdraw(f.id, f.money)
			assert.DeepEqual(t, f.expected, got)
			msg := compareErrors(f.err, err)
			if msg != "" {
				t.Errorf(msg)
			}

			DeleteTestData()
		})
	}
}

func TestWithdraw_Exceptions(t *testing.T) {}

func TestTransfer(t *testing.T) {
	type fixture struct {
		name     string
		id       int
		money    float64
		to       int
		expected []*Account
		err      error
	}

	fs := make([]*fixture, 3)

	fs[0] = &fixture{
		name:  "successfully transfer",
		id:    1001,
		money: 20,
		to:    3003,
		expected: []*Account{
			{
				Customer: Customer{
					Name:    "John",
					Address: "Los Angeles, California",
					Phone:   "(213) 444 0147",
				},
				Number:  1001,
				Balance: 80,
			},
			{
				Customer: Customer{
					Name:    "Ide Non No",
					Address: "Ta No Tsu",
					Phone:   "(0120) 117 117",
				},
				Number:  3003,
				Balance: 120,
			},
		},
		err: nil,
	}
	fs[1] = &fixture{
		name:     "Amount is grater than balance",
		id:       1001,
		money:    120,
		to:       3003,
		expected: nil,
		err:      errors.New("amount is grater than the balance. sender's amount is 120, but the balance is 100"),
	}
	fs[2] = &fixture{
		name:     "Recievers Account(ID: 404) is not found",
		id:       1001,
		money:    20,
		to:       404,
		expected: nil,
		err:      errors.New("reciever's account(ID: 404) is not found: sql: no rows in result set"),
	}

	for _, f := range fs {
		t.Run(f.name, func(t *testing.T) {
			err := InsertTestData()
			if err != nil {
				t.Errorf("failed to setup talbes: %v", err)
			}

			got, err := tnb.Transfer(f.id, f.to, f.money)
			assert.DeepEqual(t, f.expected, got)
			msg := compareErrors(f.err, err)
			if msg != "" {
				t.Errorf(msg)
			}

			DeleteTestData()
		})
	}
}

func TestStatement(t *testing.T) {
	err := InsertTestData()
	if err != nil {
		t.Errorf("failed to insertTestData(): %v", err)
	}
	defer DeleteTestData()

	id := 1001
	s, err := tnb.Statement(1001)
	if err != nil {
		t.Errorf("cannnot generate the statement of account_%v: %v", id, err)
	}

	expected := "1001 - John - 100"
	if s != expected {
		t.Errorf("got unexpected statement:\nexpected %v\ngot %v", expected, s)
	}
}

func TestGetNewId(t *testing.T) {
	rand.Seed(1001)
	id, err := tnb.GetNewId()
	if err != nil {
		t.Errorf("failed to generate a unique id: %v", err)
	}
	if id == 1001 {
		t.Errorf("generate duplicate account id: %v", id)
	}

	rand.Seed(3003)
	id, err = tnb.GetNewId()
	if err != nil {
		t.Errorf("failed to generate a unique id: %v", err)
	}
	if id == 3003 {
		t.Errorf("generate duplicate account id: %v", id)
	}
}

// tests to learn usages of database/sql
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

// test to learning database/sql
func TestDatabaseSql(t *testing.T) {
	var (
		db     *sql.DB
		driver string
		source string
		err    error
	)

	err = InsertTestData()
	if err != nil {
		t.Errorf("failed to setup tables: %v", err)
	}
	defer DeleteTestData()

	// its so hardcoding, need to modify!
	driver = "pgx"
	source = "host=localhost port=5180 user=testUser database=netbank_test password=testPassword sslmode=disable"

	db, err = sql.Open(driver, source)
	if err != nil {
		t.Errorf("failed to connect db: %v", err)
	}

	t.Run("db.Ping()", func(t *testing.T) {
		// confirme the connection with db.Ping()
		err := db.Ping()
		if err != nil {
			t.Errorf("app and db are not connected: %v", err)
		}
	})

	t.Run("db.QueryRowContext()", func(t *testing.T) {
		var (
			id      int
			balance float64
		)

		// db.QueryRowCOntext() is used to extract one row
		row := db.QueryRowContext(context.Background(), "SELECT * FROM account")
		/* any error handling is necessary!
		if err != nil {
			t.Errorf("failed to extract rows from DB: %v", err)
		}
		*/

		// assign values of row to variables
		err := row.Scan(&id, &balance)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}

		a := account{
			id:      id,
			balance: balance,
		}

		expected := account{
			id:      1001,
			balance: 100,
		}

		if reflect.DeepEqual(a, expected) == false {
			t.Errorf("rows mismatch! expected %v, got %v", expected, a)
		}
	})

	// function to check a query is successfully commited or not in INSERT, UPDATE, and DELETE.
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
		phone = "(213) 444 0147"
	)

	t.Run("INSERT", func(t *testing.T) {
		// function to check a query is successfully commited or not.

		// db.Begin() is necessary before quering create, update, and delete.
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("failed to db.Begin(): %v", err)
			return
		}
		// when any kinds of error happens in a transaction, call tx.Rollback() to cancel queries which you execute
		defer tx.Rollback()

		// check insert statement
		q := `
		INSERT INTO customer (id, username, addr, phone) 
		VALUES ($1, $2, $3, $4);
		`
		_, err = tx.ExecContext(context.Background(), q, id, name, addr, phone)
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
	})

	t.Run("UPDATE", func(t *testing.T) {
		// once you call tx.Commit(), you must call db.Begin() and set tx.Rollback() again
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("failed to db.Begin(): %v", err)
			return
		}
		defer tx.Rollback()

		// check update statement
		q := `
		UPDATE customer
		SET phone='(080) 1457 9387' 
		WHERE id=$1;
		`
		_, err = tx.ExecContext(context.Background(), q, id)
		if err != nil {
			t.Errorf("failed to tx.ExecuteContext(): %v", err)
			return
		} else {
			tx.Commit()
		}
		cs, _ := getAllRows(fmt.Sprintf("SELECT * FROM customer WHERE id=%v", id))
		expected := customer{
			id:       id,
			username: name,
			addr:     addr,
			phone:    "(080) 1457 9387",
		}
		if reflect.DeepEqual(cs[0], expected) == false {
			t.Errorf("failed to update rows: %v\n got: %v\n expected: %v\n", err, cs[0], expected)
		}
	})

	t.Run("DELETE", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("failed to db.Begin(): %v", err)
			return
		}
		defer tx.Rollback()

		// check update statement
		q := "delete FROM customer WHERE id=$1;"
		_, err = tx.ExecContext(context.Background(), q, id)
		if err != nil {
			t.Errorf("failed to tx.ExecuteContext(): %v", err)
			return
		} else {
			tx.Commit()
		}
		cs, _ := getAllRows(fmt.Sprintf("SELECT * FROM customer WHERE id=%v", id))
		if len(cs) != 0 {
			t.Errorf("id 5432 in customer shuould not exist: %v", cs)
		}
	})

	t.Run("INNER_JOIN", func(t *testing.T) {
		type sampleStruct struct {
			id      int
			balance float64
			name    string
			addr    string
			phone   string
		}

		var (
			id      int
			balance float64
			name    string
			addr    string
			phone   string
		)

		q := `
		SELECT account.id, balance, username, addr, phone 
		FROM account 
		INNER JOIN customer 
		ON account.id=customer.id 
		WHERE account.id=1001;
		`
		row := db.QueryRowContext(context.Background(), q)
		err := row.Scan(&id, &balance, &name, &addr, &phone)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}

		ss := sampleStruct{
			id:      id,
			balance: balance,
			name:    name,
			addr:    addr,
			phone:   phone,
		}

		expected := sampleStruct{
			id:      1001,
			balance: 100,
			name:    "John",
			addr:    "Los Angeles, California",
			phone:   "(213) 444 0147",
		}

		if reflect.DeepEqual(ss, expected) == false {
			t.Errorf("failed to extract row: %v\n got: %v\n expected: %v\n", err, ss, expected)
		}
	})
}

func TestGetBalance(t *testing.T) {
	type fixture struct {
		name     string
		id       int
		expected float64
		err      error
	}

	fs := make([]*fixture, 2)

	fs[0] = &fixture{
		name:     "Successfully get balance",
		id:       1001,
		expected: 100,
		err:      nil,
	}
	fs[1] = &fixture{
		name:     "Not found id",
		id:       404,
		expected: 0,
		err:      errors.New("sql: no rows in result set"),
	}

	for _, f := range fs {
		t.Run(f.name, func(t *testing.T) {
			err := InsertTestData()
			if err != nil {
				t.Errorf("failed to setup talbes: %v", err)
			}

			got, err := tnb.GetBalance(f.id)
			assert.DeepEqual(t, f.expected, got)
			msg := compareErrors(f.err, err)
			if msg != "" {
				t.Errorf(msg)
			}

			DeleteTestData()
		})
	}
}
