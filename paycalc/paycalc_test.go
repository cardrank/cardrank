package paycalc

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

func TestCalc(t *testing.T) {
	tests := []struct {
		f   float64
		e   int
		b   int64
		g   int64
		r   float64
		exp int64
	}{
		{.25, 37, 1000, 20000, .15, 7862},
		{.04, 100, 500, 60000, .15, 2400},
		{.10, 100, 500, 55000, .20, 5500},
		{.10, 101, 500, 55000, .20, 5500},
		{.10, 137, 500, 55000, .20, 5500},
		{.10, 138, 500, 55000, .20, 5520},
	}
	for i, test := range tests {
		if amt := Calc(test.f, test.e, test.b, test.g, test.r); amt != test.exp {
			t.Errorf("test %d expected %d, got: %d", i, test.exp, amt)
		}
	}
}

func TestEntries(t *testing.T) {
	tests := []struct {
		typ Type
		e   int
		col int
		exp string
	}{
		{Top10, 9000, 0, "5001+"},
		{Top10, 256, 19, "251-300"},
		{Top10, 30, 28, "11-30"},
		{Top10, 10, 29, "3-10"},
		{Top15, 1522, 6, "1501-1750"},
		{Top15, 301, 16, "301-350"},
		{Top20, 899, 10, "801-900"},
		{Top20, 900, 10, "801-900"},
	}
	for i, test := range tests {
		if col := test.typ.Entries(test.e); col != test.col {
			t.Errorf("test %d expected %d, got: %d", i, test.col, col)
		}
		if s := test.typ.EntriesTitle(test.e); s != test.exp {
			t.Errorf("test %d expected %q, got: %q", i, test.exp, s)
		}
	}
}

func TestRankings(t *testing.T) {
	tests := []struct {
		typ   Type
		start int
		end   int
		exp   []int
		s     string
	}{
		{Top10, 0, 1, []int{0}, "1st"},
		{Top15, 0, 1, []int{0}, "1st"},
		{Top20, 0, 1, []int{0}, "1st"},
		{Top10, 0, 2, []int{0, 1}, "2nd"},
		{Top15, 0, 2, []int{0, 1}, "2nd"},
		{Top20, 0, 2, []int{0, 1}, "2nd"},
		{Top10, 0, 10, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, "10th"},
		{Top15, 0, 10, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, "10th"},
		{Top20, 0, 10, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, "10th"},
		{Top10, 0, 24, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, "21-25"},
		{Top10, 27, 28, []int{13}, "26-30"},
		{Top20, 200, 220, []int{23, 24}, "201-225"},
		{Top10, 450, 550, []int{30, 31, 32}, "501-550"},
		{Top10, 501, 550, []int{32}, "501-550"},
		{Top10, 551, 552, nil, ""},
		{Top20, 1526, 1623, nil, ""},
	}
	for i, test := range tests {
		rows := test.typ.Levels(test.start, test.end)
		if !reflect.DeepEqual(rows, test.exp) {
			t.Errorf("test %d expected:\n%v\ngot:\n%v", i, test.exp, rows)
		}
		if s := test.typ.LevelsTitle(test.end); s != test.s {
			t.Errorf("test %d expected %q, got: %q", i, test.s, s)
		}
	}
}

func TestPositions(t *testing.T) {
	tests := []struct {
		typ     Type
		start   int
		end     int
		entries int
	}{
		{Top10, 0, 10, 1000},
		{Top15, 23, 37, 500},
		{Top20, 174, 251, 5000},
	}
	for i, test := range tests {
		rows, col := test.typ.Rankings(test.start, test.end), test.typ.Entries(test.entries)
		t.Logf("col: %d %q", col, test.typ.EntriesTitle(test.entries))
		if n, exp := len(rows), test.end-test.start; n != exp {
			t.Fatalf("test %d expected %d, got: %d", i, exp, n)
		}
		for pos := 0; pos < test.end-test.start; pos++ {
			t.Logf("pos: %3d row: %3d %q: %f", pos+test.start+1, rows[pos], test.typ.LevelsTitle(pos+test.start+1), test.typ.At(rows[pos], col))
		}
	}
}

func TestAmount(t *testing.T) {
	t.Parallel()
	for _, tt := range []Type{simpleTyp, Top10, Top15, Top20} {
		typ := tt
		t.Run(tt.Name(), func(t *testing.T) {
			t.Parallel()
			testAmount(t, typ)
		})
	}
}

func testAmount(t *testing.T, typ Type) {
	t.Helper()
	entriesMax, rankingMax := typ.EntriesMax(), typ.RankingMax()
	for entries := 2; entries < entriesMax; entries++ {
		t.Logf("entries %d:", entries)
		sum := 0.0
		for n := 0; n < rankingMax; n++ {
			amt := typ.Amount(n, entries)
			if amt != 0.0 {
				t.Logf("%6d: %f", n, amt)
			}
			sum += amt
		}
		t.Logf("   sum: %f", sum)
		if !EpsilonEqual(1.0, sum, 0.0000000001) {
			t.Errorf("entries %d expected sum to be 1.0, got: %f", entries, sum)
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
	n := typ.RankingMax()
	payouts := typ.Payouts(0, n, entries, buyin, guaranteed, rake)
	if i, exp := len(payouts), n; i != exp {
		t.Fatalf("expected %d, got: %d", exp, i)
	}
	sum := int64(0)
	for i := 0; i < n; i++ {
		sum += payouts[i]
	}
	tb := float64(buyin * int64(entries))
	switch total := max(int64(tb-rake*tb), guaranteed); {
	case sum < guaranteed:
		t.Errorf("sum %d < guaranteed %d ", sum, guaranteed)
	case !EpsilonEqual(total, sum, .01*tb):
		t.Errorf("sum %d != total %d", sum, total)
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
	if err := RegisterBytes(simpleTyp, []byte(simpleTable), 0.1, "simple"); err != nil {
		panic(fmt.Sprintf("expected no error, got: %v", err))
	}
}
