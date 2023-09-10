package paycalc

import (
	"bytes"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name string
		top  float64
		buf  []byte
	}{
		{"simple", .10, simpleCSV},
		{"top10", .10, top10},
		{"top15", .15, top15},
		{"top20", .20, top20},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			testLoad(t, test.name, test.top, test.buf)
		})
	}
}

func testLoad(t *testing.T, name string, top float64, buf []byte) {
	t.Helper()
	tbl, err := LoadBytes(name, top, buf)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	t.Logf("\n%s\n%c\n%m", tbl, tbl, tbl)
	b := new(bytes.Buffer)
	if err := tbl.WriteCSV(b); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if s, exp := b.String(), string(buf); s != exp {
		t.Errorf("expected generated table csv output to match input, expected:\n%s\ngot:\n%s", exp, s)
	}
}

func TestPaidUnallocated(t *testing.T) {
	tests := []struct {
		typ     Type
		entries int
		paid    int
		row     int
		col     int
		s       string
		exp     float64
	}{
		{Top10, 2, 1, 0, 30, "1st/2", 0.0},
		{Top15, 2, 1, 0, 30, "1st/2", 0.0},
		{Top20, 2, 1, 0, 27, "1st/2", 0.0},
		{Top10, 12, 2, 1, 28, "2nd/11-30", 20.0},
		{Top15, 12, 2, 1, 28, "2nd/11-20", 20.0},
		{Top20, 12, 3, 2, 25, "3rd/11-20", 15.0},
		{Top10, 20, 2, 1, 28, "2nd/11-30", 20.0},
		{Top15, 20, 3, 2, 28, "3rd/11-20", 0.0},
		{Top20, 20, 4, 3, 25, "4th/11-20", 0.0},
		{Top10, 90, 9, 8, 23, "9th/76-100", 2.5},
		{Top15, 60, 9, 8, 24, "9th/51-60", 0.0},
		{Top20, 45, 9, 8, 22, "9th/41-50", 3.2},
		{Top10, 64, 7, 6, 24, "7th/61-75", 4.5},
		{Top15, 64, 10, 9, 23, "10th/61-75", 8.8},
		{Top20, 64, 13, 10, 21, "11-13/51-75", 2.1 * 2},
		{Top10, 87, 9, 8, 23, "9th/76-100", 2.5},
		{Top15, 87, 14, 10, 22, "11-14/76-100", 1.86 * 1},
		{Top20, 87, 18, 11, 20, "16-18/76-100", 1.3 * 2},
		{Top10, 93, 10, 9, 23, "10th/76-100", 0.0},
		{Top15, 93, 14, 10, 22, "11-14/76-100", 1.86 * 1},
		{Top20, 93, 19, 11, 20, "16-19/76-100", 1.3 * 1},
		{Top10, 100, 10, 9, 23, "10th/76-100", 0.0},
		{Top15, 100, 15, 10, 22, "11-15/76-100", 0.0},
		{Top20, 100, 20, 11, 20, "16-20/76-100", 0.0},
		{Top10, 101, 11, 10, 22, "11th/101-150", 2.1 * 4},
		{Top15, 101, 16, 11, 21, "16th/101-125", 6.4},
		{Top20, 101, 21, 12, 19, "21st/101-125", 4.2},
		{Top10, 234, 24, 12, 20, "21-24/201-250", 1.0 * 1},
		{Top15, 234, 36, 15, 18, "36th/201-250", 0.45 * 4},
		{Top20, 234, 47, 16, 16, "41-47/201-250", 0.45 * 3},
		{Top10, 247, 25, 12, 20, "21-25/201-250", 0.0},
		{Top15, 247, 38, 15, 18, "36-38/201-250", 0.45 * 2},
		{Top20, 247, 50, 16, 16, "41-50/201-250", 0.0},
	}
	for i, test := range tests {
		paid, row, col := test.typ.Paid(test.entries)
		switch s := test.typ.MaxLevelTitle(paid) + "/" + test.typ.EntriesTitle(test.entries); {
		case paid != test.paid:
			t.Errorf("test %d %n %d expected paid %d, got: %d", i, test.typ, test.entries, test.paid, paid)
		case row != test.row:
			t.Errorf("test %d %n %d expected row %d, got: %d", i, test.typ, test.entries, test.row, row)
		case col != test.col:
			t.Errorf("test %d %n %d expected col %d, got: %d", i, test.typ, test.entries, test.col, col)
		case s != test.s:
			t.Errorf("test %d %n %d expected %q, got: %q", i, test.typ, test.entries, test.s, s)
		}
		if f, exp := test.typ.Unallocated(paid, row, col)*100.0, test.exp; !Equal(f, exp) {
			t.Errorf("test %d %n %d expected %f, got: %f", i, test.typ, test.entries, exp, f)
		}
	}
}
