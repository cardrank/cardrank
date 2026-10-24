# Outs

[`CalcOuts`][calc-outs] calculates each position's outs in a deal in progress:
the next board cards after which the position wins or ties, for display to a
spectator. For example:

```text
Board: K♠ 7♦ 2♥ 9♥
Player 1 (K♦ 7♠, ahead):  Two Pair, Kings over Sevens; 7 fades: any ♥ except K♥ 7♥
Player 2 (A♥ Q♥, behind): 7 outs: any ♥ except K♥ 7♥
Player 3 (3♣ 4♣, behind): drawing dead
```

Player 2's flush draw has 9 unseen hearts, but the K♥ and 7♥ also give player
1 a full house, so they aren't outs.

## Definitions

- **Outs** are the next street's cards after which a position wins or ties
  for the win. For a position that is **behind** on the current board, these
  are its outs. For a position that is **ahead**, they are the cards that keep
  it ahead, and the cards that cost it the lead are its **fades** (the "cards
  to dodge").
- Wins and ties are kept separate: a tie is a split pot, not a win.
- Outs are for the **next street only**. Runner-runner draws (needing two or
  more cards) are an equity question, answered by [`CalcOdds`][calc-odds].
- **Drawing dead** means the position can't win or tie on any remaining board,
  not just the next card. On the flop, a position with no outs on the turn can
  still win with runner-runner cards, so being dead is calculated over every
  remaining board.
- Outs are **clean**: each out is a card after which the position wins or ties
  against every active opponent, so cards that improve the position but give
  an opponent a better hand aren't outs.
- **Hi and Lo** outs are calculated separately, for types with a Lo.

## Supported types

Outs are calculated when the next street deals exactly 1 board card and no
pocket cards, and the board and active pockets are complete for the current
street. That is on the flop and the turn (for the turn and river cards) of
Hold'em, Omaha, Short Deck, and the other community board types (including
Hi/Lo). Otherwise, [`ErrInvalidCalc`][err] is returned, such as for:

- pre-flop (the flop deals 3 cards), or after the river (no next street)
- double board types ([`Double`][type], [`OmahaDouble`][type], …), where the
  next street deals a card to each board
- types that deal pocket cards on later streets ([`Fusion`][type]), and stud
  and draw types

## API

```go
// CalcOuts calculates the outs for the next street of the last of the runs,
// for every unseen card that can be dealt.
func CalcOuts(ctx context.Context, typ Type, runs []*Run, opts ...OutsOption) (*Outs, error)
func (d *Dealer) Outs(ctx context.Context, opts ...OutsOption) (*Outs, error)
func (typ Type) Outs(ctx context.Context, pockets []Cards, board Cards, opts ...OutsOption) (*Outs, error)

// WithActive, WithFolded, and WithDiscard are RunOptions, usable with both
// CalcOdds and CalcOuts.
type RunOption interface {
	OddsOption
	OutsOption
}

type Outs struct {
	Type   Type
	Street StreetDesc // the next street
	Unseen Cards      // the cards that can be dealt on the next street
	Hi     []Out      // by position
	Lo     []Out      // by position, nil when the type has no Lo
}

type Status uint8 // Inactive, Behind, Tied, or Ahead, on the current board

type Out struct {
	Status Status
	Hand   Hand      // the position's hand on the current board
	Wins   []OutCard // the cards after which the position is the only winner
	Ties   []OutCard // the cards after which the position ties for the win
	Losses []OutCard // the cards after which the position loses (fades)
	Dead   bool      // can't win or tie on any remaining board
}

func (out Out) Cards() Cards // the wins and ties

type OutCard struct {
	Card    Card
	Hand    Hand // the position's hand with the card
	Winners int  // the number of winners with the card
}

// Summary summarizes the position's outs (or fades, when ahead), such as "7
// outs: any ♥ except K♥ 7♥; ties: 4♠", "no outs", or "drawing dead".
func (outs *Outs) Summary(position int, low bool) string

// Describe describes the position's wins by resulting hand, such as "Flush:
// any ♥ except K♥ 7♥ (7); Three of a Kind: 7♣ 7♦ (2)", followed by its ties.
func (outs *Outs) Describe(position int, low bool) string

// Groups groups the cards by the resulting hand's short description (such as
// "Flush", or "Eight-low"), best first.
func Groups(cards []OutCard) []Group

type Group struct {
	Name  string   // the short description
	Rank  EvalRank // the best rank in the group
	Cards Cards
}

// DescribeCards describes cards relative to the unseen cards (see below).
func DescribeCards(cards, unseen Cards) string
```

The outs are stored as data (exact cards and hands), and described only when
formatting. Callers can format the data themselves, such as to label
positions with player names.

### Compression

Cards are described relative to the unseen cards, choosing for each suit, and
then each rank, whichever description is **shortest** when formatted:

- **List the cards:** "8♠ 7♠ 6♠ 5♠ 4♠".
- **Any:** "any ♠", when every unseen ♠ is in the cards.
- **Any, except:** "any ♠ except 2♠", listing the unseen ♠ that aren't in the
  cards as the exceptions.

Length is the formatted width (in runes): "8♠ 7♠ 6♠ 5♠ 4♠" (14) is shorter than
"any ♠ except 2♠" (15), but "any ♠ except 3♠ 2♠" is shorter than listing 8
spades. When the lengths are equal, the cards are listed, as it is more
precise.

Suits are compressed first, then ranks from the remaining cards (so a card is
only described once), and any remaining cards are listed, ordered by rank and
suit. The same rule applies to ranks: "any 7" is shorter than listing 3 or 4
sevens, but "7♥ 7♦ 7♣" is shorter than "any 7 except 7♠".

An exception is any unseen card of the suit or rank that isn't in the cards
being described: a card that isn't an out (such as one that gives an opponent
a better hand), one that only ties (ties are listed separately), or, when
describing a group, one that is in another group:

```text
Summary:  11 outs: any ♠ except 4♠, any T
Describe: Straight Flush: T♠ (1); Flush: any ♠ except T♠ 4♠ (7); Straight: T♥ T♦ T♣ (3)
```

### Example

```go
outs, err := d.Outs(ctx, cardrank.WithFolded(true))
if err != nil {
	return err
}
for i, out := range outs.Hi {
	if out.Status != cardrank.Inactive {
		fmt.Printf("%s: %s\n", names[i], outs.Summary(i, false))
	}
}

// the odds of hitting an out on the next card
out := outs.Hi[1]
pct := float64(len(out.Cards())) / float64(len(outs.Unseen)) * 100
```

## How it works

1. **Validate.** The current street is found from the last run's board length
   (the sum of the streets' board counts) and the active pockets. The next
   street must deal exactly 1 board card and no pocket cards.
2. **Unseen cards.** The cards of all runs are excluded from the deck, as
   [`CalcOdds`][calc-odds] does (with the same active, folded, and discard
   options).
3. **Status.** The active positions are evaluated on the current board, and
   ordered: a sole winner is `Ahead`, tied winners are `Tied`, and the rest are
   `Behind`.
4. **Draws.** For each unseen card, the card is added to the board, each active
   position is evaluated (with ordered best cards, for the hand descriptions),
   the Hi and Lo are ordered, and the card is added to each position's wins,
   ties, or losses. This is at most 45 cards (heads-up Hold'em on the flop)
   times the number of positions.
5. **Dead.** When the next street is the last, a position is dead when it has
   no wins or ties. Otherwise (on the flop), every remaining board is
   enumerated as `CalcOdds` does (990 boards for Hold'em), and a position is
   dead when it never wins or ties.

The tests check hand-analyzed spots, and that the outs on the turn agree with
`CalcOdds` (where each river card is one outcome) for random deals of several
types, including Hi/Lo.

## Review

The design was reviewed by OpenAI Pro, Gemini Pro, DeepSeek, and Kimi, which
agreed on these changes from the first draft:

- For a position that is ahead, its winning cards aren't outs: they are kept,
  and its losses reported as fades.
- Wins and ties are separate (the first draft had a `Split` flag).
- Per-position outs are stored, instead of a full `Result` for every draw.
- No runner-runner outs (the first draft had a `WithRunout` option).
- "Any ♥" compression is done in formatting, not in the data.
- "Drawing dead" can't be concluded from the next card alone when more streets
  follow.
- The options need a shared interface type to work with both calcs.

## Open questions

- Should double board types be supported, with outs for each board?
- Are counterfeited outs (cards that pair the board and lose a position's
  lead) worth calling out separately, beyond being in `Losses`?
- `Describe` describes a position's wins. For a position that is ahead, should
  it describe its fades, grouped by the opponent's resulting hand?

[calc-outs]: https://pkg.go.dev/github.com/cardrank/cardrank#CalcOuts
[calc-odds]: https://pkg.go.dev/github.com/cardrank/cardrank#CalcOdds
[err]: https://pkg.go.dev/github.com/cardrank/cardrank#ErrInvalidCalc
[type]: https://pkg.go.dev/github.com/cardrank/cardrank#Type
