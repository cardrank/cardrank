package cardrank

import (
	"math/rand"
	"slices"
	"testing"

	"github.com/cardrank/cardrank/internal/tpt"
)

// testTwoPlusTwo is a Two-Plus-Two eval func for tests, equivalent to the one
// registered by the twoplustwo package (which can't be imported here).
var testTwoPlusTwo = NewTwoPlusTwoEval(tpt.Table)

// TestRegisterTwoPlusTwo verifies registering the Two-Plus-Two eval func does
// not change any type's evals.
func TestRegisterTwoPlusTwo(t *testing.T) {
	type hand struct {
		typ           Type
		pocket, board []Card
	}
	var hands []hand
	r := rand.New(rand.NewSource(0x2b2))
	for _, typ := range Types() {
		boards := []int{typ.Board()}
		if 5 <= typ.Board() {
			boards = append(boards, 0, 3, 4)
		}
		for _, nb := range boards {
			for range 200 {
				d := typ.DeckType().Shuffle(r, 1)
				hands = append(hands, hand{typ, d.Draw(typ.Pocket()), d.Draw(nb)})
			}
		}
	}
	exp := make([]*Eval, len(hands))
	for i, h := range hands {
		exp[i] = h.typ.Eval(h.pocket, h.board)
	}
	var calls int
	RegisterTwoPlusTwo(func(v []Card) EvalRank {
		calls++
		return testTwoPlusTwo(v)
	})
	defer RegisterTwoPlusTwo(nil)
	for i, h := range hands {
		ev, e := h.typ.Eval(h.pocket, h.board), exp[i]
		switch {
		case ev.HiRank != e.HiRank,
			!slices.Equal(ev.HiBest, e.HiBest),
			!slices.Equal(ev.HiUnused, e.HiUnused),
			ev.LoRank != e.LoRank,
			!slices.Equal(ev.LoBest, e.LoBest),
			!slices.Equal(ev.LoUnused, e.LoUnused):
			t.Errorf("%s %v %v: expected %v %v %v / %v %v %v, got: %v %v %v / %v %v %v",
				h.typ, h.pocket, h.board,
				e.HiRank, e.HiBest, e.HiUnused, e.LoRank, e.LoBest, e.LoUnused,
				ev.HiRank, ev.HiBest, ev.HiUnused, ev.LoRank, ev.LoBest, ev.LoUnused,
			)
		}
	}
	if calls == 0 {
		t.Errorf("expected registered Two-Plus-Two eval func to be used")
	}
}
