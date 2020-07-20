package main

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"net/http"
)

func accountsHandler(w http.ResponseWriter, req *http.Request) {
	retVal := generateAccountsResponse(w, req, "GET", "/api/spot/v3/accounts")
	fmt.Fprintf(w, string(retVal))
}

func generateAccountsResponse(w http.ResponseWriter, req *http.Request, verb string, endpoint string) (retVal []byte) {

	retVal, err := checkSigHeaders(w, req)
	if err {
		return

	} else {
		setResponseHeaders(w, utils.ExpectedResponseHeadersB, map[string]string{"Strict-Transport-Security": "", "Vary": ""})

		accountsEntries := make([]utils.AccountsEntry, 1)
		accountsEntries[0] = utils.AccountsEntry{AccountID: "aid", Available: "available", Balance: "balance", CurrencyID: "cid", Frozen: "frozen", Hold: "hold", Holds: "holds"}
		retVal, _ := json.Marshal(accountsEntries)
		return retVal
	}

	return
}
