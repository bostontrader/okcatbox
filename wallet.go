package main

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"net/http"
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

	fmt.Println(req, "\n")

	retVal, err := checkSigHeaders(w, req)
	if err {
		return

	} else {
		// At this point we know there must be a valid ok_access_key.  We use the first 8 bytes as an identifier in bookwerx.
		ok_access_key := getOKAccessKey(req.Header)

		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": ""})

		walletEntries, _ := catfood(cfg, ok_access_key)
		//if err != nil {
		//return squeal(fmt.Sprintf("wallet.go generateWalletResponse 1: %v", err))
		//}

		//walletEntries := make([]utils.WalletEntry, 0)
		//walletEntries[0] = utils.WalletEntry{Available: "a", Balance: "b", CurrencyID: "c", Hold: "h"}
		retVal, _ := json.Marshal(walletEntries)
		return retVal
	}

	return
}

func getCategoryBySym(client *http.Client, url string) (retVal int32) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("wallet.go getCategoryBySym 1: %v", err)
		return -1
	}
	//req.Close = true
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("wallet.go getCategoryBySym 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		return -1
	}

	n := Category{}
	//n1, err := ioutil.ReadAll(resp.Body);
	//fmt.Println(n1);
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		fmt.Println("wallet.go getCategoryBySym 3: Error with JSON decoding.")
		return -1
	}

	return n.Id
}

func getCategoryDistSums(client *http.Client, url string) (retVal Sums) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("wallet.go getCategoryDistSums 1: %v", err)
		return Sums{}
	}
	//req.Close = true
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("wallet.go getCategoryDistSums 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		return Sums{}
	}

	n := Sums{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		fmt.Println("wallet.go getCategoryDistSums 3: Error with JSON decoding.")
		return Sums{}
	}

	//retVal, _ := json.Marshal(n)
	retVal = n
	return
	//fmt.Fprintf(w, string(retVal))
}

// Get balances, as of the current time, for all funding accounts for the given apikey category
func catfood(cfg Config, ok_access_key8 string) (retVal []utils.WalletEntry, err1 string) {

	client := GetClientA(cfg.Bookwerx.Server)

	// 1. Lookup the category_id for this user's okcatbox apikey
	url := fmt.Sprintf("%s/category/bysym/%s?apikey=%s", cfg.Bookwerx.Server, ok_access_key8, cfg.Bookwerx.APIKey)
	category_id := getCategoryBySym(client, url)
	fmt.Println(category_id)

	// 2. Get the account balances for all accounts tagged as funding and this user's okcatbox apikey.
	categories := fmt.Sprintf("%d,%s", category_id, cfg.Bookwerx.Funding)
	url = fmt.Sprintf("%s/category_dist_sums?apikey=%s&category_id=%s&decorate=true", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey, categories)
	sums := getCategoryDistSums(client, url)

	//distSums := make([]DistSum, 0)

	//dec := json.NewDecoder(resp.Body)
	//err = dec.Decode(&distSums)
	//if err != nil {
	//return nil, fmt.Sprintf("deposit.go 2.2: ", err)
	//}

	// For each item found in bookwerx, create an entry to return
	retVal = make([]utils.WalletEntry, 0)
	//for _, v := range distSums {
	for _, _ = range sums.Sums {
		//walletEntries = append(walletEntries, utils.WalletEntry{Available: "0.00000000", Balance: "0.00000000", CurrencyID: "c", Hold: "0.00000000"})
	}
	// GET dist for category OKEX-Funding-Hold
	// For each item found in bookwerx, create an entry to return
	return
}
