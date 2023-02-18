// Package github.com/cardrank/cardrank is a library of types, utilities, and
// interfaces for working with playing cards, card decks, and evaluating poker
// hand ranks.
package cardrank

import "unicode"

var (
	// DefaultEval is the default eval rank func.
	DefaultEval CactusFunc
	// DefaultCactus is the default Cactus Kev implementation.
	DefaultCactus RankFunc

	// Package rank funcs (set in z.go).
	cactus     RankFunc
	cactusFast RankFunc
	twoPlusTwo CactusFunc

	// descs are the registered type descriptions.
	descs = make(map[Type]TypeDesc)

	// evals are eval funcs.
	evals = make(map[Type]EvalFunc)
)

// Init inits the package level default variables. Must be manually called
// prior to using this package when built with the `noinit` build tag.
func Init() error {
	switch {
	case twoPlusTwo != nil && cactusFast != nil:
		DefaultEval = NewHybrid(cactusFast, twoPlusTwo)
	case cactusFast != nil:
		DefaultEval = NewRankFunc(cactusFast)
	case cactus != nil:
		DefaultEval = NewRankFunc(cactus)
	}
	switch {
	case cactusFast != nil:
		DefaultCactus = cactusFast
	case cactus != nil:
		DefaultCactus = cactus
	}
	return RegisterDefaultTypes()
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
	evals[desc.Type] = desc.Eval.New(desc.Low)
	return nil
}

// RegisterDefaultTypes registers default types.
func RegisterDefaultTypes() error {
	for _, desc := range DefaultTypes() {
		if err := RegisterType(desc); err != nil {
			return err
		}
	}
	return nil
}

// Error is a error.
type Error string

// Error satisfies the error interface.
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
func contains[T comparable](v []T, a T) bool {
	for i := 0; i < len(v); i++ {
		if v[i] == a {
			return true
		}
	}
	return false
}

// t4c2 is used for taking 4, choosing 2.
var t4c2 = [6][4]uint8{
	{0, 1, 2, 3},
	{0, 2, 1, 3},
	{0, 3, 1, 2},
	{1, 2, 0, 3},
	{1, 3, 0, 2},
	{2, 3, 0, 1},
}

// t5c2 is used for taking 5, choosing 2.
var t5c2 = [10][5]uint8{
	{0, 1, 2, 3, 4},
	{0, 2, 1, 3, 4},
	{0, 3, 1, 2, 4},
	{0, 4, 1, 2, 3},
	{1, 2, 0, 3, 4},
	{1, 3, 0, 2, 4},
	{1, 4, 0, 2, 3},
	{2, 3, 0, 1, 4},
	{2, 4, 0, 1, 3},
	{3, 4, 0, 1, 2},
}

// t5c3 is used for taking 5, choosing 3.
var t5c3 = [10][5]uint8{
	{0, 1, 2, 3, 4},
	{0, 1, 3, 2, 4},
	{0, 1, 4, 2, 3},
	{0, 2, 3, 1, 4},
	{0, 2, 4, 1, 3},
	{0, 3, 4, 1, 2},
	{1, 2, 3, 0, 4},
	{1, 2, 4, 0, 3},
	{1, 3, 4, 0, 2},
	{2, 3, 4, 0, 1},
}

// t6c2 is used for taking 6, choosing 2.
var t6c2 = [15][6]uint8{
	{0, 1, 2, 3, 4, 5},
	{0, 2, 1, 3, 4, 5},
	{0, 3, 1, 2, 4, 5},
	{0, 4, 1, 2, 3, 5},
	{0, 5, 1, 2, 3, 4},
	{1, 2, 0, 3, 4, 5},
	{1, 3, 0, 2, 4, 5},
	{1, 4, 0, 2, 3, 5},
	{1, 5, 0, 2, 3, 4},
	{2, 3, 0, 1, 4, 5},
	{2, 4, 0, 1, 3, 5},
	{2, 5, 0, 1, 3, 4},
	{3, 4, 0, 1, 2, 5},
	{3, 5, 0, 1, 2, 4},
	{4, 5, 0, 1, 2, 3},
}

// t7c5 is used for taking 7, choosing 5.
var t7c5 = [21][7]uint8{
	{0, 1, 2, 3, 4, 5, 6},
	{0, 1, 2, 3, 5, 4, 6},
	{0, 1, 2, 3, 6, 4, 5},
	{0, 1, 2, 4, 5, 3, 6},
	{0, 1, 2, 4, 6, 3, 5},
	{0, 1, 2, 5, 6, 3, 4},
	{0, 1, 3, 4, 5, 2, 6},
	{0, 1, 3, 4, 6, 2, 5},
	{0, 1, 3, 5, 6, 2, 4},
	{0, 1, 4, 5, 6, 2, 3},
	{0, 2, 3, 4, 5, 1, 6},
	{0, 2, 3, 4, 6, 1, 5},
	{0, 2, 3, 5, 6, 1, 4},
	{0, 2, 4, 5, 6, 1, 3},
	{0, 3, 4, 5, 6, 1, 2},
	{1, 2, 3, 4, 5, 0, 6},
	{1, 2, 3, 4, 6, 0, 5},
	{1, 2, 3, 5, 6, 0, 4},
	{1, 2, 4, 5, 6, 0, 3},
	{1, 3, 4, 5, 6, 0, 2},
	{2, 3, 4, 5, 6, 0, 1},
}

// primes are the first 13 prime numbers (one per card rank).
var primes = [...]uint32{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41}
