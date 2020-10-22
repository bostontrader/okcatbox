package main

import (
	"encoding/json"
	"fmt"
	utils "github.com/bostontrader/okcommon"
	"net/http"
)

// /api/account/v3/withdrawal/fee
func withdrawalFee(w http.ResponseWriter, req *http.Request) {
	retVal := generateWithdrawalFeeResponse(w, req, "GET", "/api/account/v3/withdrawal/fee")
	fmt.Fprintf(w, string(retVal))
}

func generateWithdrawalFeeResponse(w http.ResponseWriter, req *http.Request, verb string, endpoint string) (retVal []byte) {

	retVal, err := checkSigHeaders(w, req)
	if err {
		return

	} else {
		setResponseHeaders(w, utils.ExpectedResponseHeaders, map[string]string{"Strict-Transport-Security": "", "Vary": ""})

		withdrawalFees := make([]utils.WithdrawalFee, 1)
		withdrawalFees[0] = utils.WithdrawalFee{CurrencyID: "c", MinFee: "minf", MaxFee: "maxf"}
		retVal, _ = json.Marshal(withdrawalFees)

	}

	return
}
