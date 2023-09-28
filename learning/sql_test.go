package learning

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hiroyuki-takayama-RAIX/core"
	"gotest.tools/v3/assert"
)

type rightTable struct {
	Id   int
	Name string
}

var db *sql.DB

func TestMain(m *testing.M) {
	err := core.ConnectTestDB()
	if err != nil {
		msg := fmt.Sprintf("failed to connect test db: %v", err)
		panic(msg)
	}
	defer core.DisconnectTestDB()

	// sql.Open()に設定するドライバーとソースを設定する。
	driver := "pgx"
	source := "host=localhost port=5180 user=testUser database=netbank_test password=testPassword sslmode=disable"

	db, err = sql.Open(driver, source)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	code := m.Run()

	os.Exit(code)
}

// database/sqlモジュールとsqlクエリに関する学習用テスト
func TestDatabaseSql(t *testing.T) {

	t.Run("db.Ping()", func(t *testing.T) {
		// sql.Open()ではちゃんと繋がっているのか分からないため、db.Ping()を実行してデータベースとアプリケーションの接続を確かめる。
		err := db.Ping()
		if err != nil {
			t.Errorf("app and db are not connected: %v", err)
		}
	})

	// SELECT句、FROM句、INNER_JOIN句、WHERE句の学習用テスト
	// db.QueryRowContext()の学習用テスト
	t.Run("db.QueryRowContext()", func(t *testing.T) {
		err := core.InsertTestData()
		if err != nil {
			t.Errorf("failed to setup tables: %v", err)
		}
		defer core.DeleteTestData()

		var (
			id      int
			balance float64
			name    string
			address string
			phone   string
		)

		// INNERを省略しても同じ結果が返ってくる。
		//　以下の結果は`SELECT * FROM account, customer WHERE account.id=customer.id` でも同じ結果が返ってくる
		q := `
		SELECT account.id, balance, username, addr, phone
		FROM account
		INNER JOIN customer
		ON account.id=customer.id
		WHERE account.id=$1
		`

		// db.QueryRowCOntext()は1行のレコードを取得する際に使用する。
		// 複数業がヒットするクエリの場合は、先頭のレコートを取得する。
		row := db.QueryRowContext(context.Background(), q, 1001)

		// レコードとして取得したデータを各変数に格納する。
		err = row.Scan(&id, &balance, &name, &address, &phone)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}

		got := core.Account{
			Customer: core.Customer{
				Name:    name,
				Address: address,
				Phone:   phone,
			},
			Number:  id,
			Balance: balance,
		}

		want := core.Account{
			Customer: core.Customer{
				Name:    "John",
				Address: "Los Angeles, California",
				Phone:   "(213) 444 0147",
			},
			Number:  1001,
			Balance: 100,
		}

		assert.DeepEqual(t, got, want)

		core.DeleteTestData()
	})

	// INSERT句のテスト
	// tx.ExecContext()の学習用テスト
	t.Run("INSERT", func(t *testing.T) {
		// INSERT、UPDATE、DELETEを実行する前に、db.Begin()を実行してトランザクションを開始する。
		tx, err := db.Begin()
		if err != nil {
			t.Errorf("failed to db.Begin(): %v", err)
			return
		}
		// トランザクション中に何かしらのエラーが発生してプログラムが止まる際に備えるため、tx.Rollback)()をdeferによって予約して、実行したクエリを取り消す。
		defer tx.Rollback()

		want := core.Customer{
			Name:    "Mary",
			Address: "Netherland",
			Phone:   "8080",
		}
		num := 200

		// customerのテーブルには暗黙的に
		insertQuery := `
		INSERT INTO customer (id, username, addr, phone) 
		VALUES ($1, $2, $3, $4);
		`
		_, err = tx.ExecContext(context.Background(), insertQuery, num, want.Name, want.Address, want.Phone)
		if err != nil {
			t.Errorf("failed to tx.ExecuteContext(): %v", err)
			return
		} else {
			// tx.Commit()によって初めてDBに変更が反映される
			tx.Commit()
		}

		var (
			id      int
			name    string
			address string
			phone   string
		)

		readQuery := "SELECT * FROM customer WHERE id=$1"
		row := db.QueryRowContext(context.Background(), readQuery, num)
		err = row.Scan(&id, &name, &address, &phone)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		got := core.Customer{
			Name:    name,
			Address: address,
			Phone:   phone,
		}

		assert.DeepEqual(t, got, want)

		core.DeleteTestData()
	})

	// UPDATE句のテスト
	t.Run("UPDATE", func(t *testing.T) {
		err := core.InsertTestData()
		if err != nil {
			t.Errorf("failed to setup tables: %v", err)
		}
		defer core.DeleteTestData()

		tx, err := db.Begin()
		if err != nil {
			t.Errorf("failed to db.Begin(): %v", err)
			return
		}
		defer tx.Rollback()

		want := core.Customer{
			Name:    "Mary",
			Address: "South Africa",
			Phone:   "2020",
		}
		num := 1001

		q := `
		UPDATE customer
		SET username=$1, addr=$2, phone=$3
		WHERE id=$4;
		`
		_, err = tx.ExecContext(context.Background(), q, want.Name, want.Address, want.Phone, num)
		if err != nil {
			t.Errorf("failed to tx.ExecuteContext(): %v", err)
			return
		} else {
			tx.Commit()
		}

		var (
			id      int
			name    string
			address string
			phone   string
		)

		readQuery := "SELECT * FROM customer WHERE id=$1"
		row := db.QueryRowContext(context.Background(), readQuery, num)
		err = row.Scan(&id, &name, &address, &phone)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		got := core.Customer{
			Name:    name,
			Address: address,
			Phone:   phone,
		}

		assert.DeepEqual(t, got, want)
	})

	t.Run("DELETE", func(t *testing.T) {
		err := core.InsertTestData()
		if err != nil {
			t.Errorf("failed to setup tables: %v", err)
		}
		defer core.DeleteTestData()

		tx, err := db.Begin()
		if err != nil {
			t.Errorf("failed to db.Begin(): %v", err)
			return
		}
		defer tx.Rollback()

		num := 1001

		// check update statement
		q := "delete FROM account WHERE id=$1;"
		_, err = tx.ExecContext(context.Background(), q, num)
		if err != nil {
			t.Errorf("failed to tx.ExecuteContext(): %v", err)
			return
		} else {
			tx.Commit()
		}

		var (
			id      int
			balance float64
		)

		q = "SELECT * FROM account WHERE id=$1"
		row := db.QueryRowContext(context.Background(), q, num)
		err = row.Scan(&id, &balance)
		if errors.Is(err, sql.ErrNoRows) != true {
			t.Errorf("Failed to delete the recode: %v", err)
		}
	})
}

func TestDistinct(t *testing.T) {
	var (
		id   int
		name string
	)

	t.Run("DISTINCT", func(t *testing.T) {
		err := core.CreateTestTableWithDuplicateValue()
		if err != nil {
			t.Errorf("failed to setup tables: %v", err)
		}
		defer core.DropDuplicateTables()

		// DISTINCT カラム名1, カラム名２,… で選択した列の値の組み合わせの重複を削除したレコードを返す。
		// この場合だとright_tableの中の(1, 'a')のレコードのみが削除される。
		q := `
		SELECT DISTINCT id, name
		FROM duplicate_table
		ORDER BY id, name;
		`
		rows, err := db.QueryContext(context.Background(), q)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		defer rows.Close()

		got := []rightTable{}

		for rows.Next() {
			err := rows.Scan(&id, &name)
			if err != nil {
				t.Errorf("%v", err)
			}

			rt := rightTable{
				Id:   id,
				Name: name,
			}

			got = append(got, rt)
		}

		want := []rightTable{
			{
				Id:   1,
				Name: "a",
			},
			{
				Id:   2,
				Name: "b",
			},
			{
				Id:   2,
				Name: "c",
			},
			{
				Id:   3,
				Name: "c",
			},
			{
				Id:   3,
				Name: "d",
			},
			{
				Id:   4,
				Name: "a",
			},
		}

		assert.DeepEqual(t, want, got)
	})

	t.Run("DISTINCT_ON", func(t *testing.T) {
		err := core.CreateTestTableWithDuplicateValue()
		if err != nil {
			t.Errorf("failed to setup tables: %v", err)
		}
		defer core.DropDuplicateTables()

		// DISTINCT ON (カラム名) カラム名… で重複削除を行う列と重複削除を行わない列を指定する。
		// この場合だとright_tableの中の (3, 'd'), (2, 'c'), (1, 'a') が削除される。
		q := `
		SELECT DISTINCT ON (id) id, name
		FROM duplicate_table
		ORDER BY id, name;
		`
		rows, err := db.QueryContext(context.Background(), q)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		defer rows.Close()

		got := []rightTable{}

		for rows.Next() {
			err := rows.Scan(&id, &name)
			if err != nil {
				t.Errorf("%v", err)
			}

			rt := rightTable{
				Id:   id,
				Name: name,
			}

			got = append(got, rt)
		}

		want := []rightTable{
			{
				Id:   1,
				Name: "a",
			},
			{
				Id:   2,
				Name: "b",
			},
			{
				Id:   3,
				Name: "c",
			},
			{
				Id:   4,
				Name: "a",
			},
		}

		assert.DeepEqual(t, want, got)
	})
}

func TestJOIN(t *testing.T) {
	type conbined struct {
		LeftId    sql.NullInt64
		LeftName  sql.NullString
		RightId   sql.NullInt64
		RightName sql.NullString
	}

	var (
		left_id    sql.NullInt64
		left_name  sql.NullString
		right_id   sql.NullInt64
		right_name sql.NullString
	)

	t.Run("LEFT OUTER JOIN", func(t *testing.T) {
		err := core.CreateTestTable()
		if err != nil {
			t.Errorf("failed to setup tables: %v", err)
		}
		defer core.DropTables()

		var (
			left_id    sql.NullInt64
			left_name  sql.NullString
			right_id   sql.NullInt64
			right_name sql.NullString
		)

		q := `
		SELECT * 
		FROM left_table
		LEFT OUTER JOIN right_table
		ON left_table.id=right_table.id
		ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;
		`
		rows, err := db.QueryContext(context.Background(), q)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		defer rows.Close()

		got := []conbined{}

		for rows.Next() {
			err := rows.Scan(&left_id, &left_name, &right_id, &right_name)
			if err != nil {
				t.Errorf("%v", err)
			}

			c := conbined{
				LeftId:    left_id,
				LeftName:  left_name,
				RightId:   right_id,
				RightName: right_name,
			}

			got = append(got, c)
		}

		want := make([]conbined, 3)
		want[0].LeftId = sql.NullInt64{
			Int64: 1,
			Valid: true,
		}
		want[0].LeftName = sql.NullString{
			String: "d",
			Valid:  true,
		}
		want[0].RightId = sql.NullInt64{
			Int64: 1,
			Valid: true,
		}
		want[0].RightName = sql.NullString{
			String: "a",
			Valid:  true,
		}
		want[1].LeftId = sql.NullInt64{
			Int64: 2,
			Valid: true,
		}
		want[1].LeftName = sql.NullString{
			String: "e",
			Valid:  true,
		}
		want[1].RightId = sql.NullInt64{
			Int64: 2,
			Valid: true,
		}
		want[1].RightName = sql.NullString{
			String: "b",
			Valid:  true,
		}
		want[2].LeftId = sql.NullInt64{
			Int64: 6,
			Valid: true,
		}
		want[2].LeftName = sql.NullString{
			String: "f",
			Valid:  true,
		}
		want[2].RightId = sql.NullInt64{
			Int64: 0,
			Valid: false,
		}
		want[2].RightName = sql.NullString{
			String: "",
			Valid:  false,
		}

		assert.DeepEqual(t, want, got)
	})

	t.Run("RIGHT OUTER JOIN", func(t *testing.T) {
		err := core.CreateTestTable()
		if err != nil {
			t.Errorf("failed to setup tables: %v", err)
		}
		defer core.DropTables()

		q := `
		SELECT * 
		FROM left_table
		RIGHT OUTER JOIN right_table
		ON left_table.id=right_table.id
		ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;
		`
		rows, err := db.QueryContext(context.Background(), q)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		defer rows.Close()

		got := []conbined{}

		for rows.Next() {
			err := rows.Scan(&left_id, &left_name, &right_id, &right_name)
			if err != nil {
				t.Errorf("%v", err)
			}

			c := conbined{
				LeftId:    left_id,
				LeftName:  left_name,
				RightId:   right_id,
				RightName: right_name,
			}

			got = append(got, c)
		}

		want := make([]conbined, 3)
		want[0].LeftId = sql.NullInt64{
			Int64: 1,
			Valid: true,
		}
		want[0].LeftName = sql.NullString{
			String: "d",
			Valid:  true,
		}
		want[0].RightId = sql.NullInt64{
			Int64: 1,
			Valid: true,
		}
		want[0].RightName = sql.NullString{
			String: "a",
			Valid:  true,
		}
		want[1].LeftId = sql.NullInt64{
			Int64: 2,
			Valid: true,
		}
		want[1].LeftName = sql.NullString{
			String: "e",
			Valid:  true,
		}
		want[1].RightId = sql.NullInt64{
			Int64: 2,
			Valid: true,
		}
		want[1].RightName = sql.NullString{
			String: "b",
			Valid:  true,
		}
		want[2].LeftId = sql.NullInt64{
			Int64: 0,
			Valid: false,
		}
		want[2].LeftName = sql.NullString{
			String: "",
			Valid:  false,
		}
		want[2].RightId = sql.NullInt64{
			Int64: 3,
			Valid: true,
		}
		want[2].RightName = sql.NullString{
			String: "c",
			Valid:  true,
		}

		assert.DeepEqual(t, want, got)
	})

	t.Run("FULL OUTER JOIN", func(t *testing.T) {
		err := core.CreateTestTable()
		if err != nil {
			t.Errorf("failed to setup tables: %v", err)
		}
		defer core.DropTables()

		q := `
		SELECT * 
		FROM left_table
		FULL OUTER JOIN right_table
		ON left_table.id=right_table.id
		ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;
		`
		rows, err := db.QueryContext(context.Background(), q)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		defer rows.Close()

		got := []conbined{}
		var (
			left_id    sql.NullInt64
			left_name  sql.NullString
			right_id   sql.NullInt64
			right_name sql.NullString
		)

		for rows.Next() {
			err := rows.Scan(&left_id, &left_name, &right_id, &right_name)
			if err != nil {
				t.Errorf("%v", err)
			}

			c := conbined{
				LeftId:    left_id,
				LeftName:  left_name,
				RightId:   right_id,
				RightName: right_name,
			}

			got = append(got, c)
		}

		want := make([]conbined, 4)
		want[0].LeftId = sql.NullInt64{
			Int64: 1,
			Valid: true,
		}
		want[0].LeftName = sql.NullString{
			String: "d",
			Valid:  true,
		}
		want[0].RightId = sql.NullInt64{
			Int64: 1,
			Valid: true,
		}
		want[0].RightName = sql.NullString{
			String: "a",
			Valid:  true,
		}
		want[1].LeftId = sql.NullInt64{
			Int64: 2,
			Valid: true,
		}
		want[1].LeftName = sql.NullString{
			String: "e",
			Valid:  true,
		}
		want[1].RightId = sql.NullInt64{
			Int64: 2,
			Valid: true,
		}
		want[1].RightName = sql.NullString{
			String: "b",
			Valid:  true,
		}
		want[2].LeftId = sql.NullInt64{
			Int64: 6,
			Valid: true,
		}
		want[2].LeftName = sql.NullString{
			String: "f",
			Valid:  true,
		}
		want[2].RightId = sql.NullInt64{
			Int64: 0,
			Valid: false,
		}
		want[2].RightName = sql.NullString{
			String: "",
			Valid:  false,
		}
		want[3].LeftId = sql.NullInt64{
			Int64: 0,
			Valid: false,
		}
		want[3].LeftName = sql.NullString{
			String: "",
			Valid:  false,
		}
		want[3].RightId = sql.NullInt64{
			Int64: 3,
			Valid: true,
		}
		want[3].RightName = sql.NullString{
			String: "c",
			Valid:  true,
		}

		assert.DeepEqual(t, want, got)
	})

	t.Run("CROSS JOIN", func(t *testing.T) {})

	t.Run("USING", func(t *testing.T) {
		err := core.CreateTestTable()
		if err != nil {
			t.Errorf("failed to setup tables: %v", err)
		}
		defer core.DropTables()

		var (
			left_id    sql.NullInt64
			left_name  sql.NullString
			right_name sql.NullString
		)

		// 結合する二つのテーブルに同じ名前のカラムがある場合、「USING (カラム名)」を使用することで「ON テーブル１.カラム名=テーブル２.カラム名」と同じことができる。
		// また、出力される列に関して、USING()内で指定した列の重複は取り除かれる。
		q := `
		SELECT * 
		FROM left_table
		LEFT OUTER JOIN right_table
		USING (id)
		ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;
		`
		rows, err := db.QueryContext(context.Background(), q)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		defer rows.Close()

		got := []conbined{}

		for rows.Next() {
			err := rows.Scan(&left_id, &left_name, &right_name)
			if err != nil {
				t.Errorf("%v", err)
			}

			c := conbined{
				LeftId:    left_id,
				LeftName:  left_name,
				RightName: right_name,
			}

			got = append(got, c)
		}

		want := make([]conbined, 3)
		want[0].LeftId = sql.NullInt64{
			Int64: 1,
			Valid: true,
		}
		want[0].LeftName = sql.NullString{
			String: "d",
			Valid:  true,
		}
		want[0].RightName = sql.NullString{
			String: "a",
			Valid:  true,
		}
		want[1].LeftId = sql.NullInt64{
			Int64: 2,
			Valid: true,
		}
		want[1].LeftName = sql.NullString{
			String: "e",
			Valid:  true,
		}
		want[1].RightName = sql.NullString{
			String: "b",
			Valid:  true,
		}
		want[2].LeftId = sql.NullInt64{
			Int64: 6,
			Valid: true,
		}
		want[2].LeftName = sql.NullString{
			String: "f",
			Valid:  true,
		}
		want[2].RightName = sql.NullString{
			String: "",
			Valid:  false,
		}

		assert.DeepEqual(t, want, got)
	})

	t.Run("USING_ALL_ROWS", func(t *testing.T) {
		err := core.CreateTestTable()
		if err != nil {
			t.Errorf("failed to setup tables: %v", err)
		}
		defer core.DropTables()

		var (
			left_id    sql.NullInt64
			left_name  sql.NullString
			right_name sql.NullString
		)

		// USING()の中に複数列を指定すると、列のセットのレコードと一致する他のテーブルレコードと結合を行う。
		// 今回の場合はleft_tableのid列とname列のペアと一致するright_tableを探して結合させる。つまり「ON left_table.id=right_table.id AND left_table.name=right_table.name」と同じ。
		// left_tableのidとnameのペアと一致するペアはright_tableには存在しないので、left_tableのレコードのみが返ってくる。
		q := `
		SELECT * 
		FROM left_table
		LEFT OUTER JOIN right_table
		USING (id, name)
		ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;
		`
		rows, err := db.QueryContext(context.Background(), q)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		defer rows.Close()

		got := []conbined{}

		for rows.Next() {
			err := rows.Scan(&left_id, &left_name)
			if err != nil {
				t.Errorf("%v", err)
			}

			c := conbined{
				LeftId:    left_id,
				LeftName:  left_name,
				RightName: right_name,
			}

			got = append(got, c)
		}

		want := make([]conbined, 3)
		want[0].LeftId = sql.NullInt64{
			Int64: 1,
			Valid: true,
		}
		want[0].LeftName = sql.NullString{
			String: "d",
			Valid:  true,
		}
		want[1].LeftId = sql.NullInt64{
			Int64: 2,
			Valid: true,
		}
		want[1].LeftName = sql.NullString{
			String: "e",
			Valid:  true,
		}
		want[2].LeftId = sql.NullInt64{
			Int64: 6,
			Valid: true,
		}
		want[2].LeftName = sql.NullString{
			String: "f",
			Valid:  true,
		}

		assert.DeepEqual(t, want, got)
	})

	t.Run("NATURAL JOIN", func(t *testing.T) {
		err := core.CreateTestTable()
		if err != nil {
			t.Errorf("failed to setup tables: %v", err)
		}
		defer core.DropTables()

		var (
			left_id    sql.NullInt64
			left_name  sql.NullString
			right_name sql.NullString
		)

		// NATURAL ~ JOIN は USING(全カラム)と同じ挙動を行う。
		q := `
		SELECT * 
		FROM left_table
		NATURAL LEFT OUTER JOIN right_table
		ORDER BY left_table.id, left_table.name, right_table.id, right_table.name;
		`
		rows, err := db.QueryContext(context.Background(), q)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		defer rows.Close()

		got := []conbined{}

		for rows.Next() {
			err := rows.Scan(&left_id, &left_name)
			if err != nil {
				t.Errorf("%v", err)
			}

			c := conbined{
				LeftId:    left_id,
				LeftName:  left_name,
				RightName: right_name,
			}

			got = append(got, c)
		}

		want := make([]conbined, 3)
		want[0].LeftId = sql.NullInt64{
			Int64: 1,
			Valid: true,
		}
		want[0].LeftName = sql.NullString{
			String: "d",
			Valid:  true,
		}
		want[1].LeftId = sql.NullInt64{
			Int64: 2,
			Valid: true,
		}
		want[1].LeftName = sql.NullString{
			String: "e",
			Valid:  true,
		}
		want[2].LeftId = sql.NullInt64{
			Int64: 6,
			Valid: true,
		}
		want[2].LeftName = sql.NullString{
			String: "f",
			Valid:  true,
		}

		assert.DeepEqual(t, want, got)
	})
}

func TestLIMIT(t *testing.T) {
	err := core.CreateTestTableWithDuplicateValue()
	if err != nil {
		t.Errorf("failed to setup tables: %v", err)
	}
	defer core.DropDuplicateTables()

	var (
		id   int
		name string
	)

	t.Run("LIMIT", func(t *testing.T) {
		// LIMIT句を使用することで、取得したレコードの先頭~行だけを抽出できる
		q := `
		SELECT id, name
		FROM duplicate_table
		ORDER BY id, name
		LIMIT 3;
		`
		rows, err := db.QueryContext(context.Background(), q)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		defer rows.Close()

		got := []rightTable{}

		for rows.Next() {
			err := rows.Scan(&id, &name)
			if err != nil {
				t.Errorf("%v", err)
			}

			rt := rightTable{
				Id:   id,
				Name: name,
			}

			got = append(got, rt)
		}

		want := []rightTable{
			{
				Id:   1,
				Name: "a",
			},
			{
				Id:   1,
				Name: "a",
			},
			{
				Id:   2,
				Name: "b",
			},
		}

		assert.DeepEqual(t, want, got)
	})

	t.Run("OFFSET", func(t *testing.T) {
		// OFFSET句を使用することで、LIMITで抽出を始めるレコードの先頭行数を指定する。
		q := `
		SELECT id, name
		FROM duplicate_table
		ORDER BY id, name
		LIMIT 3 OFFSET 2;
		`
		rows, err := db.QueryContext(context.Background(), q)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		defer rows.Close()

		got := []rightTable{}

		for rows.Next() {
			err := rows.Scan(&id, &name)
			if err != nil {
				t.Errorf("%v", err)
			}

			rt := rightTable{
				Id:   id,
				Name: name,
			}

			got = append(got, rt)
		}

		want := []rightTable{
			{
				Id:   2,
				Name: "b",
			},
			{
				Id:   2,
				Name: "c",
			},
			{
				Id:   3,
				Name: "c",
			},
		}

		assert.DeepEqual(t, want, got)
	})
}

func TestORDERBY(t *testing.T) {
	err := core.CreateTestTableWithDuplicateValue()
	if err != nil {
		t.Errorf("failed to setup tables: %v", err)
	}
	defer core.DropDuplicateTables()

	var (
		id   int
		name string
	)

	t.Run("ASC", func(t *testing.T) {
		// ORDER BY によって取得するレコードの順番を並び替えることができる。
		// ASCを指定すると、並び替えの対象にした列を昇順で取得する。
		// SELECT句で選択して、かつORDER BY で選択されなかったものに関しては降順で並び替えが行われる。
		// デフォルトは昇順で並び替えを行うので、ASCは不要。
		q := `
		SELECT id, name
		FROM duplicate_table
		ORDER BY id ASC;
		`
		rows, err := db.QueryContext(context.Background(), q)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		defer rows.Close()

		got := []rightTable{}

		for rows.Next() {
			err := rows.Scan(&id, &name)
			if err != nil {
				t.Errorf("%v", err)
			}

			rt := rightTable{
				Id:   id,
				Name: name,
			}

			got = append(got, rt)
		}

		want := []rightTable{
			{
				Id:   1,
				Name: "a",
			},
			{
				Id:   1,
				Name: "a",
			},
			{
				Id:   2,
				Name: "c",
			},
			{
				Id:   2,
				Name: "b",
			},
			{
				Id:   3,
				Name: "d",
			},
			{
				Id:   3,
				Name: "c",
			},
			{
				Id:   4,
				Name: "a",
			},
		}

		assert.DeepEqual(t, want, got)
	})

	t.Run("DESC", func(t *testing.T) {
		// DESCを指定すると、並び替えの対象にした列を昇順で取得する。
		q := `
		SELECT id, name
		FROM duplicate_table
		ORDER BY id DESC;
		`
		rows, err := db.QueryContext(context.Background(), q)
		if err != nil {
			t.Errorf("rows.Scan() is failed: %v", err)
		}
		defer rows.Close()

		got := []rightTable{}

		for rows.Next() {
			err := rows.Scan(&id, &name)
			if err != nil {
				t.Errorf("%v", err)
			}

			rt := rightTable{
				Id:   id,
				Name: name,
			}

			got = append(got, rt)
		}

		want := []rightTable{
			{
				Id:   4,
				Name: "a",
			},
			{
				Id:   3,
				Name: "c",
			},
			{
				Id:   3,
				Name: "d",
			},
			{
				Id:   2,
				Name: "c",
			},
			{
				Id:   2,
				Name: "b",
			},
			{
				Id:   1,
				Name: "a",
			},
			{
				Id:   1,
				Name: "a",
			},
		}

		assert.DeepEqual(t, want, got)
	})
}

func TestQueryConbination(t *testing.T) {

	t.Run("UNION", func(t *testing.T) {})

	t.Run("INTERSECT", func(t *testing.T) {})

	t.Run("EXCEPT", func(t *testing.T) {})
}

func TestTableFunction(t *testing.T) {

	t.Run("UNNEST", func(t *testing.T) {})

	t.Run("LATERAL", func(t *testing.T) {})
}

func TestWHERE(t *testing.T) {

	t.Run("HAVING", func(t *testing.T) {})

	t.Run("GEOUP BY", func(t *testing.T) {})

	t.Run("GEOUPING SETS", func(t *testing.T) {})

	t.Run("CUBE", func(t *testing.T) {})

	t.Run("ROLLUP", func(t *testing.T) {})
}
