package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"github.com/gojektech/heimdall/httpclient"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

// /catbox/deposit
func catboxDepositHandler(w http.ResponseWriter, req *http.Request, cfg Config) {
	retVal := generateCatboxDepositResponse(req, cfg)
	_, _ = fmt.Fprintf(w, string(retVal))
}

func generateCatboxDepositResponse(req *http.Request, cfg Config) (retVal []byte) {

	log.Printf("%s %s %s", req.Method, req.URL, req.Header)
	methodName := "okcatbox:catbox-deposit.go:generateCatboxDepositResponse"

	if req.Method == http.MethodPost {

		type DepositRequestBody struct {
			Apikey         string
			CurrencySymbol string
			Quan           string
			Time           string
		}

		// 1. Read the body first because we may need to use it twice.
		reqBody, err := ioutil.ReadAll(req.Body)
		if s, errb := squeal(methodName, "ioutil.ReadAll", err); errb {
			return []byte(s)
		}

		// 2. Parse the request body into JSON.
		depositRequestBody := DepositRequestBody{}
		err = json.NewDecoder(bytes.NewReader(reqBody)).Decode(&depositRequestBody)
		if s, errb := squealJSONDecode(methodName, reqBody, err); errb {
			return []byte(s)
		}

		// 3. Validate the input params from the request

		// 3.1 OKCatbox apikey.
		apikey := depositRequestBody.Apikey

		// Is this key defined with this OKCatbox server?
		txn := db.Txn(false)
		defer txn.Abort()

		raw, err := txn.First("credentials", "id", apikey)
		if s, errb := squeal(methodName, "txn.First", err); errb {
			return []byte(s)
		}

		if raw == nil {
			s := fmt.Sprintf("%s: The apikey %s is not defined on this OKCatbox server.\n", methodName, apikey)
			return []byte(s)
		}

		// We'll need an HTTP client for the subsequent requests.
		timeout := 5000 * time.Millisecond
		httpClient := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

		// 3.2 currencySymbol
		currencySymbol := depositRequestBody.CurrencySymbol
		currencyID, err := getCurrencyBySym(httpClient, currencySymbol, cfg)
		if s, errb := squeal(methodName, fmt.Sprintf("The currency %s is not defined on this server.", currencySymbol), err); errb {
			return []byte(s)
		}

		// 3.3 Quantity
		quanf := depositRequestBody.Quan
		quand, err := decimal.NewFromString(quanf)
		if s, errb := squeal(methodName, fmt.Sprintf("I cannot parse %s", quanf), err); errb {
			return []byte(s)
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

		// 3.4 Time.  Just a string, no validation.

		// 4. Get the one accountID of the Hot Wallet for this currency.
		// Said account is:
		// A. Tagged with whatever category corresponds with hot wallet.
		// B. Configured to use the specified currency.
		hotWalletAccountID, err := accountDepositHistoryGetHotWalletAccountID(currencySymbol, httpClient, cfg)
		if s, errb := squeal(methodName, fmt.Sprintf("No hot wallet found for currency %s\n", currencySymbol), err); errb {
			return []byte(s)
		}

		// 5. Get the categoryID for this user.  Said category is created if necessary when the user obtains credentials so we know it must exist.
		userID := raw.(*utils.Credentials).UserID
		userCategoryID, err := getCategoryBySym(userID, httpClient, cfg)
		if s, errb := squeal(methodName, fmt.Sprintf("No category found for user %s\n", userID), err); errb {
			return []byte(s)
		}

		// 6. Get the accountID for whichever account is tagged as funding for this user and is configured for the specified currency.
		fundingAvailableAccountID, err := getFundingAvailableAccountID(userCategoryID, cfg.Bookwerx.CatFunding, currencyID, httpClient, cfg)
		if s, errb := squeal(methodName, fmt.Sprintf("No funding available account found for user %s and currency %s\n", userID, currencySymbol), err); errb {
			return []byte(s)
		}

		// 7. If said account does not exist, create it now and tag it suitably.
		if fundingAvailableAccountID == 0 {
			// 7.1 Account not found, create it
			fundingAvailableAccountID, err = postAccount(currencyID, userID, httpClient, cfg)
			if s, errb := squeal(methodName, "postAccount", err); errb {
				return []byte(s)
			}

			// 7.2 Tag with the user's category
			_, err = postAcctcat(fundingAvailableAccountID, userCategoryID, httpClient, cfg)
			if s, errb := squeal(methodName, "postAcctcat user", err); errb {
				return []byte(s)
			}

			// 7.3 Tag with funding
			_, err = postAcctcat(fundingAvailableAccountID, cfg.Bookwerx.CatFunding, httpClient, cfg)
			if s, errb := squeal(methodName, "postAcctcat funding", err); errb {
				return []byte(s)
			}

		}

		// 8. Now create the transaction and associated distributions and tag it as a deposit for this user.

		// 8.1 Create the tx
		txid, err := postTransaction("deposit", depositRequestBody.Time, httpClient, cfg)
		if s, errb := squeal(methodName, "postTransaction", err); errb {
			return []byte(s)
		}

		// 8.2 Create the DR distribution
		_, err = postDistribution(hotWalletAccountID, dramt, exp, txid, httpClient, cfg)
		if s, errb := squeal(methodName, "postDistribution DR", err); errb {
			return []byte(s)
		}

		// 8.3 Create the CR distribution
		_, err = postDistribution(fundingAvailableAccountID, cramt, exp, txid, httpClient, cfg)
		if s, errb := squeal(methodName, "postDistribution CR", err); errb {
			return []byte(s)
		}

		// 8.4 Tag with the user's category
		_, err = postTrancat(txid, userCategoryID, httpClient, cfg)
		if s, errb := squeal(methodName, "postTrancat user", err); errb {
			return []byte(s)
		}

		// 8.5 Tag as a deposit
		_, err = postTrancat(txid, cfg.Bookwerx.CatDeposit, httpClient, cfg)
		if s, errb := squeal(methodName, "postTrancat deposit", err); errb {
			return []byte(s)
		}

		return []byte("success")

	} else {
		return []byte("use POST not GET")
	}

}
