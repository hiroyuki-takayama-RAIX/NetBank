package core

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var db *sql.DB

// init is used to assign to db as global variable
func init() {
	var err error
	db, err = sql.Open("pgx", "host=localhost port=1234 user=postgres database=netbank password=passw0rd sslmode=disable")
	if err != nil {
		fmt.Printf("failed to connect db: %v\n", err)
		// You can choose to exit or handle the error accordingly.
	}
}

// Customer ...
type Customer struct {
	Name    string
	Address string
	Phone   string
}

// Account ...
type Account struct {
	Customer
	Number  int32
	Balance float64
}

type netBank struct {
}

func NewNetBank() *netBank {
	return &netBank{}
}

func (nb *netBank) Deposit(num int, money float64) error {
	// check money is more than 0
	if money <= 0 {
		return fmt.Errorf("deposit of account_%v is less than 0. you was going to deposit %v$", num, money)
	}

	// extract the account's balance
	var balance float64
	q := `SELECT balance FROM account WHERE id=$1;`
	row := db.QueryRowContext(context.Background(), q, num)
	err := row.Scan(&balance)
	if err != nil {
		return err
	}

	// check balance is more than 0
	if balance < 0 {
		return fmt.Errorf("balance of account_%v is less than 0. currenct balance is %v", num, balance)
	}

	// start the transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// update the balance
	q = "UPDATE account SET balance=$1 WHERE id=$2;"
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
	q := `SELECT balance FROM account WHERE id=$1;`
	row := db.QueryRowContext(context.Background(), q, num)
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
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// update the balance
	q = "UPDATE account SET balance=$1 WHERE id=$2;"
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
		  INNER JOIN customer ON account.id=customer.id 
		  WHERE account.id=$1;`
	row := db.QueryRowContext(context.Background(), q, num)
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
	senderQuery := `SELECT balance FROM account WHERE id=$1;`
	senderRow := db.QueryRowContext(context.Background(), senderQuery, sender)
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
	recieverRow := db.QueryRowContext(context.Background(), recieverQuery, reciever)
	err = recieverRow.Scan(&recieverBalance)
	if err != nil {
		return err
	}

	if recieverBalance < 0 {
		return fmt.Errorf("balance of account_%v is less than 0. currenct balance is %v", reciever, recieverBalance)
	}

	// start the transaction
	tx, err := db.Begin()
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
	tx, err := db.Begin()
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
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// update the balance
	q := `DELETE FROM account WHERE id=$1;`
	_, err = tx.ExecContext(context.Background(), q, num)
	if err != nil {
		return err
	} else {
		q := `DELETE FROM customer WHERE id=$1;`
		_, err = tx.ExecContext(context.Background(), q, num)
		if err != nil {
			return err
		} else {
			tx.Commit()
		}
	}
	return nil
}
