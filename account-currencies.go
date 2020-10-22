package main

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func currencies(w http.ResponseWriter, req *http.Request) {
	retVal := generateCurrenciesResponse(w, req)
	_, _ = fmt.Fprintf(w, string(retVal))
}

func generateCurrenciesResponse(w http.ResponseWriter, req *http.Request) (retVal []byte) {

	retVal, err := checkSigHeaders(w, req)
	if err {
		return // ???
	} else {
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": "", "Vary": ""})

		// Create read-only transaction
		txn := db.Txn(false)
		defer txn.Abort()

		// Get all of the defined currencies
		it, err := txn.Get("currencies", "id") // id is an alias for CurrencyID
		if err != nil {
			log.Fatalf("error: %v", err) // This should never happen
		}

		var currencies []utils.CurrenciesEntry

		for obj := it.Next(); obj != nil; obj = it.Next() {
			p := obj.(*utils.CurrenciesEntry)
			currencies = append(currencies, *p)
		}

		retVal, _ := json.Marshal(currencies)
		return retVal
	}

}
