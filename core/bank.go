package core

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
)

// Customer ...
type Customer struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

// Account ...
type Account struct {
	Customer
	Number  int32
	Balance float64
}

type netBank struct {
	db *sql.DB
}

func NewNetBank() (*netBank, error) {
	env := os.Getenv("Env")

	var (
		driver string
		source string
	)

	if env == "test" {
		driver = "pgx"
		source = "host=localhost port=5180 user=testUser database=netbank_test password=testPassword sslmode=disable"
	} else if env == "prod" {
		driver = "pgx"
		source = "host=localhost port=1234 user=postgres database=netbank password=postgres sslmode=disable"
	} else {
		msg := fmt.Sprintf("invaild environment valiable %v", env)
		panic(msg)
	}

	db, err := sql.Open(driver, source)
	if err != nil {
		return nil, err
	}
	return &netBank{db: db}, nil
}

func (nb *netBank) Ping() error {
	return nb.db.Ping()
}

func (nb *netBank) Close() error {
	return nb.db.Close()
}

func (nb *netBank) Begin() (*sql.Tx, error) {
	return nb.db.Begin()
}

func (nb *netBank) Deposit(num int, money float64) error {
	// check money is more than 0
	if money <= 0 {
		return fmt.Errorf("deposit of account_%v is less than 0. you was going to deposit %v$", num, money)
	}

	// extract the account's balance
	var balance float64
	q := `
	SELECT balance 
	FROM account 
	WHERE id=$1;
	`
	row := nb.db.QueryRowContext(context.Background(), q, num)
	err := row.Scan(&balance)
	if err != nil {
		return err
	}

	// check balance is more than 0
	if balance < 0 {
		return fmt.Errorf("balance of account_%v is less than 0. currenct balance is %v", num, balance)
	}

	// start the transaction
	tx, err := nb.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// update the balance
	q = `
	UPDATE account 
	SET balance=$1 
	WHERE id=$2;
	`
	_, err = tx.ExecContext(context.Background(), q, money+balance, num)
	if err != nil {
		return err
	} else {
		tx.Commit()
	}

	return nil
}

func (nb *netBank) Withdraw(num int, money float64) error {
	if money <= 0 {
		return fmt.Errorf("withdraw is less than zero. id_%v was going to withdraw %v", num, money)
	}

	// extract the account's balance
	var balance float64
	q := `
	SELECT balance 
	FROM account 
	WHERE id=$1;
	`
	row := nb.db.QueryRowContext(context.Background(), q, num)
	err := row.Scan(&balance)
	if err != nil {
		return err
	}

	// check balance is more than 0
	if balance < 0 {
		return fmt.Errorf("balance is less than 0. id_%v's currenct balance is %v", num, balance)
	}

	if balance-money < 0 {
		return fmt.Errorf("balance is less than withdraw. id_%v's currenct balance is %v, but its withdraw is %v", num, balance, money)
	}

	// start the transaction
	tx, err := nb.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// update the balance
	q = `
	UPDATE account 
	SET balance=$1 
	WHERE id=$2;
	`
	_, err = tx.ExecContext(context.Background(), q, balance-money, num)
	if err != nil {
		return err
	} else {
		tx.Commit()
	}

	return nil
}

func (nb *netBank) Statement(num int) (string, error) {
	var (
		id      int
		balance float64
		name    string
	)

	q := `SELECT account.id, balance, username 
	      FROM account 
		  INNER JOIN customer 
		  ON account.id=customer.id 
		  WHERE account.id=$1;`
	row := nb.db.QueryRowContext(context.Background(), q, num)
	err := row.Scan(&id, &balance, &name)
	if err != nil {
		return "", err
	}

	s := fmt.Sprintf("%v - %v - %v", id, name, balance)
	return s, nil
}

func (nb *netBank) Transfer(sender int, reciever int, money float64) error {
	/*
			nb.Withdraw() と nb.Deposit() を流用する方法もあるが、
		    トランザクションの切り替えの間に取引が行われてしまう恐れがないように
			預金残高の削減と増加を一つのトランザクションにまとめる。
	*/
	if money <= 0 {
		return fmt.Errorf("amount of transfer is less than zero. from id_%v to id_%v was going to withdraw %v", sender, reciever, money)
	}

	/*--- validation of sender ---*/
	var senderBalance float64
	senderQuery := `
	SELECT balance 
	FROM account 
	WHERE id=$1;
	`
	senderRow := nb.db.QueryRowContext(context.Background(), senderQuery, sender)
	err := senderRow.Scan(&senderBalance)
	if err != nil {
		return err
	}

	if senderBalance < 0 {
		return fmt.Errorf("balance is less than 0. id_%v's currenct balance is %v", sender, senderBalance)
	}

	if senderBalance-money < 0 {
		return fmt.Errorf("balance is less than transfer. id_%v's currenct balance is %v, but its transfer is %v", sender, senderBalance, money)
	}

	/*--- validation of reciever ---*/
	var recieverBalance float64
	recieverQuery := `SELECT balance FROM account WHERE id=$1;`
	recieverRow := nb.db.QueryRowContext(context.Background(), recieverQuery, reciever)
	err = recieverRow.Scan(&recieverBalance)
	if err != nil {
		return err
	}

	if recieverBalance < 0 {
		return fmt.Errorf("balance of account_%v is less than 0. currenct balance is %v", reciever, recieverBalance)
	}

	// start the transaction
	tx, err := nb.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// withdraw from sender's balance and deposit to reciever's balance.
	withdraw := "UPDATE account SET balance=$1 WHERE id=$2;"
	_, err = tx.ExecContext(context.Background(), withdraw, senderBalance-money, sender)
	if err != nil {
		return err
	} else {
		deposit := "UPDATE account SET balance=$1 WHERE id=$2;"
		_, err = tx.ExecContext(context.Background(), deposit, recieverBalance+money, reciever)
		if err != nil {
			return err
		} else {
			tx.Commit()
		}
	}
	return nil
}

func (nb *netBank) CreateAccount(num int, name string, addr string, phone string) error {
	tx, err := nb.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// update the balance
	q := `
	INSERT INTO customer (id, username, addr, phone) 
	VALUES ($1, $2, $3, $4);
	`
	_, err = tx.ExecContext(context.Background(), q, num, name, addr, phone)
	if err != nil {
		return err
	} else {
		q := `
		INSERT INTO account (id, balance) 
		VALUES ($1, $2);
		`
		_, err = tx.ExecContext(context.Background(), q, num, float64(0))
		if err != nil {
			return err
		} else {
			tx.Commit()
		}
	}
	return nil
}

func (nb *netBank) DeleteAccount(num int) error {
	tx, err := nb.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// update the balance
	q := `
	DELETE FROM account 
	WHERE id=$1;
	`
	_, err = tx.ExecContext(context.Background(), q, num)
	if err != nil {
		return err
	} else {
		q := `
		DELETE FROM customer 
		WHERE id=$1;
		`
		_, err = tx.ExecContext(context.Background(), q, num)
		if err != nil {
			return err
		} else {
			tx.Commit()
		}
	}
	return nil
}
