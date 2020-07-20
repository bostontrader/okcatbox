package main

import (
	"fmt"
	"github.com/gojektech/heimdall/httpclient"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// /catbox/deposit
func catbox_depositHandler(w http.ResponseWriter, req *http.Request, cfg Config) {
	retVal := generateCatboxDepositResponse(w, req, cfg)
	fmt.Fprintf(w, string(retVal))
}

func generateCatboxDepositResponse(w http.ResponseWriter, req *http.Request, cfg Config) (retVal []byte) {

	if req.Method == http.MethodPost {

		// 1. Retrieve and validate the request parameters.
		if err := req.ParseForm(); err != nil {
			s := fmt.Sprintf("deposit.go ParseForm 1 err: %v", err)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}

		log.Printf("%s %s %s %s", req.Method, req.URL, req.Header, req.Form)

		// 1.1 OKCatbox apikey.
		apikey := req.FormValue("apikey")
		ok_access_key8 := apikey[:8]
		// Is this key defined with this OKCatbox server?
		txn := db.Txn(false)
		defer txn.Abort()

		raw, err := txn.First("credentials", "id", apikey)
		if err != nil {
			s := fmt.Sprintf("deposit.go generateCatboxDepositResponse 1.1: %v", err)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}

		if raw == nil {
			s := fmt.Sprintf("deposit.go generateCatboxDepositResponse 1.1a: The apikey %s is not defined on this OKCatbox server.", apikey)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}

		// We'll need an HTTP client for the subsequent requests.
		timeout := 5000 * time.Millisecond
		client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

		// 1.2 currency_symbol
		currency_symbol := req.FormValue("currency_symbol")
		currency_id, err := getCurrencyBySym(client, currency_symbol, cfg)
		if err != nil {
			s := fmt.Sprintf("deposit.go generateCatboxDepositResponse 1.2: The currency_symbol %s is not defined on this OKCatbox server.", currency_symbol)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}

		// 1.3 Quantity
		quanf := req.FormValue("quan")
		quand, err := decimal.NewFromString(quanf)
		if err != nil {
			s := fmt.Sprintf("deposit.go 1.3 The quan %s cannot be parsed.", quanf)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
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

		// 1.4 Time.  Just a string, no validation.
		time := req.FormValue("time")

		// 2.  Get the list of all accounts along with joined currency and category info.  We'll need to search this later.
		accounts, err := getAccounts(client, cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return
		}

		// 3.  Get the account_id of the Hot Wallet for this currency.  It must exist so error if not found.
		account_id_hot, err := getHotWalletAccountID(client, accounts, currency_symbol, cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return
		}

		// 4. Get the account_id of the funding account for this user.  It might not exist yet so create it if necessary. Either way return the account_id
		account_id_api, err := getFundingAccountID(client, accounts, ok_access_key8, currency_id, currency_symbol, cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return
		}

		// 5. Now create the transaction and the two distributions using three requests.

		// 5.1 Create the tx
		txid, err := createTransaction(client, time, cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return
		}

		// 5.2 Create the DR distribution
		_, err = createDistribution(client, account_id_hot, dramt, exp, txid, cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return
		}

		// 5.3 Create the CR distribution
		_, err = createDistribution(client, account_id_api, cramt, exp, txid, cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return
		}
		return []byte("success")

	} else {
		log.Printf("%s %s %s %s", req.Method, req.URL, req.Header)
		return []byte("use post")
	}

}
