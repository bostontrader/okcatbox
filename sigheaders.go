package main

import (
	"encoding/json"
	utils "github.com/bostontrader/okcommon"
	"net/http"
)

func checkSigHeaders(w http.ResponseWriter, req *http.Request) (retVal []byte, err bool) {

	// All of these conditions produce an error, so let's assume we have an error.
	// Only if none of these conditions apply will we set this to false.
	err = true

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
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30001()) // Access key required

	} else if !asignP {
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30002()) // Signature required

	} else if !atimestampP {
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30003()) // Timestamp required

	} else if !atimestampV {
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30005()) // Invalid timestamp

	} else if atimestampEx {
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30008()) // Timestamp expired

	} else if !akeyV {
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30006()) // Invalid access key

	} else if !apassphraseP {
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30004()) // Passphrase required

	} else if !apassphraseV {
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30015()) // Invalid Passphrase

	} else if !asignV {
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30013()) // Invalid Sign

	} else {
		err = false
	}

	return
}
