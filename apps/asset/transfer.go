package asset

import (
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/common/yerror"
	. "github.com/yu-org/yu/core/context"
	. "github.com/yu-org/yu/core/tripod"
	"math/big"
	"net/http"
)

type Asset struct {
	*Tripod
	TokenName string
}

func NewAsset(tokenName string) *Asset {
	df := NewTripod()

	a := &Asset{df, tokenName}
	a.SetWritings(a.Transfer, a.CreateAccount)
	a.SetReadings(a.QueryBalance)

	//a.SetTxnChecker(func(txn *SignedTxn) error {
	//	if txn.Raw.WrCall.LeiPrice == 0 {
	//		return nil
	//	}
	//
	//	if !a.ExistAccount(txn.Raw.Caller) {
	//		return AccountNotFound(txn.Raw.Caller)
	//	}
	//
	//	balance := a.GetBalance(txn.Raw.Caller)
	//	leiPrice := new(big.Int).SetUint64(txn.Raw.WrCall.LeiPrice)
	//	if balance.Cmp(leiPrice) < 0 {
	//		return InsufficientFunds
	//	}
	//
	//	validatorsCount := len(validators)
	//	if validatorsCount > 0 {
	//		validatorsCountBigInt := new(big.Int).SetInt64(int64(validatorsCount))
	//		rewards := new(big.Int).Div(leiPrice, validatorsCountBigInt)
	//		for _, validator := range validators {
	//			err := a.transfer(txn.Raw.Caller, validator.Address(), rewards)
	//			if err != nil {
	//				return err
	//			}
	//		}
	//	}
	//
	//	return nil
	//})

	return a
}

type AccountRequest struct {
	Account string `json:"account"`
}

func (a *Asset) QueryBalance(ctx *ReadContext) {
	var req AccountRequest
	err := ctx.BindJson(&req)
	if err != nil {
		ctx.Err(http.StatusBadRequest, err)
		return
	}
	account := HexToAddress(req.Account)

	if !a.ExistAccount(&account) {
		ctx.ErrOk(AccountNotFound(account))
		return
	}
	amount := a.GetBalance(&account)
	ctx.JsonOk(map[string]*big.Int{"amount": amount})
}

func (a *Asset) Transfer(ctx *WriteContext) (err error) {
	ctx.SetLei(1)
	from := ctx.GetCaller()
	to := ctx.GetAddress("to")
	amount := big.NewInt(int64(ctx.GetUint64("amount")))

	logrus.WithField("asset", "transfer").
		Debugf("from(%s) to(%s) amount(%d)", from.String(), to.String(), amount)
	err = a.transfer(from, to, amount)
	if err != nil {
		return
	}
	ctx.EmitStringEvent("Transfer Completed!")
	return
}

func (a *Asset) transfer(from, to *Address, amount *big.Int) error {
	if !a.ExistAccount(from) {
		return AccountNotFound(*from)
	}

	fromBalance := a.GetBalance(from)
	if fromBalance.Cmp(amount) < 0 {
		return InsufficientFunds
	}

	if !a.ExistAccount(to) {
		a.SetBalance(to, amount)
	} else {
		toBalance := a.GetBalance(to)
		toAdd := new(big.Int).Add(toBalance, amount)
		a.SetBalance(to, toAdd)
	}

	fromSub := new(big.Int).Sub(fromBalance, amount)
	a.SetBalance(from, fromSub)
	return nil
}

func (a *Asset) CreateAccount(ctx *WriteContext) error {
	ctx.SetLei(10)
	addr := ctx.GetCaller()
	//if !a.isValidator(addr) {
	//	return NoPermission
	//}
	amount := big.NewInt(int64(ctx.GetUint64("amount")))

	logrus.WithField("asset", "create-account").Debugf("ACCOUNT(%s) amount(%d)", addr.String(), amount)

	if a.ExistAccount(addr) {
		ctx.EmitStringEvent("Account Exists!")
		return nil
	}

	a.SetBalance(addr, amount)
	ctx.EmitStringEvent("Account Created Success!")
	return nil
}

func (a *Asset) ExistAccount(addr *Address) bool {
	return a.Exist(addr.Bytes())
}

func (a *Asset) GetBalance(addr *Address) *big.Int {
	balanceByt, err := a.State.Get(a, addr.Bytes())
	if err != nil {
		logrus.Panic("get balance error: ", err)
	}

	if balanceByt == nil {
		return big.NewInt(0)
	}

	b := new(big.Int)
	err = b.UnmarshalText(balanceByt)
	if err != nil {
		logrus.Panic("GetBalance marshal error: ", err)
	}
	return b
}

func (a *Asset) SetBalance(addr *Address, amount *big.Int) {
	amountText, err := amount.MarshalText()
	if err != nil {
		logrus.Panic("amount marshal error: ", err)
	}

	a.Set(addr.Bytes(), amountText)
}

func (a *Asset) AddBalance(addr *Address, amount *big.Int) error {
	if amount.Sign() < 0 {
		return AmountNeg(amount)
	}
	balance := a.GetBalance(addr)
	balance.Add(balance, amount)
	a.SetBalance(addr, balance)
	return nil
}

func (a *Asset) SubBalance(addr *Address, amount *big.Int) error {
	if amount.Sign() < 0 {
		return AmountNeg(amount)
	}
	balance := a.GetBalance(addr)
	balance.Sub(balance, amount)
	a.SetBalance(addr, balance)
	return nil
}

//func (a *Asset) isValidator(addr Address) bool {
//	for _, validator := range a.validators {
//		if validator.Address() == addr {
//			return true
//		}
//	}
//	return false
//}
