package main

import (
	"fmt"
	"github.com/gojektech/heimdall/httpclient"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Distributions struct {
	Amount        int64
	AmountExp     int8
	TransactionID uint32 `json:"transaction_id"`
}

/* Two endpoints use this handler:
/api/account/v3/deposit/history    // for all currencies
/api/account/v3/deposit/history/   // filter for a specific currency
*/
func account_depositHistoryHandler(w http.ResponseWriter, req *http.Request, cfg Config, currencySymbol string) {
	errb := checkSigHeaders(w, req)
	if errb {
		return
	}
	retVal := generateAccountDepositHistoryResponse(req, cfg)
	_, _ = fmt.Fprintf(w, string(retVal))
}

func generateAccountDepositHistoryResponse(req *http.Request, cfg Config) []byte {

	log.Printf("%s %s %s", req.Method, req.URL, req.Header)
	//methodName := "okcatbox:account-deposit-address.go:generateAccountDepositHistoryResponse"

	// 1. We'll need an HTTP client for the subsequent requests.
	timeout := 5000 * time.Millisecond
	httpClient := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	// 2. We'll need the UserID associated with the catbox credentials.
	userID := getUserId(req.Header)

	// 3. We'll need the bookwerx category associated with this user.
	bwUserCat, _ := getCategoryBySym(userID, httpClient, cfg)

	// 4. Get all deposit transactions for this user
	searchCategories := make([]uint32, 2)
	searchCategories = append(searchCategories, bwUserCat)
	searchCategories = append(searchCategories, cfg.Bookwerx.CatDeposit)
	depositTransactions, _ := getTransactionsByCat(searchCategories, httpClient, cfg)
	fmt.Printf("%+v", depositTransactions)
	// 5. Get all distributions for these transactions

	// 3.
	//if currency_symbol == "" {
	// 3.1 Find all accounts tagged as funding for this api credential
	//_, _ = getTransferAccountID(categoryID, currencyID, httpClient, cfg)

	//setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": "", "Vary": ""})
	//depositHistories := make([]utils.DepositHistory, 1)
	//depositHistories[0] = utils.DepositHistory{Amount: "amount", TXID: "txid", CurrencyID: "currency", From: "from", To: "to", DepositID: 666, Timestamp: "timestamp", Status: "status"}
	//retVal, _ = json.Marshal(depositHistories)
	//return retVal

	//} else {
	// If the currency is legit
	//_, err := getCurrencyBySym(client, currency_symbol, cfg)
	//if err != nil {
	//s := fmt.Sprintf("account-deposit-history.go generateAccountDepositHistoryResponse: The currency_symbol %s is not defined on this OKCatbox server.", currency_symbol)
	//log.Error(s)
	//fmt.Fprintf(w, s)
	//return []byte{}
	//w.WriteHeader(http.StatusBadRequest)
	//retVal, _ = json.Marshal(utils.Err30031(currency_symbol))
	//return retVal
	//}

	//depositHistories := make([]utils.DepositHistory, 1)
	//depositHistories[0] = utils.DepositHistory{Amount: "amount", TXID: "txid", CurrencyID: "currency", From: "from", To: "to", DepositID: 666, Timestamp: "timestamp", Status: "status"}
	//retVal, _ = json.Marshal(depositHistories)
	//return retVal

	// find all accounts marked as funding for this api credentials and with this currency
	// else
	// error bad currency

	// find all CR distributions
	// return them all
	//}

	return []byte{}

}
