package main

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	uuid "github.com/nu7hatch/gouuid"
	"log"
	"net/http"
)

// /catbox/credentials
func catbox_credentialsHandler(w http.ResponseWriter, req *http.Request) {

	// 1. Generate the credentials
	u4, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("error: %v", err)
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
	}
	txn.Commit()

	// 3.
	retVal, _ := json.Marshal(credentials)
	fmt.Fprintf(w, string(retVal))
}
