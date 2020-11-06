package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	bwAPI "github.com/bostontrader/bookwerx-common-go"
	"github.com/gojektech/heimdall/httpclient"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strings"
)

//import (
//"bytes"
//"encoding/json"
//"fmt"
//bw_api "github.com/bostontrader/bookwerx-common-go"
//utils "github.com/bostontrader/okcommon"
//"github.com/gojektech/heimdall/httpclient"
//"github.com/shopspring/decimal"
//log "github.com/sirupsen/logrus"
//"net/http"
//"time"
//)

//import (
//"bytes"
//"encoding/json"
//bwAPI "github.com/bostontrader/bookwerx-common-go"
//"github.com/pkg/errors"
//log "github.com/sirupsen/logrus"
//"io/ioutil"
//"strings"

//"bytes"
//"encoding/json"
//"errors"
//"fmt"
//"github.com/gojektech/heimdall/httpclient"
//"io/ioutil"
//"net/http"
//"strings"
//"fmt"
//)

type AccountsID struct {
	ID uint32 `json:"accounts-id"`
}

type CategoriesID struct {
	ID uint32 `json:"categories-id"`
}

type CurrenciesID struct {
	ID uint32 `json:"currencies-id"`
}

type TransactionsID struct {
	ID uint32 `json:"transactions-id"`
}

/*



// We only need a subset of all the info returned.
type AccountJoined struct {
	AccountID     int32 `json:"id"`
	Categories    []Acctcat2
	CurrencyShort CurrencyShort `json:"currency"`
	Title         string
}

type Acctcat2 struct {
	CategorySymbol string `json:"category_symbol"`
}

type Category struct {
	Id     int32
	Apikey string
	Symbol string
	Title  string
}

// Only a subset of the available fields,,
type Currency struct {
	CurrencyID int32 `json:"id"`
	Symbol     string
}

type CurrencyShort struct {
	Symbol string
	Title  string
}

type CurrencySymbol struct {
	CurrencyID int32
	Symbol     string
}*/

type LID struct {
	LastInsertID uint32
}

/* Given a category symbol determine the bookwerx category_id.  Return said (category_id, nil) or (0, some error).
 */
func getCategoryBySym(categorySymbol string, httpClient *httpclient.Client, cfg Config) (categoryID uint32, err error) {

	methodName := "okcatbox:bookwerx-api.go:getCategoryBySym"

	// 1. Build and execute the query on the bookwerx server.
	selectt := "SELECT%20categories.id"
	from := "FROM%20categories"
	where := fmt.Sprintf("WHERE%%20categories.symbol%%3D%%22%s%%22", categorySymbol)
	query := fmt.Sprintf("%s%%20%s%%20%s", selectt, from, where)
	url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)

	responseBody, err := bwAPI.Get(httpClient, url)
	if err != nil {
		s := fmt.Sprintf("%s: bwAPI.Get error=%+v", methodName, err)
		log.Error(s, err)
		return 0, errors.New(s)
	}
	fixDot(responseBody)

	// 2. Decode the response.
	categoriesID := make([]CategoriesID, 0)
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	err = dec.Decode(&categoriesID)
	if err != nil {
		s := fmt.Sprintf("%s: JSON decode error: Err=%v\nbody=%s\n", methodName, err, string(responseBody))
		log.Error(s, err)
		return 0, errors.New(s)
	}

	// 3. The decoded response should be a slice of len either 0 or 1.
	switch len(categoriesID) {
	case 0:
		s := fmt.Sprintf("%s: Category symbol %s not found\n", methodName, categorySymbol)
		log.Error(s, err)
		return 0, errors.New(s)

	case 1:
		return categoriesID[0].ID, nil
	default:
		s := fmt.Sprintf("%s: Category symbol %s is defined more than once.  No can do error.\n", methodName, categorySymbol)
		log.Error(s, err)
		return 0, errors.New(s)
	}

}

/* Given the userCategoryID, the fundingAvailableCategoryID, and the currencyID, find all accounts that are tagged with _all_ of these categories and are configured for the given currency.  There can only be 0 or 1 of such accounts.  Return said (accountID, nil) or (0, some error)

 */
func getFundingAvailableAccountID(userCategoryID, fundingAvailableCategoryID, currencyID uint32, httpClient *httpclient.Client, cfg Config) (accountID uint32, err error) {

	methodName := "okcatbox:bookwerx-api.go:getFundingAvailableAccountID"
	accountID = 0

	selectt := "SELECT%20accounts.id"
	from := "FROM%20accounts_categories"
	join1 := "JOIN%20accounts%20ON%20accounts.id%3daccounts_categories.account_id"
	join2 := "JOIN%20currencies%20ON%20currencies.id%3daccounts.currency_id"
	where := fmt.Sprintf("WHERE%%20accounts_categories.category_id%%20IN%%20(%d,%d)",
		cfg.Bookwerx.CatFunding, userCategoryID,
	)
	and := fmt.Sprintf("AND%%20currencies.id%%3d%d", currencyID)

	// Watch out! %3d2 hardwires this query to find 2 categories.
	group := "GROUP%20BY%20accounts_categories.account_id%20HAVING%20COUNT(DISTINCT%20accounts_categories.account_id)%3d2"

	query := fmt.Sprintf("%s%%20%s%%20%s%%20%s%%20%s%%20%s%%20%s", selectt, from, join1, join2, where, and, group)
	url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)
	responseBody, err := bwAPI.Get(httpClient, url)
	//if err != nil {
	//fmt.Printf("%s: bw_api.Get error: %+v", methodName, err)
	//return
	//}
	if s, errb := squeal(methodName, "bwAPI.GET", err); errb {
		return accountID, errors.New(s)
	}

	fixDot(responseBody)

	accountsID := make([]AccountsID, 0)
	//accountsID := AccountsID{}
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	err = dec.Decode(&accountsID)
	//if err != nil {
	//fmt.Println("%s: getTransferAccountID:", methodName, err)
	//return
	//}
	if s, errb := squealJSONDecode(methodName, responseBody, err); errb {
		return accountID, errors.New(s)
	}

	// 3. The decoded response should be a slice of len either 0 or 1.
	switch len(accountsID) {
	case 0:
		//s := fmt.Sprintf("%s: Currency symbol %s not found\n", methodName, currency_symbol)
		//log.Error(s, err)
		return 0, nil // no accountID found but not an error.

	case 1:
		return accountsID[0].ID, nil
	default:
		s := fmt.Sprintf("%s: There are more than one accounts that match this criteria.  No can do error.\n", methodName)
		log.Error(s, err)
		return 0, errors.New(s)
	}

}

/* Given a currency_symbol (ie. BTC, LTC) determine the bookwerx currency_id.  Return said currency_id and nil, or 0 and some error.
Perhaps all currency_ids should be looked up when the server starts or perhaps this call should be memoized.  But optimize later.
*/
func getCurrencyBySym(client *httpclient.Client, currency_symbol string, cfg Config) (currency_id uint32, err error) {

	methodName := "okcatbox:bookwerx-api.go:getCurrencyBySym"

	// 1. Build and execute the query on the bookwerx server.
	selectt := "SELECT%20currencies.id"
	from := "FROM%20currencies"
	where := fmt.Sprintf("WHERE%%20currencies.symbol%%3D%%22%s%%22", currency_symbol)
	query := fmt.Sprintf("%s%%20%s%%20%s", selectt, from, where)
	url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)
	responseBody, err := bwAPI.Get(client, url)

	if err != nil {
		s := fmt.Sprintf("%s: bwAPI.Get error=%+v", methodName, err)
		log.Error(s, err)
		return 0, errors.New(s)
	}
	fixDot(responseBody)

	// 2. Decode the response.
	cid := make([]CurrenciesID, 0)
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	err = dec.Decode(&cid)
	if err != nil {
		s := fmt.Sprintf("%s: JSON decode error: Err=%v\nbody=%s\n", methodName, err, string(responseBody))
		log.Error(s, err)
		return 0, errors.New(s)
	}

	// 3. The decoded response should be a slice of len either 0 or 1.
	switch len(cid) {
	case 0:
		s := fmt.Sprintf("%s: Currency symbol %s not found\n", methodName, currency_symbol)
		log.Error(s, err)
		return 0, errors.New(s)

	case 1:
		return cid[0].ID, nil
	default:
		s := fmt.Sprintf("%s: Currency symbol %s is defined more than once.  No can do error.\n", methodName, currency_symbol)
		log.Error(s, err)
		return 0, errors.New(s)
	}

	//if len(n1) == 1 {
	//return n1[0].Id, nil
	//} else if len(n1) > 1 {
	//fmt.Println("bookwerx-api.go: getTransferAccountID: There are more than one suitable accounts.  This should never happen.")
	//return
	//}

	// 3. If control passes here we know that the account doesn't already exist. So create it and tag it with the two categories and the given currency.
	//if len(n1) == 0 {

	//url := fmt.Sprintf("%s/currencies?apikey=%s", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey)

	//resp, err := client.Get(url, nil)
	//if err != nil {
	//s := fmt.Sprintf("bookwerx-api.go:getCurrencyBySym 1: %v", err)
	//log.Error(s)
	//return -1, errors.New(s)
	//}
	//defer resp.Body.Close()

	//if resp.StatusCode != 200 {
	//s := fmt.Sprintf("bookwerx-api.go:getCurrencyBySym 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
	//log.Error(s)
	//return -1, errors.New(s)
	//}

	//currencies := make([]Currency, 0)
	//dec := json.NewDecoder(resp.Body)
	//err = dec.Decode(&currencies)
	//if err != nil {
	//s := fmt.Sprintf("bookwerx-api.go:getCurrencyBySym 3: %v", err)
	//log.Error(s)
	//return -1, errors.New(s)
	//}

	// Now search for the specific currency_symbol
	//found := false
	//for _, v := range currencies {
	//if strings.EqualFold(v.Symbol, currency_symbol) { // case insensitive compare
	//found = true
	//currency_id = v.CurrencyID
	//break
	//}
	//}
	//if !found {
	//s := fmt.Sprintf("bookwerx-api.go:getCurrencyBySym 4: The currency %s is not defined on this OKCatbox server.", currency_symbol)
	//log.Error(s)
	//return -1, errors.New(s)
	//}

	//return currency_id, nil
	return 0, nil
}

/*
Given a slice of transaction ID, find all distributions that are related to any of the given transactions.  Return said ([Distribution], nil) or ([0], some error).

The categoryID slice must be of length 1 or 2.
*/
func getDistributionsByTx(transactionID []uint32, httpClient *httpclient.Client, cfg Config) (transaction_id []Distributions, err error) {

	methodName := "okcatbox:account-deposit-history.go:getTransactionsByCat"

	// 1. The desired query varies according to the len(transactionID)
	var inClause string
	switch len(transactionID) {
	case 0:
		{
			return []Distributions{}, nil // no TXID, no distributions
		}

	default:
		{
			var sb strings.Builder
			sb.WriteRune('(')
			for k, _ := range transaction_id {
				if k == 0 {
					sb.WriteString("%d")
				} else {
					sb.WriteString(",%d")
				}
			}
			sb.WriteRune(')')
			inClause = sb.String()
		}
	}

	// 2. Build and execute the query.
	selectt := "SELECT%20distributions.amount, distributions.amount_exp, distributions.transaction_id"
	from := "FROM%20distributions"
	where := fmt.Sprintf("WHERE%%20distributions.transaction_id%%20IN%%20%s", inClause)
	query := fmt.Sprintf("%s%%20%s%%20%s", selectt, from, where)
	url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)

	responseBody, err := bwAPI.Get(httpClient, url)

	if err != nil {
		s := fmt.Sprintf("%s: bwAPI.Get error=%+v", methodName, err)
		log.Error(s, err)
		return []Distributions{}, errors.New(s)
	}
	//fixDot(responseBody)

	// 3. Decode the response.
	distributions := make([]Distributions, 0)
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	err = dec.Decode(&distributions)
	if err != nil {
		s := fmt.Sprintf("%s: JSON decode error: Err=%v\nbody=%s\n", methodName, err, string(responseBody))
		log.Error(s, err)
		return []Distributions{}, errors.New(s)
	}

	return distributions, nil

}

/*
Given a slice of category ID, find all transactions that are tagged with all of the given categories.  Return said ([transaction ID], nil) or ([0], some error).

The categoryID slice must be of length 1 or 2.
*/
func getTransactionsByCat(categoryID []uint32, httpClient *httpclient.Client, cfg Config) (transaction_id []uint32, err error) {

	methodName := "okcatbox:bookwerx-api.go:getTransactionsByCat"

	// 1. Validate the categoryID and build a suitable IN clause for the query.
	var inClause string
	switch len(categoryID) {
	case 1:
		{
			inClause = fmt.Sprintf("(%d)", categoryID[0])
		}
	case 2:
		{
			inClause = fmt.Sprintf("(%d,%d)", categoryID[0], categoryID[1])
		}
	default:
		{
			s := fmt.Sprintf("%s: The len(categoryID) must be 1 or 2 only. In this case the length=%d\n", methodName, len(categoryID))
			log.Error(s, err)
			return []uint32{0}, errors.New(s)
		}
	}

	// 2. Build and execute the query.
	selectt := "SELECT%20accounts.id"
	from := "FROM%20accounts_categories"
	join1 := "JOIN%20accounts%20ON%20accounts.id%3daccounts_categories.account_id"
	join2 := "JOIN%20currencies%20ON%20currencies.id%3daccounts.currency_id"
	where := fmt.Sprintf("WHERE%%20accounts_categories.category_id%%20IN%%20%s", inClause)
	//group := "GROUP%20BY%20accounts_categories.account_id%20HAVING%20COUNT(DISTINCT%20accounts_categories.account_id)%3d2"
	group := fmt.Sprintf("GROUP%%20BY%%20accounts_categories.account_id%%20HAVING%%20COUNT(DISTINCT%%20accounts_categories.account_id)%%3d%d", len(categoryID))
	query := fmt.Sprintf("%s%%20%s%%20%s%%20%s%%20%s%%20%s", selectt, from, join1, join2, where, group)
	url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)

	responseBody, err := bwAPI.Get(httpClient, url)

	if err != nil {
		s := fmt.Sprintf("%s: bwAPI.Get error=%+v", methodName, err)
		log.Error(s, err)
		return []uint32{0}, errors.New(s)
	}
	fixDot(responseBody)

	// 3. Decode the response.
	tid := make([]TransactionsID, 0)
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	err = dec.Decode(&tid)
	if err != nil {
		s := fmt.Sprintf("%s: JSON decode error: Err=%v\nbody=%s\n", methodName, err, string(responseBody))
		log.Error(s, err)
		return []uint32{0}, errors.New(s)
	}

	// 4. Map []TransactionsID -> []uint32
	retVal := make([]uint32, len(tid))
	for _, v := range tid {
		retVal = append(retVal, v.ID)
	}

	return retVal, nil

}

// Create a new Account
func postAccount(currencyID uint32, title string, httpClient *httpclient.Client, cfg Config) (accountID uint32, err error) {

	methodName := "okcatbox:bookwerx-api.go:postAccount"
	url1 := fmt.Sprintf("%s/accounts", cfg.Bookwerx.Server)
	url2 := fmt.Sprintf("apikey=%s&currency_id=%d&rarity=0&title=%s", cfg.Bookwerx.APIKey, currencyID, title)

	h := make(map[string][]string)
	h["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	resp, err := httpClient.Post(url1, bytes.NewBuffer([]byte(url2)), h)
	defer resp.Body.Close()
	if err != nil {
		s := fmt.Sprintf("%s: postAccount error %+v", methodName, err)
		log.Error(s)
		return 0, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("%s: status code error: Expected status=200, Received=%d, Body=%v", resp.StatusCode, bodyString(resp))
		log.Error(s)
		return 0, errors.New(s)
	}

	n := LID{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		s := fmt.Sprintf("%s: json.Decode error %+v", methodName, err)
		log.Error(s)
		return 0, errors.New(s)
	}

	return n.LastInsertID, nil
}

// Create a new AcctCat
func postAcctcat(accountID uint32, categoryID uint32, httpClient *httpclient.Client, cfg Config) (acctCatID uint32, err error) {

	methodName := "okcatbox:bookwerx-api.go:postAcctcat"
	url1 := fmt.Sprintf("%s/acctcats", cfg.Bookwerx.Server)
	url2 := fmt.Sprintf("apikey=%s&account_id=%d&category_id=%d", cfg.Bookwerx.APIKey, accountID, categoryID)

	h := make(map[string][]string)
	h["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	resp, err := httpClient.Post(url1, bytes.NewBuffer([]byte(url2)), h)
	defer resp.Body.Close()
	if err != nil {
		s := fmt.Sprintf("%s: %+v\n", methodName, err)
		log.Error(s)
		return 0, errors.New(s)
	}

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("%s: Expected status=200, Received=%d, Body=%v", methodName, resp.StatusCode, bodyString(resp))
		log.Error(s)
		return 0, errors.New(s)
	}

	n := LID{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		s := fmt.Sprintf("%s: json.Decode error %+v", methodName, err)
		log.Error(s)
		return 0, errors.New(s)
	}

	return n.LastInsertID, nil
}

/*
Given a symbol and a title, as well as other supporting info, post a new category to a bookwerx server.
Return the new (category id, nil) or (0, some error)
*/
func postCategory(symbol, title string, httpClient *httpclient.Client, cfg Config) (categoryID uint32, err error) {

	// 1. Init
	methodName := "okcatbox:bookwerx-api.go:postCategory"
	categoryID = 0
	url := fmt.Sprintf("%s/categories", cfg.Bookwerx.Server)
	postBody := fmt.Sprintf("apikey=%s&symbol=%s&title=%s", cfg.Bookwerx.APIKey, symbol, title)
	reqHeaders := make(map[string][]string)
	reqHeaders["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	// 2. POST request
	resp, err := httpClient.Post(url, bytes.NewBuffer([]byte(postBody)), reqHeaders)
	if s, errb := squeal(methodName, "httpClient.Post", err); errb {
		return categoryID, errors.New(s)
	}

	// 3. Read the response body in order to simply subsequent work.
	responseBody, err := ioutil.ReadAll(resp.Body)
	if s, errb := squeal(methodName, "ioutil.Readall", err); errb {
		return categoryID, errors.New(s)
	}

	// 4. Close the response body
	err = resp.Body.Close()
	if s, errb := squeal(methodName, "resp.Body.Close", err); errb {
		return categoryID, errors.New(s)
	}

	// 5. JSON decode the response
	lid := LID{}
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	err = dec.Decode(&lid)
	if s, errb := squealJSONDecode(methodName, responseBody, err); errb {
		return categoryID, errors.New(s)
	}

	return lid.LastInsertID, nil
}

/*
Given the information necessary to describe a distribution, as well as other supporting info, post a new transaction to a bookwerx server.
Return the new (distribution id, nil) or (0, some error).
*/
func postDistribution(accountID uint32, amt string, exp string, transactionID uint32, httpClient *httpclient.Client, cfg Config) (distributionID uint32, err error) {

	// 1. Init
	methodName := "okcatbox:bookwerx-api.go:postDistribution"
	distributionID = 0
	url := fmt.Sprintf("%s/distributions", cfg.Bookwerx.Server)
	postBody := fmt.Sprintf("apikey=%s&account_id=%d&amount=%s&amount_exp=%s&transaction_id=%d", cfg.Bookwerx.APIKey, accountID, amt, exp, transactionID)
	reqHeaders := make(map[string][]string)
	reqHeaders["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	// 2. POST request
	resp, err := httpClient.Post(url, bytes.NewBuffer([]byte(postBody)), reqHeaders)
	//defer resp.Body.Close()
	if err != nil {
		s := fmt.Sprintf("%s: %+v\n", methodName, err)
		log.Error(s)
		return 0, errors.New(s)
	}

	// 3. Read the response body in order to simply subsequent work.
	responseBody, err := ioutil.ReadAll(resp.Body)
	if s, errb := squeal(methodName, "ioutil.Readall", err); errb {
		return distributionID, errors.New(s)
	}

	// 4. Close the response body
	err = resp.Body.Close()
	if s, errb := squeal(methodName, "resp.Body.Close", err); errb {
		return distributionID, errors.New(s)
	}

	// 5. JSON decode the response
	var lid LID
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	err = dec.Decode(&lid)
	if s, errb := squealJSONDecode(methodName, responseBody, err); errb {
		return distributionID, errors.New(s)
	}

	return lid.LastInsertID, nil
}

// Create a new Trancat
func postTrancat(transactionID uint32, categoryID uint32, httpClient *httpclient.Client, cfg Config) (trancatID uint32, err error) {

	// 1. Init
	methodName := "okcatbox:bookwerx-api.go:postTrancat"
	trancatID = 0
	url := fmt.Sprintf("%s/trancats", cfg.Bookwerx.Server)
	postBody := fmt.Sprintf("apikey=%s&transaction_id=%d&category_id=%d", cfg.Bookwerx.APIKey, transactionID, categoryID)
	reqHeaders := make(map[string][]string)
	reqHeaders["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	// 2. POST request
	resp, err := httpClient.Post(url, bytes.NewBuffer([]byte(postBody)), reqHeaders)
	if s, errb := squeal(methodName, "httpClient.Post", err); errb {
		return trancatID, errors.New(s)
	}

	// 3. Read the response body in order to simply subsequent work.
	responseBody, err := ioutil.ReadAll(resp.Body)
	if s, errb := squeal(methodName, "ioutil.Readall", err); errb {
		return trancatID, errors.New(s)
	}

	// 4. Close the response body
	err = resp.Body.Close()
	if s, errb := squeal(methodName, "resp.Body.Close", err); errb {
		return trancatID, errors.New(s)
	}

	// 5. JSON decode the response
	var lid LID
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	err = dec.Decode(&lid)
	if s, errb := squealJSONDecode(methodName, responseBody, err); errb {
		return trancatID, errors.New(s)
	}

	return lid.LastInsertID, nil
}

/*
Given notes and a time, as well as other supporting info, post a new transaction to a bookwerx server.
Return the new (transaction id, nil) or (0, some error).
*/
func postTransaction(notes, time string, httpClient *httpclient.Client, cfg Config) (transactionID uint32, err error) {

	// 1. Init
	methodName := "okcatbox:bookwerx-api.go:postTransaction"
	transactionID = 0
	url := fmt.Sprintf("%s/transactions", cfg.Bookwerx.Server)
	postBody := fmt.Sprintf("apikey=%s&notes=%s&time=%s", cfg.Bookwerx.APIKey, notes, time)
	reqHeaders := make(map[string][]string)
	reqHeaders["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	// 2. POST request
	resp, err := httpClient.Post(url, bytes.NewBuffer([]byte(postBody)), reqHeaders)
	if s, errb := squeal(methodName, "httpClient.Post", err); errb {
		return transactionID, errors.New(s)
	}

	// 3. Read the response body in order to simply subsequent work.
	responseBody, err := ioutil.ReadAll(resp.Body)
	if s, errb := squeal(methodName, "ioutil.Readall", err); errb {
		return transactionID, errors.New(s)
	}

	// 4. Close the response body
	err = resp.Body.Close()
	if s, errb := squeal(methodName, "resp.Body.Close", err); errb {
		return transactionID, errors.New(s)
	}

	// 5. JSON decode the response
	var lid LID
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	err = dec.Decode(&lid)
	if s, errb := squealJSONDecode(methodName, responseBody, err); errb {
		return transactionID, errors.New(s)
	}

	return lid.LastInsertID, nil
}

//XXXXX
// obsolete
/*func getFundingAccountID(client *httpclient.Client, accounts []AccountJoined, ok_access_key8 string, currency_id int32, currency_symbol string, cfg Config) (account_id int32, err error) {

	// 1. Can I find an account tagged funding _and_ apikey8 using the correct currency?
	account_id, found := searchF(accounts, ok_access_key8, currency_symbol)

	// 2. If the account doesn't already exist, then create it and tag it with the customer's category and funding
	if !found {

		// 2.1 Find the category id for this apikey
		url := fmt.Sprintf("%s/category/bysym/%s?apikey=%s", cfg.Bookwerx.Server, ok_access_key8, cfg.Bookwerx.APIKey)
		category_id_api, err := getCategoryBySym(client, url)
		if err != nil {
			log.Error(err)
			return -1, err
		}

		// 2.2 Make the new account
		account_id, err = postAccount(client, currency_id, ok_access_key8, cfg)
		if err != nil {
			log.Error(err)
			return -1, err
		}

		// 2.3 Tag with the api key
		_, err = postAcctcat(client, account_id, category_id_api, cfg)
		if err != nil {
			log.Error(err)
			return -1, err
		}

		// 2.4 Tag with funding
		_, err = postAcctcat(client, account_id, cfg.Bookwerx.FundingCat, cfg)
		if err != nil {
			log.Error(err)
			return -1, err
		}

	}

	return account_id, nil
}*/

/*
Given a slice of category ID and a currency ID, find a single account that is tagged with all of the given categories and is configured to use the given currency.  Return said (account ID, nil) or (0, some error).

The categoryID slice must be of length 1 or 2.
*/
/*func getTransferAccountID(categoryID []uint32, currencyID uint32, httpClient *httpclient.Client, cfg Config) (account_id uint32, err error) {

	methodName := "okcatbox:bookwerx-api.go:getTransferAccountID"

	// 1. Validate the categoryID and build a suitable IN clause for the query.
	var inClause string
	switch len(categoryID) {
	case 1: {inClause = fmt.Sprintf("(%d)", categoryID[0])}
	case 2: {inClause = fmt.Sprintf("(%d,%d)", categoryID[0], categoryID[1])}
	default: {
		s := fmt.Sprintf("%s: The len(categoryID) must be 1 or 2 only. In this case the length=%d\n", methodName, len(categoryID))
		log.Error(s, err)
		return 0, errors.New(s)
	}
	}

	// 2. Build and execute the query.
	selectt := "SELECT%20accounts.id"
	from := "FROM%20accounts_categories"
	join1 := "JOIN%20accounts%20ON%20accounts.id%3daccounts_categories.account_id"
	join2 := "JOIN%20currencies%20ON%20currencies.id%3daccounts.currency_id"
	where := fmt.Sprintf("WHERE%%20accounts_categories.category_id%%20IN%%20%s", inClause)
	and := fmt.Sprintf("AND%%20currencies.id%%3d%d", currencyID)
	//group := "GROUP%20BY%20accounts_categories.account_id%20HAVING%20COUNT(DISTINCT%20accounts_categories.account_id)%3d2"
	group := fmt.Sprintf("GROUP%%20BY%%20accounts_categories.account_id%%20HAVING%%20COUNT(DISTINCT%%20accounts_categories.account_id)%%3d%d", len(categoryID))
	query := fmt.Sprintf("%s%%20%s%%20%s%%20%s%%20%s%%20%s%%20%s", selectt, from, join1, join2, where, and, group)
	url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)

	responseBody, err := bwAPI.Get(httpClient, url)

	if err != nil {
		s := fmt.Sprintf("%s: bwAPI.Get error=%+v", methodName, err)
		log.Error(s, err)
		return 0, errors.New(s)
	}
	fixDot(responseBody)

	// 3. Decode the response.
	aid := make([]AccountsID, 0)
	dec := json.NewDecoder(bytes.NewReader(responseBody))
	err = dec.Decode(&aid)
	if err != nil {
		s := fmt.Sprintf("%s: JSON decode error: Err=%v\nbody=%s\n", methodName, err, string(responseBody))
		log.Error(s, err)
		return 0, errors.New(s)
	}

	// 4. The decoded response should be a slice of len either 0 or 1.
	switch len(aid) {
	case 0:
		s := fmt.Sprintf("%s: This function cannot find a matching account.\n", methodName)
		log.Error(s, err)
		return 0, errors.New(s)

	case 1:
		return aid[0].ID, nil
	default:
		s := fmt.Sprintf("%s: This function found more than one matching account.  No can do error.\n", methodName)
		log.Error(s, err)
		return 0, errors.New(s)
	}



	//body, err = bwAPI.Get(client, url)
	//if err != nil {
	//fmt.Println("bookwerx-api.go:getTransferAccountID 2: get error: ", err)
	//return
	//}
	//fixDot(body)

	//n1 := make([]AId, 0)
	//err = json.NewDecoder(bytes.NewReader(body)).Decode(&n1)
	//if err != nil {
	//fmt.Println("bookwerx-api.go: getTransferAccountID:", err)
	//return
	//}
	//if len(n1) == 1 {
	//return n1[0].Id, nil
	//} else if len(n1) > 1 {
	//fmt.Println("bookwerx-api.go: getTransferAccountID: There are more than one suitable accounts.  This should never happen.")
	//return
	//}

	// 3. If control passes here we know that the account doesn't already exist. So create it and tag it with the two categories and the given currency.
	//if len(n1) == 0 {

	// 3.1 Make the new account
	//account_id, err = postAccount(client, currency_id, ok_access_key8, cfg)
	//if err != nil {
	//fmt.Println("bookwerx-api.go:getTransferAccountID 3.1: postAccount error: ", err)
	//return 0, err
	//}

	// 3.2 Tag with the api key
	//_, err = postAcctcat(client, account_id, apikey_cat_id, cfg)
	//if err != nil {
	//fmt.Println("bookwerx-api.go:getTransferAccountID 3.2: postAcctcat error: ", err)
	//return -1, err
	//}

	// 3.3 Tag with funding, spot, whatever
	//_, err = postAcctcat(client, account_id, transfer_cat_id, cfg)
	//if err != nil {
	//fmt.Println("bookwerx-api.go:getTransferAccountID 3.3: postAcctcat error: ", err)
	//return -1, err
	//}

	//}

	//return account_id, nil
}*/

/*
A transfer has a source and destination (from, to) account in bookwerx.  Each account is tagged with two categories and is configured to use
a specific currency. More particularly:

transfer_cat_id: This is the bookwerx category_id of the category that corresponds with spot, funding, etc.  This category is created
during the initial installation of okcatbox and the id is provided via okcatbox.yaml.  We therefore know that this category must exist.

apikey8: This is the first 8 digits of the user's okcatbox apikey.  This corresponds with a bookwerx categories title.  Which category?  This
function will have to determine that.  Said category is created when the user initially creates his credentials so we know the category exists
but we must look it up here.

currency_id: This is the bookwerx currency id of the currency involved in the transfer.

We cannot blindly establish these accounts in advance for each new user.  Instead, create them here if necessary.
*/
//XXXXX
// was getTransferAccountID
//func getTaggedAccountID(client *httpclient.Client, transfer_cat_id int32, ok_access_key8 string, currency_id
//int32, cfg Config) (account_id int32, err error) {
//func getTransferAccountID() (account_id int32, err error) {

// 1. Find the category id for this particular user.
//selectt := "SELECT%20categories.id"
//from := "FROM%20categories"
//where := fmt.Sprintf("WHERE%%20categories.title%%3d%%27%s%%27", ok_access_key8)
//query := fmt.Sprintf("%s%%20%s%%20%s", selectt, from, where)
//url := fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)

//body, err := bwAPI.Get(client, url)
//if err != nil {
//fmt.Println("bookwerx-api.go:getTransferAccountID 1: get error: ", err)
//return -1, err
//}
//fixDot(body) // an array of categories.id

//n := make([]CatId, 0)
//err = json.NewDecoder(bytes.NewReader(body)).Decode(&n)
//if err != nil {
//fmt.Println("bookwerx-api.go: getTransferAccountID:", err)
//return
//}
//if len(n) == 0 {
//fmt.Println("bookwerx-api.go: getTransferAccountID: Category %s is not defined in bookwerx", ok_access_key8)
//return
//} else if len(n) > 1 {
//fmt.Println("bookwerx-api.go: getTransferAccountID: There are more than one suitable categories.  This should never happen.")
//return
//}
//apikey_cat_id := n[0].Id

// 2. Find the desired account ID
//selectt = "SELECT%20accounts.id"
//from = "FROM%20accounts_categories"
//join1 := "JOIN%20accounts%20ON%20accounts.id%3daccounts_categories.account_id"
//join2 := "JOIN%20currencies%20ON%20currencies.id%3daccounts.currency_id"
//where = fmt.Sprintf("WHERE%%20accounts_categories.category_id%%20IN%%20(%d,%d)", transfer_cat_id, apikey_cat_id)
//and := fmt.Sprintf("AND%%20currencies.id%%3d%d", currency_id)
//group := "GROUP%20BY%20accounts_categories.account_id%20HAVING%20COUNT(DISTINCT%20accounts_categories.account_id)%3d2"
//query = fmt.Sprintf("%s%%20%s%%20%s%%20%s%%20%s%%20%s%%20%s", selectt, from, join1, join2, where, and, group)
//url = fmt.Sprintf("%s/sql?query=%s&apikey=%s", cfg.Bookwerx.Server, query, cfg.Bookwerx.APIKey)
//body, err = bwAPI.Get(client, url)
//if err != nil {
//fmt.Println("bookwerx-api.go:getTransferAccountID 2: get error: ", err)
//return
//}
//fixDot(body)

//n1 := make([]AId, 0)
//err = json.NewDecoder(bytes.NewReader(body)).Decode(&n1)
//if err != nil {
//fmt.Println("bookwerx-api.go: getTransferAccountID:", err)
//return
//}
//if len(n1) == 1 {
//return n1[0].Id, nil
//} else if len(n1) > 1 {
//fmt.Println("bookwerx-api.go: getTransferAccountID: There are more than one suitable accounts.  This should never happen.")
//return
//}

// 3. If control passes here we know that the account doesn't already exist. So create it and tag it with the two categories and the given currency.
//if len(n1) == 0 {

// 3.1 Make the new account
//account_id, err = postAccount(client, currency_id, ok_access_key8, cfg)
//if err != nil {
//fmt.Println("bookwerx-api.go:getTransferAccountID 3.1: postAccount error: ", err)
//return -1, err
//}

// 3.2 Tag with the api key
//_, err = postAcctcat(client, account_id, apikey_cat_id, cfg)
//if err != nil {
//fmt.Println("bookwerx-api.go:getTransferAccountID 3.2: postAcctcat error: ", err)
//return -1, err
//}

// 3.3 Tag with funding, spot, whatever
//_, err = postAcctcat(client, account_id, transfer_cat_id, cfg)
//if err != nil {
//fmt.Println("bookwerx-api.go:getTransferAccountID 3.3: postAcctcat error: ", err)
//return -1, err
//}

//}

//return account_id, nil
//}

//XXXXX
/* deprecated */
/*func getHotWalletAccountID(client *httpclient.Client, accounts []AccountJoined, currency_symbol string, cfg Config) (account_id int32, err error) {

	// Can I find an account named #{apikey} using the same currency, that is tagged with the customer funding category?
	for _, account := range accounts {
		//if accountJoined.Title == ok_access_key8 {
		if strings.EqualFold(account.CurrencyShort.Symbol, currency_symbol) { // case insensitive compare
			// is this account tagged 'H'
			for _, b1 := range account.Categories {
				if strings.EqualFold(b1.CategorySymbol, "H") { // case insensitive compare
					account_id = account.AccountID
					break
				}
			}
		}
	}

	return account_id, nil
}*/
//XXXXX
/* deprecated */
/*func getSpotAvailableAccountID(client *httpclient.Client, accounts []AccountJoined, ok_access_key8 string, currency_id int32, currency_symbol string, cfg Config) (account_id int32, err error) {

	// 1. Can I find an account tagged spot-available _and_ apikey using the correct currency?
	account_id, found := searchSA(accounts, ok_access_key8, currency_symbol)

	// 2. If the account doesn't already exist, then create it and tag it with the customer's category and funding
	if !found {

		// 2.1 Find the category id for this apikey
		url := fmt.Sprintf("%s/category/bysym/%s?apikey=%s", cfg.Bookwerx.Server, ok_access_key8, cfg.Bookwerx.APIKey)
		category_id_api, err := getCategoryBySym(client, url)
		if err != nil {
			log.Error(err)
			return -1, err
		}

		// 2.2 Make the new account
		account_id, err = postAccount(client, currency_id, ok_access_key8, cfg)
		if err != nil {
			log.Error(err)
			return -1, err
		}

		// 2.3 Tag with the api key
		_, err = postAcctcat(client, account_id, category_id_api, cfg)
		if err != nil {
			log.Error(err)
			return -1, err
		}

		// 2.4 Tag with funding
		_, err = postAcctcat(client, account_id, cfg.Bookwerx.FundingCat, cfg)
		if err != nil {
			log.Error(err)
			return -1, err
		}

	}

	return account_id, nil
}
*/

//XXXXX
/*
// HAX: Rethink and factor these search* functions.
// Can I find an account tagged funding _and_ apikey8 using the correct currency?
func searchF(accounts []AccountJoined, ok_access_key8 string, currency_symbol string) (account_id int32, found bool) {

	// Can I find an account tagged funding _and_ apikey using the correct currency?
	cat_funding := false
	cat_api := false
	for _, account := range accounts {
		if strings.EqualFold(account.CurrencyShort.Symbol, currency_symbol) {
			for _, category := range account.Categories {
				if strings.EqualFold(category.CategorySymbol, "F") { // case insensitive compare
					cat_funding = true
				} else if strings.EqualFold(category.CategorySymbol, ok_access_key8) {
					cat_api = true
				}
			}
			if cat_funding && cat_api {
				return account.AccountID, true
			}
		}

	}

	return -1, false

}
//XXXXX
// Find all accounts tagged funding _and_ apikey8
func searchFAll(accounts []AccountJoined, ok_access_key8 string) (account_id []int32) {

	var account_ids []int32
	cat_funding := false
	cat_api := false
	for _, account := range accounts {
		for _, category := range account.Categories {
			if strings.EqualFold(category.CategorySymbol, "F") { // case insensitive compare
				cat_funding = true
			} else if strings.EqualFold(category.CategorySymbol, ok_access_key8) {
				cat_api = true
			}
		}
		if cat_funding && cat_api {
			account_ids = append(account_ids, account.AccountID)
		}
	}

	return account_ids
}
//XXXXX
// Can I find an account tagged spot-available _and_ apikey using the correct currency?
func searchSA(accounts []AccountJoined, ok_access_key8 string, currency_symbol string) (account_id int32, found bool) {

	// Can I find an account tagged funding _and_ apikey using the correct currency?
	cat_funding := false
	cat_api := false
	for _, account := range accounts {
		if strings.EqualFold(account.CurrencyShort.Symbol, currency_symbol) {
			for _, category := range account.Categories {
				if strings.EqualFold(category.CategorySymbol, "F") { // case insensitive compare
					cat_funding = true
				} else if strings.EqualFold(category.CategorySymbol, ok_access_key8) {
					cat_api = true
				}
			}
			if cat_funding && cat_api {
				return account.AccountID, true
			}
		}

	}

	return -1, false

}
*/

// Some of the functions use the stock go http client.
/*func GetClientA(urlBase string) (client *http.Client) {

	if len(urlBase) >= 6 && urlBase[:6] == "https:" {
		tr := &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		}
		return &http.Client{Transport: tr}
	}

	return &http.Client{}

}*/

/*func getAccounts(client *httpclient.Client, cfg Config) (accounts []AccountJoined, err error) {

	url := fmt.Sprintf("%s/accounts?apikey=%s", cfg.Bookwerx.Server, cfg.Bookwerx.APIKey)

	resp, err := client.Get(url, nil)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:getHotWalletAccountID 1: %v", err)
		log.Error(s)
		return nil, errors.New(s)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go:getHotWalletAccountID 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return nil, errors.New(s)
	}

	accountJoineds := make([]AccountJoined, 0)
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&accountJoineds)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:getHotWalletAccountID 3: %v", err)
		log.Error(s)
		return nil, errors.New(s)
	}

	return accountJoineds, nil
}

func getCategoryBySym(client *httpclient.Client, url string) (category_id int32, err error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:getCategoryBySym 1: %v", err)
		log.Error(s)
		return -1, errors.New(s)
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s := fmt.Sprintf("bookwerx-api.go:getCategoryBySym 2: Expected status=200, Received=%d, Body=%v", resp.StatusCode, body_string(resp))
		log.Error(s)
		return -1, errors.New(s)
	}

	n := Category{}
	err = json.NewDecoder(resp.Body).Decode(&n)
	if err != nil {
		s := fmt.Sprintf("bookwerx-api.go:getCategoryBySym 3: Error with JSON decoding.")
		log.Error(s)
		return -1, errors.New(s)
	}

	return n.Id, nil
}
*/
