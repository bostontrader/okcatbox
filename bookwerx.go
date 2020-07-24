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
		return fmt.Sprintf("bookwerx.go body_string :%v", err)
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
		s := fmt.Sprintf("bookwerx.go createDistribution 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx.go createDistribution 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	var insert LID
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&insert)
	if err != nil {
		s := fmt.Sprintf("bookwerx.go createDistribution 3: %v", err)
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
		s := fmt.Sprintf("bookwerx.go createTransaction 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx.go createTransaction 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	var insert LID
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&insert)
	if err != nil {
		s := fmt.Sprintf("bookwerx.go createTransaction 3: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}
	txid = insert.LastInsertID

	return txid, nil
}

// Given a currency symbol, determine the bookwerx currency_id.
func getCurrencyBySym(client *httpclient.Client, currency_symbol string, cfg Config) (currency_id int32, err error) {

	url := fmt.Sprintf("%s/currencies?apikey=%s", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey)

	resp, err := client.Get(url, nil)
	if err != nil {
		s := fmt.Sprintf("bookwerx.go getCurrencyBySym 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx.go getCurrencyBySym 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	currencies := make([]Currency, 0)
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&currencies)
	if err != nil {
		s := fmt.Sprintf("bookwerx.go getCurrencyBySym 3: %v", err)
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
		s := fmt.Sprintf("bookwerx.go getCurrencyBySym 4: The currency %s is not defined on this OKCatbox server.", currency_symbol)
		log.Error(s)
		return -1, errors.New(s)
	}

	return currency_id, nil
}

/* func getCategoryBySym(client *http.Client, url string) (category_id int32) {

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
} */

func getCategoryBySymB(client *httpclient.Client, url string) (category_id int32, err error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s := fmt.Sprintf("bookwerx.go getCategoryBySym 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx.go getCategoryBySym 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	n := Category{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		s := fmt.Sprintf("bookwerx.go getCategoryBySym 3: Error with JSON decoding.")
		log.Error(s)
		return -1, errors.New(s)
	}

	return n.Id, nil
}

func getCategoryDistSums(client *httpclient.Client, url string) (retVal Sums, err error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s := fmt.Sprintf("wallet.go getCategoryDistSums 1: %v", err)
		log.Error(s)
		return Sums{}, nil
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("wallet.go getCategoryDistSums 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return Sums{}, nil
	}

	n := Sums{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		s := fmt.Sprintf("wallet.go getCategoryDistSums 3: Error with JSON decoding.")
		log.Error(s)
		return Sums{}, nil
	}

	return n, nil
}

func postAccount(client *httpclient.Client, currency_id int32, title string, cfg Config) (account_id int32, err error) {

	url1 := fmt.Sprintf("%s/accounts", cfg.Bookwerx.Server)
	url2 := fmt.Sprintf("apikey=%s&currency_id=%d&rarity=0&title=%s", cfg.Bookwerx.APIKey, currency_id, title)

	h := make(map[string][]string)
	h["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	resp, err := client.Post(url1, bytes.NewBuffer([]byte(url2)), h)
	defer resp.Body.Close()
	if err != nil {
		s := fmt.Sprintf("bookwerx.go postAccount 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx.go postAccount 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	n := LID{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		s := fmt.Sprintf("bookwerx.go postAccount 3: %v", err)
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
		s := fmt.Sprintf("bookwerx.go postAcctcat 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx.go postAcctcat 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	n := LID{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		s := fmt.Sprintf("bookwerx.go postAcctcat 3: %v", err)
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
		s := fmt.Sprintf("credentials.go makeCategory 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("credentials.go makeCategory 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	n := LID{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		s := fmt.Sprintf("credential.go makeCategory 3: Error with JSON decoding.")
		log.Error(s)
		return -1, errors.New(s)
	}

	return n.LastInsertID, nil
}

func searchA(accounts []AccountJoined, ok_access_key8 string, currency_symbol string) (account_id int32, found bool) {

	// Can I find an account tagged funding _and_ apikey using the correct currency?
	cat_funding := false
	cat_api := false
	//currency := false
	for _, account := range accounts {
		//currency = false
		if strings.EqualFold(account.CurrencyShort.Symbol, currency_symbol) {
			//currency = true
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

func getFundingAccountID(client *httpclient.Client, accounts []AccountJoined, ok_access_key8 string, currency_id int32, currency_symbol string, cfg Config) (account_id int32, err error) {

	// 1. Can I find an account tagged funding _and_ apikey using the correct currency?
	account_id, found := searchA(accounts, ok_access_key8, currency_symbol)

	// 2. If the account doesn't already exist, then create it and tag it with the customer's category and funding
	if !found {

		// 2.1 Find the category id for this apikey
		url := fmt.Sprintf("%s/category/bysym/%s?apikey=%s", cfg.Bookwerx.Server, ok_access_key8, cfg.Bookwerx.APIKey)
		category_id_api, err := getCategoryBySymB(client, url)
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
		_, err = postAcctcat(client, account_id, cfg.Bookwerx.Funding, cfg)
		if err != nil {
			log.Error(err)
			return -1, err
		}

	}

	return account_id, nil
}

func getAccounts(client *httpclient.Client, cfg Config) (accounts []AccountJoined, err error) {

	url := fmt.Sprintf("%s/accounts?apikey=%s", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey)

	resp, err := client.Get(url, nil)
	if err != nil {
		s := fmt.Sprintf("bookwerx.go getHotWalletAccountID 1: %v", err)
		log.Error(s)
		return nil, errors.New(s)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx.go getHotWalletAccountID 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return nil, errors.New(s)
	}

	accountJoineds := make([]AccountJoined, 0)
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&accountJoineds)
	if err != nil {
		s := fmt.Sprintf("bookwerx.go getHotWalletAccountID 3: %v", err)
		log.Error(s)
		return nil, errors.New(s)
	}

	return accountJoineds, nil
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
