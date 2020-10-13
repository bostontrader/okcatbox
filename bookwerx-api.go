// The purpose of this file is to hold items required to communicate with a bookwerx server.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gojektech/heimdall/httpclient"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type AId struct {
	Id int32 `json:"accounts-id"`
}

type CatId struct {
	Id int32 `json:"categories-id"`
}

type AccountCurrency struct {
	AccountID int32
	Title     string
	Currency  CurrencySymbol
}

// We only need a subset of all the info returned.
type AccountJoined struct {
	AccountID     int32 `json:"id"`
	Categories    []Acctcat2
	CurrencyShort CurrencyShort `json:"currency"`
	Title         string
}

type Acctcat2 struct {
	CategorySymbol string `json:"category_symbol"`
}

type BalanceResultDecorated struct {
	Account AccountCurrency
	Sum     DFP
}

type Category struct {
	Id     int32
	Apikey string
	Symbol string
	Title  string
}

// Only a subset of the available fields,,
type Currency struct {
	CurrencyID int32 `json:"id"`
	Symbol     string
}

type CurrencyShort struct {
	Symbol string
	Title  string
}

type CurrencySymbol struct {
	CurrencyID int32
	Symbol     string
}

type LID struct {
	LastInsertID int32
}

type Sums struct {
	Sums []BalanceResultDecorated
}

// Given a response object, read the body and return it as a string.  Deal with the error message if necessary.
func body_string(resp *http.Response) string {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("bookwerx-api.go:body_string :%v", err)
	}
	return string(body)
}

// Some of the functions use the stock go http client.
func GetClientA(urlBase string) (client *http.Client) {

	if len(urlBase) >= 6 && urlBase[:6] == "https:" {
		tr := &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		}
		return &http.Client{Transport: tr}
	}

	return &http.Client{}

}

func createDistribution(client *httpclient.Client, account_id int32, amt string, exp string, txid int32, cfg Config) (did int32, err error) {

	url1 := fmt.Sprintf("%s/distributions", cfg.Bookwerx.Server)
	url2 := fmt.Sprintf("apikey=%s&account_id=%d&amount=%s&amount_exp=%s&transaction_id=%d", cfg.Bookwerx.APIKey, account_id, amt, exp, txid)

	h := make(map[string][]string)
	h["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	resp, err := client.Post(url1, bytes.NewBuffer([]byte(url2)), h)
	defer resp.Body.Close()
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:createDistribution 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go:createDistribution 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	var insert LID
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&insert)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:createDistribution 3: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	return insert.LastInsertID, nil
}

func createTransaction(client *httpclient.Client, time string, cfg Config) (txid int32, err error) {

	url1 := fmt.Sprintf("%s/transactions", cfg.Bookwerx.Server)
	url2 := fmt.Sprintf("apikey=%s&notes=deposit&time=%s", cfg.Bookwerx.APIKey, time)

	h := make(map[string][]string)
	h["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	resp, err := client.Post(url1, bytes.NewBuffer([]byte(url2)), h)
	defer resp.Body.Close()
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:createTransaction 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go:createTransaction 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	var insert LID
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&insert)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:createTransaction 3: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}
	txid = insert.LastInsertID

	return txid, nil
}

func getAccounts(client *httpclient.Client, cfg Config) (accounts []AccountJoined, err error) {

	url := fmt.Sprintf("%s/accounts?apikey=%s", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey)

	resp, err := client.Get(url, nil)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:getHotWalletAccountID 1: %v", err)
		log.Error(s)
		return nil, errors.New(s)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go:getHotWalletAccountID 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return nil, errors.New(s)
	}

	accountJoineds := make([]AccountJoined, 0)
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&accountJoineds)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:getHotWalletAccountID 3: %v", err)
		log.Error(s)
		return nil, errors.New(s)
	}

	return accountJoineds, nil
}

func getCategoryBySym(client *httpclient.Client, url string) (category_id int32, err error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:getCategoryBySym 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go:getCategoryBySym 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	n := Category{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:getCategoryBySym 3: Error with JSON decoding.")
		log.Error(s)
		return -1, errors.New(s)
	}

	return n.Id, nil
}

func getCategoryDistSums(client *httpclient.Client, url string) (retVal Sums, err error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:getCategoryDistSums 1: %v", err)
		log.Error(s)
		return Sums{}, nil
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go:getCategoryDistSums 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return Sums{}, nil
	}

	n := Sums{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:getCategoryDistSums 3: Error with JSON decoding.")
		log.Error(s)
		return Sums{}, nil
	}

	return n, nil
}

func getCurrencyBySym(client *httpclient.Client, currency_symbol string, cfg Config) (currency_id int32, err error) {

	url := fmt.Sprintf("%s/currencies?apikey=%s", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey)

	resp, err := client.Get(url, nil)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:getCurrencyBySym 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go:getCurrencyBySym 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	currencies := make([]Currency, 0)
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&currencies)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:getCurrencyBySym 3: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	// Now search for the specific currency_symbol
	found := false
	for _, v := range currencies {
		if strings.EqualFold(v.Symbol, currency_symbol) { // case insensitive compare
			found = true
			currency_id = v.CurrencyID
			break
		}
	}
	if !found {
		s := fmt.Sprintf("bookwerx-api.go:getCurrencyBySym 4: The currency %s is not defined on this OKCatbox server.", currency_symbol)
		log.Error(s)
		return -1, errors.New(s)
	}

	return currency_id, nil
}

func getFundingAccountID(client *httpclient.Client, accounts []AccountJoined, ok_access_key8 string, currency_id int32, currency_symbol string, cfg Config) (account_id int32, err error) {

	// 1. Can I find an account tagged funding _and_ apikey8 using the correct currency?
	account_id, found := searchF(accounts, ok_access_key8, currency_symbol)

	// 2. If the account doesn't already exist, then create it and tag it with the customer's category and funding
	if !found {

		// 2.1 Find the category id for this apikey
		url := fmt.Sprintf("%s/category/bysym/%s?apikey=%s", cfg.Bookwerx.Server, ok_access_key8, cfg.Bookwerx.APIKey)
		category_id_api, err := getCategoryBySym(client, url)
		if err != nil {
			log.Error(err)
			return -1, err
		}

		// 2.2 Make the new account
		account_id, err = postAccount(client, currency_id, ok_access_key8, cfg)
		if err != nil {
			log.Error(err)
			return -1, err
		}

		// 2.3 Tag with the api key
		_, err = postAcctcat(client, account_id, category_id_api, cfg)
		if err != nil {
			log.Error(err)
			return -1, err
		}

		// 2.4 Tag with funding
		_, err = postAcctcat(client, account_id, cfg.Bookwerx.FundingCat, cfg)
		if err != nil {
			log.Error(err)
			return -1, err
		}

	}

	return account_id, nil
}

// The bookwerx-core server will on occasion return JSON names that contain a '.'.  This vile habit
// causes trouble here.
// A good, bad, or ugly hack is to simply change the . to a -.  Do that here.
func fixDot(b []byte) {
	for i, num := range b {
		if num == 46 { // .
			b[i] = 45 // -
		}
	}
}

// Generic get on the API
func get(client *httpclient.Client, url string) ([]byte, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("bookwerx-api.go:get: NewRequest error: %v", err)
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("bookwerx-api.go:get: client.Do error: %v", err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("bookwerx-api.go:get: ReadAll error: %v", err)
		return nil, err
	}
	_ = resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Status Code error: expected= 200, received=", resp.StatusCode)
		fmt.Println("body=", string(body))
		return nil, err
	}

	return body, nil

}

/*
A transfer has a source and destination (from, to) account in bookwerx.  Each account is tagged with two categories and is configured to use
a specific currency. More particularly:

transfer_cat_id: This is the bookwerx category_id of the category that corresponds with spot, funding, etc.  This category is created
during the initial installation of okcatbox and the id is provided via okcatbox.yaml.  We therefore know that this category must exist.

apikey8: This is the first 8 digits of the user's okcatbox apikey.  This corresponds with a bookwerx categories title.  Which category?  This
function will have to determine that.  Said category is created when the user initially creates his credentials so we know the category exists
but we must look it up here.

currency_id: This is the bookwerx currency id of the currency involved in the transfer.

We cannot blindly establish these accounts in advance for each new user.  Instead, create them here if necessary.
*/

//func getTransferAccountID(client *httpclient.Client, accounts []AccountJoined, ok_access_key8 string, currency_id int32, currency_symbol string, cfg Config) (account_id int32, err error) {
func getTransferAccountID(client *httpclient.Client, transfer_cat_id int32, ok_access_key8 string, currency_id int32, cfg Config) (account_id int32, err error) {

	// 1. Find the category id for this particular user.
	selectt := "SELECT%20categories.id"
	from := "FROM%20categories"
	where := fmt.Sprintf("WHERE%%20categories.title%%3d%%27%s%%27", ok_access_key8)
	query := fmt.Sprintf("%s%%20%s%%20%s", selectt, from, where)
	url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)

	body, err := get(client, url)
	if err != nil {
		fmt.Println("bookwerx-api.go:getTransferAccountID 1: get error: ", err)
		return -1, err
	}
	fixDot(body) // an array of categories.id

	n := make([]CatId, 0)
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&n)
	if err != nil {
		fmt.Println("bookwerx-api.go: getTransferAccountID:", err)
		return
	}
	if len(n) == 0 {
		fmt.Println("bookwerx-api.go: getTransferAccountID: Category %s is not defined in bookwerx", ok_access_key8)
		return
	} else if len(n) > 1 {
		fmt.Println("bookwerx-api.go: getTransferAccountID: There are more than one suitable categories.  This should never happen.")
		return
	}
	apikey_cat_id := n[0].Id

	// 2. Find the desired account ID
	selectt = "SELECT%20accounts.id"
	from = "FROM%20accounts_categories"
	join1 := "JOIN%20accounts%20ON%20accounts.id%3daccounts_categories.account_id"
	join2 := "JOIN%20currencies%20ON%20currencies.id%3daccounts.currency_id"
	where = fmt.Sprintf("WHERE%%20accounts_categories.category_id%%20IN%%20(%d,%d)", transfer_cat_id, apikey_cat_id)
	and := fmt.Sprintf("AND%%20currencies.id%%3d%d", currency_id)
	group := "GROUP%20BY%20accounts_categories.account_id%20HAVING%20COUNT(DISTINCT%20accounts_categories.account_id)%3d2"
	query = fmt.Sprintf("%s%%20%s%%20%s%%20%s%%20%s%%20%s%%20%s", selectt, from, join1, join2, where, and, group)
	url = fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)
	body, err = get(client, url)
	if err != nil {
		fmt.Println("bookwerx-api.go:getTransferAccountID 2: get error: ", err)
		return
	}
	fixDot(body)

	n1 := make([]AId, 0)
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&n1)
	if err != nil {
		fmt.Println("bookwerx-api.go: getTransferAccountID:", err)
		return
	}
	if len(n1) == 1 {
		return n1[0].Id, nil
	} else if len(n1) > 1 {
		fmt.Println("bookwerx-api.go: getTransferAccountID: There are more than one suitable accounts.  This should never happen.")
		return
	}

	// 3. If control passes here we know that the account doesn't already exist. So create it and tag it with the two categories and the given currency.
	if len(n1) == 0 {

		// 3.1 Make the new account
		account_id, err = postAccount(client, currency_id, ok_access_key8, cfg)
		if err != nil {
			fmt.Println("bookwerx-api.go:getTransferAccountID 3.1: postAccount error: ", err)
			return -1, err
		}

		// 3.2 Tag with the api key
		_, err = postAcctcat(client, account_id, apikey_cat_id, cfg)
		if err != nil {
			fmt.Println("bookwerx-api.go:getTransferAccountID 3.2: postAcctcat error: ", err)
			return -1, err
		}

		// 3.3 Tag with funding, spot, whatever
		_, err = postAcctcat(client, account_id, transfer_cat_id, cfg)
		if err != nil {
			fmt.Println("bookwerx-api.go:getTransferAccountID 3.3: postAcctcat error: ", err)
			return -1, err
		}

	}

	return account_id, nil
}

func getHotWalletAccountID(client *httpclient.Client, accounts []AccountJoined, currency_symbol string, cfg Config) (account_id int32, err error) {

	// Can I find an account named #{apikey} using the same currency, that is tagged with the customer funding category?
	for _, account := range accounts {
		//if accountJoined.Title == ok_access_key8 {
		if strings.EqualFold(account.CurrencyShort.Symbol, currency_symbol) { // case insensitive compare
			// is this account tagged 'H'
			for _, b1 := range account.Categories {
				if strings.EqualFold(b1.CategorySymbol, "H") { // case insensitive compare
					account_id = account.AccountID
					break
				}
			}
		}
	}

	return account_id, nil
}

func getSpotAvailableAccountID(client *httpclient.Client, accounts []AccountJoined, ok_access_key8 string, currency_id int32, currency_symbol string, cfg Config) (account_id int32, err error) {

	// 1. Can I find an account tagged spot-available _and_ apikey using the correct currency?
	account_id, found := searchSA(accounts, ok_access_key8, currency_symbol)

	// 2. If the account doesn't already exist, then create it and tag it with the customer's category and funding
	if !found {

		// 2.1 Find the category id for this apikey
		url := fmt.Sprintf("%s/category/bysym/%s?apikey=%s", cfg.Bookwerx.Server, ok_access_key8, cfg.Bookwerx.APIKey)
		category_id_api, err := getCategoryBySym(client, url)
		if err != nil {
			log.Error(err)
			return -1, err
		}

		// 2.2 Make the new account
		account_id, err = postAccount(client, currency_id, ok_access_key8, cfg)
		if err != nil {
			log.Error(err)
			return -1, err
		}

		// 2.3 Tag with the api key
		_, err = postAcctcat(client, account_id, category_id_api, cfg)
		if err != nil {
			log.Error(err)
			return -1, err
		}

		// 2.4 Tag with funding
		_, err = postAcctcat(client, account_id, cfg.Bookwerx.FundingCat, cfg)
		if err != nil {
			log.Error(err)
			return -1, err
		}

	}

	return account_id, nil
}

func postAccount(client *httpclient.Client, currency_id int32, title string, cfg Config) (account_id int32, err error) {

	url1 := fmt.Sprintf("%s/accounts", cfg.Bookwerx.Server)
	url2 := fmt.Sprintf("apikey=%s&currency_id=%d&rarity=0&title=%s", cfg.Bookwerx.APIKey, currency_id, title)

	h := make(map[string][]string)
	h["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	resp, err := client.Post(url1, bytes.NewBuffer([]byte(url2)), h)
	defer resp.Body.Close()
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:postAccount 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go:postAccount 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	n := LID{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:postAccount 3: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	return n.LastInsertID, nil
}

func postAcctcat(client *httpclient.Client, account_id int32, category_id int32, cfg Config) (acctcat_id int32, err error) {

	url1 := fmt.Sprintf("%s/acctcats", cfg.Bookwerx.Server)
	url2 := fmt.Sprintf("apikey=%s&account_id=%d&category_id=%d", cfg.Bookwerx.APIKey, account_id, category_id)

	h := make(map[string][]string)
	h["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	resp, err := client.Post(url1, bytes.NewBuffer([]byte(url2)), h)
	defer resp.Body.Close()
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:postAcctcat 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go:postAcctcat 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	n := LID{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:postAcctcat 3: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	return n.LastInsertID, nil
}

func postCategory(client *httpclient.Client, url string, cb_apikey string, bw_apikey string) (retVal int32, err error) {

	url1 := fmt.Sprintf("%s/categories", url)
	url2 := fmt.Sprintf("apikey=%s&symbol=%s&title=%s", cb_apikey, bw_apikey, bw_apikey)

	h := make(map[string][]string)
	h["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	resp, err := client.Post(url1, bytes.NewBuffer([]byte(url2)), h)
	defer resp.Body.Close()
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:makeCategory 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go:makeCategory 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	n := LID{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:makeCategory 3: Error with JSON decoding.")
		log.Error(s)
		return -1, errors.New(s)
	}

	return n.LastInsertID, nil
}

// HAX: Rethink and factor these search* functions.
// Can I find an account tagged funding _and_ apikey8 using the correct currency?
func searchF(accounts []AccountJoined, ok_access_key8 string, currency_symbol string) (account_id int32, found bool) {

	// Can I find an account tagged funding _and_ apikey using the correct currency?
	cat_funding := false
	cat_api := false
	for _, account := range accounts {
		if strings.EqualFold(account.CurrencyShort.Symbol, currency_symbol) {
			for _, category := range account.Categories {
				if strings.EqualFold(category.CategorySymbol, "F") { // case insensitive compare
					cat_funding = true
				} else if strings.EqualFold(category.CategorySymbol, ok_access_key8) {
					cat_api = true
				}
			}
			if cat_funding && cat_api {
				return account.AccountID, true
			}
		}

	}

	return -1, false

}

// Find all accounts tagged funding _and_ apikey8
func searchFAll(accounts []AccountJoined, ok_access_key8 string) (account_id []int32) {

	var account_ids []int32
	cat_funding := false
	cat_api := false
	for _, account := range accounts {
		for _, category := range account.Categories {
			if strings.EqualFold(category.CategorySymbol, "F") { // case insensitive compare
				cat_funding = true
			} else if strings.EqualFold(category.CategorySymbol, ok_access_key8) {
				cat_api = true
			}
		}
		if cat_funding && cat_api {
			account_ids = append(account_ids, account.AccountID)
		}
	}

	return account_ids
}

// Can I find an account tagged spot-available _and_ apikey using the correct currency?
func searchSA(accounts []AccountJoined, ok_access_key8 string, currency_symbol string) (account_id int32, found bool) {

	// Can I find an account tagged funding _and_ apikey using the correct currency?
	cat_funding := false
	cat_api := false
	for _, account := range accounts {
		if strings.EqualFold(account.CurrencyShort.Symbol, currency_symbol) {
			for _, category := range account.Categories {
				if strings.EqualFold(category.CategorySymbol, "F") { // case insensitive compare
					cat_funding = true
				} else if strings.EqualFold(category.CategorySymbol, ok_access_key8) {
					cat_api = true
				}
			}
			if cat_funding && cat_api {
				return account.AccountID, true
			}
		}

	}

	return -1, false

}
