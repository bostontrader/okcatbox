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

	fmt.Println(req, "\n")

	// 1. Detect whether or not the various parameters exist.  If so, detect whether or not they are valid.  Set suitable flags so that subsequent error checking can occur.

	// 1.1 Ok-Access-Key
	akeyP, akeyV := validateAccessKey(req.Header)

	// 1.2 Ok-Access-Timestamp
	atimestampP, atimestampV, atimestampEx := validateTimestamp(req.Header)

	// 1.3 Ok-Access-Passphrase
	apassphraseP, apassphraseV := validatePassPhrase(req.Header)

	// 1.4 Ok-Access-Sign
	asignP, asignV := validateSign(req)

	// The order of comparison and the boolean senses have been empirically chosen to match the order of error detection in the real OKEx server.
	if !akeyP {
		setResponseHeaders(w, utils.ExpectedResponseHeadersB, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30001()) // Access key required

	} else if !asignP {
		setResponseHeaders(w, utils.ExpectedResponseHeadersB, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30002()) // Signature required

	} else if !atimestampP {
		setResponseHeaders(w, utils.ExpectedResponseHeadersB, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30003()) // Timestamp required

	} else if !atimestampV {
		setResponseHeaders(w, utils.ExpectedResponseHeadersB, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30005()) // Invalid timestamp

	} else if atimestampEx {
		setResponseHeaders(w, utils.ExpectedResponseHeadersB, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30008()) // Timestamp expired

	} else if !akeyV {
		setResponseHeaders(w, utils.ExpectedResponseHeadersB, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30006()) // Invalid access key

	} else if !apassphraseP {
		setResponseHeaders(w, utils.ExpectedResponseHeadersB, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30004()) // Passphrase required

	} else if !apassphraseV {
		setResponseHeaders(w, utils.ExpectedResponseHeadersB, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30015()) // Invalid Passphrase

	} else if !asignV {
		setResponseHeaders(w, utils.ExpectedResponseHeadersB, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30013()) // Invalid Sign

	} else {
		setResponseHeaders(w, utils.ExpectedResponseHeadersB, map[string]string{"Strict-Transport-Security": "", "Vary": ""})

		accountsEntries := make([]utils.AccountsEntry, 1)
		accountsEntries[0] = utils.AccountsEntry{AccountID: "aid", Available: "available", Balance: "balance", CurrencyID: "cid", Frozen: "frozen", Hold: "hold", Holds: "holds"}
		retVal, _ := json.Marshal(accountsEntries)
		return retVal
	}

	return
}
