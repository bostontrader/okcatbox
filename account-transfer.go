package main

import (
	"fmt"
	"net/http"
)

// /api/account/v3/transfer
func account_transferHandler(w http.ResponseWriter, req *http.Request, cfg Config) {
	retVal := generateTransferResponse(w, req, cfg)
	fmt.Fprintf(w, string(retVal))
}

type TransferRequestBody struct {
	CurrencySymbol string `json:"currency"`
	From           string
	Amount         string
	To             string
}

type AccountTransferResult struct {
	TransferID     string `json:"transfer_id"`
	CurrencySymbol string `json:"currency"`
	From           string
	Amount         string
	To             string
	Result         bool
}

/*
This endpoint requires a POST, with Content-Type application/json, with a body that can be parsed into JSON.
It also requires the standard 4 credential headers.

There are many errors that can come from badly constructed requests to this endpoint.  Such as:

GET to this endpoint.  405 Method not allowed.
POST w/o json header and nil body.  30001 OK-ACCESS-KEY header is required
POST w/json header and nil body. 500 Internal Sever Error.
POST w/o json header and body. 30001
POST w/json header and body that is not parseable into json. 400 Bad Request. Invalid JSON format.

How the real OKEx server reacts to these errors is of minimal importance so we won't worry about them here nor
will we test for them.

*/
func generateTransferResponse(w http.ResponseWriter, req *http.Request, cfg Config) (retVal []byte) {

	/*if req.Method == http.MethodPost {

		// 1. Check the standard signature headers related to credentials
		retVal, errb := checkSigHeaders(w, req)
		if errb {
			return retVal
		}

		// 2. Parse the body into JSON.
		transferRequestBody := TransferRequestBody{}
		dec := json.NewDecoder(req.Body)
		err := dec.Decode(&transferRequestBody)
		if err != nil {
			s := fmt.Sprintf("account-transfer.go:generateTransferResponse: cannot decode request body", err)
			log.Error(s)
			//fmt.Fprintf(w, s)
			retVal, _ = json.Marshal(utils.Err30001()) // Access key required
			return retVal
		}

		// 3. Validate the OKCatbox apikey.
		apikey := req.Header["Ok-Access-Key"][0]
		ok_access_key8 := apikey
		if len(apikey) >= 8 {
			ok_access_key8 = apikey[:8]
		}

		// Is this key defined with this OKCatbox server?
		txn := db.Txn(false)
		defer txn.Abort()

		raw, err := txn.First("credentials", "id", apikey)
		if err != nil {
			s := fmt.Sprintf("account-transfer.go:generateTransferResponse 1.1:", err)
			log.Error(s)
			fmt.Fprintf(w, s)
			return []byte("catfood")
		}

		if raw == nil {
			s := fmt.Sprintf("account-transfer.go:generateTransferResponse 1.1a: The apikey %s is not defined on this OKCatbox server.", apikey)
			log.Error(s)
			fmt.Fprintf(w, s)
			return []byte("catfood")
		}

		// We'll need an HTTP client for the subsequent requests.
		timeout := 5000 * time.Millisecond
		client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

		// 1.2 Given the request currency symbol, what is the currency_id on bookwerx?
		currency_id, err := getCurrencyBySym(client, transferRequestBody.CurrencySymbol, cfg)
		if err != nil {
			s := fmt.Sprintf("account-transfer.go:generateTransferResponse 1.2: The currency_symbol %s is not defined on this OKCatbox server.", transferRequestBody.CurrencySymbol)
			log.Error(s)
			fmt.Fprintf(w, s)
			return []byte("catfood")
		}

		// 1.3 Quantity
		quand, err := decimal.NewFromString(transferRequestBody.Amount)
		if err != nil {
			s := fmt.Sprintf("account-transfer.go:generateTransferResponse 1.3: The quan %s cannot be parsed.", transferRequestBody.Amount)
			log.Error(s)
			fmt.Fprintf(w, s)
			return []byte("catfood")
		}

		quans := quand.Abs().Coefficient().String()

		exp := fmt.Sprint(quand.Exponent())
		var dramt, cramt string
		if quand.IsPositive() {
			dramt = quans
			cramt = "-" + dramt
		} else {
			cramt = quans
			dramt = "-" + cramt
		}

		// 1.4 From and To
		// Validate the source and destination code and determine the relevant bookwerx categories to use.
		catSource := cfg.Bookwerx.TransferCats[transferRequestBody.From]
		catDest := cfg.Bookwerx.TransferCats[transferRequestBody.To]

		// 3. Get the account_id of the account that is tagged with the from-available category, the user's apikey8, and
		// configured to use the given currency.  At this point, both of the categories and the given currency are guaranteed to exist
		// but the account might not exist.  So create the account if necessary. Either way return the account_id.
		account_id_from, err := getTransferAccountID(client, catSource.Available, ok_access_key8, currency_id, cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return []byte("catfood")
		}

		// 4. Do the same for the account that is tagged with the to-available category.
		account_id_to, err := getTransferAccountID(client, catDest.Available, ok_access_key8, currency_id, cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return []byte("catfood")
		}

		// 5. Now create the transaction on the OKCatbox books with its associated two distributions.

		// 5.1 Create the tx
		time := time.Now()
		txid, err := createTransaction(client, time.Format("2006-01-02T15:04:05.999Z"), cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return []byte("catfood")
		}

		// 5.2 Create the DR distribution
		_, err = createDistribution(client, account_id_from, dramt, exp, txid, cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return []byte("catfood")
		}

		// 5.3 Create the CR distribution
		_, err = createDistribution(client, account_id_to, cramt, exp, txid, cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return []byte("catfood")
		}

		// 6. Build the result and go home
		result := AccountTransferResult{
			TransferID:     time.String(),
			CurrencySymbol: transferRequestBody.CurrencySymbol,
			From:           transferRequestBody.From,
			Amount:         transferRequestBody.Amount,
			To:             transferRequestBody.To,
			Result:         true,
		}
		fmt.Println(result)
		// encode the result into a []byte and return that
		retVal, _ = json.Marshal(result)
		return []byte("success")

	} else {
		log.Printf("%s %s %s", req.Method, req.URL, req.Header)
		return []byte("use post")
	}*/
	return []byte("todo")
}
