# Migrating from v0.16

This guide covers migrating code using `cardrank` v0.16.x (the last release,
v0.16.5) to the new API. It is written for both people and coding agents: the
[checklist](#checklist) gives the order of work, the [rename
table](#renames) can be applied mechanically, and the [pattern
sections](#patterns) show how code using the removed APIs is rewritten.

The new API evaluates hands identically: for the same cards, ranks and best
and unused cards are unchanged, except for the [bug fixes](#behavior-changes)
listed below.

## Checklist

1. Require Go 1.27 or later (`go 1.27` in `go.mod`).
2. Remove the `portable`, `embedded`, `noinit`, `forcefat`, and `faketime`
   build tags from builds, and add the [`twoplustwo` import](#build-and-initialization)
   if you want its faster 7 card evals.
3. Remove calls to `Init`, `RegisterDefaultTypes`, and `paycalc.Init`.
4. Apply the [renames](#renames).
5. Rewrite the [removed APIs](#patterns): evals, ordering, the dealer loop,
   results, calcs, decks, and custom evaluators.
6. Run `go vet ./...` until it is clean, then use the [search
   patterns](#finding-remaining-usages) to find anything missed.
7. Review the [behavior changes](#behavior-changes), and update any expected
   test output that depends on them.

## Build and initialization

All build tags are removed. The package always initializes itself and always
includes `CactusFast`. The Two-Plus-Two evaluator (and its ~124 MiB lookup
table) is included only when its package is imported:

```go
import (
	"github.com/cardrank/cardrank"
	_ "github.com/cardrank/cardrank/twoplustwo"
)
```

| v0.16 | New |
|---|---|
| default build (Two-Plus-Two included) | import `github.com/cardrank/cardrank/twoplustwo` |
| `-tags portable`, `-tags embedded`, `GOOS=js` default | don't import `twoplustwo` |
| `-tags forcefat` | import `twoplustwo` |
| `-tags noinit` with `cardrank.Init()` and `cardrank.RegisterDefaultTypes()` | remove: initialization is automatic, and takes about 1 ms |
| `paycalc.Init()` | remove: the payout tables are always registered |
| `cardrank.RankCactus = f` (before `RegisterDefaultTypes`) | `cardrank.SetRankCactus(f)`, from an `init` func |
| `cardrank.NewTwoPlusTwoEval()` | `twoplustwo.New()` |

Without the `twoplustwo` import, 7 card evals rank every 5 card combination
with `CactusFast`: the results are identical, but slower. The Two-Plus-Two
table is decoded on first use instead of at init, so importing it no longer
adds startup time.

`SetRankCactus` and `RegisterTwoPlusTwo` rebuild the evaluators of all
registered types, and like `RegisterType`, must only be called during
initialization.

## Renames

| v0.16 | New |
|---|---|
| `Formatter` | `Cards` |
| `Exclude(v, ex...)` | `Cards(v).Exclude(ex...)` |
| `NewDeck()` | `DeckFrench.New()` |
| `NewShoe(n)` | `DeckFrench.Shoe(n)` |
| `deck.Shuffle(r, n)`, `typ.Shuffle(r, n)` | `deck.Shuffle(r)`, `typ.Shuffle(r)` (see below) |
| `typ.Dealer(r, shuffles, players)` | `typ.Dealer(r, players)` |
| `typ.Deal(r, shuffles, players)` | `typ.Deal(r, players)` |
| `NewDealer(desc, deck, players)` | `NewDealer(typ, deck, players)`, passing the `Type` |
| `NewShuffledDealer(desc, r, shuffles, players)` | `typ.Dealer(r, players)` |
| `ev.HiRank`, `ev.HiBest`, `ev.HiUnused` | `ev.Hi.Rank`, `ev.Hi.Best`, `ev.Hi.Unused` |
| `ev.LoRank`, `ev.LoBest`, `ev.LoUnused` | `ev.Lo.Rank`, `ev.Lo.Best`, `ev.Lo.Unused` |
| `ev.Desc(false)`, `ev.Desc(true)` | `ev.Hi`, `ev.Lo` (a `Hand` formats its own description) |
| `EvalDesc` | `Hand` |
| `a.Comp(b, false)`, `a.Comp(b, true)` | `a.Hi.Compare(b.Hi)`, `a.Lo.Compare(b.Lo)` |
| `Type.Id()` | `Type.ID()` |
| `IdToType(id)` | `ParseType(id)` |
| `StreetDesc.Id` | `StreetDesc.ID` |
| `ErrInvalidId`, `ErrMismatchedIdAndType` | `ErrInvalidID`, `ErrMismatchedIDAndType` |
| `HashKey(c0, c1)` | `StartingKey(c0, c1)` |
| `StartingExpValue(p)`, `StartingEvalRank(p)` | `StartingHand(p)`, returning both |
| `HoldemStarting()` | `StartingHand(pocket)` for each pocket |
| `AlternateEmoji[i]` | `suit.AlternateEmoji()` |
| `NewCombinGen(s, k)`, `NewCombinUnusedGen(s, k)` | `Combinations(s, k)`, `CombinationsUnused(s, k)` (see below) |

`Parse`, `Must`, `Deck.Draw`, and the `Run` fields now use `Cards` (a
`[]Card`), which is assignable to and from `[]Card`. Signatures that took
`[][]Card` now take `[]Cards`.

Repeated shuffles with a pseudo-random source add no randomness, so the
shuffle count is removed. To reproduce v0.16 output that used a count `n`
greater than 1, call `Shuffle` `n` times on a deck and pass it to `NewDealer`.

## Patterns

### Evals and hands

An `Eval` holds a Hi and a Lo `Hand`:

```go
type Hand struct {
	Rank   EvalRank // Invalid when there is no hand (such as no qualifying Lo)
	Best   Cards
	Unused Cards
}
```

```go
// v0.16
ev := cardrank.Holdem.Eval(pocket, board)
desc := ev.Desc(false)
fmt.Printf("%s %v %v %d\n", desc, desc.Best, desc.Unused, ev.HiRank)

// new
ev := cardrank.Holdem.Eval(pocket, board)
fmt.Printf("%s %v %v %d\n", ev.Hi, ev.Hi.Best, ev.Hi.Unused, ev.Hi.Rank)
```

A `Hand` formats with the same verbs as `EvalDesc`. Use `Hand.Valid()` to
check for a qualifying Lo, instead of comparing `LoRank` with `Invalid`.

### Ordering and winners

`Order(evs, low)` returned an order and a pivot. `HiOrder` and `LoOrder`
return a `Ranking`, whose `Winners` are the winning positions (`Order` now
orders a `[]Hand`):

```go
// v0.16
order, pivot := cardrank.Order(evs, false)
for i := range pivot {
	fmt.Println(order[i], "wins with", evs[order[i]].Desc(false))
}

// new
hi := cardrank.HiOrder(evs)
for _, i := range hi.Winners() {
	fmt.Println(i, "wins with", evs[i].Hi)
}
```

`hi.Pivot` and the winners in `hi.Order[:hi.Pivot]` are the same as v0.16's
pivot and winners. Positions without a hand (folded, or no qualifying Lo) may
be ordered differently among the losers. A Lo with no qualifying hands has
`Pivot == 0` (v0.16 returned a nil order).

### Dealer

The `Dealer`'s exported fields, and its embedded `TypeDesc`, are replaced by
methods, and the `Next`/`NextResult` loops by iterators:

```go
// v0.16
d := typ.Dealer(r, 1, players)
for d.Next() {
	i, run := d.Run()
	if d.HasPocket() {
		fmt.Println(run.Pockets)
	}
	if d.Id() == 'f' && i == 0 {
		if !d.ChangeRuns(2) {
			panic("unable to change runs")
		}
		d.Deactivate(3)
	}
}
for d.NextResult() {
	i, res := d.Result()
	hi, lo := res.Win(names...)
	fmt.Printf("%d: %S\n", i, hi)
	if lo != nil {
		fmt.Printf("%d: %S\n", i, lo)
	}
}

// new
d := typ.Dealer(r, players)
for i, run := range d.Deal() {
	street := d.Street()
	if street.Pocket != 0 {
		fmt.Println(run.Pockets)
	}
	if street.ID == 'f' && i == 0 {
		if err := d.ChangeRuns(2); err != nil {
			panic(err)
		}
		if err := d.Fold(3); err != nil {
			panic(err)
		}
	}
}
for i, res := range d.Results() {
	fmt.Printf("%d: %s\n", i, res.Hi.Describe(names...))
	if res.Lo.Valid() {
		fmt.Printf("%d: %s\n", i, res.Lo.Describe(names...))
	}
}
```

| v0.16 | New |
|---|---|
| `d.Next()` loop with `d.Run()` | `for i, run := range d.Deal()` |
| `d.NextResult()` loop with `d.Result()` | `for i, res := range d.Results()` |
| `d.HasNext()` | not needed: the `Deal` loop ends when dealing is done |
| `d.Street()` (the street index) | `d.StreetIndex()` (`-1` before and after dealing); `d.Street()` now returns the current `StreetDesc` |
| `d.Id()`, `d.Name()` | `d.Street().ID`, `d.Street().Name` |
| `d.Pocket()`, `d.PocketUp()`, `d.PocketDiscard()`, `d.PocketDraw()` | the `d.Street()` fields of the same names |
| `d.Board()`, `d.BoardDiscard()` | `d.Street().Board`, `d.Street().BoardDiscard` |
| `d.HasPocket()`, `d.HasBoard()` | `d.Street().Pocket != 0`, `d.Street().Board != 0` |
| `d.NextId()` | `typ.Streets()[d.StreetIndex()+1].ID`, when `d.StreetIndex()+1 < len(typ.Streets())` |
| `d.Deactivate(p...)` (returned `bool`) | `d.Fold(p...)` (returns `error`) |
| `d.Active[i]` | `d.Active(i)` |
| `d.Inactive()` | `d.Active(i)` for each position |
| `d.ChangeRuns(n)` (returned `bool`) | `d.ChangeRuns(n)` (returns `error`) |
| `d.Calc(ctx, folded, opts...)` | `d.Odds(ctx, cardrank.WithFolded(folded), opts...)` |
| `d.Deck`, `d.Count` | `d.Deck()`, `d.Players()` |
| `d.Type`, `d.Low`, `d.Double`, `d.Max`, … (`TypeDesc` fields) | `d.Type()`, and its methods: `d.Type().Low()`, `d.Type().Double()`, `d.Type().Desc()` |
| `d.Runs`, `d.Results` | the `Deal` and `Results` iterators |
| `d.Deal(street, run)` | removed (internal) |

### Results and wins

`Result.Win(names...)` returned `(*Win, *Win)`. The Hi and Lo wins are now
fields of the `Result`:

| v0.16 | New |
|---|---|
| `res.HiOrder`, `res.HiPivot` | `res.Hi.Order`, `res.Hi.Pivot` |
| `res.LoOrder`, `res.LoPivot` | `res.Lo.Order`, `res.Lo.Pivot` |
| `hi, lo := res.Win(names...)` | `res.Hi`, `res.Lo` |
| `lo != nil` | `res.Lo.Valid()` |
| `fmt.Sprintf("%S", hi)` (using names) | `res.Hi.Describe(names...)` |
| `win.Invalid()` | `!win.Valid()` |
| `win.Desc()` | `win.Hands()` |
| `NewWin(...)` | removed |

`Win` still formats with the `%d`, `%s`, `%S`, `%v`, and `%V` verbs, where
`%S` uses the position numbers.

### Calcs

`OddsCalc`, `ExpValueCalc`, and their constructors are replaced by functions.
Options are typed (an option for the wrong calc no longer compiles), and calcs
return an `error` instead of a `bool`:

```go
// v0.16
hi, lo, ok := cardrank.Omaha.Odds(ctx, pockets, board)
expv, ok := cardrank.Holdem.ExpValue(ctx, pocket, cardrank.WithBoard(board))
hi, lo, ok = cardrank.NewOddsCalc(typ, cardrank.WithRuns(runs), cardrank.WithActive(active, true)).Calc(ctx)

// new
odds, err := cardrank.Omaha.Odds(ctx, pockets, board) // odds.Hi, odds.Lo
expv, err := cardrank.Holdem.ExpValue(ctx, pocket, board)
odds, err = cardrank.CalcOdds(ctx, typ, runs, cardrank.WithActive(active), cardrank.WithFolded(true))
```

| v0.16 | New |
|---|---|
| `NewOddsCalc(typ, opts...).Calc(ctx)` | `CalcOdds(ctx, typ, runs, opts...)` |
| `NewExpValueCalc(typ, pocket, opts...).Calc(ctx)` | `CalcExpValue(ctx, typ, pocket, board, opts...)` |
| `(*Odds, *Odds, bool)` | `(RunOdds, error)`, with `RunOdds.Hi` and `RunOdds.Lo` |
| `(*ExpValue, bool)` | `(ExpValue, error)` |
| `ok == false` (context done) | `errors.Is(err, context.Canceled)`, or `context.DeadlineExceeded` |
| `ok == false` (invalid input) | `ErrInvalidCalc` |
| `WithRuns(runs)` | the `runs` argument of `CalcOdds` |
| `WithPocketsBoard(pockets, board)` | `typ.Odds(ctx, pockets, board)` |
| `WithBoard(board)` | the `board` argument of `CalcExpValue` and `Type.ExpValue` |
| `WithActive(active map[int]bool, folded)` | `WithActive(active []bool)` and `WithFolded(folded)` |
| `CalcOption` (`func(any)`) | `OddsOption`, `ExpValueOption`, and `CalcOption` (both) |

When the context is done, the partial result is returned along with the
context's error, as v0.16 returned the partial result with `false`.

### Custom evaluators

The `New*Eval` constructors, `EvalFunc`, `EvalType.New`, the `Eval.Hi5` …
`Eval.Max7` helpers, and the per-type description funcs (`CactusDesc`,
`LowDesc`, …) are no longer exported. Evaluation is done by an `Evaluator`:

```go
type Evaluator interface {
	Rank(ev *Eval, pocket, board []Card) // ranks only (used by the calcs)
	Eval(ev *Eval, pocket, board []Card) // ranks, and orders best and unused cards
}
```

| v0.16 | New |
|---|---|
| `NewEval(f)` | `RankEvaluator(f)` |
| `NewHybridEval(normalize, low)`, `NewCactusEval(...)`, … for a built-in type | `typ.Evaluator()` |
| `EvalType.New(board, normalize, low)` | `TypeDesc.Eval` selects the built-in evaluator |
| a custom `EvalFunc` | set `TypeDesc.Evaluator` before `RegisterType` |
| `f(ev, pocket, board)` with `normalize == false` | `e.Rank(ev, pocket, board)` |
| `f(ev, pocket, board)` with `normalize == true` | `e.Eval(ev, pocket, board)` |
| `ev.Eval(pocket, board)` | unchanged |
| `CactusDesc(f, verb, rank, best, unused)`, … | format a `Hand` (its `DescType` comes from the `Type`) |

A `TypeDesc.Evaluator` is kept when the built-in evaluators are rebuilt by
`SetRankCactus` or `RegisterTwoPlusTwo`.

### Combinations

```go
// v0.16
for g, v := cardrank.NewCombinGen(cards, 2); g.Next(); {
	use(v)
}
for g, v := cardrank.NewCombinUnusedGen(cards, 5); g.Next(); {
	use(v[:5], v[5:])
}

// new
for v := range cardrank.Combinations(cards, 2) {
	use(v)
}
for best, unused := range cardrank.CombinationsUnused(cards, 5) {
	use(best, unused)
}
```

The order of the combinations is unchanged. As with v0.16, the yielded slices
are reused between iterations, so clone them (such as with `slices.Clone`) to
retain them. Collecting the iterator with `slices.Collect` without cloning
yields the same slice repeatedly.

### Decks

| v0.16 | New |
|---|---|
| `deck.All()` (`[]Card`) | `slices.Collect(deck.All())` (`All` is an `iter.Seq[Card]`) |
| `typ.Exclude(ex...)` with `[]Card` arguments | unchanged (`Cards` arguments accept `[]Card` values) |

### Removed internals

These had no external use, and are removed without a replacement:
`(*Odds).Add`, `NewOdds`, `(*Run).Eval`, `(*Run).CalcStart`, `NewResult`,
`BinGen`, `(*Dealer).HasActive`, and the `TypeDesc.Apply` method promoted
through the `Dealer`. Use `Type.Desc()` to inspect a dealer's type
description.

## Behavior changes

These change results or output for some inputs:

- **Hold'em flop evals.** `Holdem.Eval` (and `Split`) with a 2 card pocket
  and a 3 card board now evaluates the 5 cards. v0.16 returned the pocket's
  starting rank, ignoring the board. Fewer than 5 cards (such as pre-flop)
  still return the starting rank.
- **5 card Hi/Lo best cards.** For 5 card Hi/Lo evals (such as `DrawHiLo`, and
  `Split` on the flop), the Hi best cards are now ordered Ace-high. v0.16
  shared the Hi and Lo best cards, so ordering the Lo also reordered the Hi,
  and the Hi description was wrong (for example, "Eight-high, kickers Four,
  Three, Two, Ace" instead of "Ace-high, kickers Eight, Four, Three, Two").
  Ranks were not affected.
- **Ordering with no valid hands.** A `Ranking` never counts `Invalid` hands
  as winners. v0.16 counted equally `Invalid` Hi hands as winners (only
  possible with `Video` types).
- **`Fold`.** Returns `ErrInvalidPosition` for a position that doesn't exist,
  where `Deactivate` silently ignored it.
- **Single remaining player.** When all but one player folds, the `Result`'s
  `Evals` has an entry for every position (nil for folded positions). v0.16's
  `Evals` had a single entry, which could panic when formatting the win.
- **`Deck.Limit`.** Clamps the limit to the deck size (v0.16 could panic when
  drawing past the end of the deck).
- **`FromIndex`.** Returns `InvalidCard` for negative indexes (v0.16 could
  panic).
- **`RegisterType`.** Returns `ErrInvalidID` for duplicate street IDs (v0.16
  never detected them).
- **`DeckFrenchJoker`.** Its value changed from 244 to 252 (it was a typo), and
  it now has a name ("French Joker").
- **Expected value calcs.** On cancel, `CalcExpValue` waits for its workers to
  stop before returning, and uses a fixed number of workers
  (`runtime.GOMAXPROCS`), instead of a goroutine per board.

## Finding remaining usages

After `go vet ./...` is clean, search for patterns that still compile but may
need attention, such as reproducing v0.16 shuffles, or comments and strings
naming the old APIs:

```sh
grep -rnE '\b(HiRank|LoRank|HiBest|LoBest|HiUnused|LoUnused|HiOrder|LoOrder|HiPivot|LoPivot)\b|\.Comp\(|\.Desc\((true|false)\)|\b(Init|RegisterDefaultTypes|NewOddsCalc|NewExpValueCalc|NewCombinGen|NewCombinUnusedGen|NewDeck|NewShoe|IdToType|HashKey|Deactivate|NextResult|EvalFunc|Formatter)\b|-tags' --include='*.go' --include='*.md' --include='Makefile' --include='*.yml' .
```

`HiOrder` and `LoOrder` match both the removed `Result` fields and the new
functions, so check each match.
