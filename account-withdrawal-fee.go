package main

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// /api/account/v3/withdrawal/fee
func withdrawalFee(w http.ResponseWriter, req *http.Request) {
	retVal := generateWithdrawalFeeResponse(w, req, "GET", "/api/account/v3/withdrawal/fee")
	fmt.Fprintf(w, string(retVal))
}

func generateWithdrawalFeeResponse(w http.ResponseWriter, req *http.Request, verb string, endpoint string) (retVal []byte) {

	retVal, err := checkSigHeaders(w, req)
	if err {
		return

	}

	// Check for the existence of a currency.
	currencies, ok := req.URL.Query()["currency"]
	if ok {
		// Currency specified in the query string.  Is it valid?
		txn := db.Txn(false)
		defer txn.Abort()
		raw, err := txn.First("withdrawalFees", "id", currencies[0])
		if err != nil {
			log.Fatalf("error: %v", err) // This should never happen
			return
		}
		if raw == nil {
			// Currency not valid, return 400 error
			//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": ""})
			w.WriteHeader(400)
			retVal, _ = json.Marshal(utils.Err30031(currencies[0]))
			return retVal
		} else {
			// The currency is valid, return just that record.
			retVal, _ = json.Marshal(raw)
			return retVal
		}

	} else {
		// No currency specified in the query string.  Return all records.
		// Create read-only transaction
		txn := db.Txn(false)
		defer txn.Abort()

		// Get all of the defined withdrawlFee records
		it, err := txn.Get("withdrawalFees", "id") // id is an alias for CurrencyID
		if err != nil {
			log.Fatalf("error: %v", err) // This should never happen
			return
		}

		var withdrawalFees []utils.WithdrawalFee

		for obj := it.Next(); obj != nil; obj = it.Next() {
			p := obj.(*utils.WithdrawalFee)
			withdrawalFees = append(withdrawalFees, *p)
		}

		retVal, _ = json.Marshal(withdrawalFees)
		return retVal
	}

	return
}
