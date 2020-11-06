package main

import (
	"fmt"
	utils "github.com/bostontrader/okcommon"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

// Various random things that are widely used.
// Given a response object, read the body and return it as a string.  Deal with the error message if necessary.
func bodyString(resp *http.Response) string {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("okcatbox:bookwerx-api.go:body_string :%v", err)
	}
	return string(body)
}

/* The bookwerx-core server will on occasion return JSON names that contain a '.'.  This vile habit causes trouble here. A good, bad, or ugly hack is to simply change the . to a -.  Do that here.
 */
func fixDot(b []byte) {
	for i, num := range b {
		if num == 46 { // .
			b[i] = 45 // -
		}
	}
}

func getOKAccessKey(headers map[string][]string) string {
	if value, ok := headers["Ok-Access-Key"]; ok {
		return value[0][:8]
	}
	return ""
}

// Given the ok access key, determine the user_id
func getUserId(headers map[string][]string) string {
	if apikey, ok := headers["Ok-Access-Key"]; ok {
		// Is this key defined with this OKCatbox server?
		txn := db.Txn(false)
		defer txn.Abort()

		raw, err := txn.First("credentials", "id", apikey[0])
		if err != nil {
			s := fmt.Sprintf("wallet.go:getUserId:", err)
			log.Error(s)
			return "error"
		}

		if raw == nil {
			s := fmt.Sprintf("wallet.go:getUserId: The apikey %s is not defined on this OKCatbox server.", apikey)
			log.Error(s)
			return "error"
		}
		return raw.(*utils.Credentials).UserID

	}
	return ""
}

/*
The squealer functions are conveniences to reduce the boilerplate required to handle the myriad of errors that we encounter as well as to make said handling more consistent.  Each squealer will return false if there's no error or true if there is.  If there is an error the squealer will also build an error message, log it, and return it.  In this way all the error-handling boilerplate can be reduce to a single line.

In order to localize the error a squealer gets an argument methodName.  The squealor also gets other arguments as necessary in order to better describe the error.
*/
func squeal(methodName, note string, err error) (errMsg string, errb bool) {
	errb = false
	if err != nil {
		errMsg := fmt.Sprintf("%s:%s error: %+v\n", methodName, note, err)
		log.Error(errMsg)
		errb = true
	}

	return
}

/* Squealor for JSON Decode errors.  The input argument provides the purported JSON data that couldn't be decoded.  */
func squealJSONDecode(methodName string, input []byte, err error) (errMsg string, errb bool) {
	errb = false
	if err != nil {
		errMsg := fmt.Sprintf("%s:JSON Decode error: %+v\ninput=%s\n", methodName, string(input), err)
		log.Error(errMsg)
		errb = true
	}

	return
}

/* Squealor for JSON Marshall errors.  The input argument provides the purported JSON data that couldn't be marshalled. */
func squealJSONMarshal(methodName string, input interface{}, err error) (errMsg string, errb bool) {
	errb = false
	if err != nil {
		errMsg := fmt.Sprintf("%s:JSON Marshal error: %+v\ninput=%+v\n", methodName, input, err)
		log.Error(errMsg)
		errb = true
	}

	return
}
