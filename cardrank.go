// Package cardrank is a library of types, utilities, and interfaces for
// working with playing cards, card decks, evaluating poker hand ranks, and
// managing deals and run outs for different game types.
//
// # Cards and Decks
//
// A [Card] is a [Rank] and [Suit], and can be created with [New], parsed from
// strings with [Parse] or [Must], or created from a unicode playing card rune
// with [FromRune]. A [Deck] is created from a [DeckType], such as the standard
// 52 card [DeckFrench], and is shuffled with any [Shuffler] (such as
// [math/rand.Rand]).
//
// # Evaluation
//
// Each registered [Type] (such as [Holdem] or [OmahaHiLo]) evaluates pocket
// and board cards with [Type.Eval], returning an [Eval] containing the Hi and
// Lo [EvalRank] values and the best and unused cards. Lower ranks are better. Evals
// are compared with [Eval.Comp], and winners are determined with [Order].
//
// Ranks are calculated with pure Go implementations of well-known poker hand
// evaluators: [Cactus], [CactusFast], and the Two-Plus-Two lookup table (see
// [NewTwoPlusTwoEval]).
//
// # Dealing
//
// A [Dealer] deals the streets of a [Type], handles multiple run outs, and
// determines the winners of each run. [OddsCalc] and [ExpValueCalc] calculate
// a run's odds and a pocket's expected value.
//
// # Two-Plus-Two
//
// Importing the [twoplustwo] package registers the Two-Plus-Two evaluator,
// used by the built-in types for faster 7 card evals. Its large lookup table
// is only included in binaries that import it:
//
//	import _ "github.com/cardrank/cardrank/twoplustwo"
//
// [twoplustwo]: https://pkg.go.dev/github.com/cardrank/cardrank/twoplustwo
package cardrank

//go:generate go run ./internal/cmd/cgen

import (
	"cmp"
	"maps"
	"slices"
	"unicode"
)

var (
	// RankCactus is the Cactus Kev rank func used by the built-in types,
	// [CactusFast] by default. Use [SetRankCactus] to change it.
	RankCactus RankFunc = CactusFast

	// twoPlusTwo is the registered Two-Plus-Two eval func (see
	// [RegisterTwoPlusTwo]).
	twoPlusTwo func([]Card) EvalRank

	// descs are the registered type descriptions.
	descs = make(map[Type]TypeDesc)

	// calcs are calc funcs.
	calcs = make(map[Type]EvalFunc)

	// evals are eval funcs.
	evals = make(map[Type]EvalFunc)
)

// SetRankCactus sets the [RankCactus] rank func, and rebuilds the eval funcs of
// all registered types to use it. Must only be called during program
// initialization (such as from an init func), as it is not safe for
// concurrent use.
func SetRankCactus(f RankFunc) {
	RankCactus = f
	rebuild()
}

// RegisterTwoPlusTwo registers the Two-Plus-Two eval func f, and rebuilds the
// eval funcs of all registered types to use it for 7 card evals. Called when
// importing the [twoplustwo] package. Must only be called during program
// initialization (such as from an init func), as it is not safe for
// concurrent use.
//
// See [NewTwoPlusTwoEval] and [NewHybridEval].
//
// [twoplustwo]: https://pkg.go.dev/github.com/cardrank/cardrank/twoplustwo
func RegisterTwoPlusTwo(f func([]Card) EvalRank) {
	twoPlusTwo = f
	rebuild()
}

// registerDefaultTypes registers the default types (see [DefaultTypes]).
func registerDefaultTypes() error {
	for _, desc := range DefaultTypes() {
		if err := RegisterType(desc); err != nil {
			return err
		}
	}
	return nil
}

// RegisterType registers a type.
func RegisterType(desc TypeDesc) error {
	if _, ok := descs[desc.Type]; ok {
		return ErrInvalidId
	}
	// check street ids
	m := make(map[byte]bool)
	for _, street := range desc.Streets {
		if (!unicode.IsLetter(rune(street.Id)) && !unicode.IsNumber(rune(street.Id))) || m[street.Id] {
			return ErrInvalidId
		}
		m[street.Id] = true
	}
	desc.Num = len(descs)
	descs[desc.Type] = desc
	desc.build()
	return nil
}

// rebuild rebuilds the calc and eval funcs of all registered types.
func rebuild() {
	for _, desc := range descs {
		desc.build()
	}
}

// build builds the type's calc and eval funcs.
func (desc TypeDesc) build() {
	calcs[desc.Type] = desc.Eval.New(desc.board, false, desc.Low)
	evals[desc.Type] = desc.Eval.New(desc.board, true, desc.Low)
}

// Types returns registered types.
func Types() []Type {
	return slices.SortedFunc(maps.Keys(descs), func(a, b Type) int {
		return cmp.Compare(descs[a].Num, descs[b].Num)
	})
}

// Error is an error.
type Error string

// Error satisfies the [error] interface.
func (err Error) Error() string {
	return string(err)
}

// Error values.
const (
	// ErrInvalidId is the invalid id error.
	ErrInvalidId Error = "invalid id"
	// ErrMismatchedIdAndType is the mismatched id and type error.
	ErrMismatchedIdAndType Error = "mismatched id and type"
	// ErrInvalidCard is the invalid card error.
	ErrInvalidCard Error = "invalid card"
	// ErrInvalidType is the invalid type error.
	ErrInvalidType Error = "invalid type"
)

// primes are the first 13 prime numbers (one per card rank).
var primes = [...]uint32{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41}
