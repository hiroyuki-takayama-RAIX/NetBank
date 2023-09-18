package core

import (
	"context"
)

var tnb *netBank

const (
	driver = "pgx"
	source = "host=localhost port=5180 user=testUser database=netbank_test password=testPassword sslmode=disable"
)

func connectTestDB() error {
	var err error

	tnb, err = NewNetBank(driver, source)
	if err != nil {
		return err
	}
	return nil
}

func disconnectTestDB() error {
	err := tnb.Close()
	if err != nil {
		return err
	}
	return nil
}

func insertTestData() error {
	tx, err := tnb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := `
	INSERT INTO customer (id, username, addr, phone) 
	VALUES (1001, 'John', 'Los Angeles, California', '(213) 555 0147');

	INSERT INTO account (id, balance) 
	VALUES (1001, 100);

	INSERT INTO customer (id, username, addr, phone) 
	VALUES (3003, 'Ide Non No', 'Ta No Tsu', '(0120) 117 117');

	INSERT INTO account (id, balance) 
	VALUES (3003, 100);
    `
	_, err = tx.ExecContext(context.Background(), q)
	if err != nil {
		return err
	} else {
		tx.Commit()
	}

	return nil
}

func deleteTestData() error {
	tx, err := tnb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := `
	DELETE FROM account;
	DELETE FROM customer;
	`
	_, err = tx.ExecContext(context.Background(), q)
	if err != nil {
		return err
	} else {
		tx.Commit()
	}

	return nil
}
