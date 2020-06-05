package main

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// /catbox/deposit
func catbox_depositHandler(w http.ResponseWriter, req *http.Request, cfg Config) {
	retVal := generateCatboxDepositResponse(w, req, "POST", "/catbox/deposit", cfg)
	fmt.Fprintf(w, string(retVal))
}

func GetClient(urlBase string) (client *http.Client) {

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

/*
Recall that this method is a convenience method for the OKCatbox.  It doesn't exist in
the real OKEx API.
*/
func generateCatboxDepositResponse(w http.ResponseWriter, req *http.Request, verb string, endpoint string, cfg Config) (retVal []byte) {

	// Only a subset of the available fields,,
	type Currency struct {
		CurrencyID int `json:"id"`
		Symbol     string
	}

	type CurrencyShort struct {
		Symbol string
		Title  string
	}

	// We only need a subset of all the info returned.
	type AccountJoined struct {
		AccountID     int           `json:"id"`
		CurrencyShort CurrencyShort `json:"currency"`
		Title         string
	}

	type Insert struct {
		LID int `json:"last_insert_id"`
	}

	type Data struct {
		Data Insert `json:"data"`
	}

	if req.Method == http.MethodPost {

		// 1. Retrieve and validate the request parameters

		// 1.1 api_key
		api_key := req.FormValue("api_key") // OKCatbox key
		txn := db.Txn(false)
		defer txn.Abort()

		raw, err := txn.First("credentials", "id", api_key)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if raw == nil {
			return []byte("the api_key is not defined on this OKCatbox server.")
		}

		// 1.2 currency_symbol
		currency_symbol := req.FormValue("currency_symbol")

		// Determine the Bookwerx currency_id, thereby verifying that said currency is defined.
		url := fmt.Sprintf("%s/currencies?apikey=%s", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey)
		client := GetClient(url)
		req1, err := http.NewRequest("GET", url, nil)
		resp, err := client.Do(req1)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if resp.StatusCode != 200 {
			fmt.Println("error:\nexpected= ", 200, "\nreceived=", resp.StatusCode)
		}

		currencies := make([]Currency, 0)

		dec := json.NewDecoder(resp.Body)
		//dec.DisallowUnknownFields()
		err = dec.Decode(&currencies)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		resp.Body.Close()

		// Now search for the specific currency_symbol
		found := false
		var currency_id string
		for _, v := range currencies {
			if strings.EqualFold(v.Symbol, currency_symbol) { // case insensitive compare
				found = true
				currency_id = strconv.Itoa(v.CurrencyID)
				break
			}
		}
		if !found {
			return []byte(fmt.Sprintf("The currency %s is not defined on this OKCatbox server.", currency_symbol))
		}

		// 1.3 Quantity
		quanf := req.FormValue("quan")
		quand, err := decimal.NewFromString(quanf)
		if err != nil {
			return []byte(fmt.Sprintf("The quan %s cannot be parsed.", quanf))
		}

		quans := quand.Abs().Coefficient().String()

		exp := fmt.Sprint(quand.Exponent())
		var dramt, cramt string
		if quand.IsPositive() {
			dramt = quans
			cramt = "-" + dramt
		} else {
			cramt = quans
			dramt = "-" + cramt
		}
		fmt.Println(quans)

		// 1.4 Time.  Just a string, no validation.
		time := req.FormValue("time")

		// 2. Does the funding account for this api_key, currency exist?
		url = fmt.Sprintf("%s/accounts?apikey=%s", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey)
		req, err = http.NewRequest("GET", url, nil)
		resp, err = client.Do(req)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if resp.StatusCode != 200 {
			fmt.Println("error 2:\nexpected= ", 200, "\nreceived=", resp.StatusCode)
		}

		accountJoineds := make([]AccountJoined, 0)

		dec = json.NewDecoder(resp.Body)
		//dec.DisallowUnknownFields()
		err = dec.Decode(&accountJoineds)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		resp.Body.Close()

		// Can I find an account named #{api_key} using the same currency, that is tagged with the customer funding category?
		var account_id string
		account_exists := false
		for _, accountJoined := range accountJoineds {
			if accountJoined.Title == api_key {
				if strings.EqualFold(accountJoined.CurrencyShort.Symbol, currency_symbol) { // case insensitive compare
					// is this account tagged 'funding'  Figure this out later.
					account_id = strconv.Itoa(accountJoined.AccountID)
					account_exists = true
				}
			}
		}

		// If the account doesn't already exist, then create it.
		if !account_exists {
			url = fmt.Sprintf("%s/accounts", cfg.Bookwerx.Server)

			req, err = http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf("apikey=%s&currency_id=%s&rarity=0&title=%s", cfg.Bookwerx.APIKey, currency_id, api_key)))

			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			resp, err = client.Do(req)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			if resp.StatusCode != 200 {
				fmt.Println("error 2.1:\nexpected= ", 200, "\nreceived=", resp.StatusCode)
			}

			var insert Data
			dec = json.NewDecoder(resp.Body)
			//dec.DisallowUnknownFields()
			err = dec.Decode(&insert)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			resp.Body.Close()
		}

		// 3. Now create the transaction and the two distributions using three requests.

		//time.Sleep(1000 * time.Millisecond)

		// 3.1 Create the tx
		url = fmt.Sprintf("%s/transactions", cfg.Bookwerx.Server)
		req, err = http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf("apikey=%s&notes=deposit&time=%s", cfg.Bookwerx.APIKey, time)))

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Close = true
		resp, err = client.Do(req)
		if err != nil {
			log.Fatalf("error 3.1: %v", err)
		}

		if resp.StatusCode != 200 {
			fmt.Println("error 3.1:\nexpected= ", 200, "\nreceived=", resp.StatusCode, resp.Body)
		}

		var insert Data
		dec = json.NewDecoder(resp.Body)
		//dec.DisallowUnknownFields()
		err = dec.Decode(&insert)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		txid := strconv.Itoa(insert.Data.LID)
		resp.Body.Close()

		// 3.2 Create the DR distributions
		url = fmt.Sprintf("%s/distributions", cfg.Bookwerx.Server)

		// HACK! Hardwired hot-wallet account_id.  Fix this!
		req, err = http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf("apikey=%s&account_id=117&amount=%s&amount_exp=%s&transaction_id=%s", cfg.Bookwerx.APIKey, dramt, exp, txid)))

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Close = true
		resp, err = client.Do(req)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if resp.StatusCode != 200 {
			fmt.Println("error 3.2:\nexpected= ", 200, "\nreceived=", resp.StatusCode)
		}

		dec = json.NewDecoder(resp.Body)
		//dec.DisallowUnknownFields()
		err = dec.Decode(&insert)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		resp.Body.Close()

		// 3.3 Create the CR distributions
		url = fmt.Sprintf("%s/distributions", cfg.Bookwerx.Server)
		req, err = http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf("apikey=%s&account_id=%s&amount=%s&amount_exp=%s&transaction_id=%s", cfg.Bookwerx.APIKey, account_id, cramt, exp, txid)))

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Close = true
		resp, err = client.Do(req)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		if resp.StatusCode != 200 {
			fmt.Println("error 3.3:\nexpected= ", 200, "\nreceived=", resp.StatusCode)
		}

		dec = json.NewDecoder(resp.Body)
		//dec.DisallowUnknownFields()
		err = dec.Decode(&insert)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		resp.Body.Close()

		return []byte("success")

	} else {
		return []byte("use post")
	}

}
