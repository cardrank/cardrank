# About

Package `github.com/cardrank/cardrank` is a library of types, utilities, and
interfaces for working with playing cards, card decks, evaluating poker hand
ranks, managing deals and run outs for different game types.

<!-- START supports -->
Supports [Texas Holdem][holdem-example], [Texas Holdem Short (6+)][short-example],
[Texas Holdem Royal (10+)][royal-example], [Omaha][omaha-example],
[Omaha Hi/Lo][omaha-hi-lo-example], [Stud][stud-example], [Stud Hi/Lo][stud-hi-lo-example],
[Razz][razz-example], and [Badugi][badugi-example].
<!-- END -->

[![Tests](https://github.com/cardrank/cardrank/workflows/Test/badge.svg)](https://github.com/cardrank/cardrank/actions?query=workflow%3ATest)
[![Go Report Card](https://goreportcard.com/badge/github.com/cardrank/cardrank)](https://goreportcard.com/report/github.com/cardrank/cardrank)
[![Reference](https://pkg.go.dev/badge/github.com/cardrank/cardrank.svg)](https://pkg.go.dev/github.com/cardrank/cardrank)
[![Releases](https://img.shields.io/github/v/release/cardrank/cardrank?display_name=tag&sort=semver)](https://github.com/cardrank/cardrank/releases)

## Overview

The `github.com/cardrank/cardrank` package contains types for [cards][card],
[card suits][suit], [card ranks][rank], [card decks][deck], [evaluating poker
hands][eval], and [managing deals and run outs][dealer].

A single [package level type][type] provides a standard interface for dealing
cards for, and [evaluating the relative poker ranks][eval] of the following:

<!-- START overview -->
* [Texas Holdem][holdem-example]
* [Texas Holdem Short (6+)][short-example]
* [Texas Holdem Royal (10+)][royal-example]
* [Omaha][omaha-example]
* [Omaha Hi/Lo][omaha-hi-lo-example]
* [Stud][stud-example]
* [Stud Hi/Lo][stud-hi-lo-example]
* [Razz][razz-example]
* [Badugi][badugi-example]
<!-- END -->

[Evaluation and ranking][eval-ranking] of the supported [poker hand
types][type] is accomplished through pure Go implementations of [well-known
poker rank evaluation algorithms](#cactus-kev). [Poker hands][eval] can be
compared and [ordered to determine winner(s)][winners].

## Using

To use within a Go package:

```sh
go get github.com/cardrank/cardrank
```

See package level [Go documentation][pkg] for in-depth overviews of APIs.

### Quickstart

Complete [examples for all `Type`'s][examples] are available, including
additional examples showing use of various types, utilities, and interfaces is
available in the [Go package documentation][pkg].

Below are quick examples for [`Dealer`][dealer], [`Holdem`][type] and
[`OmahaHiLo`][type] included in the [example directory](/_example):

* [dealer](/_example/dealer) - shows use of the [`Dealer`][dealer], to handle dealing cards, handling multiple run outs, and determining winners using for all [`Type`'s][type]
* [holdem](/_example/holdem) - shows using types and utilities to deal a hand of [`Holdem`][holdem-example]
* [omahahilo](/_example/omahahilo) - shows using types and utilities to deal a hand of [`OmahaHiLo`][omaha-hi-lo-example], showcasing "lo" winners

### Eval Ranking

Poker ranks are determined on a low-to-high basis by evaluating the ranking
using a number of different registered [poker types][type].
[`EvalRank`'s][eval-rank] are relative/comparable:

```go
fmt.Printf("%t\n", cardrank.StraightFlush < cardrank.FullHouse)

// Output:
true
```

Pocket and board [`Card`'s][card] can be passed to a [`Type`'s][type] `Eval`
method, which in turn uses the `Type`'s registered [`EvalFunc`][eval-func] and
returns an [`Eval`uated][eval] value:

```go
v := cardrank.Must("Ah Kh Qh Jh Th")
ev := cardrank.Holdem.Eval(v, nil)
fmt.Printf("%s - %d\n", ev, ev.HiRank)

// Output:
// Straight Flush, Ace-high, Royal [Ah Kh Qh Jh Th] - 1
```

When evaluating cards, one usually passes 5, 6, or 7 cards, but some `Type`'s
are capable of evaluating fewer `Card`'s:

```go
v := cardrank.Must("2h 3s 4c")
ev := cardrank.Badugi.Eval(v, nil)
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

Different [poker types][type] may have both a "hi" and "lo" values, such as
[double board Holdem][double-example], or Hi/Lo variants such as [Omaha
Hi/Lo][omaha-hi-lo-example].

When a [`Eval`][eval] is created, both the "hi" and "lo" (if applicable) values
are evaluated, and stored in `Eval` as the [`HiRank`][eval.hi-rank] and and
[`LoRank`][eval.lo-rank] values, respectively.

#### Cactus Kev

For regular poker hand types, poker hand rank is determined by Go
implementations of different [Cactus Kev][cactus-kev] evaluators:

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

The [`TwoPlusTwo`][two-plus-two] makes use of a large (approximately 130 MiB)
lookup table to accomplish extremely fast 5, 6 and 7 card hand rank evaluation.
Due to the large size of the lookup table, the `TwoPlusTwo` will be excluded
when using the using the [`portable` or `embedded` build tags][build-tags].

The `TwoPlusTwo` is disabled by default for `GOOS=js` (ie, WASM) builds, but
can be enabled using the [`forcefat` build tag][build-tags].

### Winner Determination

Winner(s) are determined by the [lowest possible `EvalRank`][eval-rank] for
either the "hi" or "lo" value for the [`Type`][type]. Two or more hands having
a `EvalRank` of equal value indicate that the hands have equivalent ranks, and
have both won.

`Eval`'s can thus be sorted (low-to-high) by `HiRank` or `LoRank`, the
winner(s) of a hand will be the hands in the lowest position and having
equivalent `HiRank`'s or `LoRank`'s.

#### Comparing Eval Ranks

A [`Eval`][eval] can be compared to another `Eval` using
[`Comp`][eval-comp].

`Comp` returns `-1`, `0`, or `+1`, making it easy to compare or sort hands:

```go
// Compare hi evals:
if ev1.Comp(ev2, false) < 0 {
	fmt.Printf("%s is a winner!", ev1)
}

// Compare lo evals:
if ev1.Comp(ev2, true) == 0 {
	fmt.Printf("%s and %s are equal!", ev1, ev2)
}

// Sort hi evals:
hi := evs[0].NewComp(false)
sort.Slice(evs, func(i, j int) bool {
	return hi(evs[i], evs[j]) < 0
})

// Sort lo evals:
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
// Order hi evals, low to high:
hiOrder, hiPivot := cardrank.Order(evs, false)
```

For a [Type][type] with a lo value:

```go
// Order lo evals, low to high (applicable only for types with lo values):
loOrder, loPivot := cardrank.Order(evs, true)
```

A `Eval` whose index is in position `i < pivot` is considered to be the
winner(s) of the eval. Hi hands are guaranteed to have 1 or more winner(s),
while lo hands have 0 or more winner(s):

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

See [the examples][examples] for more in-depth use of `Order`.

### Build Tags

Build tags can be used with `go build` to change the package's build
configuration. Available tags:

#### `portable`

The `portable` tag disables the `TwoPlusTwo`, in effect excluding the the
[large lookup table](#two-plus-two), and creating significantly smaller
binaries but at the cost of more expensive poker hand rank evaluation. Useful
when building for portable or embedded environments, such as a client
application:

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
[examples]: https://pkg.go.dev/github.com/cardrank/cardrank#pkg-examples
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
[init]: https://pkg.go.dev/github.com/cardrank/cardrank#Init
[order]: https://pkg.go.dev/github.com/cardrank/cardrank#Order
[eval-rank]: https://pkg.go.dev/github.com/cardrank/cardrank#EvalRank
[eval-func]: https://pkg.go.dev/github.com/cardrank/cardrank#EvalFunc
[eval-comp]: https://pkg.go.dev/github.com/cardrank/cardrank#Eval.Comp
[rank-cactus]: https://pkg.go.dev/github.com/cardrank/cardrank#RankCactus
[cactus]: https://pkg.go.dev/github.com/cardrank/cardrank#Cactus
[cactus-fast]: https://pkg.go.dev/github.com/cardrank/cardrank#CactusFast
[two-plus-two]: https://pkg.go.dev/github.com/cardrank/cardrank#TwoPlusTwo
[hybrid]: https://pkg.go.dev/github.com/cardrank/cardrank#Hybrid

<!-- START links -->
[holdem-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Holdem
[double-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Double
[short-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Short
[royal-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Royal
[omaha-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Omaha
[omaha-hi-lo-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-OmahaHiLo
[stud-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Stud
[stud-hi-lo-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-StudHiLo
[razz-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Razz
[badugi-example]: https://pkg.go.dev/github.com/cardrank/cardrank#example-package-Badugi
<!-- END -->
