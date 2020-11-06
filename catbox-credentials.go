package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"github.com/gojektech/heimdall/httpclient"
	uuid "github.com/nu7hatch/gouuid"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

const letterBytes = "ABCDEF0123456789"
const (
	letterIdxBits = 4                    // 4 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// /catbox/credentials
/* Warning! This is not very secure at all!  Anybody can request new credentials using an existing user's name. */
func catboxCredentialsHandler(w http.ResponseWriter, req *http.Request, cfg Config) {

	log.Printf("%s %s %s", req.Method, req.URL, req.Header)
	methodName := "okcatbox:catbox-credentials.go:catboxCredentialsHandler"

	if req.Method == http.MethodPost {

		type CredentialsRequestBody struct {
			UserID string
			Type   string
		}

		// 1. Read the body first because we may need to use it twice.
		reqBody, err := ioutil.ReadAll(req.Body)
		if s, errb := squeal(methodName, "ioutil.ReadAll", err); errb {
			_, _ = fmt.Fprintf(w, s)
			return
		}

		// 2. Parse the request body into JSON.
		credentialsRequestBody := CredentialsRequestBody{}
		dec := json.NewDecoder(bytes.NewReader(reqBody))
		err = dec.Decode(&credentialsRequestBody)
		if s, errb := squealJSONDecode(methodName, reqBody, err); errb {
			_, _ = fmt.Fprintf(w, s)
			return
		}

		// 3. Validate the input params from the request

		// 3.1 credentialsType
		credentialsType := credentialsRequestBody.Type
		switch credentialsType {
		case "read", "read-trade", "read-withdraw": // must be one of these
		default:
			s := fmt.Sprintf("%s: type must be read, read-trade, or read-withdraw\n", methodName)
			_, _ = fmt.Fprintf(w, s)
			return
		}

		// 3.2 userID
		userID := credentialsRequestBody.UserID
		if userID == "" {
			s := fmt.Sprintf("%s: you must specify a user_id\n", methodName)
			_, _ = fmt.Fprintf(w, s)
			return
		}

		// 4. Generate the credentials
		u4, err := uuid.NewV4()
		if s, errb := squeal(methodName, "uuid.NewV4", err); errb {
			_, _ = fmt.Fprintf(w, s)
			return
		}

		var tryArr = make([]string, 1)
		for i := range tryArr {
			tryArr[i] = RandStringBytesMaskImprSrc(32)
		}

		credentials := &utils.Credentials{Key: u4.String(), SecretKey: tryArr[0], Passphrase: "valid passphrase",
			Type: credentialsType, UserID: userID}

		// 5. Save the credentials in the in-memory db.
		txn := db.Txn(true)
		err = txn.Insert("credentials", credentials)
		if s, errb := squeal(methodName, "txn.Insert", err); errb {
			_, _ = fmt.Fprintf(w, s)
			return
		}

		txn.Commit()

		// 6. Now update bookwerx with a category for the user's userID.  We will create accounts and tag them with this category, later, as necessary.

		// 6.1 We'll need an HTTP client for the subsequent requests.
		timeout := 5000 * time.Millisecond
		httpClient := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

		// 6.2 Make a new category for this user's userID.
		// It's tempting to save the newly created category_id in the in-memory db now because we need it later.  However, doing so
		// would require a fair bit of re-engineering and we can always query bookwerx to get this category_id when we need it.
		// So therefore let's just take the ez path for now.
		_, err = postCategory(userID, userID, httpClient, cfg)
		if s, errb := squeal(methodName, "postCategory", err); errb {
			_, _ = fmt.Fprintf(w, s)
			return
		}

		// 7. All done!
		retVal, err := json.Marshal(credentials)
		if s, errb := squealJSONMarshal(methodName, credentials, err); errb {
			_, _ = fmt.Fprintf(w, s)
			return
		}

		_, _ = fmt.Fprintf(w, "%s\n", string(retVal))
	} else {
		_, _ = fmt.Fprintf(w, "use POST not GET")
	}
}
