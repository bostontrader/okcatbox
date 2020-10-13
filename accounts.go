package main

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"github.com/gojektech/heimdall/httpclient"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// /api/spot/v3/accounts
func spot_accountsHandler(w http.ResponseWriter, req *http.Request, cfg Config) {
	retVal := generateAccountsResponse(w, req, cfg)
	fmt.Fprintf(w, string(retVal))
}

func generateAccountsResponse(w http.ResponseWriter, req *http.Request, cfg Config) (retVal []byte) {

	retVal, err := checkSigHeaders(w, req)
	if err {
		return

	} else {
		// At this point we know there must be a valid ok_access_key.  We use the first 8 bytes as an identifier in bookwerx.
		ok_access_key8 := getOKAccessKey(req.Header)

		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": ""})

		// We'll need an HTTP client for the subsequent requests.
		timeout := 5000 * time.Millisecond
		client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

		// 1. Lookup the category_id for this user's okcatbox apikey
		url := fmt.Sprintf("%s/category/bysym/%s?apikey=%s", cfg.Bookwerx.Server, ok_access_key8, cfg.Bookwerx.APIKey)
		category_id, err := getCategoryBySym(client, url)
		if err != nil {
			log.Error(err)
			return []byte(fmt.Sprintf("%v", err))
		}

		// 2. Get the account balances for all accounts tagged as spot_available_cat and this user's okcatbox apikey.
		categories := fmt.Sprintf("%d,%d", category_id, cfg.Bookwerx.SpotAvailableCat)
		url = fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey, categories)
		sumsA, err := getCategoryDistSums(client, url)
		if err != nil {
			log.Error(err)
			return []byte(fmt.Sprintf("%v", err))
		}

		// 3. Get the account balances for all accounts tagged as spot_hold_cat and this user's okcatbox apikey.
		//categories = fmt.Sprintf("%d,%d", category_id, cfg.Bookwerx.SpotHoldCat)
		//url = fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey, categories)
		//sumsH, err := getCategoryDistSums(client, url)
		//if err != nil {
		//log.Error(err)
		//return []byte(fmt.Sprintf("%v", err))
		//}

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

		retVal, _ := json.Marshal(values)
		return retVal

	}

	return
}
