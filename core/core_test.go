package core

// without import bank.go, you can use objects ans functions because core_test.go and bank.go are in the same module.
import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
)

// init is used to assign to db as global variable
func init() {
	var err error
	db, err = sql.Open("pgx", "host=localhost port=1234 user=postgres database=netbank password=passw0rd sslmode=disable")
	if err != nil {
		fmt.Printf("failed to connect db: %v\n", err)
		// You can choose to exit or handle the error accordingly.
	}
}

func TestNewNetBank(t *testing.T) {
	got := NewNetBank()
	expected := &netBank{}
	if reflect.DeepEqual(got, expected) != true {
		t.Errorf("failed to NewNetBank:\ngot %v\nexpected %v", got, expected)
	}
}

func TestCreateAccount(t *testing.T) {
	err := BeforeEach()
	if err != nil {
		t.Errorf("failed to BeforeEach(): %v", err)
	}
	defer AfterEach()

	nb := netBank{}
	id := 2002
	name := "C.J."
	addr := "Los Santos"
	phone := "(080) 1457 9387"

	err = nb.CreateAccount(id, name, addr, phone)
	if err != nil {
		t.Errorf("failed to create a new account_%v: %v", id, err)
	}

	s, err := nb.Statement(id)
	if err != nil {
		t.Errorf("cannnot generate the statement of account_%v: %v", id, err)
	}

	expected := "2002 - C.J. - 0"
	if s != expected {
		t.Errorf("got unexpected statement:\nexpected %v\ngot %v", expected, s)
	}
}

func TestDeleteAccount(t *testing.T) {
	err := BeforeEach()
	if err != nil {
		t.Errorf("failed to BeforeEach(): %v", err)
	}
	defer AfterEach()

	nb := netBank{}
	id := 2002

	err = nb.DeleteAccount(id)
	if err != nil {
		t.Errorf("failed to create a new account_%v: %v", id, err)
	}

	_, err = nb.Statement(id)
	if err == nil {
		t.Errorf("failed to delete account_%v", id)
	}
}

func TestDeposit(t *testing.T) {
	err := BeforeEach()
	if err != nil {
		t.Errorf("failed to BeforeEach(): %v", err)
	}
	defer AfterEach()

	nb := netBank{}
	id := 1001
	money := float64(100)

	err = nb.Deposit(id, money)
	if err != nil {
		t.Errorf("failed to deposit on account_%v: %v", id, err)
	}

	s, err := nb.Statement(id)
	if err != nil {
		t.Errorf("cannnot generate the statement of account_%v: %v", id, err)
	}

	expected := "1001 - John - 200"
	if s != expected {
		t.Errorf("got unexpected statement:\nexpected %v\ngot %v", expected, s)
	}
}

// Invalid pattern test

func TestWithdraw(t *testing.T) {
	err := BeforeEach()
	if err != nil {
		t.Errorf("failed to setup talbes: %v", err)
	}
	defer AfterEach()

	id := 1001
	money := float64(100)
	nb := NewNetBank()

	err = nb.Withdraw(id, money)
	if err != nil {
		t.Errorf("failed to withdraw: %v", err)
	}

	s, err := nb.Statement(id)
	if err != nil {
		t.Errorf("cannnot generate the statement: %v", err)
	}

	expected := "1001 - John - 0"
	if s != expected {
		t.Errorf("got unexpected statement:\nexpected %v\ngot %v", expected, s)
	}
}

func TestTransfer(t *testing.T) {
	err := BeforeEach()
	if err != nil {
		t.Errorf("failed to BeforeEach(): %v", err)
	}
	defer AfterEach()

	nb := NewNetBank()
	sender_id := 1001
	reciever_id := 3003
	money := float64(50)

	err = nb.Transfer(sender_id, reciever_id, money)
	if err != nil {
		t.Errorf("failed to withdraw: %v", err)
	}

	s, err := nb.Statement(sender_id)
	if err != nil {
		t.Errorf("cannnot generate the statement of account_%v: %v", sender_id, err)
	}
	expected := "1001 - John - 50"
	if s != expected {
		t.Errorf("got unexpected statement:\nexpected %v\ngot %v", expected, s)
	}

	s, err = nb.Statement(reciever_id)
	if err != nil {
		t.Errorf("cannnot generate the statement of account_%v: %v", reciever_id, err)
	}
	expected = "3003 - Ide Non No - 150"
	if s != expected {
		t.Errorf("got unexpected statement:\nexpected %v\ngot %v", expected, s)
	}
}

func TestStatement(t *testing.T) {
	err := BeforeEach()
	if err != nil {
		t.Errorf("failed to BeforeEach(): %v", err)
	}
	defer AfterEach()

	nb := NewNetBank()
	id := 1001
	s, err := nb.Statement(1001)
	if err != nil {
		t.Errorf("cannnot generate the statement of account_%v: %v", id, err)
	}

	expected := "1001 - John - 100"
	if s != expected {
		t.Errorf("got unexpected statement:\nexpected %v\ngot %v", expected, s)
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

func TestDatabaseSql(t *testing.T) {
	err := BeforeEach()
	if err != nil {
		t.Errorf("failed to setup tables: %v", err)
	}
	defer AfterEach()

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
		phone = "(213) 555 0147"
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
			phone:   "(213) 555 0147",
		}

		if reflect.DeepEqual(ss, expected) == false {
			t.Errorf("failed to extract row: %v\n got: %v\n expected: %v\n", err, ss, expected)
		}
	})
}
