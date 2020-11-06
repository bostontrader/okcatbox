package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	bw_api "github.com/bostontrader/bookwerx-common-go"
	utils "github.com/bostontrader/okcommon"
	"github.com/gojektech/heimdall/httpclient"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

func accountLedgerHandler(w http.ResponseWriter, req *http.Request, cfg Config) {
	errb := checkSigHeaders(w, req)
	if errb {
		return
	}
	retVal := generateAccountLedgerResponse(w, req, cfg)
	_, _ = fmt.Fprintf(w, string(retVal))
}

/* This endpoint is a real monstrosity.  Superficially, producing a list of transactions along with resulting
account balances, _after_ said transaction seems to be useful. However, in real life this endpoint either
serves up too much unrelated stuff or omits other important information.


In order to implement this endpoint, we must be really careful to start with certain basic operations and then build from there.

1. Basic principles:

* We must tag transactions with a user name and a transaction type using the notes field.
This is a hack because we can't presently tag transactions with categories as we can with accounts.

* The operation of this endpoint will produce info for all currencies, or only a single particular currency. Don't need to bother with a list of desired currencies.

* The operation of this endpoint will produce info for all types, or only a single particular type.  Don't need to bother with a list of desired types.

* This endpoint can mix and match currencies and types thus: all currencies, all types, all currencies, one type
one currency all types, one currency one type.


2. Suppose we want to invoke this endpoint for "type 1", "deposit" for a single currency.  In this example
let's use BTC.  Realize that this means external deposits into the user's funding account.  The output
will give us a list of deposits, as well as the BTC funding account balance after each deposit is made.

Also realize you cannot necessarily use this list of transactions and resulting balances to reconcile your
internal bookkeeping records because there may be other transactions, of other types, that are not listed
in a particular query.

2.1 Determine which account is the funding account, for this user, for BTC.

2.2 Get a list of all transactions along with the resulting running balance.

2.3 Filter this list to only include distributions from a transaction of the correct type, along with their balances.  Done.  Easy peasy.

3. Recall that we have four variations of currency and type specification.  We can generalize #2 as follows:

3.1 Determine the list of relevant accounts.  Any account for the supported types or the single account implied
by the requested type, filtered by the requested currency if any, and tagged for this customer.

3.2 For each particular account, execute op 2.

3.3 Merge all of these, sort by timestamp, and filter according to after, before, and limit.


*/
func generateAccountLedgerResponse(w http.ResponseWriter, req *http.Request, cfg Config) []byte {

	// by default, currency=all, type=all, before=all, after=all
	// Check for the existence of a currency in the query string.
	currencies, ok := req.URL.Query()["currency"]
	if ok {
		// Currency specified in the query string.  Is it valid?
		txn := db.Txn(false)
		defer txn.Abort()
		raw, err := txn.First("currencies", "id", currencies[0])
		if err != nil {
			log.Fatalf("error: %v", err) // This should never happen
			return []byte{}
		}
		if raw == nil {
			// Currency not valid, return 400 error
			w.WriteHeader(400)
			retVal, _ := json.Marshal(utils.Err30031(currencies[0]))
			return retVal
		} else {
			// The currency is valid, use it later as the default.
		}
	}

	// Check for the existence of a type in the query string.
	ttype, ok := req.URL.Query()["type"]
	n := make(map[string]string)
	n["1"] = "deposit"
	n["2"] = "withdraw"
	n["31"] = "into C2C"
	n["32"] = "out of C2C"
	n["37"] = "into spot"
	n["38"] = "out of spot"
	// 1 all distributions that dr funding and no other okex account
	// 2 all distributions that cr funding and no other okex account
	// 31 all distributions that dr c2c
	if ok {
		// Type specified in the query string.  Is it valid?  Only those types that are on the list are valid.
		_, ok := n[ttype[0]]
		if ok {
			// value is a valid type. use it later.
		} else {
			// Not a valid type.
			w.WriteHeader(400)
			retVal, _ := json.Marshal(utils.Err30024("Invalid type type"))
			return retVal
		}

	}

	// Check for the existence of a before in the query string.
	before, ok := req.URL.Query()["type"]
	if ok {
		// before specified in the query string.  Is it valid?  Only those befores that can parse into an int32 are valid.
		_, err := strconv.ParseUint(before[0], 10, 32)
		if err != nil {
			// Not a valid before.
			w.WriteHeader(400)
			retVal, _ := json.Marshal(utils.Err30025("before parameter format is error"))
			return retVal
		}
		// u contains the correctly parsed before

	}

	// Check for the existence of an after in the query string.
	after, ok := req.URL.Query()["type"]
	if ok {
		// after specified in the query string.  Is it valid?  Only those afters that can parse into an int32 are valid.
		_, err := strconv.ParseUint(after[0], 10, 32)
		if err != nil {
			// Not a valid after.
			w.WriteHeader(400)
			retVal, _ := json.Marshal(utils.Err30025("after parameter format is error"))
			return retVal
		}
		// u contains the correctly parsed before

	}

	// Check for the existence of a limit in the query string.
	limit, ok := req.URL.Query()["type"]
	if ok {
		// limit specified in the query string.  Is it valid?  Only those limits that can parse into an int32 are valid.
		_, err := strconv.ParseUint(limit[0], 10, 32)
		if err != nil {
			// Not a valid limit.
			w.WriteHeader(400)
			retVal, _ := json.Marshal(utils.Err30025("limit parameter format is error"))
			return retVal
		}
		// u contains the correctly parsed limit

	}

	// We'll need an HTTP client for the subsequent requests.
	timeout := 5000 * time.Millisecond
	httpClient := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))

	// 1. Find all transactions where note like "%type 1%" or "%type 2%" ... where tx id > before and tx id < after
	// the actual meaning of "before" and "after" in this endpoint is to omit anything <= before and >= after
	selectt := "SELECT%20transactions.id,%20transactions.time"
	from := "FROM%20transactions"
	where := "WHERE notes like %user moe% and notes like %type 2%"
	//where := fmt.Sprintf("WHERE%%20categories.title%%3d%%27%s%%27", ok_access_key8)
	query := fmt.Sprintf("%s%%20%s%%20%s", selectt, from, where)
	url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)
	body, err1 := bw_api.Get(httpClient, url)
	if err1 != nil {
		fmt.Printf("okcatbox:account-ledger.go:generateAccountLedgerResponse: get error: %v\n", err1)
		return []byte{}
	}

	type Transaction struct {
		TransactionID uint32 `json:"id"`
		Notes         string
		Time          string
	}

	n2 := make([]Transaction, 0)
	err1 = json.NewDecoder(bytes.NewReader(body)).Decode(&n2)
	if err1 != nil {
		//fmt.Println("bookwerx-api.go: getTransferAccountID:", err)
		return []byte{}
	}

	ledgerEntriesMap := make(map[uint32]utils.LedgerEntry)
	for _, tx := range n2 {
		le := utils.LedgerEntry{
			Amount:    "-1",
			Balance:   "-1",
			Currency:  "currency",
			Fee:       "fee",
			LedgerID:  string(tx.TransactionID),
			Typename:  "typename", // what is the type of this transaction? and what is said type name?
			Timestamp: tx.Time,
		}
		ledgerEntriesMap[tx.TransactionID] = le

	}

	// For each item found in bookwerx, create an entry to return
	//walletEntries := make([]utils.WalletEntry, 0)
	//for _, brd := range sums.Sums {

	// Negate the sign.  BW reports this balance as a liability of okcatbox.  The ordinary CR balance is represented using a - sign.  But the user expects a DR value to match the asset on his books.
	//n1 := brd.Sum
	//n2 := DFP{-n1.Amount, n1.Exp}
	//n3 := dfp_fmt(n2, -8) // here we lose info re: extra digits hidden by roundoff
	//walletEntries = append(walletEntries, utils.WalletEntry{
	//Available:  n3.s,
	//Balance:    n3.s,
	//CurrencyID: brd.Account.Currency.Symbol, // bw currency.symbol is used as the okex CurrencyID
	//Hold:       "0.00000000"})
	//}

	//retVal, _ := json.Marshal(walletEntries)
	//return retVal

	return []byte{}

}
