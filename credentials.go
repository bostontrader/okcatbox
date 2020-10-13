package main

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"github.com/gojektech/heimdall/httpclient"
	uuid "github.com/nu7hatch/gouuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// /catbox/credentials
func catbox_credentialsHandler(w http.ResponseWriter, req *http.Request, cfg Config) {

	log.Printf("%s %s %s", req.Method, req.URL, req.Header)

	// 1. Generate the credentials
	u4, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("error: %v", err)
		return
	}

	var tryArr = make([]string, 1)
	for i := range tryArr {
		tryArr[i] = RandStringBytesMaskImprSrc(32)
	}

	credentials := &utils.Credentials{Key: u4.String(), SecretKey: tryArr[0], Passphrase: "valid passphrase"}

	// 2. Save the credentials in the in-memory db
	txn := db.Txn(true)
	if err := txn.Insert("credentials", credentials); err != nil {
		log.Fatalf("error: %v", err)
		return
	}
	txn.Commit()

	// 3. Now update bookwerx with a category for the user's okcatbox apikey.  We will create accounts and tag them with categories, later, if necessary.

	// 3.1 We'll need an HTTP client for the subsequent requests.
	timeout := 5000 * time.Millisecond
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	// 3.2 Make a new category for this user's apikey.
	// It's tempting to save the newly created category_id in the in-memory db now because we need it later.  However, doing so would require a fair bit of re-engineering and we can always query bookwerx to get this category_id when we need it.  So therefore let's just take the ez path for now.
	_, err = postCategory(
		client,
		cfg.Bookwerx.Server,
		cfg.Bookwerx.APIKey, // The apikey that cb uses with bw
		credentials.Key[:8]) // The apikey that the user uses with cb
	if err != nil {
		log.Fatalf("error: %v", err)
		return
	}

	// 4.
	retVal, _ := json.Marshal(credentials)
	fmt.Fprintf(w, string(retVal))
}
