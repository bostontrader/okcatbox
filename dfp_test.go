package main

import (
	"testing"
)

func Test(t *testing.T) {

	if !(abs(-1) == 1) {
		t.Error()
	}
	if !(abs(1) == 1) {
		t.Error()
	}

	if !(dfp_fmt(DFP{555, -4}, -5) == DFPFmt{"0.05550", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -4}, -4) == DFPFmt{"0.0555", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -4}, -3) == DFPFmt{"0.056", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -4}, -2) == DFPFmt{"0.06", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -4}, -1) == DFPFmt{"0.1", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -4}, 0) == DFPFmt{"0", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -3}, -5) == DFPFmt{"0.55500", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -3}, -4) == DFPFmt{"0.5550", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -3}, -3) == DFPFmt{"0.555", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -3}, -2) == DFPFmt{"0.56", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -3}, -1) == DFPFmt{"0.6", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -3}, 0) == DFPFmt{"1", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -3}, 1) == DFPFmt{"0", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -2}, -5) == DFPFmt{"5.55000", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -2}, -4) == DFPFmt{"5.5500", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -2}, -3) == DFPFmt{"5.550", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -2}, -2) == DFPFmt{"5.55", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -2}, -1) == DFPFmt{"5.6", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -2}, 0) == DFPFmt{"6", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -2}, 1) == DFPFmt{"10", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -2}, 2) == DFPFmt{"0", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -1}, -5) == DFPFmt{"55.50000", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -1}, -4) == DFPFmt{"55.5000", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -1}, -3) == DFPFmt{"55.500", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -1}, -2) == DFPFmt{"55.50", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -1}, -1) == DFPFmt{"55.5", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -1}, 0) == DFPFmt{"56", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -1}, 1) == DFPFmt{"60", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -1}, 2) == DFPFmt{"100", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, -1}, 3) == DFPFmt{"0", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 0}, -5) == DFPFmt{"555.00000", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 0}, -4) == DFPFmt{"555.0000", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 0}, -3) == DFPFmt{"555.000", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 0}, -2) == DFPFmt{"555.00", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 0}, -1) == DFPFmt{"555.0", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 0}, 0) == DFPFmt{"555", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 0}, 1) == DFPFmt{"560", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 0}, 2) == DFPFmt{"600", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 0}, 3) == DFPFmt{"1000", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 0}, 4) == DFPFmt{"0", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 1}, -5) == DFPFmt{"5550.00000", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 1}, -4) == DFPFmt{"5550.0000", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 1}, -3) == DFPFmt{"5550.000", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 1}, -2) == DFPFmt{"5550.00", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 1}, -1) == DFPFmt{"5550.0", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 1}, 0) == DFPFmt{"5550", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 1}, 1) == DFPFmt{"5550", false}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 1}, 2) == DFPFmt{"5600", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 1}, 3) == DFPFmt{"6000", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 1}, 4) == DFPFmt{"10000", true}) {
		t.Error()
	}
	if !(dfp_fmt(DFP{555, 1}, 5) == DFPFmt{"0", true}) {
		t.Error()
	}

	if !(dfp_round(-5, DFP{555, -4}) == DFP{555, -4}) {
		t.Error()
	}
	if !(dfp_round(-4, DFP{555, -4}) == DFP{555, -4}) {
		t.Error()
	}
	if !(dfp_round(-3, DFP{555, -4}) == DFP{56, -3}) {
		t.Error()
	}
	if !(dfp_round(-2, DFP{555, -4}) == DFP{6, -2}) {
		t.Error()
	}
	if !(dfp_round(-1, DFP{555, -4}) == DFP{1, -1}) {
		t.Error()
	}
	if !(dfp_round(0, DFP{555, -4}) == DFP{0, 0}) {
		t.Error()
	}
	if !(dfp_round(-4, DFP{555, -3}) == DFP{555, -3}) {
		t.Error()
	}
	if !(dfp_round(-3, DFP{555, -3}) == DFP{555, -3}) {
		t.Error()
	}
	if !(dfp_round(-2, DFP{555, -3}) == DFP{56, -2}) {
		t.Error()
	}
	if !(dfp_round(-1, DFP{555, -3}) == DFP{6, -1}) {
		t.Error()
	}
	if !(dfp_round(0, DFP{555, -3}) == DFP{1, 0}) {
		t.Error()
	}
	if !(dfp_round(1, DFP{555, -3}) == DFP{0, 1}) {
		t.Error()
	}
	if !(dfp_round(-3, DFP{555, -2}) == DFP{555, -2}) {
		t.Error()
	}
	if !(dfp_round(-2, DFP{555, -2}) == DFP{555, -2}) {
		t.Error()
	}
	if !(dfp_round(-1, DFP{555, -2}) == DFP{56, -1}) {
		t.Error()
	}
	if !(dfp_round(0, DFP{555, -2}) == DFP{6, 0}) {
		t.Error()
	}
	if !(dfp_round(1, DFP{555, -2}) == DFP{1, 1}) {
		t.Error()
	}
	if !(dfp_round(2, DFP{555, -2}) == DFP{0, 2}) {
		t.Error()
	}
	if !(dfp_round(-2, DFP{555, -1}) == DFP{555, -1}) {
		t.Error()
	}
	if !(dfp_round(-1, DFP{555, -1}) == DFP{555, -1}) {
		t.Error()
	}
	if !(dfp_round(0, DFP{555, -1}) == DFP{56, 0}) {
		t.Error()
	}
	if !(dfp_round(1, DFP{555, -1}) == DFP{6, 1}) {
		t.Error()
	}
	if !(dfp_round(2, DFP{555, -1}) == DFP{1, 2}) {
		t.Error()
	}
	if !(dfp_round(3, DFP{555, -1}) == DFP{0, 3}) {
		t.Error()
	}
	if !(dfp_round(-1, DFP{555, 0}) == DFP{555, 0}) {
		t.Error()
	}
	if !(dfp_round(0, DFP{555, 0}) == DFP{555, 0}) {
		t.Error()
	}
	if !(dfp_round(1, DFP{555, 0}) == DFP{56, 1}) {
		t.Error()
	}
	if !(dfp_round(2, DFP{555, 0}) == DFP{6, 2}) {
		t.Error()
	}
	if !(dfp_round(3, DFP{555, 0}) == DFP{1, 3}) {
		t.Error()
	}
	if !(dfp_round(4, DFP{555, 0}) == DFP{0, 4}) {
		t.Error()
	}
	if !(dfp_round(0, DFP{555, 1}) == DFP{555, 1}) {
		t.Error()
	}
	if !(dfp_round(1, DFP{555, 1}) == DFP{555, 1}) {
		t.Error()
	}
	if !(dfp_round(2, DFP{555, 1}) == DFP{56, 2}) {
		t.Error()
	}
	if !(dfp_round(3, DFP{555, 1}) == DFP{6, 3}) {
		t.Error()
	}
	if !(dfp_round(4, DFP{555, 1}) == DFP{1, 4}) {
		t.Error()
	}
	if !(dfp_round(5, DFP{555, 1}) == DFP{0, 5}) {
		t.Error()
	}

}
