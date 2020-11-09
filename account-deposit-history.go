package main

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"github.com/gojektech/heimdall/httpclient"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Distributions struct {
	CurrencySymbol string `json:"currencies.symbol"`
	Amount         int64  `json:"distributions.amount"`
	AmountExp      int8   `json:"distributions.amount_exp"`
	TransactionID  uint32 `json:"distributions.transaction_id"`
	Time           string `json:"transactions.time"`
}

/* Two endpoints use this handler:
/api/account/v3/deposit/history    // for all currencies
/api/account/v3/deposit/history/   // filter for a specific currency
*/
func accountDepositHistoryHandler(w http.ResponseWriter, req *http.Request, cfg Config, currencySymbol string) {
	errb := checkSigHeaders(w, req)
	if errb {
		return
	}
	retVal := generateAccountDepositHistoryResponse(currencySymbol, req, cfg)
	_, _ = fmt.Fprintf(w, string(retVal))
}

func generateAccountDepositHistoryResponse(currencySymbol string, req *http.Request, cfg Config) []byte {

	log.Printf("%s %s %s", req.Method, req.URL, req.Header)
	methodName := "okcatbox:account-deposit-history.go:generateAccountDepositHistoryResponse"

	//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": "", "Vary": ""})

	// 1. We'll need an HTTP client for the subsequent requests.
	timeout := 60000 * time.Millisecond
	httpClient := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	// 2. We'll need the UserID associated with the catbox credentials.
	userID := getUserId(req.Header)

	// 3. We'll need the bookwerx category associated with this user.
	bwUserCat, _ := getCategoryBySym(userID, httpClient, cfg)

	// 4. Get all deposit transactions for this user
	searchCategories := make([]uint32, 0)
	searchCategories = append(searchCategories, bwUserCat)
	searchCategories = append(searchCategories, cfg.Bookwerx.CatDeposit)
	depositTransactions, err := accountDepositHistoryGetTransactionsByCat(searchCategories, httpClient, cfg)
	if s, errb := squeal(methodName, "getTransactionsByCat", err); errb {
		return []byte(s)
	}

	// 5. Get all distributions for these transactions, that affect a hot wallet
	distributions, _ := accountDepositHistoryGetDistributionsByTx(depositTransactions, httpClient, cfg)
	if s, errb := squeal(methodName, "getDistributionsByTx", err); errb {
		return []byte(s)
	}

	// Loop over all distributions and build the return result.
	depositHistories := make([]utils.DepositHistory, 0)
	for _, v := range distributions {

		// If a currency symbol has been specified, filter all but that.
		if currencySymbol == "" || currencySymbol == v.CurrencySymbol {
			// here we potentially lose info because extra digits may be hidden by rounding
			amt := dfp_fmt(DFP{v.Amount, v.AmountExp}, -8)

			dh := utils.DepositHistory{
				Amount:     amt.s,
				TXID:       "blockchain txid",
				CurrencyID: v.CurrencySymbol,
				From:       "",
				To:         "okex address",
				DepositID:  v.TransactionID,
				Timestamp:  v.Time,
				Status:     "2", //  2: deposit successful
			}

			depositHistories = append(depositHistories, dh)
		}

	}

	retVal, err := json.Marshal(depositHistories)
	if s, errb := squealJSONMarshal(methodName, depositHistories, err); errb {
		return []byte(s)
	}
	return retVal

}
