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

func fundingDepositAddressHandler(w http.ResponseWriter, req *http.Request, cfg Config) {
	retVal := generateDepositAddressResponse(w, req, cfg)
	_, _ = fmt.Fprintf(w, string(retVal))
}

func generateDepositAddressResponse(w http.ResponseWriter, req *http.Request, cfg Config) []byte {

	// This pattern of error handling will return an expected server response.  Don't change this.
	retVal, errb := checkSigHeaders(w, req)
	if errb {
		return retVal
	}

	// Check for the existence of a currency.
	currencies, ok := req.URL.Query()["currency"]
	if !ok {
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": ""})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30023("currency cannot be blank"))
		return retVal
	}

	// We'll need an HTTP client for the subsequent requests.
	timeout := 5000 * time.Millisecond
	httpClient := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	// Ensure that it's a valid currency.
	currencySymbol := currencies[0] // we know there must be at least one currency.  Only care about the first one.
	_, err := getCurrencyBySym(httpClient, currencySymbol, cfg)
	if err != nil {
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": ""})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30031(currencySymbol)) // invalid param
		return retVal
	}

	// The request is fully validated.
	setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": ""})

	depositAddresses := make([]utils.DepositAddress, 0)
	depositAddresses = append(depositAddresses, utils.DepositAddress{Address: "deposit address", CurrencyID: currencySymbol, To: 6})

	retVal, err = json.Marshal(depositAddresses)
	if err != nil {
		log.Error(err)
		return []byte(err.Error())
	}

	return retVal
}
