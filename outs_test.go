package cardrank

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"testing"
)

func TestOuts(t *testing.T) {
	tests := []struct {
		typ     Type
		board   string
		pockets []string
		street  string
		status  []Status
		summary []string
		desc    []string
	}{
		{
			Holdem, "Ks 7d 2h 9h", []string{"Kd 7s", "Ah Qh", "3c 4c"}, "River",
			[]Status{Ahead, Behind, Behind},
			[]string{
				"7 fades: any ♥ except K♥ 7♥",
				"7 outs: any ♥ except K♥ 7♥",
				"drawing dead",
			},
			[]string{
				"Full House: K♥ K♣ 7♥ 7♣ (4); Two Pair: any ♠, any ♦, any ♣ except K♣ 7♣ (31)",
				"Flush: any ♥ except K♥ 7♥ (7)",
				"drawing dead",
			},
		},
		{
			// no outs on the turn, but runner-runner on the river
			Holdem, "Ks 7d 2h", []string{"Kd 7s", "Ah Qh", "3c 4c"}, "Turn",
			[]Status{Ahead, Behind, Behind},
			[]string{"no fades", "no outs", "no outs"},
			[]string{
				"Full House: K♥ K♣ 7♥ 7♣ (4); Two Pair: any ♠, any ♥ except K♥ 7♥, any ♦, any ♣ except K♣ 7♣ (39)",
				"no outs",
				"no outs",
			},
		},
		{
			Holdem, "As Ks 4d", []string{"Qs Js", "Ah Ad"}, "Turn",
			[]Status{Behind, Ahead},
			[]string{
				"11 outs: any ♠ except 4♠, any T",
				"11 fades: any ♠ except 4♠, any T",
			},
			[]string{
				"Straight Flush: T♠ (1); Flush: any ♠ except T♠ 4♠ (7); Straight: T♥ T♦ T♣ (3)",
				"Four of a Kind: A♣ (1); Full House: any K, any 4 (6); Three of a Kind: any ♥ except K♥ T♥ 4♥, any ♦ except K♦ T♦, any ♣ except A♣ K♣ T♣ 4♣ (27)",
			},
		},
		{
			// chopped on the board, with outs to win
			Holdem, "Ah Kd Qc Js", []string{"Tc 3d", "Th 4s"}, "River",
			[]Status{Tied, Tied},
			[]string{"no outs; ties: any ♠, any ♥, any ♦, any ♣", "no outs; ties: any ♠, any ♥, any ♦, any ♣"},
			[]string{"no outs; ties: any ♠, any ♥, any ♦, any ♣", "no outs; ties: any ♠, any ♥, any ♦, any ♣"},
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			pockets := make([]Cards, len(test.pockets))
			for j, p := range test.pockets {
				pockets[j] = Must(p)
			}
			outs, err := test.typ.Outs(context.Background(), pockets, Must(test.board))
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			logOuts(t, outs, pockets, Must(test.board))
			if outs.Street.Name != test.street {
				t.Errorf("expected street %q, got: %q", test.street, outs.Street.Name)
			}
			if outs.Lo != nil {
				t.Errorf("expected no Lo outs")
			}
			for j := range pockets {
				if s := outs.Hi[j].Status; s != test.status[j] {
					t.Errorf("position %d expected status %s, got: %s", j, test.status[j], s)
				}
				if s := outs.Summary(j, false); s != test.summary[j] {
					t.Errorf("position %d expected summary %q, got: %q", j, test.summary[j], s)
				}
				if s := outs.Describe(j, false); s != test.desc[j] {
					t.Errorf("position %d expected description %q, got: %q", j, test.desc[j], s)
				}
				// every unseen card is a win, tie, or loss
				out := outs.Hi[j]
				if n := len(out.Wins) + len(out.Ties) + len(out.Losses); n != len(outs.Unseen) {
					t.Errorf("position %d expected %d draws, got: %d", j, len(outs.Unseen), n)
				}
			}
		})
	}
}

func TestOutsLo(t *testing.T) {
	pockets := []Cards{Must("Kd 7s 3c 4c"), Must("Ah Qh 2s 5d")}
	outs, err := OmahaHiLo.Outs(context.Background(), pockets, Must("Ks 7d 2h 9h"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	logOuts(t, outs, pockets, Must("Ks 7d 2h 9h"))
	if len(outs.Lo) != 2 {
		t.Fatalf("expected Lo outs")
	}
	for i, exp := range []string{"14 outs: any A, any 8, any 6, any 5", "6 outs: any 4, any 3"} {
		if s := outs.Summary(i, true); s != exp {
			t.Errorf("position %d expected %q, got: %q", i, exp, s)
		}
		// no qualifying low on the board
		if s := outs.Lo[i].Status; s != Behind {
			t.Errorf("position %d expected behind, got: %s", i, s)
		}
	}
	if s, exp := outs.Describe(0, true), "Seven-low: any A, any 6, any 5 (10); Eight-low: any 8 (4)"; s != exp {
		t.Errorf("expected %q, got: %q", exp, s)
	}
}

// TestOutsOdds verifies the outs on the turn agree with the odds, where each
// river card is a single outcome.
func TestOutsOdds(t *testing.T) {
	r := rand.New(rand.NewSource(0x0a75))
	for _, typ := range []Type{Holdem, Short, Omaha, OmahaHiLo, OmahaFive, Split, Courchevel} {
		t.Run(typ.Name(), func(t *testing.T) {
			board := 0
			for _, street := range typ.Streets()[:len(typ.Streets())-1] {
				board += street.Board
			}
			for n := range 50 {
				players := 2 + r.Intn(4)
				d := typ.Deck()
				d.Shuffle(r)
				pockets := make([]Cards, players)
				for i := range pockets {
					pockets[i] = d.Draw(typ.Pocket())
				}
				b := d.Draw(board)
				outs, err := typ.Outs(context.Background(), pockets, b)
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				odds, err := typ.Odds(context.Background(), pockets, b)
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				// log the first few deals
				if n < 3 {
					logOuts(t, outs, pockets, b)
					for i := range players {
						t.Logf("  %d: odds %*s", i, i, odds.Hi)
					}
				}
				for i := range players {
					out := outs.Hi[i]
					if n, exp := len(out.Wins)+len(out.Ties), odds.Hi.Counts[i]; n != exp {
						t.Errorf("%v %v position %d: expected %d outs, got: %d", pockets, b, i, exp, n)
					}
					if out.Dead != (odds.Hi.Counts[i] == 0) {
						t.Errorf("%v %v position %d: expected dead %t", pockets, b, i, odds.Hi.Counts[i] == 0)
					}
					if typ.Low() {
						if n, exp := len(outs.Lo[i].Wins)+len(outs.Lo[i].Ties), odds.Lo.Counts[i]; n != exp {
							t.Errorf("%v %v position %d: expected %d Lo outs, got: %d", pockets, b, i, exp, n)
						}
					}
				}
			}
		})
	}
}

func TestOutsDealer(t *testing.T) {
	d := Holdem.Dealer(rand.New(rand.NewSource(0x0de)), 4)
	for range d.Deal() {
		if d.Street().ID != 't' {
			continue
		}
		if err := d.Fold(2); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		outs, err := d.Outs(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		run := d.runs[d.r]
		logOuts(t, outs, run.Pockets, run.Hi)
		if s := outs.Hi[2].Status; s != Inactive {
			t.Errorf("expected folded position to be inactive, got: %s", s)
		}
		if outs.Hi[2].Dead || len(outs.Hi[2].Wins) != 0 {
			t.Errorf("expected folded position to have no outs")
		}
		// the folded pocket is unseen, unless the folded cards are known
		folded, err := d.Outs(context.Background(), WithFolded(true))
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if n, exp := len(folded.Unseen), len(outs.Unseen)-2; n != exp {
			t.Errorf("expected %d unseen with folded cards known, got: %d", exp, n)
		}
		for _, c := range d.runs[0].Pockets[2] {
			if !slices.Contains(outs.Unseen, c) || slices.Contains(folded.Unseen, c) {
				t.Errorf("expected folded %s to be unseen only without WithFolded", c)
			}
		}
		return
	}
	t.Fatal("expected to deal the turn")
}

func TestOutsInvalid(t *testing.T) {
	ctx := context.Background()
	holdem := []Cards{Must("Ah Kh"), Must("2c 2d")}
	tests := []struct {
		typ     Type
		pockets []Cards
		board   string
	}{
		{Holdem, holdem, ""},               // pre-flop
		{Holdem, holdem, "3s 4s 5s 6s 7s"}, // river dealt
		{Holdem, holdem, "3s 4s"},          // incomplete flop
		{Holdem, []Cards{Must("Ah"), Must("2c 2d")}, "3s 4s 5s"},
		{Holdem, nil, "3s 4s 5s"},
		{Double, holdem, "3s 4s 5s"},
		{Stud, []Cards{Must("Ah Kh Qh Jh"), Must("2c 2d 3c 3d")}, ""},
		{Fusion, []Cards{Must("Ah Kh 9c"), Must("2c 2d 9d")}, "3s 4s 5s"},
	}
	for i, test := range tests {
		if _, err := test.typ.Outs(ctx, test.pockets, Must(test.board)); !errors.Is(err, ErrInvalidCalc) {
			t.Errorf("test %d expected ErrInvalidCalc, got: %v", i, err)
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := Holdem.Outs(canceled, holdem, Must("3s 4s 5s")); !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestDescribeCards(t *testing.T) {
	unseen := Must("2s 3s 4s 5s 6s 7s 8s 9s Ts Js 2h 7h 7c 7d Kh Kc")
	tests := []struct {
		cards string
		exp   string
	}{
		{"", ""},
		{"4s", "4♠"},
		// listing is shorter than "any ♠ except ..."
		{"4s 5s 6s 7s 8s", "8♠ 7♠ 6♠ 5♠ 4♠"},
		{"2s 3s 4s 5s 6s 7s 8s 9s Ts Js", "any ♠"},
		{"3s 4s 5s 6s 7s 8s 9s Ts Js", "any ♠ except 2♠"},
		{"4s 5s 6s 7s 8s 9s Ts Js", "any ♠ except 3♠ 2♠"},
		// ranks, after suits
		{"7s 7h 7c 7d", "any 7"},
		{"7h 7c 7d", "7♥ 7♦ 7♣"},
		{"Kh Kc", "K♥ K♣"},
		{"2s 3s 4s 5s 6s 7s 8s 9s Ts Js 7h 7c 7d", "any ♠, any 7"},
		// cards that aren't unseen are listed
		{"As Ad", "A♠ A♦"},
	}
	t.Logf("unseen: %b", unseen)
	for i, test := range tests {
		s := DescribeCards(Must(test.cards), unseen)
		t.Logf("%2d: %-30s => %s", i, fmt.Sprintf("%b", Must(test.cards)), s)
		if s != test.exp {
			t.Errorf("test %d %q expected %q, got: %q", i, test.cards, test.exp, s)
		}
	}
}

// logOuts logs the outs for the pockets and board.
func logOuts(t *testing.T, outs *Outs, pockets []Cards, board Cards) {
	t.Helper()
	t.Logf("%s: next %s (%d unseen)", outs.Type, outs.Street.Name, len(outs.Unseen))
	t.Logf("  Board: %b", board)
	for i, pocket := range pockets {
		hi := outs.Hi[i]
		if hi.Status == Inactive {
			t.Logf("  %d: %b inactive", i, pocket)
			continue
		}
		t.Logf("  %d: %b %s, %s", i, pocket, hi.Status, hi.Hand)
		t.Logf("       Summary:  %s", outs.Summary(i, false))
		t.Logf("       Describe: %s", outs.Describe(i, false))
		if outs.Lo != nil {
			t.Logf("       Lo: %s, %s", outs.Lo[i].Status, outs.Lo[i].Hand)
			t.Logf("       Summary:  %s", outs.Summary(i, true))
			t.Logf("       Describe: %s", outs.Describe(i, true))
		}
	}
}
