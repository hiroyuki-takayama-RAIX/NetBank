package core

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
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
	Number  int     `json:"id"`
	Balance float64 `json:"balance"`
}

// a field name in a struct must have capital initial when its encoded as json.
type Trade struct {
	Class  string  `json:"class"`
	Amount float64 `json:"amount"`
	From   int     `json:"from"`
	To     int     `json:"to"`
}

type netBank struct {
	db *sql.DB
}

func (a *Account) SetUniqueID(nb *netBank) error {
	id, err := nb.GetNewId()
	if err != nil {
		return err
	}
	a.Number = id
	return nil
}

func NewNetBank() (*netBank, error) {
	env := os.Getenv("Env")

	var (
		driver string
		source string
	)

	if env == "prod" {
		// デプロイ用のコンテナにアプリケーションを立ち上げて、本番用のDBが立ち上がっているコンテナに接続する。故にproduction_db:5432にぬけて接続する。
		driver = "pgx"
		source = "host=production_db port=5432 user=postgres database=netbank password=postgres sslmode=disable"
	} else {
		// ローカル環境でアプリケーションを立ち上げて、テスト用のDBが立ち上がっているコンテナに接続する。故にlocalhost:5180に向けて接続する。
		driver = "pgx"
		source = "host=localhost port=5180 user=testUser database=netbank_test password=testPassword sslmode=disable"
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

func (nb *netBank) Deposit(num int, money float64) (*Account, error) {
	// check money is more than 0
	if money <= 0 {
		return nil, fmt.Errorf("deposit of account_%v is less than 0. you was going to deposit %v$", num, money)
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
		return nil, err
	}

	// check balance is more than 0
	if balance < 0 {
		return nil, fmt.Errorf("balance of account_%v is less than 0. currenct balance is %v", num, balance)
	}

	// start the transaction
	tx, err := nb.db.Begin()
	if err != nil {
		return nil, err
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
		return nil, err
	} else {
		tx.Commit()
	}

	account, err := nb.GetAccount(num)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (nb *netBank) Withdraw(num int, money float64) (*Account, error) {
	if money <= 0 {
		return nil, fmt.Errorf("withdraw is less than zero. id_%v was going to withdraw %v", num, money)
	}

	// extract the account's balance
	balance, err := nb.GetBalance(num)
	if err != nil {
		return nil, err
	}

	// check which balance is more than 0 or not.
	if balance < 0 {
		return nil, fmt.Errorf("balance is less than 0. id_%v's currenct balance is %v", num, balance)
	}

	if balance-money < 0 {
		return nil, fmt.Errorf("amount is grater than the balance. your amount is %v, but the balance is %v", money, balance)
	}

	// start the transaction
	tx, err := nb.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// update the balance
	q := `
	UPDATE account 
	SET balance=$1 
	WHERE id=$2;
	`
	_, err = tx.ExecContext(context.Background(), q, balance-money, num)
	if err != nil {
		return nil, err
	} else {
		tx.Commit()
	}

	account, err := nb.GetAccount(num)
	if err != nil {
		return nil, err
	}

	return account, nil
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

func (nb *netBank) Transfer(sender int, reciever int, money float64) ([]*Account, error) {
	/*
			nb.Withdraw() と nb.Deposit() を流用する方法もあるが、
		    トランザクションの切り替えの間に取引が行われてしまう恐れがないように
			預金残高の削減と増加を一つのトランザクションにまとめる。
	*/
	if money <= 0 {
		return nil, fmt.Errorf("amount of transfer is less than zero. from id_%v to id_%v was going to withdraw %v", sender, reciever, money)
	}

	/*--- validation of sender ---*/
	senderBalance, err := nb.GetBalance(sender)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("sender's account(ID: %v) is not found: %w", sender, err)
		} else {
			return nil, err
		}
	}

	if senderBalance < 0 {
		return nil, fmt.Errorf("balance is less than 0. id_%v's currenct balance is %v", sender, senderBalance)
	}

	if senderBalance-money < 0 {
		return nil, fmt.Errorf("amount is grater than the balance. sender's amount is %v, but the balance is %v", money, senderBalance)
	}

	/*--- validation of reciever ---*/
	recieverBalance, err := nb.GetBalance(reciever)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("reciever's account(ID: %v) is not found: %w", reciever, err)
		} else {
			return nil, err
		}
	}

	if recieverBalance < 0 {
		return nil, fmt.Errorf("balance of account_%v is less than 0. currenct balance is %v", reciever, recieverBalance)
	}

	// start the transaction
	tx, err := nb.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// withdraw from sender's balance and deposit to reciever's balance.
	withdraw := "UPDATE account SET balance=$1 WHERE id=$2;"
	_, err = tx.ExecContext(context.Background(), withdraw, senderBalance-money, sender)
	if err != nil {
		return nil, err
	} else {
		deposit := "UPDATE account SET balance=$1 WHERE id=$2;"
		_, err = tx.ExecContext(context.Background(), deposit, recieverBalance+money, reciever)
		if err != nil {
			return nil, err
		} else {
			tx.Commit()
		}
	}

	accounts := make([]*Account, 2)
	from, err := nb.GetAccount(sender)
	if err != nil {
		return nil, err
	}
	to, err := nb.GetAccount(reciever)
	if err != nil {
		return nil, err
	}
	accounts[0] = from
	accounts[1] = to
	return accounts, nil
}

func (nb *netBank) CreateAccount(c *Customer) (*Account, error) {
	tx, err := nb.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	id, err := nb.GetNewId()
	if err != nil {
		return nil, err
	}

	// update the balance
	q := `
	INSERT INTO customer (id, username, addr, phone) 
	VALUES ($1, $2, $3, $4);
	`
	_, err = tx.ExecContext(context.Background(), q, id, c.Name, c.Address, c.Phone)
	if err != nil {
		return nil, err
	} else {
		q := `
		INSERT INTO account (id, balance) 
		VALUES ($1, $2);
		`
		_, err = tx.ExecContext(context.Background(), q, id, float64(0))
		if err != nil {
			return nil, err
		} else {
			tx.Commit()
		}
	}

	account, err := nb.GetAccount(id)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (nb *netBank) DeleteAccount(num int) error {
	tx, err := nb.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// check the existence of account having num as id.
	_, err = nb.GetAccount(num)
	if err != nil {
		return err
	}

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

func (nb *netBank) GetAccounts(min float64, max float64) ([]*Account, error) {
	var (
		name    string
		address string
		phone   string
		id      int
		balance float64
	)

	q := `SELECT username, addr, phone, account.id, balance 
	      FROM account 
		  INNER JOIN customer 
		  ON account.id=customer.id
		  WHERE balance>=$1 AND balance<=$2;`
	rows, err := nb.db.QueryContext(context.Background(), q, min, max)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []*Account{}

	// Iterate through the result set
	for rows.Next() {
		err := rows.Scan(&name, &address, &phone, &id, &balance)
		if err != nil {
			return nil, err
		}

		account := &Account{
			Customer: Customer{
				Name:    name,
				Address: address,
				Phone:   phone,
			},
			Number:  id,
			Balance: balance,
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (nb *netBank) GetAccount(num int) (*Account, error) {
	var (
		name    string
		address string
		phone   string
		id      int
		balance float64
	)

	q := `SELECT username, addr, phone, account.id, balance 
	      FROM account 
		  INNER JOIN customer 
		  ON account.id=customer.id
		  WHERE account.id=$1;`
	row := nb.db.QueryRowContext(context.Background(), q, num)

	err := row.Scan(&name, &address, &phone, &id, &balance)
	if err != nil {
		return nil, err
	}

	account := Account{
		Customer: Customer{
			Name:    name,
			Address: address,
			Phone:   phone,
		},
		Number:  id,
		Balance: balance,
	}

	return &account, nil
}

func (nb *netBank) GetBalance(id int) (float64, error) {
	account, err := nb.GetAccount(id)
	if err != nil {
		return 0, err
	} else {
		return account.Balance, nil
	}
}

func (nb *netBank) GetNewId() (int, error) {
	var id int

	q := `SELECT id
	      FROM account;
		  `
	rows, err := nb.db.QueryContext(context.Background(), q)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	ids := []int{}

	// Iterate through the result set
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return 0, err
		}
		ids = append(ids, id)
	}

	var newID int
	for _, id := range ids {
		newID = rand.Intn(2147483647)
		if newID != id {
			break
		}
	}
	return newID, nil
}

func (nb *netBank) UpdateAccount(id int, c *Customer) (*Account, error) {
	tx, err := nb.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// check update statement
	q := `
	UPDATE customer
	SET username=$1, addr=$2, phone=$3 
	WHERE id=$4;
	`
	_, err = tx.ExecContext(context.Background(), q, c.Name, c.Address, c.Phone, id)
	if err != nil {
		return nil, err
	} else {
		tx.Commit()
	}

	account, err := nb.GetAccount(id)
	if err != nil {
		return nil, err
	}

	return account, nil
}
