package main

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"github.com/gojektech/heimdall/httpclient"
	"net/http"
	"time"
)

// /api/account/v3/wallet
func accountWalletHandler(w http.ResponseWriter, req *http.Request, cfg Config) {
	errb := checkSigHeaders(w, req)
	if errb {
		return
	}
	retVal := generateWalletResponse(req, cfg)
	_, _ = fmt.Fprintf(w, string(retVal))
}

func generateWalletResponse(req *http.Request, cfg Config) (retVal []byte) {

	methodName := "okcatbox:account-wallet.go:generateWalletResponse"

	//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": ""})

	// We'll need an HTTP client for the subsequent requests.
	timeout := 60000 * time.Millisecond
	httpClient := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	// 1. At this point we know there must be a valid ok_access_key.  Given said access key, lookup the userID and
	// use that to find the categoryID for said user. Said category is created if necessary when the user obtains credentials so we know it must exist.
	userID := getUserId(req.Header)
	userCategoryID, err := getCategoryBySym(userID, httpClient, cfg)
	if s, errb := squeal(methodName, fmt.Sprintf("No category found for user %s\n", userID), err); errb {
		return []byte(s)
	}

	// 2. Get the account balances for all accounts tagged as funding available and this user.
	categories := fmt.Sprintf("%d,%d", userCategoryID, cfg.Bookwerx.CatFunding)
	url := fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey, categories)
	sums, err := getCategoryDistSums(httpClient, url)
	if s, errb := squeal(methodName, "getCategoryDistSums", err); errb {
		return []byte(s)
	}

	// For each item found in bookwerx, create an entry to return
	walletEntries := make([]utils.WalletEntry, 0)
	for _, brd := range sums.Sums {

		// Negate the sign.  BW reports this balance as a liability of okcatbox.  The ordinary CR balance is represented using a - sign.  But the user expects a DR value to match the asset on his books.
		n1 := brd.Sum
		n2 := DFP{-n1.Amount, n1.Exp}
		n3 := dfp_fmt(n2, -8) // here we lose info re: extra digits hidden by roundoff
		walletEntries = append(walletEntries, utils.WalletEntry{
			Available:  n3.s,
			Balance:    n3.s,
			CurrencyID: brd.Account.Currency.Symbol, // bw currency.symbol is used as the okex CurrencyID
			Hold:       "0.00000000",
		})
	}

	retVal, err = json.Marshal(walletEntries)
	if s, errb := squealJSONMarshal(methodName, walletEntries, err); errb {
		return []byte(s)
	}
	return retVal

}
