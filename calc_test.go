package cardrank

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestStartingExpValue(t *testing.T) {
	m, count := make(map[string]bool), 0
	for r1 := Ace; r1 != InvalidRank; r1-- {
		for r2 := Ace; r2 != InvalidRank; r2-- {
			a, b := r1, r2
			if a < b {
				a, b = b, a
			}
			c0, c1, c2 := New(a, Heart), New(b, Heart), New(b, Spade)
			if a != b {
				if key := HashKey(c0, c1); !m[key] {
					m[key], count = true, count+1
					expv := StartingExpValue([]Card{c0, c2})
					t.Logf("%3d: %- 3v %v", count, key, expv)
				}
			}
			if key := HashKey(c0, c2); !m[key] {
				m[key], count = true, count+1
				expv := StartingExpValue([]Card{c0, c2})
				t.Logf("%3d: %- 3v %v", count, key, expv)
			}
		}
	}
}

func TestStartingExpValueOmaha(t *testing.T) {
	expv, ok := Omaha.ExpValue(context.Background(), Must("Ah Kh 9s Jd"))
	if !ok {
		t.Fatalf("expected ok")
	}
	t.Logf("%v", expv)
}

func TestOddsCalc(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tests := []struct {
		typ      Type
		pockets  []string
		board    string
		v        []int
		n        int
		inactive []int
	}{
		{
			Holdem,
			[]string{
				"Ah Qc",
				"Qd Qh",
				"Th Tc",
			},
			"7d Kc Td",
			[]int{
				120, 87, 707,
			},
			914,
			nil,
		},
		{
			Holdem,
			[]string{
				"Jc 9c",
				"Ac 4h",
			},
			"7d 4c Qc",
			[]int{
				483, 507,
			},
			990,
			nil,
		},
		{
			Holdem,
			[]string{
				"Jc 9c",
				"Ac 4h",
			},
			"7d 4c Qc Kh",
			[]int{
				17, 27,
			},
			44,
			nil,
		},
		{
			Holdem,
			[]string{
				"Jc 9c",
				"Ac 4h",
			},
			"7d 4c Qc Kh 8h",
			[]int{
				0, 1,
			},
			1,
			nil,
		},
		{
			Holdem,
			[]string{
				"Ad 6c",
				"As 4h",
			},
			"Qh Ah 5s 9h",
			[]int{
				32, 36,
			},
			68,
			nil,
		},
		{
			Holdem,
			[]string{
				"Ah Jc",
				"Qh Jh",
				"Qs 9c",
				"6h 4s",
				"3s 3d",
				"Jd 8h",
				"Kc Td",
				"Js 4h",
			},
			"",
			[]int{
				65642, 47822, 40591, 50138, 73099, 36467, 69067, 9447,
			},
			392273,
			nil,
		},
		{
			Holdem,
			[]string{
				"Jd Jc",
				"6c 6s", // folded
				"Qd 4h", // folded
				"5c 2s", // folded
				"Kd 9d",
				"Qc 9c",
				"Kc Qs",
			},
			"",
			[]int{
				285949, 0, 0, 0, 76137, 44172, 123268,
			},
			529526,
			[]int{1, 2, 3},
		},
		{
			Holdem,
			[]string{
				"Jd Jc",
				"6c 6s", // folded
				"Qd 4h", // folded
				"5c 2s", // folded
				"Kd 9d",
				"Qc 9c",
				"Kc Qs",
			},
			"Kh Qh 9s",
			[]int{
				179, 0, 0, 0, 45, 8, 393,
			},
			625,
			[]int{1, 2, 3},
		},
		{
			Omaha,
			[]string{
				"Ah Qc 7s 7h",
				"Kd 7c 9c 5c",
				"8h Th Ac As",
			},
			"2h 2c 5d",
			[]int{
				64, 96, 506,
			},
			666,
			nil,
		},
		{
			Omaha,
			[]string{
				"Ah Kh Jd 6c",
				"Qh Tc 6s 6h",
			},
			"9h Jh 6d",
			[]int{
				295, 525,
			},
			820,
			nil,
		},
		{
			Omaha,
			[]string{
				"Kh Qh 2c 2h",
				"Ac Jc Kd 4h",
				"Qd Qs Jh Jd",
				"8h 7c Td 3h",
				"6d 5d Th Qc",
			},
			"",
			[]int{
				32924, 45033, 35036, 53714, 37559,
			},
			204266,
			nil,
		},
		{
			Omaha,
			[]string{
				"2h 5h Ts Js",
				"Kh Kc 9s Ac",
			},
			"Jc Qc 2d",
			[]int{
				361, 459,
			},
			820,
			nil,
		},
		{
			Omaha,
			[]string{
				"Jc Th Td Tc",
				"Ah Kh 5d 4h",
				"Ts 8s 6c 5c",
			},
			"4c Qc 6s",
			[]int{
				320, 125, 235,
			},
			680,
			nil,
		},
		{
			Omaha,
			[]string{
				"As Ah Kd Td",
				"Js 8h 8d 3h", // folded
				"Kc Qs Qd 3d", // folded
				"9s 9h 6c 3s", // folded
				"Tc 9d 7s 5h", // folded
				"Ac 4s 4c 2s", // folded
				"7c 5d 4d 3c",
			},
			"7h Jc 2d",
			[]int{
				138, 0, 0, 0, 0, 0, 72,
			},
			210,
			[]int{1, 2, 3, 4, 5},
		},
		{
			OmahaFive,
			[]string{
				"Kh Qh 2c 2h 2c",
				"Ac Jc Kd 4h Ad",
				"Qd Qs Jh Jd 9c",
				"8h 7c Td 3h 6h",
				"6d 5d Th Qc 3d",
			},
			"",
			[]int{
				17987, 22634, 14518, 31961, 17426,
			},
			104526,
			nil,
		},
		{
			OmahaSix,
			[]string{
				"Kh Qh 2c 2h 2c 4d",
				"Ac Jc Kd 4h Ad 9h",
				"Qd Qs Jh Jd 9c 5c",
				"8h 7c Td 3h 6h Ah",
				"6d 5d Th Qc 3d As",
			},
			"",
			[]int{
				7597, 3621, 7073, 13035, 5883,
			},
			37209,
			nil,
		},
	}
	for i, tt := range tests {
		test := tt
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			pockets := make([][]Card, len(test.pockets))
			for i := 0; i < len(test.pockets); i++ {
				pockets[i] = Must(test.pockets[i])
			}
			var active map[int]bool
			if len(test.inactive) != 0 {
				active = make(map[int]bool)
				for i := 0; i < len(test.pockets); i++ {
					active[i] = true
				}
				for i := 0; i < len(test.inactive); i++ {
					active[test.inactive[i]] = false
				}
			}
			testOddsCalc(t, ctx, test.typ, pockets, Must(test.board), test.v, test.n, active)
		})
	}
}

func testOddsCalc(t *testing.T, ctx context.Context, typ Type, pockets [][]Card, board []Card, v []int, n int, active map[int]bool) {
	t.Helper()
	odds, _, ok := NewOddsCalc(
		typ,
		WithPocketsBoard(pockets, board),
		WithDeep(true),
		WithActive(active, true),
	).Calc(ctx)
	switch {
	case !ok:
		t.Fatalf("expected ok == true")
	case odds == nil:
		t.Fatalf("expected non-nil odds")
	case !reflect.DeepEqual(odds.Counts, v):
		t.Errorf("expected %v, got: %v", v, odds.Counts)
	case odds.Total != n:
		t.Errorf("expected %d, got: %d", n, odds.Total)
	}
	t.Logf("board: %v", board)
	total := 0
	for i := 0; i < len(pockets); i++ {
		total += odds.Counts[i]
		t.Logf("%d: %v %*s", i, pockets[i], i, odds)
	}
	if total != n {
		t.Errorf("expected %d, got: %d", n, total)
	}
}

func TestExpValueCalc(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tests := []struct {
		typ    Type
		pocket string
		board  string
		opp    int
		wins   uint64
		splits uint64
		losses uint64
		total  uint64
	}{
		{
			Holdem,
			"Ah As",
			"7d Kc Td",
			1,
			883595, 2207, 184388, 1070190,
		},
		{
			Holdem,
			"Ah As",
			"7d Kc Td Kd",
			1,
			35823, 35, 9682, 45540,
		},
		{
			Holdem,
			"Ah As",
			"7d Kc Td Kd Ks",
			1,
			945, 1, 44, 990,
		},
		{
			Holdem,
			"Ah Kh",
			"7d Kc Td Kd Ks",
			1,
			990, 0, 0, 990,
		},
		{
			Holdem,
			"2s 2h",
			"3h 4h 5s",
			1,
			554489, 48872, 466829, 1070190,
		},
		{
			Holdem,
			"Jc Ah",
			"Jh 9s Ks 3c",
			1,
			34560, 255, 10725, 45540,
		},
		{
			Holdem,
			"Jc Ah",
			"Jh 9s Ks 3c Jd",
			1,
			953, 3, 34, 990,
		},
		{
			Omaha,
			"2s 2h Ad Js",
			"3h 4d 5s",
			1,
			85224285, 4436558, 32515057, 122175900,
		},
		{
			Omaha,
			"2s 2h Ad Js",
			"3h 4d 5s Kh",
			1,
			3916369, 190437, 1323234, 5430040,
		},
		{
			Omaha,
			"2s 2h Ad Js",
			"3h 4d 5s Kh Qd",
			1,
			103699, 3677, 16034, 123410,
		},
		{
			OmahaFive,
			"2s 2h Kd Js Ks",
			"3h 4d 5s Kh",
			1,
			25613788, 0, 10964936, 36578724,
		},
		{
			OmahaFive,
			"2s 2h Kd Js Ks",
			"3h 4d 4s Kh",
			1,
			35839274, 0, 739450, 36578724,
		},
		/*
			{
				Holdem,
				"Qh 7s",
				"",
				1,
				1046780178, 78084287, 972707935, 2097572400,
			},
		*/
		/*
			{
				Omaha,
				"Ah As Kd 9s",
				"",
				1,
				1046780178, 78084287, 972707935, 2097572400,
			},
		*/
	}
	for i, tt := range tests {
		test := tt
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			testExpValueCalc(t, ctx, test.typ, Must(test.pocket), Must(test.board), test.opp, test.wins, test.splits, test.losses, test.total)
		})
	}
}

func testExpValueCalc(t *testing.T, ctx context.Context, typ Type, pocket, board []Card, opponents int, wins, splits, losses, total uint64) {
	t.Helper()
	t.Logf("type: %v pocket: %v board: %v opponents: %d", typ, pocket, board, opponents)
	expv, ok := NewExpValueCalc(
		typ,
		pocket,
		WithBoard(board),
		WithOpponents(opponents),
	).Calc(ctx)
	t.Logf("wins/splits/losses/total: %d/%d/%d/%d", expv.Wins, expv.Splits, expv.Losses, expv.Total)
	switch {
	case !ok:
		t.Fatalf("expected ok == true")
	case expv.Wins != wins:
		t.Errorf("expected wins to be equal: %d != %d", wins, expv.Wins)
	case expv.Splits != splits:
		t.Errorf("expected splits to be equal: %d != %d", splits, expv.Splits)
	case expv.Losses != losses:
		t.Errorf("expected losses to be equal: %d != %d", losses, expv.Losses)
	case expv.Total != total:
		t.Errorf("expected total to be equal: %d != %d", total, expv.Total)
	default:
		t.Logf("expected value: %v (%f)", expv, expv)
	}
}

func TestExpValueCalcStartingPockets(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	if s := os.Getenv("TESTS"); !strings.Contains(s, "starting") {
		t.Skip("skipping: $ENV{TESTS} does not contain 'starting'")
	}
	m, ch, count, wait := make(map[string]bool), make(chan *expValueRes, 1), 0, int64(0)
	for r1 := Ace; r1 != InvalidRank; r1-- {
		for r2 := Ace; r2 != InvalidRank; r2-- {
			a, b := r1, r2
			if a < b {
				a, b = b, a
			}
			c0, c1, c2 := New(a, Heart), New(b, Heart), New(b, Spade)
			if a != b {
				if key := HashKey(c0, c1); !m[key] {
					m[key], count = true, count+1
					atomic.AddInt64(&wait, 1)
					go testExpValue(t, ctx, c0, c1, &wait, ch)
				}
			}
			if key := HashKey(c0, c2); !m[key] {
				m[key], count = true, count+1
				atomic.AddInt64(&wait, 1)
				go testExpValue(t, ctx, c0, c2, &wait, ch)
			}
		}
	}
	go func() {
		start := time.Now()
		t.Logf("started: %d", count)
		tick := time.NewTicker(1 * time.Minute)
		for {
			w := atomic.LoadInt64(&wait)
			if w == 0 {
				close(ch)
				return
			}
			select {
			case <-ctx.Done():
			case <-tick.C:
				t.Logf("wait: %d elapsed: %v", w, time.Now().Sub(start).Round(time.Second))
			case <-time.After(50 * time.Millisecond):
			}
		}
	}()
	var v []*expValueRes
	for res := range ch {
		t.Logf("%v/%v: %v %f", res.c0, res.c1, res, res.Float64())
		v = append(v, res)
	}
	sort.Slice(v, func(i, j int) bool {
		return v[j].Float64() < v[i].Float64()
	})
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%s,%s,%s,%s,%s\n", "name", "wins", "splits", "losses", "calc")
	for _, res := range v {
		fmt.Fprintf(buf, "%s,%d,%d,%d,%f\n", HashKey(res.c0, res.c1), res.Wins, res.Splits, res.Losses, res.Float64())
	}
	if err := os.WriteFile("starting.csv", buf.Bytes(), 0o644); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func testExpValue(t *testing.T, ctx context.Context, c0, c1 Card, wait *int64, ch chan *expValueRes) {
	t.Helper()
	expv, ok := NewExpValueCalc(Holdem, []Card{c0, c1}).Calc(ctx)
	ch <- &expValueRes{
		c0:       c0,
		c1:       c1,
		ExpValue: *expv,
		ok:       ok,
	}
	atomic.AddInt64(wait, -1)
}

type expValueRes struct {
	c0, c1 Card
	ExpValue
	ok bool
}

func TestHashKey(t *testing.T) {
	tests := []struct {
		s   string
		exp string
	}{
		{"As Ah", "AA"},
		{"As Ks", "AKs"},
		{"2h 2s", "22"},
		{"3h 2c", "32o"},
		{"7h Jc", "J7o"},
	}
	for i, test := range tests {
		v := Must(test.s)
		if s := HashKey(v[0], v[1]); s != test.exp {
			t.Errorf("test %d expected %q, got: %q", i, test.exp, s)
		}
		if s := HashKey(v[1], v[0]); s != test.exp {
			t.Errorf("test %d expected %q, got: %q", i, test.exp, s)
		}
	}
}

/*
func TestStarting(t *testing.T) {
	type vv struct {
		key string
		val int
	}
	var res []vv
	for k, v := range starting {
		if len(v) == 1 {
			res = append(res, vv{k, int(v[0])})
		} else {
			res = append(res, vv{k + "s", int(v[0])}, vv{k + "o", int(v[1])})
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].val < res[j].val
	})
	for i, v := range res {
		t.Logf("% 3d %v: %d", i, v.key, v.val)
	}
}
*/
