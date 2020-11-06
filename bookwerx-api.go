package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	bwAPI "github.com/bostontrader/bookwerx-common-go"
	"github.com/gojektech/heimdall/httpclient"
	//log "github.com/sirupsen/logrus"
	//"net/http"
)

type AccountCurrency struct {
	AccountID uint32
	Title     string
	Currency  CurrencySymbol
}

type BalanceResultDecorated struct {
	Account AccountCurrency
	Sum     DFP
}

type CurrencySymbol struct {
	CurrencyID uint32
	Symbol     string
}

type Sums struct {
	Sums []BalanceResultDecorated
}

func getCategoryDistSums(httpClient *httpclient.Client, url string) (retVal Sums, err error) {

	methodName := "okcatbox:bookwerx-api.go:getCategoryDistSums"
	//req, err := http.NewRequest("GET", url, nil)
	//if err != nil {
	//s := fmt.Sprintf("bookwerx-api.go:getCategoryDistSums 1: %v", err)
	//log.Error(s)
	//return Sums{}, nil
	//}
	//resp, err := client.Do(req)
	//defer resp.Body.Close()
	// 2. Submit the query
	responseBody, err := bwAPI.Get(httpClient, url)
	if s, errb := squeal(methodName, "bwAPI.GET", err); errb {
		return Sums{}, errors.New(s)
	}

	sums := Sums{}
	err = json.NewDecoder(bytes.NewReader(responseBody)).Decode(&sums)
	//if err != nil {
	//s := fmt.Sprintf("bookwerx-api.go:getCategoryDistSums 3: Error with JSON decoding.")
	//log.Error(s)
	//return Sums{}, nil
	//}
	if s, errb := squealJSONDecode(methodName, responseBody, err); errb {
		return Sums{}, errors.New(s)
	}

	return sums, nil
}

/*
Given a currencySymbol, return the account ID of the account that is tagged as a hot wallet and is configured to use the given currency, or 0 if a suitable account is not found. Return said (accountID, nil) or (0, some error).
*/
func getHotWalletAccountID(currencySymbol string, httpClient *httpclient.Client, cfg Config) (accountID uint32, err error) {

	methodName := "okcatbox:bookwerx-api.go:getHotWalletAccountID"
	accountID = 0

	// 1. Build the query
	// The desired account is:
	// A. Tagged with whatever category corresponds with hot wallet.
	// B. Configured to use the specified currency.
	selectt := "SELECT%20accounts.id"
	from := "FROM%20accounts_categories"
	join1 := "JOIN%20accounts%20ON%20accounts.id%3daccounts_categories.account_id"
	join2 := "JOIN%20currencies%20ON%20currencies.id%3daccounts.currency_id"
	where := fmt.Sprintf("WHERE%%20category_id%%3d%d%%20AND%%20currencies.symbol%%3d'%s'", cfg.Bookwerx.CatHotWallet, currencySymbol)
	query := fmt.Sprintf("%s%%20%s%%20%s%%20%s%%20%s", selectt, from, join1, join2, where)
	url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)

	// 2. Submit the query
	responseBody, err := bwAPI.Get(httpClient, url)
	if s, errb := squeal(methodName, "bwAPI.GET", err); errb {
		return accountID, errors.New(s)
	}

	fixDot(responseBody)

	// 3. Decode the response.
	accountsID := make([]AccountsID, 0)
	err = json.NewDecoder(bytes.NewReader(responseBody)).Decode(&accountsID)
	if s, errb := squealJSONDecode(methodName, responseBody, err); errb {
		return accountID, errors.New(s)
	}

	// 4. Figure out what exactly to send back.
	switch len(accountsID) {
	case 0:
		{
			errMsg := fmt.Sprintf("%s: Bookwerx does not have any account properly configured.\n", methodName)
			return 0, errors.New(errMsg)
		}
	case 1:
		{
			return accountsID[0].ID, nil
		} // found it!
	default:
		{
			errMsg := fmt.Sprintf("%s: Bookwerx has more than one suitable account.  This should never happen.\n", methodName)
			return 0, errors.New(errMsg)
		}
	}

}
