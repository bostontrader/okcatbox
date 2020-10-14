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

	if req.Method == http.MethodPost {

		type CredentialsRequestBody struct {
			UserID string `json:"user_id"`
			Type   string
		}

		type User struct {
			UserID string
		}

		// 1. Parse the body into JSON.
		credentialsRequestBody := CredentialsRequestBody{}
		dec := json.NewDecoder(req.Body)
		err := dec.Decode(&credentialsRequestBody)
		if err != nil {
			s := fmt.Sprintf("credentials.go:catbox_credentialsHandler: cannot JSON decode request body %v\n", err)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}

		// 2. Validate the input params from the body

		// 2.1 credentialsType
		credentialsType := credentialsRequestBody.Type
		switch credentialsType {
		case "read", "read-trade", "read-withdraw":
		default:
			s := fmt.Sprintf("credentials.go:catbox_credentialsHandler: type must be read, read-trade, or read-withdraw\n")
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}

		// 2.2 userID
		userID := credentialsRequestBody.UserID
		if userID == "" {
			s := fmt.Sprintf("credentials.go:catbox_credentialsHandler: you must specify a user_id\n")
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}

		// 3. Generate the credentials
		u4, err := uuid.NewV4()
		if err != nil {
			s := fmt.Sprintf("credentials.go:catbox_credentialsHandler: uuid error: %v\n", err)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}

		var tryArr = make([]string, 1)
		for i := range tryArr {
			tryArr[i] = RandStringBytesMaskImprSrc(32)
		}

		credentials := &utils.Credentials{Key: u4.String(), SecretKey: tryArr[0], Passphrase: "valid passphrase",
			Type: credentialsType, UserID: userID}

		// 4. Save the credentials in the in-memory db.
		txn := db.Txn(true)
		if err := txn.Insert("credentials", credentials); err != nil {
			s := fmt.Sprintf("credentials.go:catbox_credentialsHandler: credentials save error: %v\n", err)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}
		txn.Commit()

		// 5. Now update bookwerx with a category for the user's userID.  We will create accounts and tag them with this category, later, if necessary.

		// 5.1 We'll need an HTTP client for the subsequent requests.
		timeout := 5000 * time.Millisecond
		client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

		// 5.2 Make a new category for this user's userID.
		// It's tempting to save the newly created category_id in the in-memory db now because we need it later.  However, doing so
		// would require a fair bit of re-engineering and we can always query bookwerx to get this category_id when we need it.
		// So therefore let's just take the ez path for now.
		_, err = postCategory(
			client,
			cfg.Bookwerx.Server,
			cfg.Bookwerx.APIKey, // The apikey that cb uses with bw
			userID)              // The new category will use this as title and symbol
		if err != nil {
			log.Fatalf("error: %v", err)
			return
		}

		// 5.
		retVal, _ := json.Marshal(credentials) // Ignore the error, assume this always works.
		fmt.Fprintf(w, "%s\n", string(retVal))
	} else {
		fmt.Fprintf(w, "use POST not GET")
	}
}
