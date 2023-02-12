package cardrank

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Hand contains hand eval info.
type Hand struct {
	Type     Type
	Pocket   []Card
	Board    []Card
	HiRank   HandRank
	HiBest   []Card
	HiUnused []Card
	LoRank   HandRank
	LoBest   []Card
	LoUnused []Card
}

// NewUnevaluatedHand creates an unevaluated hand for the type, pocket, and
// board.
func NewUnevaluatedHand(typ Type, pocket, board []Card) *Hand {
	p, b := make([]Card, len(pocket)), make([]Card, len(board))
	copy(p, pocket)
	copy(b, board)
	h := &Hand{
		Type:   typ,
		Pocket: p,
		Board:  b,
		HiRank: Invalid,
		LoRank: Invalid,
	}
	return h
}

// NewHand creates a hand for the type, pocket, and board and evaluates the
// hand's rank.
func NewHand(typ Type, pocket, board []Card) *Hand {
	h := NewUnevaluatedHand(typ, pocket, board)
	typ.Eval(h)
	return h
}

// Init inits best, unused.
func (h *Hand) Init(n, m int, loMax HandRank) {
	if 0 < n {
		h.HiBest = make([]Card, n)
	}
	if 0 < m {
		h.HiUnused = make([]Card, m)
	}
	if loMax != Invalid {
		if 0 < n {
			h.LoBest = make([]Card, n)
		}
		if 0 < m {
			h.LoUnused = make([]Card, m)
		}
	}
}

// Hand returns the combined pocket, board.
func (h *Hand) Hand() []Card {
	hand := make([]Card, len(h.Pocket)+len(h.Board))
	copy(hand, h.Pocket)
	copy(hand[len(h.Pocket):], h.Board)
	return hand
}

// Fixed returns the hand's fixed rank.
func (h *Hand) Fixed() HandRank {
	return h.HiRank.Fixed()
}

// LowValid returns true if is a valid low hand.
func (h *Hand) LowValid() bool {
	return h.LoRank != Invalid
}

// Format satisfies the fmt.Formatter interface.
func (h *Hand) Format(f fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		fmt.Fprintf(f, "%s %s", h.Description(), h.HiBest)
	case 'q':
		fmt.Fprintf(f, "\"%s %s\"", h.Description(), h.HiBest)
	case 'S':
		fmt.Fprintf(f, "%s %S", h.Description(), CardFormatter(h.HiBest))
	case 'b':
		fmt.Fprintf(f, "%s %b", h.Description(), h.HiBest)
	case 'h':
		fmt.Fprintf(f, "%s %h", h.Description(), CardFormatter(h.HiBest))
	case 'c':
		fmt.Fprintf(f, "%s %c", h.Description(), h.HiBest)
	case 'C':
		fmt.Fprintf(f, "%s %C", h.Description(), CardFormatter(h.HiBest))
	case 'f':
		for _, c := range h.HiBest {
			c.Format(f, 's')
		}
	default:
		fmt.Fprintf(f, "%%!%c(ERROR=unknown verb, Hand<%s>: %s/%s %d)", verb, h.Type, h.Pocket, h.Board, h.HiRank)
	}
}

// Description describes the hand's best-five cards.
//
// Examples:
//
//	Straight Flush, Ace-high, Royal
//	Straight Flush, King-high
//	Straight Flush, Five-high, Steel Wheel
//	Four of a Kind, Nines, kicker Jack
//	Full House, Sixes full of Fours
//	Flush, Ten-high
//	Straight, Eight-high
//	Three of a Kind, Fours, kickers Ace, King
//	Two Pair, Nines over Sixes, kicker Jack
//	Pair, Aces, kickers King, Queen, Nine
//	Nothing, Seven-high, kickers Six, Five, Three, Two
func (h *Hand) Description() string {
	r := h.HiRank
	switch {
	case h.Type == Badugi,
		h.Type == Razz && h.HiRank < rankLowMax:
		s := make([]string, len(h.HiBest))
		for i := 0; i < len(h.HiBest); i++ {
			s[i] = h.HiBest[i].Rank().Name()
		}
		return strings.Join(s, ", ") + "-low"
	case h.Type == Razz:
		r = Invalid - r
	case h.Type == Lowball,
		h.Type == LowballTriple:
		if r = rankMax - r; r > Pair {
			s := make([]string, len(h.HiBest))
			for i := 0; i < len(h.HiBest); i++ {
				s[i] = h.HiBest[i].Rank().Name()
			}
			str := strings.Join(s, ", ") + "-low"
			if r == Nothing {
				str += ", Wheel"
			}
			return str
		}
	}
	switch r.Fixed() {
	case StraightFlush:
		switch r := h.HiBest[0].Rank(); {
		case r == Ace:
			return fmt.Sprintf("Straight Flush, %N-high, Royal", h.HiBest[0])
		case r == Nine && h.Type == Short:
			return fmt.Sprintf("Straight Flush, %N-high, Iron Maiden", h.HiBest[0])
		case r == Five:
			return fmt.Sprintf("Straight Flush, %N-high, Steel Wheel", h.HiBest[0])
		}
		return fmt.Sprintf("Straight Flush, %N-high", h.HiBest[0])
	case FourOfAKind:
		return fmt.Sprintf("Four of a Kind, %P, kicker %N", h.HiBest[0], h.HiBest[4])
	case FullHouse:
		return fmt.Sprintf("Full House, %P full of %P", h.HiBest[0], h.HiBest[3])
	case Flush:
		return fmt.Sprintf("Flush, %N-high", h.HiBest[0])
	case Straight:
		return fmt.Sprintf("Straight, %N-high", h.HiBest[0])
	case ThreeOfAKind:
		return fmt.Sprintf("Three of a Kind, %P, kickers %N, %N", h.HiBest[0], h.HiBest[3], h.HiBest[4])
	case TwoPair:
		return fmt.Sprintf("Two Pair, %P over %P, kicker %N", h.HiBest[0], h.HiBest[2], h.HiBest[4])
	case Pair:
		return fmt.Sprintf("Pair, %P, kickers %N, %N, %N", h.HiBest[0], h.HiBest[2], h.HiBest[3], h.HiBest[4])
	}
	return fmt.Sprintf("Nothing, %N-high, kickers %N, %N, %N, %N", h.HiBest[0], h.HiBest[1], h.HiBest[2], h.HiBest[3], h.HiBest[4])
}

// LowDescription describes the hands best-five low cards.
func (h *Hand) LowDescription() string {
	if h.LoRank == Invalid {
		return "None"
	}
	s := make([]string, len(h.LoBest))
	for i := 0; i < len(h.LoBest); i++ {
		s[i] = h.LoBest[i].Rank().Name()
	}
	return strings.Join(s, ", ") + "-low"
}

// HiComp compares the hi hand rank.
func (h *Hand) HiComp(b *Hand) int {
	return h.Type.HiComp()(h, b)
}

// LoComp compares the lo hand rank.
func (h *Hand) LoComp(b *Hand) int {
	return h.Type.LoComp()(h, b)
}

// HiOrder orders hands by HiRank, low to high, returning 'pivot' of winning vs
// losing hands. Pivot will always be 1 or higher.
func HiOrder(hands []*Hand) ([]int, int) {
	if len(hands) == 0 {
		return nil, 0
	}
	i, n := 0, len(hands)
	m, h := make(map[int]*Hand, n), make([]int, n)
	for ; i < n; i++ {
		m[i], h[i] = hands[i], i
	}
	f := hands[0].Type.HiComp()
	sort.SliceStable(h, func(j, k int) bool {
		return f(m[h[j]], m[h[k]]) < 0
	})
	for i = 1; i < n && m[h[i-1]].HiRank == m[h[i]].HiRank; i++ {
	}
	return h, i
}

// LoOrder orders hands by LoRank, low to high, returning 'pivot' of winning vs
// losing hands. If there are no low hands the pivot will be 0.
func LoOrder(hands []*Hand) ([]int, int) {
	if len(hands) == 0 {
		return nil, 0
	}
	i, n := 0, len(hands)
	m, h := make(map[int]*Hand, n), make([]int, n)
	for ; i < n; i++ {
		m[i], h[i] = hands[i], i
	}
	f := hands[0].Type.LoComp()
	sort.SliceStable(h, func(j, k int) bool {
		return f(m[h[j]], m[h[k]]) < 0
	})
	if m[h[0]].LoRank == Invalid {
		return nil, 0
	}
	for i = 1; i < n && m[h[i-1]].LoRank == m[h[i]].LoRank; i++ {
	}
	return h, i
}

// Win describes a win.
type Win struct {
	Hi      []int
	HiPivot int
	Lo      []int
	LoPivot int
	Low     bool
}

// NewWin creates a new win.
func NewWin(h1, h2 []*Hand, low bool) Win {
	h, hp := HiOrder(h1)
	var l []int
	var lp int
	switch {
	case low:
		l, lp = LoOrder(h1)
	case h2 != nil:
		l, lp = HiOrder(h2)
	}
	return Win{
		Hi:      h,
		HiPivot: hp,
		Lo:      l,
		LoPivot: lp,
		Low:     low,
	}
}

// Format satisfies the fmt.Formatter interface.
func (win Win) Format(f fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		_, _ = f.Write([]byte(win.HiDesc(func(_, i int) string {
			return strconv.Itoa(i + 1)
		})))
	}
}

// HiDesc returns a description.
func (win Win) HiDesc(f func(int, int) string) string {
	return win.HiJoin(f, ", ") + " " + win.HiVerb()
}

// LoDescribe returns a low description.
func (win Win) LoDesc(f func(int, int) string) string {
	return win.LoJoin(f, ", ") + " " + win.LoVerb()
}

// Scoop returns true when a pot is scooped.
func (win Win) Scoop() bool {
	switch {
	case win.Low && win.HiPivot == 1 && win.LoPivot == 0:
		return true
	case win.HiPivot == 1 && win.LoPivot == 1:
		return win.Hi[0] == win.Lo[0]
	}
	return false
}

// HiVerb returns the win verb.
func (win Win) HiVerb() string {
	return WinVerb(win.HiPivot, win.Scoop())
}

// LoVerb returns the win verb.
func (win Win) LoVerb() string {
	return WinVerb(win.LoPivot, win.Scoop())
}

// HiJoin joins strings.
func (win Win) HiJoin(f func(int, int) string, sep string) string {
	var v []string
	for i := 0; i < win.HiPivot; i++ {
		v = append(v, f(i, win.Hi[i]))
	}
	return strings.Join(v, sep)
}

// LoJoin joins strings.
func (win Win) LoJoin(f func(int, int) string, sep string) string {
	var v []string
	for i := 0; i < win.LoPivot; i++ {
		v = append(v, f(i, win.Lo[i]))
	}
	return strings.Join(v, sep)
}

// WinVerb returns the win verb.
func WinVerb(n int, scoop bool) string {
	switch {
	case scoop:
		return "scoops"
	case n > 2:
		return "push"
	case n == 2:
		return "split"
	}
	return "wins"
}
