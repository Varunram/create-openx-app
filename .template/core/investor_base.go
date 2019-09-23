package core

import (
	//"log"
	"github.com/pkg/errors"

	utils "github.com/Varunram/essentials/utils"
	tickers "github.com/YaleOpenLab/openx/chains/exchangetickers"
	xlm "github.com/YaleOpenLab/openx/chains/xlm"
	openxconsts "github.com/YaleOpenLab/openx/consts"
	openx "github.com/YaleOpenLab/openx/database"

	consts "github.com/org/plat/consts"
)

// Investor defines the investor structure
type Investor struct {
	// U is the base User class inherited from openx
	U *openx.User

	// VotingBalance is the balance associated with the particular investor (equal to the amount of USD he possesses)
	VotingBalance float64

	// AmountInvested is the total amount invested by the investor
	AmountInvested float64

	// InvestedProjects is a list of the investor assets of the opensolar projects the investor has invested in
	InvestedProjects []string

	// InvestedProjectsIndices is an integer list of the projects the investor has invested in
	InvestedProjectsIndices []int

	// InvestedProjects is a list of the investor assets of the opensolar projects the investor has invested in
	SeedInvestedProjects []string

	// InvestedProjectsIndices is an integer list of the projects the investor has invested in
	SeedInvestedProjectsIndices []int
}

// NewInvestor creates a new investor based on params passed
func NewInvestor(uname string, pwd string, seedpwd string, Name string) (Investor, error) {
	var a Investor
	var err error
	user, err := NewUser(uname, utils.SHA3hash(pwd), seedpwd, Name)
	if err != nil {
		return a, errors.Wrap(err, "error while creating a new user")
	}
	a.U = &user
	a.AmountInvested = -1.0
	err = a.Save()
	return a, err
}

// CanInvest checks whether an investor has the required funds to invest in a project
func (a *Investor) CanInvest(targetBalance float64) bool {
	if !consts.Mainnet {
		// testnet
		usdBalance, err := xlm.GetAssetBalance(a.U.StellarWallet.PublicKey, "STABLEUSD")
		if err != nil {
			usdBalance = 0
		}

		xlmBalance, err := xlm.GetNativeBalance(a.U.StellarWallet.PublicKey)
		if err != nil {
			xlmBalance = 0
		}

		// need to fetch the oracle price here for the order
		oraclePrice := tickers.ExchangeXLMforUSD(xlmBalance)
		if usdBalance > targetBalance || oraclePrice > targetBalance {
			// return true since the user has enough USD balance to pay for the order
			return true
		}
		return false
	}

	// mainnet
	usdBalance, err := xlm.GetAssetBalance(a.U.StellarWallet.PublicKey, openxconsts.AnchorUSDCode)
	if err != nil {
		usdBalance = 0
	}

	return usdBalance > targetBalance
}
