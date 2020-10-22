package main

import (
	"flag"
	"github.com/bostontrader/okcommon"
	"github.com/hashicorp/go-memdb"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

// 1. Random necessities

// A transfer requires a source and a destination (from,to).  Each endpoint has an available and hold amount.
// Here we store the bookwerx category_ids for a particular transfer endpoint.
type AH struct {
	Available int32
	Hold      int32
}

// The OKCatbox will use a Bookwerx server for its internal operation.
type Bookwerx struct {
	APIKey string
	Server string

	// Customer accounts shall be tagged with these categories where applicable.
	// These are deprecated.  Use TransferCats instead.
	FundingCat       int32 `yaml:"funding_cat"`
	SpotAvailableCat int32 `yaml:"spot_available_cat"`
	SpotHoldCat      int32 `yaml:"spot_hold_cat"`

	// Any hot wallet shall be tagged with this category.
	HotWalletCat int32 `yaml:"hot_wallet_cat"`

	// The OKEx API specifies transfer endpoints using strings.  ie. "1" = spot, "6" = funding, etc.
	TransferCats map[string]AH `yaml:"transfer_cats"`
}

// When the OKCatbox executes it needs some configuration.
type Config struct {
	Bookwerx   Bookwerx
	ListenAddr string
}

// Most calls the OKCatbox API need some credentials
var obj utils.Credentials

// This is our in-memory db
var db *memdb.MemDB

func setResponseHeaders(w http.ResponseWriter, expectedResponseHeaders, extraResponseHeaders map[string]string) {
	for k, v := range expectedResponseHeaders {
		// Even if we expect Content-Length, don't send it
		if k != "Content-Length" {
			w.Header().Set(k, v)
		}
	}
	for k, v := range extraResponseHeaders {
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
			obj.Key = raw.(*utils.Credentials).Key
			obj.SecretKey = raw.(*utils.Credentials).SecretKey
			obj.Passphrase = raw.(*utils.Credentials).Passphrase
			valid = true
		}
	}
	return
}

/*func validateCurrencyParam(req *http.Request) (exists, valid bool) {

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
*/

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

func main() {

	// 1. Setup CLI parsing
	help := flag.Bool("help", false, "Guess what this flag does.")
	config := flag.String("config", "/path/to/okcatbox.yaml", "The config file for the OKCatbox")

	// Args[0] is the path to the program
	// Args[1] is okcatbox
	// Args[2:] are any remaining args.
	if len(os.Args) < 2 { // Invoke w/o any args
		flag.Usage()
		return
	}

	flag.Parse()

	if *help == true {
		flag.Usage()
		return
	}

	log.Println("The OKCatbox is using the following runtime args:")
	log.Println("help:", *help)
	log.Println("config:", *config)

	// Try to read the config file.
	data, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	cfg := Config{}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// 2. Create the in-memory DB schema and db
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"credentials": {
				Name: "credentials",
				Indexes: map[string]*memdb.IndexSchema{
					// We must have an id index so we use the Key field as the id
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Key"},
					},
				},
			},
			"currencies": {
				Name: "currencies",
				Indexes: map[string]*memdb.IndexSchema{
					// We must have an id index so we use the Key field as the id
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "CurrencyID"},
					},
				},
			},
			"withdrawalFees": {
				Name: "withdrawalFees",
				Indexes: map[string]*memdb.IndexSchema{
					// We must have an id index so we use the CurrencyID field as the id
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "CurrencyID"},
					},
				},
			},
		},
	}

	db, err = memdb.NewMemDB(schema)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// 3. Hardwire some initial values

	// 3.1 Currencies
	txn := db.Txn(true)

	if err := txn.Insert("currencies", &utils.CurrenciesEntry{CanDeposit: "1", CanWithdraw: "1", CurrencyID: "BTC", Name: "", MinWithdrawal: "0.0001"}); err != nil {
		log.Fatalf("error: %v", err)
	}
	if err := txn.Insert("currencies", &utils.CurrenciesEntry{CanDeposit: "1", CanWithdraw: "1", CurrencyID: "LTC", Name: "", MinWithdrawal: "0.0001"}); err != nil {
		log.Fatalf("error: %v", err)
	}

	// 3.2 Withdrawal fees
	if err := txn.Insert("withdrawalFees", &utils.WithdrawalFee{CurrencyID: "BTC", MinFee: "0.00040000", MaxFee: "0.01000000"}); err != nil {
		log.Fatalf("error: %v", err)
	}
	if err := txn.Insert("withdrawalFees", &utils.WithdrawalFee{CurrencyID: "LTC", MinFee: "0.00100000", MaxFee: "0.00200000"}); err != nil {
		log.Fatalf("error: %v", err)
	}

	txn.Commit()

	// 4. Setup request handlers

	// 4.1 Unique to the Catbox
	http.HandleFunc("/catbox/credentials", func(w http.ResponseWriter, req *http.Request) {
		catboxCredentialsHandler(w, req, cfg)
	})

	http.HandleFunc("/catbox/deposit", func(w http.ResponseWriter, req *http.Request) {
		catbox_depositHandler(w, req, cfg)
	})

	// 4.2 Funding
	http.HandleFunc("/api/account/v3/currencies", currencies)

	http.HandleFunc("/api/account/v3/deposit/address", func(w http.ResponseWriter, req *http.Request) {
		fundingDepositAddressHandler(w, req, cfg)
	})

	// Get deposits for all currencies
	http.HandleFunc("/api/account/v3/deposit/history", func(w http.ResponseWriter, req *http.Request) {
		account_depositHistoryHandler(w, req, cfg, "")
	})

	// Get deposits for only the currency specified in the URL .../history/<currency>
	http.HandleFunc("/api/account/v3/deposit/history/", func(w http.ResponseWriter, req *http.Request) {
		n := strings.Split(req.URL.Path, "/")
		currency := n[len(n)-1]
		account_depositHistoryHandler(w, req, cfg, currency)
	})

	http.HandleFunc("/api/account/v3/transfer", func(w http.ResponseWriter, req *http.Request) {
		account_transferHandler(w, req, cfg)
	})

	http.HandleFunc("/api/account/v3/wallet", func(w http.ResponseWriter, req *http.Request) {
		account_walletHandler(w, req, cfg)
	})

	http.HandleFunc("/api/account/v3/withdrawal/fee", withdrawalFee)

	// 4.3 Spot
	http.HandleFunc("/api/spot/v3/accounts", func(w http.ResponseWriter, req *http.Request) {
		spot_accountsHandler(w, req, cfg)
	})

	// 5. Let er rip!
	log.Printf("The Catbox is listening to %s\n", cfg.ListenAddr)
	err = http.ListenAndServe(cfg.ListenAddr, nil)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
