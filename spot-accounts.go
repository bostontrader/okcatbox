package main

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"github.com/gojektech/heimdall/httpclient"
	"net/http"
	"time"
)

// /api/spot/v3/accounts
func spotAccountsHandler(w http.ResponseWriter, req *http.Request, cfg Config) {
	errb := checkSigHeaders(w, req)
	if errb {
		return
	}
	retVal := generateAccountsResponse(req, cfg)
	_, _ = fmt.Fprintf(w, string(retVal))
}

func generateAccountsResponse(req *http.Request, cfg Config) (retVal []byte) {

	methodName := "okcatbox:account-wallet.go:generateAccountsResponse"

	//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": ""})

	// We'll need an HTTP client for the subsequent requests.
	timeout := 5000 * time.Millisecond
	httpClient := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	// 1. At this point we know there must be a valid ok_access_key.  Given said access key, lookup the userID and
	// use that to find the categoryID for said user. Said category is created if necessary when the user obtains credentials so we know it must exist.
	userID := getUserId(req.Header)
	userCategoryID, err := getCategoryBySym(userID, httpClient, cfg)
	if s, errb := squeal(methodName, fmt.Sprintf("No category found for user %s\n", userID), err); errb {
		return []byte(s)
	}

	// 2. Get the account balances for all accounts tagged as spot_available_cat and this user's okcatbox apikey.
	//categories := fmt.Sprintf("%d,%d", category_id, cfg.Bookwerx.SpotAvailableCat)
	//url = fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey, categories)
	//sumsA, err := getCategoryDistSums(client, url)
	//if err != nil {
	//log.Error(err)
	//return []byte(fmt.Sprintf("%v", err))
	//}
	// 2. Get the account balances for all accounts tagged as spot available and this user.
	categories := fmt.Sprintf("%d,%d", userCategoryID, cfg.Bookwerx.CatSpotAvailable)
	url := fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey, categories)
	sumsA, err := getCategoryDistSums(httpClient, url)
	if s, errb := squeal(methodName, "getCategoryDistSums spot available", err); errb {
		return []byte(s)
	}

	// 3. Get the account balances for all accounts tagged as spot_hold_cat and this user's okcatbox apikey.
	//categories = fmt.Sprintf("%d,%d", category_id, cfg.Bookwerx.SpotHoldCat)
	//url = fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey, categories)
	//sumsH, err := getCategoryDistSums(client, url)
	//if err != nil {
	//log.Error(err)
	//return []byte(fmt.Sprintf("%v", err))
	//}
	// 3. Get the account balances for all accounts tagged as spot hold and this user.
	categories = fmt.Sprintf("%d,%d", userCategoryID, cfg.Bookwerx.CatSpotHold)
	url = fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey, categories)
	//sumsH, err := getCategoryDistSums(httpClient, url)
	//if s, errb := squeal(methodName, "getCategoryDistSums spot hold", err); errb { return []byte(s) }

	// 4. Now build the return value.  In order to do this we will build a map of AccountsEntry.  We find the bits 'n' pieces of this separately (in hold and available) and we need to be able to access these AccountsEntry via a currency symbol key.
	accountsEntries := make(map[string]utils.AccountsEntry)

	// 4.1 For each item found in bookwerx for spot available, create the initial entry.
	for _, brd := range sumsA.Sums {

		// Negate the sign.  BW reports this balance as a liability of the OKCatbox.  The ordinary CR balance is represented using a - sign.  But the user expects a DR value to match the asset on his books.
		n1 := brd.Sum
		n2 := DFP{-n1.Amount, n1.Exp}
		n3 := dfp_fmt(n2, n2.Exp) // here we lose info re: extra digits hidden by roundoff

		accountsEntry := utils.AccountsEntry{
			AccountID:  "",
			Available:  n3.s,
			Balance:    n3.s,
			CurrencyID: brd.Account.Currency.Symbol, // bw currency.symbol is used as the okex CurrencyID
			Frozen:     "0",
			Hold:       "0",
			Holds:      "0",
		}

		accountsEntries[brd.Account.Currency.Symbol] = accountsEntry
	}

	// 4.2 For each item found in bookwerx for spot hold, update the existing entry or create a new one if necessary.
	// TODO

	// 4.3 Now extract the final return value as an array of the hashmap values priorly created.
	i := 0
	values := make([]utils.AccountsEntry, len(accountsEntries))
	for _, v := range accountsEntries {
		values[i] = v
		i++
	}

	retVal, err = json.Marshal(values)
	if s, errb := squealJSONMarshal(methodName, values, err); errb {
		return []byte(s)
	}
	return retVal

}
