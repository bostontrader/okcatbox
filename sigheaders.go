package main

import (
	"encoding/json"
	utils "github.com/bostontrader/okcommon"
	"net/http"
)

/* Examine the request headers and validate the four signature headers.
If there is an error with any of these headers, build a suitable error message, output to the ResponseWriter with a suitable status code, and return err=True.
If there is no error, merely return err = False
*/
func checkSigHeaders(w http.ResponseWriter, req *http.Request) (errb bool) {

	// Assume we have an error.
	// Only if none of these conditions apply will we set this to false.
	errb = true

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
		//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		retVal, _ := json.Marshal(utils.Err30001()) // Access key required
		w.WriteHeader(401)
		_, _ = w.Write(retVal)

	} else if !asignP {
		//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		retVal, _ := json.Marshal(utils.Err30002()) // Signature required
		w.WriteHeader(400)
		_, _ = w.Write(retVal)

	} else if !atimestampP {
		//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		retVal, _ := json.Marshal(utils.Err30003()) // Timestamp required
		w.WriteHeader(400)
		_, _ = w.Write(retVal)

	} else if !atimestampV {
		//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		retVal, _ := json.Marshal(utils.Err30005()) // Invalid timestamp
		w.WriteHeader(400)
		_, _ = w.Write(retVal)

	} else if atimestampEx {
		//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		retVal, _ := json.Marshal(utils.Err30008()) // Timestamp expired
		w.WriteHeader(400)
		_, _ = w.Write(retVal)

	} else if !akeyV {
		//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		retVal, _ := json.Marshal(utils.Err30006()) // Invalid access key
		w.WriteHeader(401)
		_, _ = w.Write(retVal)

	} else if !apassphraseP {
		//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		retVal, _ := json.Marshal(utils.Err30004()) // Passphrase required
		w.WriteHeader(400)
		_, _ = w.Write(retVal)

	} else if !apassphraseV {
		//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		retVal, _ := json.Marshal(utils.Err30015()) // Invalid Passphrase
		w.WriteHeader(400)
		_, _ = w.Write(retVal)

	} else if !asignV {
		//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		retVal, _ := json.Marshal(utils.Err30013()) // Invalid Sign
		w.WriteHeader(401)
		_, _ = w.Write(retVal)

	} else {
		errb = false
	}

	return
}
