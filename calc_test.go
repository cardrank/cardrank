package cardrank

import (
	"context"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestCalcPocket(t *testing.T) {
	v := Must("Jh 9h")
	f, _ := CalcStart(v)
	t.Logf("%0.2f%%", f*100)
}

func TestCalc(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tests := []struct {
		typ     Type
		pockets []string
		board   string
		v       []int
		n       int
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
			testCalc(t, ctx, test.typ, pockets, Must(test.board), test.v, test.n)
		})
	}
}

func testCalc(t *testing.T, ctx context.Context, typ Type, pockets [][]Card, board []Card, v []int, n int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	odds, _ := NewCalc(typ, WithCalcPockets(pockets, board)).Calc(ctx)
	switch {
	case odds == nil:
		t.Fatalf("expected non-nil odds")
	case !reflect.DeepEqual(odds.V, v):
		t.Errorf("expected %v, got: %v", v, odds.V)
	case odds.N != n:
		t.Errorf("expected %d, got: %d", n, odds.N)
	}
	t.Logf("board: %v", board)
	total := 0
	for i := 0; i < len(pockets); i++ {
		total += odds.V[i]
		t.Logf("%d: %v %*s", i, pockets[i], i, odds)
	}
	if total != n {
		t.Errorf("expected %d, got: %d", n, total)
	}
}
