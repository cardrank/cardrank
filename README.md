# About

Package `cardrank` is a library of types, utilities, and interfaces for working
with playing cards, card decks, evaluating poker hand ranks, and managing deals
and run outs for different game types.

[![Tests](https://github.com/cardrank/cardrank/workflows/Test/badge.svg)](https://github.com/cardrank/cardrank/actions?query=workflow%3ATest)
[![Go Report Card](https://goreportcard.com/badge/github.com/cardrank/cardrank)](https://goreportcard.com/report/github.com/cardrank/cardrank)
[![Reference](https://pkg.go.dev/badge/github.com/cardrank/cardrank.svg)][pkg]
[![Releases](https://img.shields.io/github/v/release/cardrank/cardrank?display_name=tag&sort=semver)](https://github.com/cardrank/cardrank/releases)

## Overview

The `cardrank` package contains types for working with [cards][card],
[suits][suit], [ranks][rank], and [decks][deck], [evaluating poker
ranks][eval], and [managing deals and run outs][dealer].

In most cases, the high-level [`Dealer`][dealer] with any [registered
`Type`][type] is all that is needed. An in-depth example is provided [in the
package documentation][pkg-example].

A [`Type`][type] wraps a [type description][type-desc] defining the type's
[streets][street-desc], [deck][deck-type], [eval][eval-type], [Hi/Lo
description][desc-type], and other metadata needed for [dealing streets and
managing run outs][dealer].

[Evaluation and ranking][eval-ranking] of [types][type] is done with pure Go
implementations of [well-known poker hand evaluators](#cactus-kev), which are
among the [fastest available in Go](#benchmarks). [Evals][eval] can be compared
and [ordered to determine the winners][winners] of a hand.

The [`paycalc`][paycalc] package contains tournament payout tables, and
calculates tournament payouts.

## Supported Types

Supports [evaluating and ranking][eval] the following [types][type]:

| Holdem Variants    | Omaha Variants           | Hybrid Variants        | Draw Variants      | Other                   |
| ------------------ | ------------------------ | ---------------------- | ------------------ | ----------------------- |
| [`Holdem`][type]   | [`Omaha`][type]          | [`Dallas`][type]       | [`Video`][type]    | [`Soko`][type]          |
| [`Split`][type]    | [`OmahaHiLo`][type]      | [`DallasDouble`][type] | [`Draw`][type]     | [`SokoHiLo`][type]      |
| [`Short`][type]    | [`OmahaDouble`][type]    | [`Houston`][type]      | [`DrawHiLo`][type] | [`Lowball`][type]       |
| [`Manila`][type]   | [`OmahaFive`][type]      | [`Fusion`][type]       | [`Stud`][type]     | [`LowballTriple`][type] |
| [`Spanish`][type]  | [`OmahaSix`][type]       | [`FusionHiLo`][type]   | [`StudHiLo`][type] | [`Razz`][type]          |
| [`Royal`][type]    | [`Jakarta`][type]        |                        | [`StudFive`][type] | [`Badugi`][type]        |
| [`Double`][type]   | [`Courchevel`][type]     |                        |                    |                         |
| [`Showtime`][type] | [`CourchevelHiLo`][type] |                        |                    |                         |
| [`Swap`][type]     |                          |                        |                    |                         |
| [`River`][type]    |                          |                        |                    |                         |

See the package's [`Type`][type] documentation for a description of each.

## Using

Requires Go 1.27 or later. To use within a Go package:

```sh
go get github.com/cardrank/cardrank
```

See the [Go package documentation][pkg] for in-depth overviews of the APIs.

### Quickstart

Examples of the types, utilities, and interfaces are available in the [Go
package documentation][pkg].

Additional examples for a [`Dealer`][dealer] and the [`Holdem`][type] and
[`OmahaHiLo`][type] types are included in the [example directory](/_example):

- [dealer](/_example/dealer) - uses the [`Dealer`][dealer] to deal cards, handle multiple run outs, and determine winners for any [`Type`][type]
- [holdem](/_example/holdem) - uses types and utilities to deal [`Holdem`][type]
- [omahahilo](/_example/omahahilo) - uses types and utilities to deal [`OmahaHiLo`][type], splitting Hi and Lo wins

### Eval Ranking

An [`EvalRank`][eval-rank] is determined using the [`EvalFunc`][eval-func]
registered for a [`Type`][type]. `EvalRank` values are ordered low to high, and
a lower rank is a better hand:

```go
fmt.Printf("%t\n", cardrank.StraightFlush < cardrank.FullHouse)

// Output:
// true
```

Pocket and board [cards][card] can be passed to a type's `Eval` method, which
uses the type's registered [`EvalFunc`][eval-func] and returns an
[`Eval`][eval]:

```go
pocket, board := cardrank.Must("Ah Kh"), cardrank.Must("Qh Jh Th 2s 3s")
ev := cardrank.Holdem.Eval(pocket, board)
fmt.Printf("%s - %d\n", ev, ev.HiRank)

// Output:
// Straight Flush, Ace-high, Royal [Ah Kh Qh Jh Th] - 1
```

Usually 5, 6, or 7 cards are evaluated, but some types can evaluate fewer
cards:

```go
pocket := cardrank.Must("2h 3s 4c")
ev := cardrank.Badugi.Eval(pocket, nil)
fmt.Printf("%s\n", ev)

// Output:
// Four, Three, Two-low [4c 3s 2h]
```

If an invalid number of cards is passed to a type's `EvalFunc`, the `Eval`'s
[`HiRank`][eval.hi-rank] and [`LoRank`][eval.lo-rank] are set to
[`Invalid`][invalid].

Hands are compared by their [`Eval.HiRank`][eval.hi-rank] or
[`Eval.LoRank`][eval.lo-rank] values to determine a winner.

#### Hi/Lo

Some types have both a Hi and Lo [`EvalRank`][eval-rank], such as
[`Double`][type] board [`Holdem`][type] and the `*HiLo` variants like
[`OmahaHiLo`][type]. An [`Eval`][eval] contains both values as its
[`HiRank`][eval.hi-rank] and [`LoRank`][eval.lo-rank].

#### Cactus Kev

For most types, the [`EvalRank`][eval-rank] is determined by Go implementations
of a few well-known [Cactus Kev][cactus-kev] algorithms:

- [`Cactus`][cactus] - the original [Cactus Kev][cactus-kev] poker hand evaluator
- [`CactusFast`][cactus-fast] - the [Fast Cactus][senzee] poker hand evaluator, using Paul Senzee's perfect hash lookup
- [`TwoPlusTwo`][two-plus-two] - the [2+2 forum][tangentforks] poker hand evaluator, using a 124 MiB lookup table (see [below](#two-plus-two))

#### Rank Cactus Func

The package-level [`RankCactus`][rank-cactus] variable is the rank func used for
standard poker evaluation, and defaults to [`CactusFast`][cactus-fast].

Use [`SetRankCactus`][set-rank-cactus] to change the rank func used by the
built-in types, such as when using a third-party implementation, or
experimenting with new Cactus implementations. Like registering types, it
should only be called during initialization:

```go
func init() {
	cardrank.SetRankCactus(cardrank.Cactus)
}
```

#### Two-Plus-Two

The [`twoplustwo`][twoplustwo] package provides the Two-Plus-Two evaluator,
which uses a large (124 MiB) lookup table for extremely fast 5, 6, and 7 card
evaluation. Importing the package registers the evaluator, and the built-in
types then use it for 7 card evals:

```go
import (
	"github.com/cardrank/cardrank"
	_ "github.com/cardrank/cardrank/twoplustwo"
)
```

The lookup table is embedded in the `twoplustwo` package, so it is only
included in binaries that import it: a small program is about 3 MB without the
import, and about 134 MB with it. The table is decoded into memory the first
time it is used (about 140 ms), so the import adds no startup cost.

Without the import, 7 card evals use [`RankCactus`][rank-cactus] to rank every
5 card combination, which is slower (see [benchmarks](#benchmarks)) but gives
identical results. Leaving out the import is useful for client applications
and constrained environments, such as WASM.

[`twoplustwo.New`][twoplustwo-new] returns the Two-Plus-Two eval func for
direct use.

### Winner Determination

Winners are determined by the [lowest `EvalRank`][eval-rank] of either the Hi
or Lo value for the [`Type`][type]. Hands with equal `EvalRank` values are
equivalent, and split the win.

#### Comparing Eval Ranks

An [`Eval`][eval] can be compared to another `Eval` using
[`Comp`][eval.comp]. `Comp` returns `-1`, `0`, or `+1`, making it easy to
compare or sort hands:

```go
// Compare a and b's Hi:
if a.Comp(b, false) < 0 {
	fmt.Printf("%s is a winner!", a)
}

// Compare a and b's Lo:
if a.Comp(b, true) == 0 {
	fmt.Printf("%s and %s are equal!", a, b)
}

// Sort []*Eval by Hi:
slices.SortFunc(evs, func(a, b *cardrank.Eval) int {
	return a.Comp(b, false)
})

// Sort []*Eval by Lo:
slices.SortFunc(evs, func(a, b *cardrank.Eval) int {
	return a.Comp(b, true)
})
```

The package-level [`Order`][order] func orders evals and determines winners.
See [ordering evals](#ordering-evals) below.

#### Ordering Evals

[`Order`][order] determines the winners of a hand by ordering the indexes of a
[`[]*Eval`][eval]. It returns the ordered indexes as a `[]int`, and an `int`
pivot marking the cutoff for a win:

```go
// Order by HiRank:
hiOrder, hiPivot := cardrank.Order(evs, false)
```

For a [`Type`][type] with a Lo value:

```go
// Order by LoRank:
loOrder, loPivot := cardrank.Order(evs, true)
```

The evals whose indexes are before the pivot are the winners. When ordering by
`HiRank` there is always at least 1 winner (except for [`Video`][type] types),
but when ordering by `LoRank` there may be no winners:

```go
for _, i := range hiOrder[:hiPivot] {
	fmt.Printf("%s is a Hi winner!", evs[i])
}
```

Similarly, for Lo winners:

```go
for _, i := range loOrder[:loPivot] {
	fmt.Printf("%s is a Lo winner!", evs[i])
}
```

## Benchmarks

The [`test`](/test) directory is a separate module that benchmarks `cardrank`
against other well-known, pure Go poker hand evaluators, and verifies that
every evaluator ranks the same randomly dealt Hold'em and Omaha hands
identically:

```sh
cd test && go test -bench .
```

Results with Go 1.27.3 on `linux/amd64`, with the [`twoplustwo`](#two-plus-two)
package imported (time per hand, with allocations per hand in parentheses when
non-zero).

Best 5 of 5, 6, or 7 cards:

| Evaluator                                  | 5 cards      | 6 cards      | 7 cards       |
| ------------------------------------------ | -----------: | -----------: | ------------: |
| `cardrank` [`CactusFast`][cactus-fast]     | 2.7 ns       | 61 ns (3)    | 120 ns (3)    |
| `cardrank` [`Cactus`][cactus]              | 7.9 ns       | 92 ns (3)    | 234 ns (3)    |
| `cardrank` [`TwoPlusTwo`][two-plus-two]    | 16.9 ns      | 22.0 ns      | 24.7 ns       |
| `cardrank` [`NewHybridEval`][hybrid-eval]  | 18.5 ns (1)  | 62 ns (3)    | 54.9 ns (1)   |
| `cardrank` [`Holdem.Eval`][type]           | 104 ns (2)   | 159 ns (4)   | 115 ns (2)    |
| [paulhankin/poker][paulhankin]             | 5.4 ns       |              | 14.3 ns       |
| [alexclewontin/riverboat][riverboat]       | 5.0 ns       | 56 ns (1)    | 380 ns (7)    |
| [chehsunliu/poker][chehsunliu]             | 10.3 ns      | 73 ns        | 522 ns        |
| [notnil/joker][joker]                      | 1.27 µs (48) | 7.1 µs (269) | 25.4 µs (984) |

Omaha, best of any 2 pocket and 3 board cards, with 5 board cards:

| Evaluator                              | 4 pocket      | 5 pocket      | 6 pocket      |
| -------------------------------------- | ------------: | ------------: | ------------: |
| `cardrank` [`CactusFast`][cactus-fast] | 225 ns        | 372 ns        | 549 ns        |
| `cardrank` [`Omaha.Eval`][type]        | 364 ns (2)    | 498 ns (2)    | 658 ns (2)    |
| [alexclewontin/riverboat][riverboat]   | 306 ns        | 505 ns        | 749 ns        |
| [chehsunliu/poker][chehsunliu]         | 1.26 µs       | 2.10 µs       | 3.15 µs       |
| [paulhankin/poker][paulhankin]         | 1.27 µs       | 2.10 µs       | 3.13 µs       |
| [notnil/joker][joker]                  | 73.6 µs (2973) | 120 µs (4941) | 181 µs (7397) |

Notes:

- `CactusFast` and `Cactus` evaluate 6 and 7 cards by ranking every 5 card
  combination with [`NewEval`][new-eval].
- `NewHybridEval` uses [`RankCactus`][rank-cactus] (`CactusFast`) for 5 and 6
  cards, and `TwoPlusTwo` for 7 cards when the `twoplustwo` package is
  imported. It is the eval used by `Holdem.Eval`.
- `Holdem.Eval` and `Omaha.Eval` are the high-level APIs ([`Omaha`][type],
  [`OmahaFive`][type], and [`OmahaSix`][type]): they also determine and order
  the best five and unused cards, and allocate the returned [`Eval`][eval].
- None of the other evaluators support Omaha, so their Omaha results (and the
  `CactusFast` Omaha result) rank every combination of 2 pocket and 3 board
  cards with the evaluator's 5 card eval: 60, 100, and 150 evals for 4, 5, and
  6 pocket cards.
- paulhankin/poker does not evaluate 6 cards, and decodes its lookup tables
  when the package is initialized (about 150 ms and 40 MB).
- The `TwoPlusTwo` lookup table is decoded on first use (about 140 ms).

## Development

Generated code and data can be regenerated with `go generate ./...`, using the
commands in [`internal/cmd`](/internal/cmd):

- [`cgen`](/internal/cmd/cgen) generates the combination tables in [`combin.go`](combin.go)
- [`twoplustwogen`](/internal/cmd/twoplustwogen) generates the `internal/tpt/twoplustwo*.dat` lookup table, verifying it against the reference table's MD5 hash
- [`paycalcgen`](/internal/cmd/paycalcgen) generates the `paycalc/top*.csv` payout tables from the WSOP website

Card parsing has fuzz tests:

```sh
go test -fuzz FuzzParse
```

Some long-running tests are skipped unless enabled with the `TESTS`
environment variable:

```sh
TESTS=all go test ./...
```

## Links

- [Overview of Cactus Kev][cactus-kev] - original Cactus Kev article
- [Coding the Wheel][coding-the-wheel] - article covering various poker hand evaluators
- [Paul Senzee Perfect Hash for Cactus Kev][senzee] - overview of the "Fast Cactus" perfect hash by Paul Senzee
- [TwoPlusTwoHandEvaluator][tangentforks] - original implementation of the Two-Plus-Two evaluator

[cactus-kev]: https://archive.is/G6GZg
[coding-the-wheel]: https://www.codingthewheel.com/archives/poker-hand-evaluator-roundup/
[senzee]: http://senzee.blogspot.com/2006/06/some-perfect-hash.html
[tangentforks]: https://github.com/tangentforks/TwoPlusTwoHandEvaluator
[pkg]: https://pkg.go.dev/github.com/cardrank/cardrank
[eval-ranking]: #eval-ranking
[winners]: #winner-determination
[card]: https://pkg.go.dev/github.com/cardrank/cardrank#Card
[suit]: https://pkg.go.dev/github.com/cardrank/cardrank#Suit
[rank]: https://pkg.go.dev/github.com/cardrank/cardrank#Rank
[deck]: https://pkg.go.dev/github.com/cardrank/cardrank#Deck
[dealer]: https://pkg.go.dev/github.com/cardrank/cardrank#Dealer
[eval]: https://pkg.go.dev/github.com/cardrank/cardrank#Eval
[invalid]: https://pkg.go.dev/github.com/cardrank/cardrank#Invalid
[type]: https://pkg.go.dev/github.com/cardrank/cardrank#Type
[type-desc]: https://pkg.go.dev/github.com/cardrank/cardrank#TypeDesc
[deck-type]: https://pkg.go.dev/github.com/cardrank/cardrank#DeckType
[eval-type]: https://pkg.go.dev/github.com/cardrank/cardrank#EvalType
[desc-type]: https://pkg.go.dev/github.com/cardrank/cardrank#DescType
[street-desc]: https://pkg.go.dev/github.com/cardrank/cardrank#StreetDesc
[order]: https://pkg.go.dev/github.com/cardrank/cardrank#Order
[eval-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#EvalRank
[eval-func]: https://pkg.go.dev/github.com/cardrank/cardrank#EvalFunc
[eval.comp]: https://pkg.go.dev/github.com/cardrank/cardrank#Eval.Comp
[eval.hi-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#Eval.HiRank
[eval.lo-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#Eval.LoRank
[rank-cactus]: https://pkg.go.dev/github.com/cardrank/cardrank#RankCactus
[cactus]: https://pkg.go.dev/github.com/cardrank/cardrank#Cactus
[cactus-fast]: https://pkg.go.dev/github.com/cardrank/cardrank#CactusFast
[two-plus-two]: https://pkg.go.dev/github.com/cardrank/cardrank/twoplustwo
[pkg-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package
[paycalc]: https://pkg.go.dev/github.com/cardrank/cardrank/paycalc
[new-eval]: https://pkg.go.dev/github.com/cardrank/cardrank#NewEval
[hybrid-eval]: https://pkg.go.dev/github.com/cardrank/cardrank#NewHybridEval
[paulhankin]: https://github.com/paulhankin/poker
[riverboat]: https://github.com/alexclewontin/riverboat
[chehsunliu]: https://github.com/chehsunliu/poker
[joker]: https://github.com/notnil/joker
[set-rank-cactus]: https://pkg.go.dev/github.com/cardrank/cardrank#SetRankCactus
[twoplustwo]: https://pkg.go.dev/github.com/cardrank/cardrank/twoplustwo
[twoplustwo-new]: https://pkg.go.dev/github.com/cardrank/cardrank/twoplustwo#New
