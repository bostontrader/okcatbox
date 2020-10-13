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

func account_depositHistoryHandler(w http.ResponseWriter, req *http.Request, cfg Config, currency string) {
	retVal := generateAccountDepositHistoryResponse(w, req, cfg, currency)
	_, _ = fmt.Fprintf(w, string(retVal))
}

func generateAccountDepositHistoryResponse(w http.ResponseWriter, req *http.Request, cfg Config, currency_symbol string) []byte {

	// 1.
	retVal, err := checkSigHeaders(w, req)
	if err {
		return retVal
	}

	// 2. We'll need an HTTP client for the subsequent requests.
	timeout := 5000 * time.Millisecond
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	// 3.
	if currency_symbol == "" {
		// 3.1 Find all accounts marked as funding for this api credential

		//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": "", "Vary": ""})
		depositHistories := make([]utils.DepositHistory, 1)
		depositHistories[0] = utils.DepositHistory{Amount: "amount", TXID: "txid", CurrencyID: "currency", From: "from", To: "to", DepositID: 666, Timestamp: "timestamp", Status: "status"}
		retVal, _ = json.Marshal(depositHistories)
		return retVal

	} else {
		// If the currency is legit
		_, err := getCurrencyBySym(client, currency_symbol, cfg)
		if err != nil {
			s := fmt.Sprintf("account-deposit-history.go generateAccountDepositHistoryResponse: The currency_symbol %s is not defined on this OKCatbox server.", currency_symbol)
			log.Error(s)
			//fmt.Fprintf(w, s)
			//return []byte{}
			w.WriteHeader(http.StatusBadRequest)
			retVal, _ = json.Marshal(utils.Err30031(currency_symbol))
			return retVal
		}

		depositHistories := make([]utils.DepositHistory, 1)
		depositHistories[0] = utils.DepositHistory{Amount: "amount", TXID: "txid", CurrencyID: "currency", From: "from", To: "to", DepositID: 666, Timestamp: "timestamp", Status: "status"}
		retVal, _ = json.Marshal(depositHistories)
		return retVal

		// find all accounts marked as funding for this api credentials and with this currency
		// else
		// error bad currency

		// find all CR distributions
		// return them all
	}

	return []byte{}

}
