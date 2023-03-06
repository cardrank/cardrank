# About

Package `cardrank` is a library of types, utilities, and interfaces for working
with playing cards, card decks, evaluating poker ranks, managing deals and run
outs for different game types.

[![Tests](https://github.com/cardrank/cardrank/workflows/Test/badge.svg)](https://github.com/cardrank/cardrank/actions?query=workflow%3ATest)
[![Go Report Card](https://goreportcard.com/badge/github.com/cardrank/cardrank)](https://goreportcard.com/report/github.com/cardrank/cardrank)
[![Reference](https://pkg.go.dev/badge/github.com/cardrank/cardrank.svg)][pkg]
[![Releases](https://img.shields.io/github/v/release/cardrank/cardrank?display_name=tag&sort=semver)](https://github.com/cardrank/cardrank/releases)

## Overview

The `cardrank` package contains types for working with [`Card`'s][card],
[`Suit`'s][suit], [`Rank`'s][rank], [`Deck`'s][deck], [evaluating poker
ranks][eval], and [managing deals and run outs][dealer].

In most cases, using [the high-level `Dealer`][dealer] with any [registered
`Type`][type] should be sufficient for most purposes. An in-depth example is
provided [in the package documentation][pkg-example].

A [`Type`][type] wraps a [type description][type-desc] defining a type's [deal
streets][street-desc], [deck][deck-type], [eval][eval-type], [Hi/Lo
description][desc-type] and other meta-data needed for [dealing streets and
managing run outs][dealer].

[Evaluation and ranking][eval-ranking] of the [types][type] is accomplished
through pure Go implementations of [well-known poker rank evaluation
algorithms](#cactus-kev). [Evaluation][eval] of cards can be compared and
[ordered to determine winner(s)][winners].

## Supported Types

Supports [evaluating and ranking][eval] the following [`Type`][type]'s:

| Holdem Variants    | Omaha Variants           | Hybrid Variants      | Draw Variants      | Other                   |
|--------------------|--------------------------|----------------------|--------------------|-------------------------|
| [`Holdem`][type]   | [`Omaha`][type]          | [`Dallas`][type]     | [`Video`][type]    | [`Soko`][type]          |
| [`Split`][type]    | [`OmahaHiLo`][type]      | [`Houston`][type]    | [`Draw`][type]     | [`SokoHiLo`][type]      |
| [`Short`][type]    | [`OmahaDouble`][type]    | [`Fusion`][type]     | [`DrawHiLo`][type] | [`Lowball`][type]       |
| [`Manila`][type]   | [`OmahaFive`][type]      | [`FusionHiLo`][type] | [`Stud`][type]     | [`LowballTriple`][type] |
| [`Spanish`][type]  | [`OmahaSix`][type]       |                      | [`StudHiLo`][type] | [`Razz`][type]          |
| [`Royal`][type]    | [`Courchevel`][type]     |                      | [`StudFive`][type] | [`Badugi`][type]        |
| [`Double`][type]   | [`CourchevelHiLo`][type] |                      |                    |                         |
| [`Showtime`][type] |                          |                      |                    |                         |
| [`Swap`][type]     |                          |                      |                    |                         |
| [`River`][type]    |                          |                      |                    |                         |

See the package's [`Type`][type] documentation for an overview of the above.

## Using

To use within a Go package:

```sh
go get github.com/cardrank/cardrank
```

See package level [Go package documentation][pkg] for in-depth overviews of
APIs.

### Quickstart

Various examples are available in the [Go package documentation][pkg] showing
use of various types, utilities, and interfaces.

Additional examples for a [`Dealer`][dealer] and the [`Holdem`][type] and
[`OmahaHiLo`][type] types are included in the [example directory](/_example):

* [dealer](/_example/dealer) - shows use of the [`Dealer`][dealer], to handle dealing cards, handling multiple run outs, and determining winners using any [`Type`'s][type]
* [holdem](/_example/holdem) - shows using types and utilities to deal [`Holdem`][type]
* [omahahilo](/_example/omahahilo) - shows using types and utilities to [`OmahaHiLo`][type], demonstrating splitting Hi and Lo wins

### Eval Ranking

[`EvalRank`][eval-rank]'s are determined using a registered
[`EvalFunc`][eval-func] associated with the [`Type`][type]. `EvalRank`'s are
always ordered, low to high, and are relative/comparable:

```go
fmt.Printf("%t\n", cardrank.StraightFlush < cardrank.FullHouse)

// Output:
true
```

Pocket and board [`Card`'s][card] can be passed to a [`Type`'s][type] `Eval`
method, which in turn uses the `Type`'s registered [`EvalFunc`][eval-func] and
returns an [`Eval`uated][eval] value:

```go
pocket, board := cardrank.Must("Ah Kh"), cardrank.Must("Qh Jh Th 2s 3s")
ev := cardrank.Holdem.Eval(pocket, board)
fmt.Printf("%s - %d\n", ev, ev.HiRank)

// Output:
// Straight Flush, Ace-high, Royal [Ah Kh Qh Jh Th] - 1
```

When evaluating cards, usually the eval is for 5, 6, or 7 cards, but some
`Type`'s are capable of evaluating fewer `Card`'s:

```go
pocket := cardrank.Must("2h 3s 4c")
ev := cardrank.Badugi.Eval(pocket, nil)
fmt.Printf("%s\n", ev)

// Output:
// Four, Three, Two-low [4c 3s 2h]
```

If an invalid number of cards is passed to a `Type`'s `EvalFunc`, the `Eval`'s
[`HiRank`][eval.hi-rank] and [`LoRank`][eval.lo-rank] values will be set to
[`Invalid`][invalid].

[`Eval`][eval]'s can be used to compare different hands of `Card`'s in order to
determine a winner, by comparing the [`Eval.HiRank`][eval.hi-rank] or
[`Eval.LoRank`][eval.lo-rank] values.

#### Hi/Lo

Different [`Type`'s][type] may have both a Hi and Lo [`EvalRank`][eval-rank],
such as [`Double`][type] board [`Holdem`][type], and various `*HiLo` variants,
such as [`OmahaHiLo`][type].

When a [`Eval`][eval] is created, both the Hi and Lo values will be made
available in the resulting `Eval` as the [`HiRank`][eval.hi-rank] and
[`LoRank`][eval.lo-rank], respectively.

#### Cactus Kev

For most [`Type`'s][type], the [`EvalRank`][eval-rank] is determined by Go
implementations of a few well-known [Cactus Kev][cactus-kev] algorithms:

* [`Cactus`][cactus] - the original [Cactus Kev][cactus-kev] poker hand evaluator
* [`CactusFast`][cactus-fast] - the [Fast Cactus][senzee] poker hand evaluator, using Paul Senzee's perfect hash lookup
* [`TwoPlusTwo`][two-plus-two] - the [2+2 forum][tangentforks] poker hand evaluator, using a 130 MiB lookup table

See [below for more information](#rank-cactus-func) on the default rank func in
use by the package, and for information on [using build tags][build-tags] to
enable/disable functionality for different target runtime environments.

#### Rank Cactus Func

The package-level [`RankCactus`][rank-cactus] variable is used for regular
poker evaluation, and can be set externally when wanting to build new game
types, or trying new algorithms.

#### Two-Plus-Two

[`NewTwoPlusTwoEval`][two-plus-two] makes use of a large (approximately 130
MiB) lookup table to accomplish extremely fast 5, 6 and 7 card hand rank
evaluation. Due to the large size of the lookup table, the lookup table can be
excluded when using the using the [`portable` or `embedded` build tags][build-tags],
with slightly degraded performance for when evaluating 7 cards.

Note: it is disabled by default when `GOOS=js` (ie, WASM) builds, but can be
forced to be included with the [`forcefat` build tag][build-tags].

### Winner Determination

Winner(s) are determined by the [lowest possible `EvalRank`][eval-rank] for
either the Hi or Lo value for the [`Type`][type]. Two or more hands having a
`EvalRank` of equal value indicate that the hands have equivalent ranks, and
have both won.

`Eval`'s can be sorted (low-to-high) by it's `HiRank` and `LoRank` members.
Winner(s) of a hand will be the hands in the lowest position and having
equivalent `HiRank`'s or `LoRank`'s.

#### Comparing Eval Ranks

A [`Eval`][eval] can be compared to another `Eval` using [`Comp`][eval.comp].

`Comp` returns `-1`, `0`, or `+1`, making it easy to compare or sort hands:

```go
// Compare Hi:
if ev1.Comp(ev2, false) < 0 {
	fmt.Printf("%s is a winner!", ev1)
}

// Compare Lo:
if ev1.Comp(ev2, true) == 0 {
	fmt.Printf("%s and %s are equal!", ev1, ev2)
}

// Sort by Hi:
hi := evs[0].NewComp(false)
sort.Slice(evs, func(i, j int) bool {
	return hi(evs[i], evs[j]) < 0
})

// Sort by Lo:
lo := evs[0].NewComp(true)
sort.Slice(evs, func(i, j int) bool {
	return lo(evs[i], evs[j]) < 0
})
```

#### Ordering Evals

[`Order`][order] can determine the winner(s) of a hand by ordering the indexes
of a [`[]*Eval`][eval] and returning the list of ordered evals as a `[]int` and
an `int` pivot indicating the position within the returned `[]int` as a cutoff
for a win:

```go
// Order by HiRank:
hiOrder, hiPivot := cardrank.Order(evs, false)
```

For a [Type][type] with a lo value:

```go
// Order by LoRank:
loOrder, loPivot := cardrank.Order(evs, true)
```

A `Eval` whose index is in position `i < pivot` is considered to be the
winner(s). When ordering by `HiRank`, there will be 1 or more winner(s) (with
exception for [`Video`][type] types), but when ordering by `LoRank` there may
be 0 or more winner(s):

```go
for i := 0; i < hiPivot; i++ {
	fmt.Printf("%s is a Hi winner!", evs[hiOrder[i]])
}
```

Similarly, for lo winners:

```go
for i := 0; i < loPivot; i++ {
	fmt.Printf("%s is a Lo winner!", evs[loOrder[i]])
}
```

### Build Tags

Build tags can be used with `go build` to change the package's build
configuration. Available tags:

#### `portable`

The `portable` tag disables inclusion of the [Two-plus-two lookup
tables][two-plus-two], and creating significantly smaller binaries but at the
cost of more expensive poker hand rank evaluation. Useful when building for
portable or embedded environments, such as a client application:

```sh
go build -tags portable
```

#### `embedded`

The `embedded` tag disables the `CactusFast` and the `TwoPlusTwo`, creating the
smallest possible binaries. Useful when either embedding the package in another
application, or in constrained runtime environments such as WASM:

```sh
GOOS=js GOARCH=wasm go build -tags embedded
```

#### `noinit`

The `noinit` tag disables the package level initialization. Useful when
applications need the fastest possible startup times and can defer
initialization, or when using a third-party algorithm:

```sh
GOOS=js GOARCH=wasm go build -tags 'embedded noinit' -o cardrank.wasm
```

When using the `noinit` build tag, the user will need to call the [`Init`
func][init] to set `RankCactus` and to register the default types
automatically:

```go
// Set DefaultCactus, DefaultRank based on available implementations:
cardrank.Init()
```

Alternatively, the `RankCactus` can be set manually. After `RankCactus` has
been set, call `RegisterDefaultTypes` to register built in types:

```go
// Set when using a third-party implementation, or experimenting with new
// Cactus implementations:
cardrank.RankCactus = cardrank.CactusFast

// Call RegisterDefaultTypes to register default types
if err := cardrank.RegisterDefaultTypes(); err != nil {
	panic(err)
}
```

#### `forcefat`

The `forcefat` tag forces a "fat" binary build, including the `TwoPlusTwo`'s
large lookup table, irrespective of other build tags:

```sh
GOOS=js GOARCH=wasm go build -tags 'forcefat' -o cardrank.wasm
```

## Links

* [Overview of Cactus Kev][cactus-kev] - original Cactus Kev article
* [Coding the Wheel][coding-the-wheel] - article covering various poker hand evaluators
* [Paul Senzee Perfect Hash for Cactus Kev][senzee] - overview of the "Fast Cactus" perfect hash by Paul Senzee
* [TwoPlusTwoHandEvaluator][tangentforks] - original implementation of the Two-Plus-Two evaluator

[cactus-kev]: https://archive.is/G6GZg
[coding-the-wheel]: https://www.codingthewheel.com/archives/poker-hand-evaluator-roundup/
[senzee]: http://senzee.blogspot.com/2006/06/some-perfect-hash.html
[tangentforks]: https://github.com/tangentforks/TwoPlusTwoHandEvaluator

[pkg]: https://pkg.go.dev/github.com/cardrank/cardrank
[eval-ranking]: #eval-ranking
[build-tags]: #build-tags
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
[init]: https://pkg.go.dev/github.com/cardrank/cardrank#Init
[order]: https://pkg.go.dev/github.com/cardrank/cardrank#Order
[eval-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#EvalRank
[eval-func]: https://pkg.go.dev/github.com/cardrank/cardrank#EvalFunc
[eval.comp]: https://pkg.go.dev/github.com/cardrank/cardrank#Eval.Comp
[eval.hi-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#Eval.HiRank
[eval.lo-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#Eval.LoRank
[rank-cactus]: https://pkg.go.dev/github.com/cardrank/cardrank#RankCactus
[cactus]: https://pkg.go.dev/github.com/cardrank/cardrank#Cactus
[cactus-fast]: https://pkg.go.dev/github.com/cardrank/cardrank#CactusFast
[two-plus-two]: https://pkg.go.dev/github.com/cardrank/cardrank#NewTwoPlusTwoEval

[pkg-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package
