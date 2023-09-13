package paycalc

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestEntries(t *testing.T) {
	tests := []struct {
		typ     Type
		entries int
		col     int
		exp     string
	}{
		{Top10, 1, -1, ""},
		{Top15, 1, -1, ""},
		{Top20, 1, -1, ""},
		{Top10, 2, 30, "2"},
		{Top10, 5000, 1, "4501-5000"},
		{Top10, 256, 19, "251-300"},
		{Top10, 30, 28, "11-30"},
		{Top10, 10, 29, "3-10"},
		{Top15, 1522, 6, "1501-1750"},
		{Top15, 301, 16, "301-350"},
		{Top20, 899, 10, "801-900"},
		{Top20, 900, 10, "801-900"},
		{Top10, 5001, 0, "5001+"},
		{Top15, 7000, 0, "3001+"},
		{Top20, 3001, 0, "3001+"},
	}
	for i, test := range tests {
		if col := test.typ.Entries(test.entries); col != test.col {
			t.Errorf("test %d expected %d, got: %d", i, test.col, col)
		}
		if s := test.typ.EntriesTitle(test.entries); s != test.exp {
			t.Errorf("test %d expected %q, got: %q", i, test.exp, s)
		}
	}
}

func TestLevel(t *testing.T) {
	tests := []struct {
		level int
		exp   int
		s     string
	}{
		{7000, -1, ""},
		{0, 0, "1st"},
		{1, 1, "2nd"},
		{2, 2, "3rd"},
		{3, 3, "4th"},
		{4, 4, "5th"},
		{5, 5, "6th"},
		{6, 6, "7th"},
		{7, 7, "8th"},
		{8, 8, "9th"},
		{9, 9, "10th"},
		{10, 10, "11-15"},
		{15, 11, "16-20"},
		{16, 11, "16-20"},
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i, test := range tests {
		for _, typ := range []Type{Top10, Top15, Top20} {
			entries := 150 + int(r.Int31n(1000))
			t.Logf("test %d level: %d entries: %d", i, test.level, entries)
			if level, exp := typ.Level(test.level), test.exp; level != exp {
				t.Errorf("test %d type %n level %d expected %d, got: %d", i, typ, test.level, exp, level)
			}
			if s, exp := typ.LevelTitle(test.level), test.s; s != exp {
				t.Errorf("test %d type %n level %d expected %q, got: %q", i, typ, test.level, exp, s)
			}
		}
	}
}

func TestPayouts(t *testing.T) {
	t.Parallel()
	for _, tt := range []Type{simpleTyp, Top10, Top15, Top20} {
		typ := tt
		for _, gg := range []int64{7, 10, 20} {
			gtd := gg
			t.Run(fmt.Sprintf("%s/%d", typ.Name(), gtd), func(t *testing.T) {
				t.Parallel()
				testPayouts(t, typ, 158, 10000, gtd*100000, 0.15)
			})
		}
	}
}

func testPayouts(t *testing.T, typ Type, entries int, buyin, guaranteed int64, rake float64) {
	t.Helper()
	payouts, total := typ.Payouts(entries, buyin, guaranteed, rake)
	paid, _, _ := typ.Paid(entries)
	if i, exp := len(payouts), paid; i != exp {
		t.Fatalf("expected %d, got: %d", exp, i)
	}
	var sum int64
	for _, amt := range payouts {
		sum += amt
	}
	prize := Prize(entries, buyin, guaranteed, rake)
	switch {
	case !EqualEpsilon(total, prize, .01*float64(prize)):
		t.Errorf("prize %d != total %d", prize, total)
	case sum != total:
		t.Errorf("sum %d != total %d", sum, total)
	case total < guaranteed:
		t.Errorf("total %d < guaranteed %d ", total, guaranteed)
	case !EqualEpsilon(total, sum, .01*float64(total)):
		t.Errorf("total %d != sum %d", total, sum)
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	for _, typ := range []Type{Top10, Top15, Top20} {
		m := map[string]Type{
			"a": typ,
		}
		buf, err := json.Marshal(m)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		p := strconv.Itoa(int(typ.Top() * 100))
		if s, exp := string(buf), fmt.Sprintf(`{"a":"top`+p+`"}`); s != exp {
			t.Errorf("expected:\n%s\ngot:\n%s", exp, s)
		}
		var v map[string]Type
		if err := json.Unmarshal(buf, &v); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if tt, exp := v["a"], typ; tt != exp {
			t.Errorf("expected %d, got: %d", int(exp), int(tt))
		}
	}
}

var simpleTyp = Top20 + 1

func init() {
	simpleTyp := Top20 + 1
	if err := RegisterBytes(simpleTyp, "simpl", 0.10, simpl); err != nil {
		panic(fmt.Sprintf("expected no error, got: %v", err))
	}
}

//go:embed simple.csv
var simpl []byte
