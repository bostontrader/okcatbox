package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	bw_api "github.com/bostontrader/bookwerx-common-go"
	utils "github.com/bostontrader/okcommon"
	"github.com/gojektech/heimdall/httpclient"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// /catbox/deposit
func catbox_depositHandler(w http.ResponseWriter, req *http.Request, cfg Config) {
	retVal := generateCatboxDepositResponse(w, req, cfg)
	fmt.Fprintf(w, string(retVal))
}

func generateCatboxDepositResponse(w http.ResponseWriter, req *http.Request, cfg Config) (retVal []byte) {

	log.Printf("%s %s %s %s", req.Method, req.URL, req.Header, req.Form)

	if req.Method == http.MethodPost {

		// 1. Retrieve and validate the request parameters.
		if err := req.ParseForm(); err != nil {
			s := fmt.Sprintf("catbox-deposit.go:generateCatboxDepositResponse: ParseForm err: %v\n", err)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}

		// 1.1 OKCatbox apikey.
		apikey := req.FormValue("apikey")

		// Is this key defined with this OKCatbox server?
		txn := db.Txn(false)
		defer txn.Abort()

		raw, err := txn.First("credentials", "id", apikey)
		if err != nil {
			s := fmt.Sprintf("catbox-deposit.go:generateCatboxDepositResponse: Error searching for APIKEY: %v\n", err)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}

		if raw == nil {
			s := fmt.Sprintf("catbox-deposit.go:generateCatboxDepositResponse: The apikey %s is not defined on this OKCatbox server.\n", apikey)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}

		// We'll need an HTTP client for the subsequent requests.
		timeout := 5000 * time.Millisecond
		client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

		// 1.2 currency_symbol
		currency_symbol := req.FormValue("currency_symbol")
		currency_id, err := getCurrencyBySym(client, currency_symbol, cfg)
		if err != nil {
			s := fmt.Sprintf("catbox-deposit.go:generateCatboxDepositResponse: The currency_symbol %s is not defined on this OKCatbox server.\n", currency_symbol)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}

		// 1.3 Quantity
		quanf := req.FormValue("quan")
		quand, err := decimal.NewFromString(quanf)
		if err != nil {
			s := fmt.Sprintf("catbox-deposit.go:generateCatboxDepositResponse: The quan %s cannot be parsed.", quanf)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
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

		// 1.4 Time.  Just a string, no validation.
		time := req.FormValue("time")

		// 2. Get the account_id of the Hot Wallet for this currency.  It must exist so error if not found.
		// Said account is:
		// A. Tagged with whatever category corresponds with hot wallet.
		// B. Configured to use the specified currency.
		selectt := "SELECT%20accounts.id"
		from := "FROM%20accounts_categories"
		join1 := "JOIN%20accounts%20ON%20accounts.id%3daccounts_categories.account_id"
		join2 := "JOIN%20currencies%20ON%20currencies.id%3daccounts.currency_id"
		where := fmt.Sprintf("WHERE%%20category_id%%3d%d%%20AND%%20currencies.symbol%%3d'%s'", cfg.Bookwerx.HotWalletCat, currency_symbol)
		query := fmt.Sprintf("%s%%20%s%%20%s%%20%s%%20%s", selectt, from, join1, join2, where)
		url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)

		body, err := bw_api.Get(client, url)
		if err != nil {
			s := fmt.Sprintf("catbox-deposit.go:generateCatboxDepositResponse: %v\n", err)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}
		fixDot(body)

		n1 := make([]AId, 0)
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&n1)
		if err != nil {
			s := fmt.Sprintf("catbox-deposit.go:generateCatboxDepositResponse: %v\n", err)
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}

		if len(n1) == 0 {
			s := fmt.Sprintf("catbox-deposit.go:generateCatboxDepositResponse: Bookwerx does not have any account properly configured.")
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		} else if len(n1) > 1 {
			s := fmt.Sprintf("catbox-deposit.go:generateCatboxDepositResponse: Bookwerx has more than one suitable account.  This should never happen.")
			log.Error(s)
			fmt.Fprintf(w, s)
			return
		}
		account_id_hot := n1[0].Id

		// 3. Get the category_id for the user.  Said category is created if necessary when the user obtains credentials.
		// So we know it must exist.
		url = fmt.Sprintf("%s/category/bysym/%s?apikey=%s", cfg.Bookwerx.Server, raw.(*utils.Credentials).UserID, cfg.Bookwerx.APIKey)
		user_category_id, err := getCategoryBySym(client, url)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return
		}

		// 4. Get the account_id of the funding account for this user and currency.  It might not exist yet so create it if necessary. Either way return the account_id.  Said account is:
		// A. tagged as funding
		// B. tagged as user
		// C. configured for currency
		selectt = "SELECT%20accounts.id"
		from = "FROM%20accounts_categories"
		join1 = "JOIN%20accounts%20ON%20accounts.id%3daccounts_categories.account_id"
		join2 = "JOIN%20currencies%20ON%20currencies.id%3daccounts.currency_id"
		where = fmt.Sprintf("WHERE%%20accounts_categories.category_id%%20IN%%20(%d,%d)",
			cfg.Bookwerx.FundingCat, user_category_id,
		)
		and := fmt.Sprintf("AND%%20currencies.id%%3d%d", currency_id)
		group := "GROUP%20BY%20accounts_categories.account_id%20HAVING%20COUNT(DISTINCT%20accounts_categories.account_id)%3d2"
		query = fmt.Sprintf("%s%%20%s%%20%s%%20%s%%20%s%%20%s%%20%s", selectt, from, join1, join2, where, and, group)
		url = fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)
		body, err = bw_api.Get(client, url)
		if err != nil {
			fmt.Println("catbox-deposit.go:generateCatboxDepositResponse 2: get error: ", err)
			return
		}
		fixDot(body)

		n2 := make([]AId, 0)
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&n1)
		if err != nil {
			fmt.Println("bookwerx-api.go: getTransferAccountID:", err)
			return
		}

		var account_id_user int32
		if len(n2) == 0 {
			// 4.1 Account not found, create it
			account_id_user, err = postAccount(client, currency_id, raw.(*utils.Credentials).UserID, cfg)
			if err != nil {
				fmt.Println("catbox-deposit.go:generateCatboxDepositResponse 3.1: postAccount error: ", err)
				return
			}

			// 4.2 Tag with the user's category
			_, err = postAcctcat(client, account_id_user, user_category_id, cfg)
			if err != nil {
				fmt.Println("catbox-deposit.go:generateCatboxDepositResponse 3.2: postAcctcat error: ", err)
			}

			// 4.3 Tag with funding
			_, err = postAcctcat(client, account_id_user, cfg.Bookwerx.FundingCat, cfg)
			if err != nil {
				fmt.Println("catbox-deposit.go:generateCatboxDepositResponse 3.3: postAcctcat error: ", err)
				return
			}
		} else if len(n2) == 1 {
			account_id_user = n2[0].Id // account found
		} else if len(n2) > 1 {
			fmt.Println("bookwerx-api.go: getTransferAccountID: There are more than one suitable accounts.  This should never happen.")
			return
		}

		// 5. Now create the transaction and associated distributions.

		// 5.1 Create the tx
		txid, err := createTransaction(client, time, cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return
		}

		// 5.2 Create the DR distribution
		_, err = createDistribution(client, account_id_hot, dramt, exp, txid, cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return
		}

		// 5.3 Create the CR distribution
		_, err = createDistribution(client, account_id_user, cramt, exp, txid, cfg)
		if err != nil {
			log.Error(err)
			fmt.Fprintf(w, err.Error())
			return
		}
		return []byte("success")

	} else {
		return []byte("use POST not GET")
	}

}
