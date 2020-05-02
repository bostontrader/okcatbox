package main

import (
	"encoding/json"
	"fmt"
	"github.com/bostontrader/okcommon"
	"github.com/hashicorp/go-memdb"
	uuid "github.com/nu7hatch/gouuid"
	"math/rand"
	"net/http"
	"time"
)

// 1. Random necessities

var obj utils.Credentials

// This is our in-memory db
var db *memdb.MemDB

func setResponseHeaders(w http.ResponseWriter, extraHeaders map[string]string) {
	w.Header().Set("Cache-Control", "")
	w.Header().Set("Expires", "")
	w.Header().Set("Pragma", "")
	w.Header().Set("Set-Cookie", "")
	w.Header().Set("Via", "")
	w.Header().Set("X-Brokerid", "")
	w.Header().Set("X-Content-Type-Options", "")
	w.Header().Set("X-Frame-Options", "")
	w.Header().Set("X-Kong-Proxy-Latency", "")
	w.Header().Set("X-Kong-Upstream-Latency", "")
	w.Header().Set("X-Ratelimit-Limit-Second", "")
	w.Header().Set("X-Ratelimit-Remaining-Second", "")
	w.Header().Set("X-Xss-Protection", "")

	for k, v := range extraHeaders {
		w.Header().Set(k, v)
	}
}

func validateAccessKey(headers map[string][]string) (exists, valid bool) {
	if value, ok := headers["Ok-Access-Key"]; ok {
		// Ok-Access-Key is in the headers.  Now try to validate it.
		exists = true

		// Create read-only transaction
		txn := db.Txn(false)
		defer txn.Abort()

		// Lookup by id/key
		raw, err := txn.First("credentials", "id", value[0])
		if err != nil {
			panic(err)
		}
		if raw == nil {
			// not found, do nothing, flag already correctly set
		} else {
			fmt.Printf(raw.(*utils.Credentials).Key)
			obj.Key = raw.(*utils.Credentials).Key
			obj.SecretKey = raw.(*utils.Credentials).SecretKey
			obj.Passphrase = raw.(*utils.Credentials).Passphrase
			valid = true
		}
	}
	return
}

func validateCurrencyParam(req *http.Request) (exists, valid bool) {

	if value, ok := req.URL.Query()["currency"]; ok {
		// currency exists as a parameter.  Now try to validate it.
		exists = true

		// Create read-only transaction
		txn := db.Txn(false)
		defer txn.Abort()

		// Lookup by id/CurrencyID
		raw, err := txn.First("withdrawalFees", "id", value[0])
		if err != nil {
			panic(err)
		}
		if raw == nil {
			// not found, do nothing, flag already correctly set
		} else {
			fmt.Printf(raw.(*utils.WithdrawalFee).CurrencyID)
			//obj.Key = raw.(*utils.Credentials).Key
			//obj.SecretKey = raw.(*utils.Credentials).SecretKey
			//obj.Passphrase = raw.(*utils.Credentials).Passphrase
			valid = true
		}
	}
	return
}

func validatePassPhrase(headers map[string][]string) (exists, valid bool) {
	if value, ok := headers["Ok-Access-Passphrase"]; ok {
		// Ok-Access-Passphrase is in the headers.  Now try to validate it.
		exists = true
		if value[0] == "valid passphrase" {
			valid = true
		}
	}
	return
}

func validateSign(req *http.Request) (exists, valid bool) {
	if sigValue, ok := req.Header["Ok-Access-Sign"]; ok {
		exists = true

		// Ok-Access-Sign is in the headers.  Now try to validate it.
		if tsValue, ok := req.Header["Ok-Access-Timestamp"]; ok {
			// Now build a signature
			timestamp := tsValue[0]
			prehash := timestamp + req.Method + req.RequestURI
			fmt.Println("prehash: ", prehash)
			encoded, _ := utils.HmacSha256Base64Signer(prehash, obj.SecretKey)
			if sigValue[0] == encoded {
				valid = true
			}
		} else {
			// No timestamp, therefore the signature cannot be right.  Do nothing cuz the valid flag is already set correctly.
		}

	}
	return
}

func validateTimestamp(headers map[string][]string) (exists, valid, expired bool) {
	expired = true
	if value, ok := headers["Ok-Access-Timestamp"]; ok {
		// Ok-Access-Timestamp is in the headers.  Try to validate it.
		exists = true

		timestamp, err := time.Parse(time.RFC3339, value[0])
		if err == nil {
			valid = true
			age := time.Now().Add(time.Duration(-30) * time.Second)
			if timestamp.After(age) {
				expired = false
			}
		}
	}
	return
}

func generateCurrenciesResponse(w http.ResponseWriter, req *http.Request, verb string, endpoint string) (retVal []byte) {

	fmt.Println(req, "\n")

	// 1. Detect whether or not the various parameters exist.  If so, detect whether or not they are valid.  Set suitable flags so that subsequent error checking can occur.

	// 1.1 Ok-Access-Key
	akeyP, akeyV := validateAccessKey(req.Header)

	// 1.2 Ok-Access-Timestamp
	atimestampP, atimestampV, atimestampEx := validateTimestamp(req.Header)

	// 1.3 Ok-Access-Passphrase
	apassphraseP, apassphraseV := validatePassPhrase(req.Header)

	// 1.4 Ok-Access-Sign
	asignP, asignV := validateSign(req)

	// The order of comparison and the boolean senses have been empirically chosen to match the order of error detection in the real OKEx server.
	if !akeyP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30001()) // Access key required

	} else if !asignP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30002()) // Signature required

	} else if !atimestampP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30003()) // Timestamp required

	} else if !atimestampV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30005()) // Invalid timestamp

	} else if atimestampEx {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30008()) // Timestamp expired

	} else if !akeyV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30006()) // Invalid access key

	} else if !apassphraseP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30004()) // Passphrase required

	} else if !apassphraseV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30015()) // Invalid Passphrase

	} else if !asignV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30013()) // Invalid Sign

	} else {
		//retVal, _ = getJsonResponseGood()
		setResponseHeaders(w, map[string]string{"Strict-Transport-Security": "", "Vary": ""})

		currencies := make([]utils.CurrenciesEntry, 1)
		currencies[0] = utils.CurrenciesEntry{CanDeposit: "cd", CanWithdraw: "cw", CurrencyID: "c", Name: "n", MinWithdrawal: "mw"}
		retVal, _ := json.Marshal(currencies)
		return retVal
	}

	return
}

func generateDepositAddressResponse(w http.ResponseWriter, req *http.Request, verb string, endpoint string) (retVal []byte) {

	fmt.Println(req, "\n")

	// 1. Detect whether or not the various parameters exist.  If so, detect whether or not they are valid.  Set suitable flags so that subsequent error checking can occur.

	// 1.1 Ok-Access-Key
	akeyP, akeyV := validateAccessKey(req.Header)

	// 1.2 Ok-Access-Timestamp
	atimestampP, atimestampV, atimestampEx := validateTimestamp(req.Header)

	// 1.3 Ok-Access-Passphrase
	apassphraseP, apassphraseV := validatePassPhrase(req.Header)

	// 1.4 Ok-Access-Sign
	asignP, asignV := validateSign(req)

	// 1.5 currency parameter
	curParamP, curParamV := validateCurrencyParam(req)
	fmt.Println(curParamP, curParamV)
	// The order of comparison and the boolean senses have been empirically chosen to match the order of error detection in the real OKEx server.
	if !akeyP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30001()) // Access key required

	} else if !asignP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30002()) // Signature required

	} else if !atimestampP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30003()) // Timestamp required

	} else if !atimestampV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30005()) // Invalid timestamp

	} else if atimestampEx {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30008()) // Timestamp expired

	} else if !akeyV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30006()) // Invalid access key

	} else if !apassphraseP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30004()) // Passphrase required

	} else if !apassphraseV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30015()) // Invalid Passphrase

	} else if !asignV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30013()) // Invalid Sign
	} else if !curParamP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30023()) // currency cannot be blank
	} else if curParamP && !curParamV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)

		// The currency param should exist
		if value, ok := req.URL.Query()["currency"]; ok {
			retVal, _ = json.Marshal(utils.Err30031(value[0])) // Invalid token
		} else {
			// This should never happen
			retVal, _ = json.Marshal(utils.Err30031(""))
		}

	} else { // currency must be present and valid to get here
		setResponseHeaders(w, map[string]string{"Strict-Transport-Security": ""})

		depositAddresses := make([]utils.DepositAddress, 1)
		depositAddresses[0] = utils.DepositAddress{Address: "a", CurrencyID: "c", To: 6}
		retVal, _ = json.Marshal(depositAddresses)
	}

	return
}

func generateWalletResponse(w http.ResponseWriter, req *http.Request, verb string, endpoint string) (retVal []byte) {

	fmt.Println(req, "\n")

	// 1. Detect whether or not the various parameters exist.  If so, detect whether or not they are valid.  Set suitable flags so that subsequent error checking can occur.

	// 1.1 Ok-Access-Key
	akeyP, akeyV := validateAccessKey(req.Header)

	// 1.2 Ok-Access-Timestamp
	atimestampP, atimestampV, atimestampEx := validateTimestamp(req.Header)

	// 1.3 Ok-Access-Passphrase
	apassphraseP, apassphraseV := validatePassPhrase(req.Header)

	// 1.4 Ok-Access-Sign
	asignP, asignV := validateSign(req)

	// The order of comparison and the boolean senses have been empirically chosen to match the order of error detection in the real OKEx server.
	if !akeyP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30001()) // Access key required

	} else if !asignP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30002()) // Signature required

	} else if !atimestampP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30003()) // Timestamp required

	} else if !atimestampV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30005()) // Invalid timestamp

	} else if atimestampEx {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30008()) // Timestamp expired

	} else if !akeyV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30006()) // Invalid access key

	} else if !apassphraseP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30004()) // Passphrase required

	} else if !apassphraseV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30015()) // Invalid Passphrase

	} else if !asignV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30013()) // Invalid Sign

	} else {
		setResponseHeaders(w, map[string]string{"Strict-Transport-Security": ""})

		walletEntries := make([]utils.WalletEntry, 1)
		walletEntries[0] = utils.WalletEntry{Available: "a", Balance: "b", CurrencyID: "c", Hold: "h"}
		retVal, _ := json.Marshal(walletEntries)
		return retVal
	}

	return
}

func generateWithdrawalFeeResponse(w http.ResponseWriter, req *http.Request, verb string, endpoint string) (retVal []byte) {

	fmt.Println(req, "\n")

	// 1. Detect whether or not the various parameters exist.  If so, detect whether or not they are valid.  Set suitable flags so that subsequent error checking can occur.

	// 1.1 Ok-Access-Key
	akeyP, akeyV := validateAccessKey(req.Header)

	// 1.2 Ok-Access-Timestamp
	atimestampP, atimestampV, atimestampEx := validateTimestamp(req.Header)

	// 1.3 Ok-Access-Passphrase
	apassphraseP, apassphraseV := validatePassPhrase(req.Header)

	// 1.4 Ok-Access-Sign
	asignP, asignV := validateSign(req)

	// 1.5 currency parameter
	curParamP, curParamV := validateCurrencyParam(req)
	fmt.Println("curParamP:", curParamP, "curParamV", curParamV)
	// The order of comparison and the boolean senses have been empirically chosen to match the order of error detection in the real OKEx server.
	if !akeyP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30001()) // Access key required

	} else if !asignP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30002()) // Signature required

	} else if !atimestampP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30003()) // Timestamp required

	} else if !atimestampV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30005()) // Invalid timestamp

	} else if atimestampEx {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30008()) // Timestamp expired

	} else if !akeyV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30006()) // Invalid access key

	} else if !apassphraseP {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30004()) // Passphrase required

	} else if !apassphraseV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)
		retVal, _ = json.Marshal(utils.Err30015()) // Invalid Passphrase

	} else if !asignV {
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(401)
		retVal, _ = json.Marshal(utils.Err30013()) // Invalid Sign
	} else if curParamP && !curParamV {
		fmt.Println("parameter exists but is not valid")
		setResponseHeaders(w, map[string]string{})
		w.WriteHeader(400)

		// The currency param should exist
		if value, ok := req.URL.Query()["currency"]; ok {
			// currency exists as a parameter.  Now try to validate it.
			retVal, _ = json.Marshal(utils.Err30031(value[0])) // Invalid token
		} else {
			// This should never happen
			retVal, _ = json.Marshal(utils.Err30031(""))
		}
	} else if curParamP && curParamV {
		setResponseHeaders(w, map[string]string{"Strict-Transport-Security": ""})

		withdrawalFees := make([]utils.WithdrawalFee, 1)
		withdrawalFees[0] = utils.WithdrawalFee{CurrencyID: "c", MinFee: "minf", MaxFee: "maxf"}
		retVal, _ = json.Marshal(withdrawalFees)

	} else {
		setResponseHeaders(w, map[string]string{"Strict-Transport-Security": "", "Vary": ""})

		withdrawalFees := make([]utils.WithdrawalFee, 1)
		withdrawalFees[0] = utils.WithdrawalFee{CurrencyID: "c", MinFee: "minf", MaxFee: "maxf"}
		retVal, _ = json.Marshal(withdrawalFees)

	}

	return
}

const letterBytes = "ABCDEF0123456789"
const (
	letterIdxBits = 4                    // 4 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func getCredentials() ([]byte, error) {

	u4, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	var tryArr = make([]string, 1)
	for i := range tryArr {
		tryArr[i] = RandStringBytesMaskImprSrc(32)
	}

	n1 := &utils.Credentials{Key: u4.String(), SecretKey: tryArr[0], Passphrase: "valid passphrase"}

	// Save this in the db
	txn := db.Txn(true)
	if err := txn.Insert("credentials", n1); err != nil {
		panic(err)
	}
	txn.Commit()

	return json.Marshal(n1)
}

// 2. Request handlers

// /api/account/v3/currencies
func currencies(w http.ResponseWriter, req *http.Request) {
	retVal := generateCurrenciesResponse(w, req, "GET", "/api/account/v3/currencies")
	fmt.Fprintf(w, string(retVal))
}

// /api/account/v3/deposit/address
func depositAddress(w http.ResponseWriter, req *http.Request) {
	retVal := generateDepositAddressResponse(w, req, "GET", "/api/account/v3/deposit/address")
	fmt.Fprintf(w, string(retVal))
}

// /api/account/v3/wallet
func wallet(w http.ResponseWriter, req *http.Request) {
	retVal := generateWalletResponse(w, req, "GET", "/api/account/v3/wallet")
	fmt.Fprintf(w, string(retVal))
}

// /api/account/v3/withdrawal/fee
func withdrawalFee(w http.ResponseWriter, req *http.Request) {
	retVal := generateWithdrawalFeeResponse(w, req, "GET", "/api/account/v3/withdrawal/fee")
	fmt.Fprintf(w, string(retVal))
}

// /credentials
func credentials(w http.ResponseWriter, req *http.Request) {
	retVal, _ := getCredentials()
	fmt.Fprintf(w, string(retVal))
}

func main() {

	// 1. Create the in-memory DB schema and db
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"credentials": &memdb.TableSchema{
				Name: "credentials",
				Indexes: map[string]*memdb.IndexSchema{
					// We must have an id index so we use the Key field as the id
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Key"},
					},
				},
			},
			"withdrawalFees": &memdb.TableSchema{
				Name: "withdrawalFees",
				Indexes: map[string]*memdb.IndexSchema{
					// We must have an id index so we use the CurrencyID field as the id
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "CurrencyID"},
					},
				},
			},
		},
	}

	var err error
	db, err = memdb.NewMemDB(schema)
	if err != nil {
		panic(err)
	}

	// 2. Hardwire a first set of credentials
	txn := db.Txn(true)
	n1 := &utils.Credentials{"47477ba4-74ad-4649-4c71-36c587a82c7d", "4790CA744289696413598ECBAB430B79", "valid passphrase"}
	if err := txn.Insert("credentials", n1); err != nil {
		panic(err)
	}
	txn.Commit()

	// 3. Hardwire a first set of withdrawal fees
	txn = db.Txn(true)
	//n2 := []*utils.WithdrawalFee {
	//&utils.WithdrawalFee{"BTC", "0.00040000", "0.01000000"},
	//&utils.WithdrawalFee{"LTC", "0.00100000", "0.00200000"},
	//}
	n2 := &utils.WithdrawalFee{"BTC", "0.00040000", "0.01000000"}
	n3 := &utils.WithdrawalFee{"LTC", "0.00100000", "0.00200000"}

	if err := txn.Insert("withdrawalFees", n2); err != nil {
		panic(err)
	}
	if err := txn.Insert("withdrawalFees", n3); err != nil {
		panic(err)
	}

	txn.Commit()

	// 4. Setup request handlers
	http.HandleFunc("/api/account/v3/wallet", wallet)
	http.HandleFunc("/api/account/v3/deposit/address", depositAddress)
	http.HandleFunc("/api/account/v3/currencies", currencies)
	http.HandleFunc("/api/account/v3/withdrawal/fee", withdrawalFee)
	http.HandleFunc("/credentials", credentials)

	// 5. Let er rip!
	http.ListenAndServe(":8090", nil)
}