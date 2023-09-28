package core

import (
	"context"
	"fmt"
)

var (
	tnb *netBank
)

func ConnectTestDB() error {

	var err error

	tnb, err = NewNetBank()
	if err != nil {
		return err
	}
	return nil
}

func DisconnectTestDB() error {
	err := tnb.Close()
	if err != nil {
		return err
	}
	return nil
}

func InsertTestData() error {
	tx, err := tnb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := `

	INSERT INTO customer (id, username, addr, phone) 
	VALUES (1001, 'John', 'Los Angeles, California', '(213) 444 0147');

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

func DeleteTestData() error {
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

func compareErrors(want error, got error) string {
	if want == nil {
		if got == want {
			return ""
		} else {
			msg := fmt.Sprintf("Actual error value is not match to want:\nwant%v\ngot :%v", want, got.Error())
			return msg
		}
	} else {
		if want.Error() == got.Error() {
			return ""
		} else {
			msg := fmt.Sprintf("Actual error value is not match to want:\nwant:%v\ngot :%v", want.Error(), got.Error())
			return msg
		}
	}
}

func InsertLargeTestData() error {
	tx, err := tnb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := `

	-- 顧客テーブルにランダムなデータを1000行挿入
	INSERT INTO customer (id, username, addr, phone)
	SELECT
		generate_series(1, 1000) AS id,
		md5(random()::text) AS username,
		substr(md5(random()::text), 1, 15) AS addr,
		'+1-' || floor(random() * 1000000000)::bigint AS phone;
	
	-- アカウントテーブルにランダムなデータを50行挿入
	INSERT INTO account (id, balance)
	SELECT
		generate_series(1, 500) AS id,
		random() * 10000 AS balance;
    `
	_, err = tx.ExecContext(context.Background(), q)
	if err != nil {
		return err
	} else {
		tx.Commit()
	}

	return nil
}

func CreateTestTable() error {
	tx, err := tnb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := `
	CREATE TABLE right_table (
		id INT PRIMARY KEY,
		name varchar(80)
	);

	INSERT INTO right_table 
	VALUES (1, 'a'), (2, 'b'), (3, 'c');
	
	CREATE TABLE left_table (
		id INT PRIMARY KEY,
		name varchar(80)
	);

	INSERT INTO left_table 
	VALUES (1, 'd'), (2, 'e'), (6, 'f');	
	`
	_, err = tx.ExecContext(context.Background(), q)
	if err != nil {
		return err
	} else {
		tx.Commit()
	}

	return nil
}

func CreateTestTableWithDuplicateValue() error {
	tx, err := tnb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := `
	CREATE TABLE duplicate_table (
		id INT,
		name varchar(80)
	);

	INSERT INTO duplicate_table 
	VALUES (1, 'a'), (2, 'b'), (3, 'c'), (3, 'd'), (2, 'c'), (4, 'a'), (1, 'a');
	`
	_, err = tx.ExecContext(context.Background(), q)
	if err != nil {
		return err
	} else {
		tx.Commit()
	}

	return nil
}

func DropTables() error {
	tx, err := tnb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := `
	DROP TABLE right_table, left_table;
	`
	_, err = tx.ExecContext(context.Background(), q)
	if err != nil {
		return err
	} else {
		tx.Commit()
	}

	return nil
}

func DropDuplicateTables() error {
	tx, err := tnb.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := `
	DROP TABLE duplicate_table;
	`
	_, err = tx.ExecContext(context.Background(), q)
	if err != nil {
		return err
	} else {
		tx.Commit()
	}

	return nil
}
