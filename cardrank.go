// Package cardrank is a library of types, utilities, and interfaces for
// working with playing cards, card decks, evaluating poker hand ranks, and
// managing deals and run outs for different game types.
//
// [noinit]: https://pkg.go.dev/github.com/cardrank/cardrank#readme-noinit
package cardrank

import (
	"sort"
	"unicode"
)

var (
	// RankCactus is the default Cactus Kev func.
	RankCactus RankFunc

	// Package rank funcs (set in z.go).
	cactus     RankFunc
	cactusFast RankFunc
	twoPlusTwo func([]Card) EvalRank

	// descs are the registered type descriptions.
	descs = make(map[Type]TypeDesc)

	// calcs are calc funcs.
	calcs = make(map[Type]EvalFunc)

	// evals are eval funcs.
	evals = make(map[Type]EvalFunc)
)

// Init inits the package level default variables. Must be manually called
// prior to using the package when built with the [noinit] build tag.
func Init() {
	if RankCactus == nil {
		switch {
		case cactusFast != nil:
			RankCactus = cactusFast
		case cactus != nil:
			RankCactus = cactus
		}
	}
}

// RegisterDefaultTypes registers default types.
//
// See [DefaultTypes].
func RegisterDefaultTypes() error {
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
	}
	desc.Num = len(descs)
	descs[desc.Type] = desc
	calcs[desc.Type] = desc.Eval.New(false, desc.Low)
	evals[desc.Type] = desc.Eval.New(true, desc.Low)
	return nil
}

// Types returns registered types.
func Types() []Type {
	var v []TypeDesc
	for _, desc := range descs {
		v = append(v, desc)
	}
	sort.Slice(v, func(i, j int) bool {
		return v[i].Num < v[j].Num
	})
	types := make([]Type, len(v))
	for i := 0; i < len(types); i++ {
		types[i] = v[i].Type
	}
	return types
}

// Error is a error.
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

// ordered is the ordered constraint.
type ordered interface {
	~float32 | ~float64 | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// min returns the min of a, b.
func min[T ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// max returns the max of a, b.
func max[T ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// insert inserts a at i in v.
func insert[T any](v []T, i int, a ...T) []T {
	return append(v[:i], append(a, v[i:]...)...)
}

// equals returns true when a equals b.
func equals[T comparable](a, b []T) bool {
	n := len(a)
	if n != len(b) {
		return false
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// contains returns true when v contains a.
func contains[T comparable](a T, v ...T) bool {
	for i := 0; i < len(v); i++ {
		if v[i] == a {
			return true
		}
	}
	return false
}

// primes are the first 13 prime numbers (one per card rank).
var primes = [...]uint32{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41}
