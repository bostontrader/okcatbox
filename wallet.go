package main

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"net/http"
)

func walletHandler(w http.ResponseWriter, req *http.Request) {
	retVal := generateWalletResponse(w, req, "GET", "/api/account/v3/wallet")
	fmt.Fprintf(w, string(retVal))
}

func generateWalletResponse(w http.ResponseWriter, req *http.Request, verb string, endpoint string) (retVal []byte) {

	fmt.Println(req, "\n")

	retVal, err := checkSigHeaders(w, req)
	if err {
		return

	} else {
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": ""})

		walletEntries := make([]utils.WalletEntry, 0)
		//walletEntries[0] = utils.WalletEntry{Available: "a", Balance: "b", CurrencyID: "c", Hold: "h"}
		retVal, _ := json.Marshal(walletEntries)
		return retVal
	}

	return
}
