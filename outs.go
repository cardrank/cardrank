package cardrank

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// OutsOption is an option for [CalcOuts].
type OutsOption interface {
	applyOuts(*outsCalc)
}

// Status is a position's status on the current board.
type Status uint8

// Statuses.
const (
	// Inactive is a folded (not active) position.
	Inactive Status = iota
	// Behind is a position that loses on the current board.
	Behind
	// Tied is a position that ties for the win on the current board.
	Tied
	// Ahead is a position that wins on the current board.
	Ahead
)

// String satisfies the [fmt.Stringer] interface.
func (status Status) String() string {
	switch status {
	case Behind:
		return "behind"
	case Tied:
		return "tied"
	case Ahead:
		return "ahead"
	}
	return "inactive"
}

// Outs are the outs for the next street of a run: the cards that can be
// dealt, and each position's resulting wins, ties, and losses.
type Outs struct {
	// Type is the type.
	Type Type
	// Street is the next street.
	Street StreetDesc
	// Unseen are the cards that can be dealt on the next street.
	Unseen Cards
	// Hi are the Hi outs, by position.
	Hi []Out
	// Lo are the Lo outs, by position, or nil when the type has no Lo.
	Lo []Out
}

// Out are a position's outs for the next street.
//
// For a position that is [Behind] (or [Tied]), its outs are its wins (and
// ties). For a position that is [Ahead], its wins are the cards that keep it
// ahead, and its losses are its fades (the cards that cost it the lead).
type Out struct {
	// Status is the position's status on the current board.
	Status Status
	// Hand is the position's hand on the current board.
	Hand Hand
	// Wins are the cards after which the position is the only winner.
	Wins []OutCard
	// Ties are the cards after which the position ties for the win.
	Ties []OutCard
	// Losses are the cards after which the position loses.
	Losses []OutCard
	// Dead is true when the position can't win or tie on any remaining
	// board.
	Dead bool
}

// Cards returns the out cards: the wins and ties.
func (out Out) Cards() Cards {
	return append(drawCards(out.Wins), drawCards(out.Ties)...)
}

// OutCard is a drawn card, and a position's resulting hand.
type OutCard struct {
	// Card is the drawn card.
	Card Card
	// Hand is the position's hand with the drawn card.
	Hand Hand
	// Winners is the number of winners with the drawn card.
	Winners int
}

// drawCards returns the cards of the draws.
func drawCards(draws []OutCard) Cards {
	v := make(Cards, len(draws))
	for i, d := range draws {
		v[i] = d.Card
	}
	return v
}

// CalcOuts calculates the outs for the next street of the last of the runs,
// for every unseen card that can be dealt. Cards in all the runs are unseen,
// except for the active positions' pockets, and the folded pockets (with
// [WithFolded]) and discards (with [WithDiscard]).
//
// Outs are calculated when the next street deals exactly 1 board card and no
// pocket cards, such as the turn and river of [Holdem] and [Omaha]. Otherwise
// returns [ErrInvalidCalc], such as pre-flop, after the river, or for double
// board, stud, and draw types.
func CalcOuts(ctx context.Context, typ Type, runs []*Run, opts ...OutsOption) (*Outs, error) {
	c := &outsCalc{
		oddsCalc: oddsCalc{
			typ:  typ,
			runs: runs,
		},
	}
	for _, o := range opts {
		o.applyOuts(c)
	}
	return c.calc(ctx)
}

// outsCalc calculates outs.
type outsCalc struct {
	oddsCalc
}

// calc calculates the outs.
func (c *outsCalc) calc(ctx context.Context) (*Outs, error) {
	if len(c.runs) == 0 {
		return nil, ErrInvalidCalc
	}
	run := c.runs[len(c.runs)-1]
	next := c.next(run)
	if next == -1 {
		return nil, ErrInvalidCalc
	}
	streets, low, n := c.typ.Streets(), c.typ.Low(), len(run.Pockets)
	outs := &Outs{
		Type:   c.typ,
		Street: streets[next],
		Unseen: c.u(),
		Hi:     make([]Out, n),
	}
	if low {
		outs.Lo = make([]Out, n)
	}
	// status on the current board
	evs := run.eval(c.typ, c.active, false)
	status(outs.Hi, evs, false)
	if low {
		status(outs.Lo, evs, true)
	}
	// draw each unseen card
	board := make(Cards, len(run.Hi)+1)
	copy(board, run.Hi)
	draw := &Run{
		Pockets: run.Pockets,
		Hi:      board,
	}
	for _, card := range outs.Unseen {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		board[len(board)-1] = card
		evs := draw.eval(c.typ, c.active, false)
		addDraws(outs.Hi, evs, card, false)
		if low {
			addDraws(outs.Lo, evs, card, true)
		}
	}
	// a position is dead when it can't win or tie on any remaining board
	if next == len(streets)-1 {
		for i := range n {
			outs.Hi[i].Dead = dead(outs.Hi[i])
			if low {
				outs.Lo[i].Dead = dead(outs.Lo[i])
			}
		}
		return outs, nil
	}
	odds := c.oddsCalc
	odds.deep = true
	hi, lo, ok := odds.calc(ctx)
	switch {
	case ctx.Err() != nil:
		return nil, ctx.Err()
	case !ok:
		return nil, ErrInvalidCalc
	}
	for i := range n {
		outs.Hi[i].Dead = outs.Hi[i].Status != Inactive && hi.Counts[i] == 0
		if low {
			outs.Lo[i].Dead = outs.Lo[i].Status != Inactive && lo.Counts[i] == 0
		}
	}
	return outs, nil
}

// next returns the index of the run's next street, or -1 when the next street
// does not deal exactly 1 board card and no pocket cards, or the run's board
// and active pockets don't match a street.
func (c *outsCalc) next(run *Run) int {
	if c.typ.Double() || c.typ.Board() == 0 || len(run.Pockets) == 0 || len(run.Lo) != 0 {
		return -1
	}
	streets, current := c.typ.Streets(), -1
	for i, board, pocket := 0, 0, 0; i < len(streets); i++ {
		board, pocket = board+streets[i].Board, pocket+streets[i].Pocket
		if board != len(run.Hi) {
			continue
		}
		dealt := true
		for j, p := range run.Pockets {
			if c.active == nil || c.active[j] {
				dealt = dealt && len(p) == pocket
			}
		}
		if dealt {
			current = i
		}
	}
	if current == -1 || len(streets) <= current+1 {
		return -1
	}
	if street := streets[current+1]; street.Board != 1 || street.Pocket != 0 {
		return -1
	}
	return current + 1
}

// status sets the status and hand of each position from the evals.
func status(outs []Out, evs []*Eval, low bool) {
	r := Order(hands(evs, low))
	for i, ev := range evs {
		switch {
		case ev == nil:
			outs[i].Status = Inactive
			continue
		case !slices.Contains(r.Winners(), i):
			outs[i].Status = Behind
		case r.Pivot == 1:
			outs[i].Status = Ahead
		default:
			outs[i].Status = Tied
		}
		outs[i].Hand = hand(ev, low)
	}
}

// addDraws adds the drawn card to each position's wins, ties, or losses.
func addDraws(outs []Out, evs []*Eval, card Card, low bool) {
	r := Order(hands(evs, low))
	for i, ev := range evs {
		if ev == nil {
			continue
		}
		d := OutCard{
			Card:    card,
			Hand:    hand(ev, low),
			Winners: r.Pivot,
		}
		switch {
		case !slices.Contains(r.Winners(), i):
			outs[i].Losses = append(outs[i].Losses, d)
		case r.Pivot == 1:
			outs[i].Wins = append(outs[i].Wins, d)
		default:
			outs[i].Ties = append(outs[i].Ties, d)
		}
	}
}

// hand returns the eval's Hi or Lo hand.
func hand(ev *Eval, low bool) Hand {
	if low {
		return ev.Lo
	}
	return ev.Hi
}

// dead returns true when an active position has no wins or ties.
func dead(out Out) bool {
	return out.Status != Inactive && len(out.Wins) == 0 && len(out.Ties) == 0
}

// Outs calculates the outs for the dealer's current run (see [CalcOuts]).
func (d *Dealer) Outs(ctx context.Context, opts ...OutsOption) (*Outs, error) {
	if d.r < 0 || d.n <= d.r {
		return nil, ErrInvalidState
	}
	return CalcOuts(ctx, d.desc.Type, d.runs[:d.r+1], append([]OutsOption{WithActive(d.active)}, opts...)...)
}

// Outs calculates the outs for the pockets and board (see [CalcOuts]).
func (typ Type) Outs(ctx context.Context, pockets []Cards, board Cards, opts ...OutsOption) (*Outs, error) {
	return CalcOuts(ctx, typ, []*Run{{Pockets: pockets, Hi: board}}, opts...)
}

// out returns the position's Hi or Lo out.
func (outs *Outs) out(position int, low bool) (Out, bool) {
	v := outs.Hi
	if low {
		v = outs.Lo
	}
	if position < 0 || len(v) <= position {
		return Out{}, false
	}
	return v[position], true
}

// Summary summarizes the position's outs (or fades, when ahead) for a
// spectator, such as "7 outs: any ♥ except K♥ 7♥", "2 fades: 7♣ 7♦; ties:
// 4♠", or "drawing dead". Cards are described with [DescribeCards].
func (outs *Outs) Summary(position int, low bool) string {
	out, ok := outs.out(position, low)
	switch {
	case !ok, out.Status == Inactive:
		return "inactive"
	case out.Dead:
		return "drawing dead"
	}
	s := outs.count(out.Wins, "out")
	if out.Status == Ahead {
		s = outs.count(out.Losses, "fade")
	}
	if len(out.Ties) != 0 {
		s += "; ties: " + DescribeCards(drawCards(out.Ties), outs.Unseen)
	}
	return s
}

// Describe describes the position's wins by resulting hand, such as "Flush:
// any ♥ except K♥ 7♥ (7); Three of a Kind: 7♣ 7♦ (2)", followed by its ties.
// Cards are described with [DescribeCards].
func (outs *Outs) Describe(position int, low bool) string {
	out, ok := outs.out(position, low)
	switch {
	case !ok, out.Status == Inactive:
		return "inactive"
	case out.Dead:
		return "drawing dead"
	}
	var v []string
	for _, g := range Groups(out.Wins) {
		v = append(v, fmt.Sprintf("%s: %s (%d)", g.Name, DescribeCards(g.Cards, outs.Unseen), len(g.Cards)))
	}
	if len(v) == 0 {
		v = append(v, "no outs")
	}
	if len(out.Ties) != 0 {
		v = append(v, "ties: "+DescribeCards(drawCards(out.Ties), outs.Unseen))
	}
	return strings.Join(v, "; ")
}

// count describes the number of draws and their cards, such as "7 outs: any ♥
// except K♥ 7♥", or "no outs".
func (outs *Outs) count(draws []OutCard, noun string) string {
	switch len(draws) {
	case 0:
		return "no " + noun + "s"
	case 1:
		return "1 " + noun + ": " + DescribeCards(drawCards(draws), outs.Unseen)
	}
	return strconv.Itoa(len(draws)) + " " + noun + "s: " + DescribeCards(drawCards(draws), outs.Unseen)
}

// Group is a group of draws making hands with the same short description
// (such as "Flush", or "Eight-low").
type Group struct {
	// Name is the hands' short description.
	Name string
	// Rank is the best rank of the hands.
	Rank EvalRank
	// Cards are the drawn cards.
	Cards Cards
}

// Groups groups the draws by the short description of the resulting hand
// (such as "Flush", or "Eight-low"), best first.
func Groups(draws []OutCard) []Group {
	var groups []Group
	index := make(map[string]int)
	for _, d := range draws {
		name := fmt.Sprintf("%e", d.Hand)
		i, ok := index[name]
		if !ok {
			i, index[name] = len(groups), len(groups)
			groups = append(groups, Group{Name: name, Rank: d.Hand.Rank})
		}
		groups[i].Rank = min(groups[i].Rank, d.Hand.Rank)
		groups[i].Cards = append(groups[i].Cards, d.Card)
	}
	slices.SortStableFunc(groups, func(a, b Group) int {
		return cmp.Compare(a.Rank, b.Rank)
	})
	return groups
}

// DescribeCards describes the cards relative to the unseen cards, such as
// "any ♥ except K♥ 7♥, 7♣ 7♦".
//
// For each suit, and then each rank of the remaining cards, the shortest of
// listing the cards ("8♥ 6♥ 5♥"), "any ♥" (when every unseen ♥ is in the
// cards), or "any ♥ except K♥ 7♥" (listing the unseen ♥ that aren't) is used,
// preferring the list when they are the same length. Remaining cards are
// listed, ordered by rank and suit.
func DescribeCards(cards, unseen Cards) string {
	in, described := make(map[Card]bool, len(cards)), make(map[Card]bool, len(cards))
	for _, c := range cards {
		in[c] = true
	}
	var v []string
	// describe matches cards (not yet described) having the suit or rank,
	// returning true when described as "any" (and marking them described)
	describe := func(name string, match func(Card) bool) {
		var have, missing Cards
		for _, c := range cards {
			if match(c) && !described[c] {
				have = append(have, c)
			}
		}
		for _, c := range unseen {
			if match(c) && !in[c] {
				missing = append(missing, c)
			}
		}
		if len(have) == 0 {
			return
		}
		// the cards must all be unseen for "any" to describe them
		for _, c := range have {
			if !slices.Contains(unseen, c) {
				return
			}
		}
		phrase := "any " + name
		if len(missing) != 0 {
			phrase += " except " + listCards(missing)
		}
		if utf8.RuneCountInString(listCards(have)) <= utf8.RuneCountInString(phrase) {
			return
		}
		v = append(v, phrase)
		for _, c := range have {
			described[c] = true
		}
	}
	for _, suit := range []Suit{Spade, Heart, Diamond, Club} {
		describe(string(suit.UnicodeBlack()), func(c Card) bool {
			return c.Suit() == suit
		})
	}
	for rank := Ace; rank != InvalidRank; rank-- {
		describe(string(rank.Byte()), func(c Card) bool {
			return c.Rank() == rank
		})
	}
	var rest Cards
	for _, c := range cards {
		if !described[c] {
			rest = append(rest, c)
		}
	}
	if len(rest) != 0 {
		v = append(v, listCards(rest))
	}
	return strings.Join(v, ", ")
}

// listCards lists the cards, ordered by rank and suit.
func listCards(cards Cards) string {
	v := slices.Clone(cards)
	slices.SortFunc(v, func(a, b Card) int {
		return cmp.Or(cmp.Compare(b.Rank(), a.Rank()), cmp.Compare(a.SuitIndex(), b.SuitIndex()))
	})
	s := make([]string, len(v))
	for i, c := range v {
		s[i] = string(c.RankByte()) + string(c.Suit().UnicodeBlack())
	}
	return strings.Join(s, " ")
}
