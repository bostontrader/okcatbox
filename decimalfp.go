package main

import (
	"fmt"
	"strings"
)

type DFP struct {
	Amount int64
	Exp    int8
}

// Not really a decimal floating point function, just a convenience.
func abs(n int8) (abs int8) {
	if n >= 0 {
		return n
	} else {
		return -n
	}
}

// The dfp_fmt function will return its results using a DFPFmt struct.  Said struct contains the desired string format as well as a boolean to indicate whether or not the formatted string omits any digits due to rounding.
type DFPFmt struct {
	s string // the result
	r bool   // True = loss of precision displayed due to rounding
}

// Given a DFP and an int8 p, round the DFP to the 10^p position and produce a formatted string with a decimal point and suitable leading and trailing zeros as necessary.  Tamper with this at your own peril.  Feel free to figure this out, refactor, or simplify.
func dfp_fmt(dfp DFP, p int8) DFPFmt {

	negative := dfp.Amount < 0

	sign := ""
	if negative {
		sign = "-"
	}

	//  Ensure that the quantity of digits available does not exceed the quantity desired to display.
	//  If dfp_norm != norm dfp then there are digits obscured by rounding.
	dfp_norm := DFPNorm(dfp_round(p, dfp))

	samt := fmt.Sprintf("%d", dfp_norm.Amount)
	if negative {
		samt = fmt.Sprintf("%d", dfp_norm.Amount)[1:]
	}

	slen := len(samt)

	// A sequence of tedious contortions best divided and conquered thus
	s2 := s2(dfp_norm, int8(slen), p, samt)
	s3 := s3(s2)

	// If dfp_norm != norm(dfp) then there are digits obscured by rounding.
	return DFPFmt{sign + s3, dfp_norm != (DFPNorm(dfp))}

}

func insertDp(pos int, samt string) (retVal string) {
	lhs := samt[0:pos]
	rhs := samt[pos:]
	return lhs + "." + rhs
}

func s2(dfp_norm DFP, slen int8, p int8, samt string) (retVal string) {
	if dfp_norm.Exp < 0 {
		// move dp to the left
		if abs(dfp_norm.Exp) < slen {
			// split samt and insert dp
			return insertDp(int(slen-abs(dfp_norm.Exp)),
				samt+strings.Repeat("0", int(abs(p)+dfp_norm.Exp)))
		} else if abs(dfp_norm.Exp) == slen {
			return "0." + samt + strings.Repeat("0", int(abs(p)-slen))
		} else {
			//-- abs(dfp_norm.Exp) > slen
			n1 := abs(dfp_norm.Exp) - slen
			n2 := abs(p) - slen - n1
			return "0." + strings.Repeat("0", int(n1)) + samt + strings.Repeat("0", int(n2))
		}

	} else if dfp_norm.Exp == 0 {

		//-- don't move the dp
		if p > 0 {
			return samt
		} else {
			return samt + "." + strings.Repeat("0", int(-p))
		}
	} else {
		//-- dfp_norm.Exp > 0, move dp to the right
		if p > 0 {
			return samt + strings.Repeat("0", int(dfp_norm.Exp))
		} else {
			return samt + strings.Repeat("0", int(dfp_norm.Exp)) + "." + strings.Repeat("0", int(-p))
		}
	}
}

func s3(s2 string) string {
	if s2[len(s2)-1:] == "." {
		return s2[0 : len(s2)-1]
	} else {
		return s2
	}
}

func DFPNorm(d DFP) (retVal DFP) {

	if d.Amount == 0 {
		return DFP{0, 0}
	} else if d.Amount%10 == 0 {
		return DFPNorm(DFP{d.Amount / 10, d.Exp + 1})
	} else {
		//-- already in normal form
		return d
	}
}

//-- Given an int8 p and a DFP d, round d as necessary such that 10^p is the least significant digit, and return the final rounded DFP result

func dfp_round(p int8, d DFP) (retVal DFP) {

	if p <= d.Exp {
		//-- already rounded enough
		return d
	} else {
		last_digit := d.Amount % 10
		new_Amount := d.Amount / 10
		if last_digit >= 5 {
			new_Amount = d.Amount/10 + 1
		}
		new_dfp := DFP{new_Amount, d.Exp + 1}
		return dfp_round(p, new_dfp)
	}
}
