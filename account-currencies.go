package main

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func accountCurrenciesHandler(w http.ResponseWriter, req *http.Request) {
	errb := checkSigHeaders(w, req)
	if errb {
		return
	}
	retVal := generateCurrenciesResponse(req)
	_, _ = fmt.Fprintf(w, string(retVal))
}

func generateCurrenciesResponse(req *http.Request) []byte {

	log.Printf("%s %s %s", req.Method, req.URL, req.Header)
	methodName := "okcatbox:account-currencies.go:generateCurrenciesResponse"

	//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": "", "Vary": ""})

	// 1. Get all of the currencies defined in the in-memory db
	txn := db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("currencies", "id") // id is an alias for CurrencyID
	if err != nil {
		s := fmt.Sprintf("%s: txn.Get error :%v", methodName, err)
		log.Error(s)
		return []byte(s)
	}

	var currencies []utils.CurrenciesEntry

	for obj := it.Next(); obj != nil; obj = it.Next() {
		p := obj.(*utils.CurrenciesEntry)
		currencies = append(currencies, *p)
	}

	// 2. Return the results.
	retVal, err := json.Marshal(currencies)
	if err != nil {
		s := fmt.Sprintf("%s: json.Marshal error: %v\ncurrencies=%+v\n", methodName, err, currencies)
		log.Error(s)
		return []byte(s)
	}

	return retVal

}
