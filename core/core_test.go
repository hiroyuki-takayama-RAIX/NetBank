package core

// without import bank.go, you can use objects ans functions because core_test.go and bank.go are in the same module.
import (
	"testing"
)

// to state account uotside TestMain(), account can be accessed every test functions.
var account Account

/*func TestMain(m *testing.M) {
	setup()
	exitCode := m.Run()
	teardown()
	account.Balance = 0
	os.Exit(exitCode)
}*/

func setupAccount(t *testing.T) Account {
	// set test parallelly
	t.Parallel()

	// makeig test data
	account = Account{
		Customer: Customer{
			Name:    "John",
			Address: "Los Angeles, California",
			Phone:   "(213) 555 0147",
		},
		Number:  1001,
		Balance: 0,
	}

	return account
}

func setupSenderAndReciever(t *testing.T) (sender Account, reciever Account) {
	// set test parallelly
	t.Parallel()

	sender = Account{
		Customer: Customer{
			Name:    "John",
			Address: "Los Angeles, California",
			Phone:   "(213) 555 0147",
		},
		Number:  1001,
		Balance: 100,
	}

	reciever = Account{
		Customer: Customer{
			Name:    "C.J.",
			Address: "Los Santos, San And Reas",
			Phone:   "(080) 1457 9387",
		},
		Number:  2002,
		Balance: 100,
	}

	return sender, reciever
}

//func teardown() {}

func TestAccount(t *testing.T) {
	account = setupAccount(t)
	// golang uses if statent and deifine a failure condition there instead of assert statment.
	if account.Name == "" {
		t.Error("can't create an Account object")
	}
}

func TestDeposit(t *testing.T) {
	account = setupAccount(t)

	if err := account.Deposit(10); err != nil {
		t.Error(err)
	}

	if account.Balance != 10 {
		t.Error("the calculation of withdraw get something wrong! balance :", account.Balance)
	}
}

// Invalid pattern test
func TestDepositInvalid(t *testing.T) {
	account = setupAccount(t)
	if err := account.Deposit(-10); err == nil {
		t.Error(err)
	}
}

func TestWithdraw(t *testing.T) {
	account = setupAccount(t)

	account.Deposit(20)

	if err := account.Withdraw(10); err != nil {
		t.Error(err)
	}

	if account.Balance != 10 {
		t.Error("the calculation of withdraw get something wrong! balance :", account.Balance)
	}
}

func TestTransfer(t *testing.T) {
	sender, reciever := setupSenderAndReciever(t)

	sender.Transfer(&reciever, 50)

	// sender.Balance == 50 && reciever.Balance == 150 is more simple!
	senderStatement := sender.Statement()
	if senderStatement != "1001 - John - 50" {
		t.Error("statement doesn't have the proper format! statement :", senderStatement)
	}

	recieverStatement := reciever.Statement()
	if recieverStatement != "2002 - C.J. - 150" {
		t.Error("statement doesn't have the proper format! statement :", recieverStatement)
	}
}

func TestStatement(t *testing.T) {
	account := Account{
		Customer: Customer{
			Name:    "John",
			Address: "Los Angeles, California",
			Phone:   "(213) 555 0147",
		},
		Number:  1001,
		Balance: 0,
	}

	account.Deposit(100)
	statement := account.Statement()
	if statement != "1001 - John - 100" {
		t.Error("statement doesn't have the proper format")
	}
}

/*
"{\"Name\":\"John\",\"Address\":\"Los Angeles, California\",\"Phone\":\"(213) 555 0147\",\"Number\":1001,\"Balance\":0}"
*/
