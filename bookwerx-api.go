/*
Structs and functions used to access the bookwerx API should go here.  Although many of these things are unique to
a specific endpoint and perhaps ought to be defined in that code instead, we put them here anyway so that we can
better observe duplication and relevant abstractions as well as consistently handle error conditions.
*/
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	bwAPI "github.com/bostontrader/bookwerx-common-go"
	"github.com/gojektech/heimdall/httpclient"
	"strings"
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

/*
Given a slice of transaction ID, find all distributions that are related to any of the given transactions.  Return said ([Distribution], nil) or ([], some error).
*/
func accountDepositHistoryGetDistributionsByTx(transactionID []uint32, httpClient *httpclient.Client, cfg Config) ([]Distributions, error) {

	methodName := "okcatbox:bookwerx-api.go:accountDepositHistoryGetDistributionsByTx"

	// 1. The desired query varies according to the len(transactionID)
	var inClause string
	switch len(transactionID) {
	case 0:
		{
			return []Distributions{}, nil // no TXID, no distributions
		}

	default:
		{
			var sb strings.Builder
			sb.WriteRune('(')
			for k, v := range transactionID {
				if k == 0 {
					sb.WriteString(fmt.Sprintf("%d", v))
				} else {
					sb.WriteString(fmt.Sprintf(",%d", v))
				}
			}
			sb.WriteRune(')')
			inClause = sb.String()
		}
	}

	// 2. Build and execute the query.
	selectt := "SELECT%20currencies.symbol,%20distributions.amount,%20distributions.amount_exp,%20distributions.transaction_id,%20transactions.time"
	from := "FROM%20distributions"
	join1 := "JOIN%20accounts%20ON%20accounts.id%3ddistributions.account_id"
	join2 := "JOIN%20currencies%20ON%20currencies.id%3daccounts.currency_id"
	join3 := "JOIN%20transactions%20ON%20transactions.id%3ddistributions.transaction_id"
	join4 := "JOIN%20accounts_categories%20ON%20accounts_categories.account_id%3daccounts.id"
	where1 := fmt.Sprintf("WHERE%%20distributions.transaction_id%%20IN%%20%s", inClause)
	where2 := fmt.Sprintf("AND%%20accounts_categories.category_id%%3d%d", cfg.Bookwerx.CatHotWallet)

	query := fmt.Sprintf("%s%%20%s%%20%s%%20%s%%20%s%%20%s%%20%s%%20%s", selectt, from, join1, join2, join3, join4, where1, where2)
	url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)

	responseBody, err := bwAPI.Get(httpClient, url)

	if err != nil {
		if s, errb := squeal(methodName, "bwAPI.GET", err); errb {
			return []Distributions{}, errors.New(s)
		}
	}

	// 3. Decode the response.
	distributions := make([]Distributions, 0)
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	err = dec.Decode(&distributions)
	if err != nil {
		if s, errb := squealJSONDecode(methodName, responseBody, err); errb {
			return []Distributions{}, errors.New(s)
		}
	}

	return distributions, nil

}

/*
Given a currencySymbol, return the account ID of the account that is tagged as a hot wallet and is configured to use the given currency, or 0 if a suitable account is not found. Return said (accountID, nil) or (0, some error).
*/
func accountDepositHistoryGetHotWalletAccountID(currencySymbol string, httpClient *httpclient.Client, cfg Config) (accountID uint32, err error) {

	methodName := "okcatbox:bookwerx-api.go:accountDepositHistoryGetHotWalletAccountID"
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

/*
Given a slice of category ID, find all transactions that are tagged with all of the given categories.  Return said ([transaction ID], nil) or ([], some error).

The categoryID slice must be of length 0, 1 or 2.
*/
func accountDepositHistoryGetTransactionsByCat(categoryID []uint32, httpClient *httpclient.Client, cfg Config) ([]uint32, error) {

	methodName := "okcatbox:bookwerx-api.go:accountDepositHistoryGetTransactionsByCat"

	// 1. Validate the categoryID and build a suitable IN clause for the query.
	var inClause string
	switch len(categoryID) {
	case 0:
		{
			return []uint32{}, nil // no categoryID, no transactions
		}
	case 1:
		{
			inClause = fmt.Sprintf("(%d)", categoryID[0])
		}
	case 2:
		{
			inClause = fmt.Sprintf("(%d,%d)", categoryID[0], categoryID[1])
		}
	default:
		{
			if s, errb := squeal(methodName, fmt.Sprintf("%s: The len(categoryID) must be 0, 1 or 2 only. In this case the length=%d\n", methodName, len(categoryID)), errors.New("")); errb {
				return []uint32{}, errors.New(s)
			}
		}
	}

	// 2. Build and execute the query.
	selectt := "SELECT%20transactions_categories.transaction_id"
	from := "FROM%20transactions_categories"
	where := fmt.Sprintf("WHERE%%20transactions_categories.category_id%%20IN%%20%s", inClause)
	group := fmt.Sprintf("GROUP%%20BY%%20transactions_categories.transaction_id")

	query := fmt.Sprintf("%s%%20%s%%20%s%%20%s", selectt, from, where, group)
	url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)

	responseBody, err := bwAPI.Get(httpClient, url)
	if err != nil {
		if s, errb := squeal(methodName, "bwAPI.GET", err); errb {
			return []uint32{}, errors.New(s)
		}
	}
	//fixDot(responseBody)

	// 3. Decode the response.
	tid := make([]TransactionsID, 0)
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	err = dec.Decode(&tid)
	if err != nil {
		if s, errb := squealJSONDecode(methodName, responseBody, err); errb {
			return []uint32{}, errors.New(s)
		}
	}

	// 4. Map []TransactionsID -> []uint32
	retVal := make([]uint32, 0)
	for _, v := range tid {
		retVal = append(retVal, v.ID)
	}

	return retVal, nil

}
