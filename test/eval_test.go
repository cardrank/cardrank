package test

import (
	"cmp"
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"testing"

	riverboat "github.com/alexclewontin/riverboat/eval"
	"github.com/cardrank/cardrank"
	"github.com/cardrank/cardrank/twoplustwo"
	chehsunliu "github.com/chehsunliu/poker"
	joker "github.com/notnil/joker/pkg/hand"
	paulhankin "github.com/paulhankin/poker/v2/poker"
)

// count is the number of hands generated (must be a power of 2).
const count = 1 << 12

// hands are randomly dealt hands of 11 cards, converted to each evaluator's
// card representation.
//
// 5, 6, and 7 card evals use the first 5, 6, or 7 cards of a hand. Omaha evals
// use the first 4, 5, or 6 cards as the pocket, and the last 5 cards as the
// board.
var hands = deal(count, 0x5eed)

// hand is a dealt hand in each evaluator's representation.
type hand struct {
	cardrank   [11]cardrank.Card
	chehsunliu [11]chehsunliu.Card
	paulhankin [11]paulhankin.Card
	riverboat  [11]riverboat.Card
	joker      [11]joker.Card
}

// deal deals n random hands of 11 cards.
func deal(n int, seed uint64) []hand {
	r, deck := rand.New(rand.NewPCG(seed, seed)), cardrank.DeckFrench.Unshuffled()
	suits := strings.NewReplacer("s", "♠", "h", "♥", "d", "♦", "c", "♣")
	v := make([]hand, n)
	for i := range v {
		r.Shuffle(len(deck), func(j, k int) {
			deck[j], deck[k] = deck[k], deck[j]
		})
		for j, c := range deck[:11] {
			s := c.String()
			v[i].cardrank[j] = c
			v[i].chehsunliu[j] = chehsunliu.NewCard(s)
			v[i].paulhankin[j] = paulhankin.NameToCard[strings.ToUpper(s[1:]+s[:1])]
			v[i].riverboat[j] = riverboat.MustParseCardString(s)
			if err := v[i].joker[j].UnmarshalText([]byte(s[:1] + suits.Replace(s[1:]))); err != nil {
				panic(err)
			}
		}
	}
	return v
}

// evaluator is a poker hand evaluator. Each benchmark func contains its own
// loop so that the only per-hand overhead is the evaluator itself.
type evaluator struct {
	name   string
	bench5 func(*testing.B)
	bench6 func(*testing.B)
	bench7 func(*testing.B)
}

// evaluators are the benchmarked evaluators.
var evaluators = []evaluator{
	{"cardrank/TwoPlusTwo", benchTwoPlusTwo(5), benchTwoPlusTwo(6), benchTwoPlusTwo(7)},
	{"cardrank/CactusFast", benchRankFunc(cardrank.CactusFast), benchEvalFunc(cardrank.CactusFast, 6), benchEvalFunc(cardrank.CactusFast, 7)},
	{"cardrank/Cactus", benchRankFunc(cardrank.Cactus), benchEvalFunc(cardrank.Cactus, 6), benchEvalFunc(cardrank.Cactus, 7)},
	{"cardrank/NewHybridEval", benchHybrid(5), benchHybrid(6), benchHybrid(7)},
	{"cardrank/Holdem.Eval", benchHoldem(5), benchHoldem(6), benchHoldem(7)},
	{"chehsunliu/poker", benchChehsunliu(5), benchChehsunliu(6), benchChehsunliu(7)},
	{"paulhankin/poker", benchPaulhankin5, nil, benchPaulhankin7},
	{"alexclewontin/riverboat", benchRiverboat5, benchRiverboat6, benchRiverboat7},
	{"notnil/joker", benchJoker(5), benchJoker(6), benchJoker(7)},
}

func BenchmarkEval(b *testing.B) {
	for n := 5; n <= 7; n++ {
		for _, e := range evaluators {
			f := [...]func(*testing.B){e.bench5, e.bench6, e.bench7}[n-5]
			if f == nil {
				continue
			}
			b.Run(e.name+"/"+strconv.Itoa(n), func(b *testing.B) {
				b.ReportAllocs()
				f(b)
			})
		}
	}
}

// omahaTypes are the Omaha types for 4, 5, and 6 pocket cards.
var omahaTypes = map[int]cardrank.Type{
	4: cardrank.Omaha,
	5: cardrank.OmahaFive,
	6: cardrank.OmahaSix,
}

// omahaEvaluators are the benchmarked Omaha evaluators. Other than
// Omaha.Eval, each ranks every combination of 2 pocket and 3 board cards with
// the evaluator's 5 card eval, keeping the best.
var omahaEvaluators = []struct {
	name  string
	bench func(int) func(*testing.B)
}{
	{"cardrank/Omaha.Eval", benchOmahaType},
	{"cardrank/CactusFast", benchOmaha(func(v *hand) *[11]cardrank.Card { return &v.cardrank }, rankCardrank)},
	{"chehsunliu/poker", benchOmaha(func(v *hand) *[11]chehsunliu.Card { return &v.chehsunliu }, rankChehsunliu)},
	{"paulhankin/poker", benchOmaha(func(v *hand) *[11]paulhankin.Card { return &v.paulhankin }, rankPaulhankin)},
	{"alexclewontin/riverboat", benchOmaha(func(v *hand) *[11]riverboat.Card { return &v.riverboat }, riverboat.HandValue)},
	{"notnil/joker", benchOmahaJoker},
}

func BenchmarkOmaha(b *testing.B) {
	for k := 4; k <= 6; k++ {
		for _, e := range omahaEvaluators {
			b.Run(e.name+"/"+strconv.Itoa(k), func(b *testing.B) {
				b.ReportAllocs()
				e.bench(k)(b)
			})
		}
	}
}

// sinks prevent the compiler from optimizing away evals.
var (
	sinkCardrank cardrank.EvalRank
	sinkInt      int
)

// twoPlusTwo is the Two-Plus-Two eval func. Importing the twoplustwo package
// also registers it with cardrank.
var twoPlusTwo = twoplustwo.New()

func benchTwoPlusTwo(n int) func(*testing.B) {
	return func(b *testing.B) {
		twoPlusTwo(hands[0].cardrank[:n]) // load lookup table
		for i := 0; b.Loop(); i++ {
			sinkCardrank = twoPlusTwo(hands[i&(count-1)].cardrank[:n])
		}
	}
}

func benchRankFunc(f cardrank.RankFunc) func(*testing.B) {
	return func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			v := &hands[i&(count-1)].cardrank
			sinkCardrank = f(v[0], v[1], v[2], v[3], v[4])
		}
	}
}

func benchEvalFunc(f cardrank.RankFunc, n int) func(*testing.B) {
	return benchCardrankEvalFunc(cardrank.NewEval(f), n)
}

func benchHybrid(n int) func(*testing.B) {
	return benchCardrankEvalFunc(cardrank.NewHybridEval(false, false), n)
}

func benchCardrankEvalFunc(eval cardrank.EvalFunc, n int) func(*testing.B) {
	return func(b *testing.B) {
		ev := cardrank.EvalOf(cardrank.Holdem)
		eval(ev, hands[0].cardrank[:2], hands[0].cardrank[2:n]) // load lookup table
		for i := 0; b.Loop(); i++ {
			v := hands[i&(count-1)].cardrank[:n]
			eval(ev, v[:2], v[2:])
			sinkCardrank = ev.HiRank
		}
	}
}

func benchHoldem(n int) func(*testing.B) {
	return func(b *testing.B) {
		cardrank.Holdem.Eval(hands[0].cardrank[:2], hands[0].cardrank[2:n]) // load lookup table
		for i := 0; b.Loop(); i++ {
			v := hands[i&(count-1)].cardrank[:n]
			sinkCardrank = cardrank.Holdem.Eval(v[:2], v[2:]).HiRank
		}
	}
}

func benchChehsunliu(n int) func(*testing.B) {
	return func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			sinkInt = int(chehsunliu.Evaluate(hands[i&(count-1)].chehsunliu[:n]))
		}
	}
}

func benchPaulhankin5(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		sinkInt = int(paulhankin.Eval5((*[5]paulhankin.Card)(hands[i&(count-1)].paulhankin[:5])))
	}
}

func benchPaulhankin7(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		sinkInt = int(paulhankin.Eval7((*[7]paulhankin.Card)(hands[i&(count-1)].paulhankin[:7])))
	}
}

func benchRiverboat5(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		v := &hands[i&(count-1)].riverboat
		sinkInt = riverboat.HandValue(v[0], v[1], v[2], v[3], v[4])
	}
}

func benchRiverboat6(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		v := &hands[i&(count-1)].riverboat
		_, sinkInt = riverboat.BestFiveOfSix(v[0], v[1], v[2], v[3], v[4], v[5])
	}
}

func benchRiverboat7(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		v := &hands[i&(count-1)].riverboat
		_, sinkInt = riverboat.BestFiveOfSeven(v[0], v[1], v[2], v[3], v[4], v[5], v[6])
	}
}

func benchJoker(n int) func(*testing.B) {
	return func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			sinkInt = int(joker.New(hands[i&(count-1)].joker[:n]).Ranking())
		}
	}
}

func benchOmahaType(k int) func(*testing.B) {
	typ := omahaTypes[k]
	return func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			v := &hands[i&(count-1)].cardrank
			sinkCardrank = typ.Eval(v[:k], v[6:]).HiRank
		}
	}
}

func benchOmaha[C any](cards func(*hand) *[11]C, f func(c0, c1, c2, c3, c4 C) int) func(int) func(*testing.B) {
	return func(k int) func(*testing.B) {
		return func(b *testing.B) {
			for i := 0; b.Loop(); i++ {
				sinkInt = omaha(cards(&hands[i&(count-1)]), k, f)
			}
		}
	}
}

func benchOmahaJoker(k int) func(*testing.B) {
	return func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			sinkInt = int(omahaJoker(&hands[i&(count-1)].joker, k).Ranking())
		}
	}
}

// pocketPairs are the pocket card index pairs for 4, 5, and 6 pocket cards.
var pocketPairs = [7][][2]int{
	4: combin2(4),
	5: combin2(5),
	6: combin2(6),
}

// boardTriples are the board card index triples for 5 board cards.
var boardTriples = combin3(5)

func combin2(n int) [][2]int {
	var v [][2]int
	for i := range n {
		for j := i + 1; j < n; j++ {
			v = append(v, [2]int{i, j})
		}
	}
	return v
}

func combin3(n int) [][3]int {
	var v [][3]int
	for i := range n {
		for j := i + 1; j < n; j++ {
			for k := j + 1; k < n; k++ {
				v = append(v, [3]int{i, j, k})
			}
		}
	}
	return v
}

// omaha returns the best (lowest) rank of every combination of 2 of the first
// k cards and 3 of the last 5 cards in v, ranked with f.
func omaha[C any](v *[11]C, k int, f func(c0, c1, c2, c3, c4 C) int) int {
	best := math.MaxInt
	for _, p := range pocketPairs[k] {
		for _, t := range boardTriples {
			best = min(best, f(v[p[0]], v[p[1]], v[6+t[0]], v[6+t[1]], v[6+t[2]]))
		}
	}
	return best
}

// omahaJoker returns the best hand of every combination of 2 of the first k
// cards and 3 of the last 5 cards in v.
func omahaJoker(v *[11]joker.Card, k int) *joker.Hand {
	var best *joker.Hand
	for _, p := range pocketPairs[k] {
		for _, t := range boardTriples {
			h := joker.New([]joker.Card{v[p[0]], v[p[1]], v[6+t[0]], v[6+t[1]], v[6+t[2]]})
			if best == nil || h.CompareTo(best) > 0 {
				best = h
			}
		}
	}
	return best
}

// 5 card rank funcs returning an int, where lower is better.

func rankCardrank(c0, c1, c2, c3, c4 cardrank.Card) int {
	return int(cardrank.CactusFast(c0, c1, c2, c3, c4))
}

func rankChehsunliu(c0, c1, c2, c3, c4 chehsunliu.Card) int {
	v := [5]chehsunliu.Card{c0, c1, c2, c3, c4}
	return int(chehsunliu.Evaluate(v[:]))
}

func rankPaulhankin(c0, c1, c2, c3, c4 paulhankin.Card) int {
	v := [5]paulhankin.Card{c0, c1, c2, c3, c4}
	return -int(paulhankin.Eval5(&v))
}

// TestAgreement verifies every evaluator ranks the dealt hands identically to
// cardrank.
func TestAgreement(t *testing.T) {
	for n := 5; n <= 7; n++ {
		evals := make([]cardrank.EvalRank, count)
		for i := range hands {
			v := hands[i].cardrank[:n]
			evals[i] = twoPlusTwo(v)
			if r := cardrank.Holdem.Eval(v[:2], v[2:]).HiRank; r != evals[i] {
				t.Fatalf("%v: Holdem.Eval %d != TwoPlusTwo %d", v, r, evals[i])
			}
			ev := cardrank.EvalOf(cardrank.Holdem)
			if cardrank.NewHybridEval(false, false)(ev, v[:2], v[2:]); ev.HiRank != evals[i] {
				t.Fatalf("%v: NewHybridEval %d != TwoPlusTwo %d", v, ev.HiRank, evals[i])
			}
			if r := chehsunliu.Evaluate(hands[i].chehsunliu[:n]); int(r) != int(evals[i]) {
				t.Errorf("%v: chehsunliu %d != cardrank %d", v, r, evals[i])
			}
			if r := riverboatEval(&hands[i].riverboat, n); r != int(evals[i]) {
				t.Errorf("%v: riverboat %d != cardrank %d", v, r, evals[i])
			}
		}
		// paulhankin (higher is better) and joker use different scales, so
		// compare each adjacent pair of hands ordered by cardrank
		for i, j := range adjacent(evals) {
			exp := cmp.Compare(evals[j], evals[i]) // > 0 when hand i is better
			if n != 6 {
				if c := cmp.Compare(paulhankinEval(&hands[i].paulhankin, n), paulhankinEval(&hands[j].paulhankin, n)); c != exp {
					t.Errorf("%v vs %v: paulhankin compare %d, expected %d", hands[i].cardrank[:n], hands[j].cardrank[:n], c, exp)
				}
			}
			if c := cmp.Compare(joker.New(hands[i].joker[:n]).CompareTo(joker.New(hands[j].joker[:n])), 0); c != exp {
				t.Errorf("%v vs %v: joker compare %d, expected %d", hands[i].cardrank[:n], hands[j].cardrank[:n], c, exp)
			}
		}
	}
}

// TestOmahaAgreement verifies every evaluator ranks the dealt Omaha hands
// identically to cardrank.
func TestOmahaAgreement(t *testing.T) {
	for k, typ := range omahaTypes {
		evals := make([]cardrank.EvalRank, count)
		for i := range hands {
			v := &hands[i].cardrank
			evals[i] = typ.Eval(v[:k], v[6:]).HiRank
			if r := omaha(v, k, rankCardrank); r != int(evals[i]) {
				t.Fatalf("%s %v %v: CactusFast %d != %d", typ, v[:k], v[6:], r, evals[i])
			}
			if r := omaha(&hands[i].chehsunliu, k, rankChehsunliu); r != int(evals[i]) {
				t.Errorf("%s %v %v: chehsunliu %d != %d", typ, v[:k], v[6:], r, evals[i])
			}
			if r := omaha(&hands[i].riverboat, k, riverboat.HandValue); r != int(evals[i]) {
				t.Errorf("%s %v %v: riverboat %d != %d", typ, v[:k], v[6:], r, evals[i])
			}
		}
		for i, j := range adjacent(evals) {
			exp := cmp.Compare(evals[j], evals[i]) // > 0 when hand i is better
			ph := cmp.Compare(omaha(&hands[j].paulhankin, k, rankPaulhankin), omaha(&hands[i].paulhankin, k, rankPaulhankin))
			if ph != exp {
				t.Errorf("%s %v vs %v: paulhankin compare %d, expected %d", typ, hands[i].cardrank, hands[j].cardrank, ph, exp)
			}
			if c := cmp.Compare(omahaJoker(&hands[i].joker, k).CompareTo(omahaJoker(&hands[j].joker, k)), 0); c != exp {
				t.Errorf("%s %v vs %v: joker compare %d, expected %d", typ, hands[i].cardrank, hands[j].cardrank, c, exp)
			}
		}
	}
}

// adjacent returns an iterator over the indexes of each adjacent pair of
// hands, ordered by evals.
func adjacent(evals []cardrank.EvalRank) func(func(int, int) bool) {
	order := make([]int, len(evals))
	for i := range order {
		order[i] = i
	}
	slices.SortFunc(order, func(a, b int) int {
		return cmp.Compare(evals[a], evals[b])
	})
	return func(yield func(int, int) bool) {
		for k := 1; k < len(order); k++ {
			if !yield(order[k-1], order[k]) {
				return
			}
		}
	}
}

func riverboatEval(v *[11]riverboat.Card, n int) int {
	switch n {
	case 5:
		return riverboat.HandValue(v[0], v[1], v[2], v[3], v[4])
	case 6:
		_, r := riverboat.BestFiveOfSix(v[0], v[1], v[2], v[3], v[4], v[5])
		return r
	}
	_, r := riverboat.BestFiveOfSeven(v[0], v[1], v[2], v[3], v[4], v[5], v[6])
	return r
}

func paulhankinEval(v *[11]paulhankin.Card, n int) int16 {
	if n == 5 {
		return paulhankin.Eval5((*[5]paulhankin.Card)(v[:5]))
	}
	return paulhankin.Eval7((*[7]paulhankin.Card)(v[:7]))
}
