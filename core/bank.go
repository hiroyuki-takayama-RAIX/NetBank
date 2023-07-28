package core

import (
	"errors"
	"fmt"
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

type jsonStatement interface {
	Statement() string
}

// custome error must be initialized by 'var ~ = errors.New("・・・")
var (
	BALANCE_LESS_THAN_ZERO_ERROR        = errors.New("A balance must be zore or more!")
	AMOUNT_LESS_THAN_ZERO_ERROR         = errors.New("Amount must be more than zero!")
	TRANSFER_GREATER_THAN_DEPOSIT_ERROR = errors.New("transfer is greater than deposit!")
	WITHDRAW_GREATER_THAN_DEPOSIT_ERROR = errors.New("Amount of withdraw must be more than deposit!")
)

// any function ans method should have error handling programm at least one
// Deposit ...
func (a *Account) Deposit(f float64) error {
	if f <= 0 {
		return AMOUNT_LESS_THAN_ZERO_ERROR
	}

	if 0 > a.Balance+f {
		return BALANCE_LESS_THAN_ZERO_ERROR
	} else {
		a.Balance += f
		return nil
	}
}

func (a *Account) Withdraw(f float64) error {
	if f <= 0 {
		return AMOUNT_LESS_THAN_ZERO_ERROR
	}

	if 0 > a.Balance-f {
		return WITHDRAW_GREATER_THAN_DEPOSIT_ERROR
	} else {
		a.Balance -= f
		return nil
	}
}

func (a *Account) Statement() string {
	// useing %f returns 100.0000…
	return fmt.Sprintf("%v - %s - %v", a.Number, a.Name, a.Balance)
}

func Statement(js jsonStatement) {
}

func (a *Account) Transfer(reciever *Account, f float64) error {
	//programms before transfer
	if f <= 0 {
		return AMOUNT_LESS_THAN_ZERO_ERROR
	} else if f > a.Balance {
		return TRANSFER_GREATER_THAN_DEPOSIT_ERROR
	}

	//programms after transfer
	if err := a.Withdraw(f); err != nil {
		return err
	} else {
		if err = reciever.Deposit(f); err != nil {
			return err
		} else {
			return nil
		}
	}
}
