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
	odds, _, ok := NewCalc(typ, WithCalcPockets(pockets, board), WithCalcDeep(true)).Calc(ctx)
	switch {
	case !ok:
		t.Fatalf("expected ok")
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
		t.Logf("%d: %v %*s - %*o", i, pockets[i], i, odds, i, odds)
	}
	if total != n {
		t.Errorf("expected %d, got: %d", n, total)
	}
}
