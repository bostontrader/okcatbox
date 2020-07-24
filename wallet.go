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

type CurrencySymbol struct {
	CurrencyID int32
	Symbol     string
}

type AccountCurrency struct {
	AccountID int32
	Title     string
	Currency  CurrencySymbol
}

type BalanceResultDecorated struct {
	Account AccountCurrency
	Sum     DFP
}

type Sums struct {
	Sums []BalanceResultDecorated
}

// /catbox/wallet
func catbox_walletHandler(w http.ResponseWriter, req *http.Request, cfg Config) {
	retVal := generateWalletResponse(w, req, "GET", "/api/account/v3/wallet", cfg)
	fmt.Fprintf(w, string(retVal))
}
func getOKAccessKey(headers map[string][]string) string {
	if value, ok := headers["Ok-Access-Key"]; ok {
		return value[0][:8]
	}
	return ""
}

func generateWalletResponse(w http.ResponseWriter, req *http.Request, verb string, endpoint string, cfg Config) (retVal []byte) {

	retVal, err := checkSigHeaders(w, req)
	if err {
		return

	} else {
		// At this point we know there must be a valid ok_access_key.  We use the first 8 bytes as an identifier in bookwerx.
		ok_access_key8 := getOKAccessKey(req.Header)

		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": ""})

		//walletEntries, _ := catfood(cfg, ok_access_key)
		//client := GetClientA(cfg.Bookwerx.Server)
		// We'll need an HTTP client for the subsequent requests.
		timeout := 5000 * time.Millisecond
		client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

		// 1. Lookup the category_id for this user's okcatbox apikey
		url := fmt.Sprintf("%s/category/bysym/%s?apikey=%s", cfg.Bookwerx.Server, ok_access_key8, cfg.Bookwerx.APIKey)
		category_id, err := getCategoryBySymB(client, url)
		if err != nil {
			log.Error(err)
			return []byte(fmt.Sprintf("%v", err))
		}

		// 2. Get the account balances for all accounts tagged as funding and this user's okcatbox apikey.
		categories := fmt.Sprintf("%d,%d", category_id, cfg.Bookwerx.Funding)
		url = fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey, categories)
		sums, err := getCategoryDistSums(client, url)
		if err != nil {
			log.Error(err)
			return []byte(fmt.Sprintf("%v", err))
		}

		// For each item found in bookwerx, create an entry to return
		walletEntries := make([]utils.WalletEntry, 0)
		for _, _ = range sums.Sums {
			walletEntries = append(walletEntries, utils.WalletEntry{
				Available:  "0.00000000",
				Balance:    "0.00000000",
				CurrencyID: "c",
				Hold:       "0.00000000"})
		}

		retVal, _ := json.Marshal(walletEntries)
		return retVal
	}

	return
}

// Get balances, as of the current time, for all funding accounts for the given apikey category
//func catfood(cfg Config, ok_access_key8 string) (retVal []utils.WalletEntry, err1 string) {

//client := GetClientA(cfg.Bookwerx.Server)

// 1. Lookup the category_id for this user's okcatbox apikey
//url := fmt.Sprintf("%s/category/bysym/%s?apikey=%s", cfg.Bookwerx.Server, ok_access_key8, cfg.Bookwerx.APIKey)
//category_id := getCategoryBySym(client, url)

// 2. Get the account balances for all accounts tagged as funding and this user's okcatbox apikey.
//categories := fmt.Sprintf("%d,%s", category_id, cfg.Bookwerx.Funding)
//url = fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey, categories)
//sums := getCategoryDistSums(client, url)

// For each item found in bookwerx, create an entry to return
//retVal = make([]utils.WalletEntry, 0)
//for _, _ = range sums.Sums {
//walletEntries = append(walletEntries, utils.WalletEntry{Available: "0.00000000", Balance: "0.00000000", CurrencyID: "c", Hold: "0.00000000"})
//}
// GET dist for category OKEX-Funding-Hold
// For each item found in bookwerx, create an entry to return
//return
//}
