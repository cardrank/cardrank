package cardrank_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/cardrank/cardrank"
)

func ExampleType_Eval() {
	pocket, board := cardrank.Must("2s 5d"), cardrank.Must("Ks As 7h")
	ev := cardrank.Holdem.Eval(pocket, board)
	fmt.Printf("%s\n", ev)
	// Output:
	// Ace-high, kickers King, Seven, Five, Two [As Ks 7h 5d 2s]
}

func ExampleFromRune() {
	c := cardrank.FromRune('🂡')
	fmt.Printf("%b\n", c)
	// Output:
	// A♠
}

func ExampleFromString() {
	c := cardrank.FromString("Ah")
	fmt.Printf("%N of %L (%b)\n", c, c, c)
	// Output:
	// Ace of Hearts (A♥)
}

func ExampleMust() {
	v := cardrank.Must("Ah K♠ 🃍 J♤ 10h")
	fmt.Printf("%b", v)
	// Output:
	// [A♥ K♠ Q♦ J♠ T♥]
}

func ExampleCard_unmarshal() {
	var v []cardrank.Card
	if err := json.Unmarshal([]byte(`["3s", "4c", "5c", "Ah", "2d"]`), &v); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", v)
	// Output:
	// [3s 4c 5c Ah 2d]
}

func ExampleDeck_Draw() {
	d := cardrank.NewDeck()
	// note: use a real random source
	r := rand.New(rand.NewSource(52))
	d.Shuffle(r, 1)
	v := d.Draw(7)
	fmt.Printf("%b\n", v)
	// Output:
	// [9♣ 6♥ Q♠ 3♠ J♠ 9♥ K♣]
}

func ExampleHoldem_Eval() {
	d := cardrank.NewDeck()
	// note: use a real random source
	r := rand.New(rand.NewSource(26076))
	d.Shuffle(r, 1)
	ev := cardrank.Holdem.Eval(d.Draw(2), d.Draw(5))
	fmt.Printf("%b\n", ev)
	fmt.Printf("%S\n", ev)
	// Output:
	// Straight Flush, Five-high, Steel Wheel [5♣ 4♣ 3♣ 2♣ A♣]
	// Straight Flush, Five-high
}

func ExampleSpanish_Eval() {
	d := cardrank.DeckSpanish.New()
	// note: use a real random source
	r := rand.New(rand.NewSource(2231))
	d.Shuffle(r, 1)
	ev := cardrank.Spanish.Eval(d.Draw(2), d.Draw(5))
	fmt.Printf("%b\n", ev)
	// Output:
	// Straight Flush, Jack-high, Bronze Fist [J♦ T♦ 9♦ 8♦ A♦]
}

func Example() {
	for i, game := range []struct {
		seed     int64
		players  int
		change   byte
		runs     int
		inactive []int
		names    []string
	}{
		{566, 2, 't', 3, nil, []string{"Alice", "Bob"}},
		{1039, 5, 'f', 2, []int{0, 3, 4}, []string{"Alice", "Bob", "Carl", "Dave", "Elizabeth"}},
		{2087, 6, 't', 2, []int{0, 5}, []string{"Alice", "Bob", "Carl", "Dave", "Elizabeth", "Frank"}},
		{4022, 6, 'p', 2, []int{0, 1, 4}, []string{"Alice", "Bob", "Carl", "Dave", "Elizabeth", "Fenny"}},
	} {
		// note: use a real random source
		r := rand.New(rand.NewSource(game.seed))
		fmt.Printf("------ FusionHiLo %d ------\n", i+1)
		// setup dealer and display
		d := cardrank.FusionHiLo.Dealer(r, 1, game.players)
		// display deck
		deck := d.Deck.All()
		fmt.Println("Deck:")
		for i := 0; i < len(deck); i += 8 {
			n := min(i+8, len(deck))
			fmt.Printf("  %v\n", deck[i:n])
		}
		last := -1
		for d.Next() {
			i, run := d.Run()
			if last != i {
				fmt.Printf("Run %d:\n", i)
			}
			last = i
			fmt.Printf("  %s\n", d)
			// display pockets
			if d.HasPocket() {
				for i := range game.players {
					fmt.Printf("    %d %s: %v\n", i, game.names[i], run.Pockets[i])
				}
			}
			// display discarded cards
			if v := d.Discarded(); len(v) != 0 {
				fmt.Printf("    Discard: %v\n", v)
			}
			// display board
			if d.HasBoard() {
				fmt.Printf("    Board: %v\n", run.Hi)
				if d.Double {
					fmt.Printf("           %v\n", run.Lo)
				}
			}
			// change runs, deactivate positions
			if d.Id() == game.change && i == 0 {
				if success := d.ChangeRuns(game.runs); !success {
					panic("unable to change runs")
				}
				// deactivate
				if success := d.Deactivate(game.inactive...); !success {
					panic("unable to deactivate positions")
				}
			}
		}
		fmt.Println("Showdown:")
		for d.NextResult() {
			n, res := d.Result()
			fmt.Printf("  Run %d:\n", n)
			for i := range game.players {
				if d.Active[i] {
					hi := res.Evals[i].Desc(false)
					fmt.Printf("    %d: %v %v %s\n", i, hi.Best, hi.Unused, hi)
					if d.Low || d.Double {
						lo := res.Evals[i].Desc(true)
						fmt.Printf("       %v %v %s\n", lo.Best, lo.Unused, lo)
					}
				} else {
					fmt.Printf("    %d: inactive\n", i)
				}
			}
			hi, lo := res.Win(game.names...)
			fmt.Printf("    Result: %S\n", hi)
			if lo != nil {
				fmt.Printf("            %S\n", lo)
			}
		}
	}
	// Output:
	// ------ FusionHiLo 1 ------
	// Deck:
	//   [4h Qs 5c 4c 5d 8d 8c As]
	//   [Ks 6h 7s 9s 3h Ac Js 9h]
	//   [4s 7d 2h 8s 2s Ad Ts Qh]
	//   [Qc 5h 6s 9d 9c 6c Kd 2d]
	//   [3s Ah Kh 5s Jd Jc 2c Td]
	//   [3c Jh 8h 4d Th 7c 7h 3d]
	//   [6d Tc Kc Qd]
	// Run 0:
	//   p: Pre-Flop (p: 2)
	//     0 Alice: [4h 5c]
	//     1 Bob: [Qs 4c]
	//   f: Flop (p: 1, d: 1, b: 3)
	//     0 Alice: [4h 5c 5d]
	//     1 Bob: [Qs 4c 8d]
	//     Discard: [8c]
	//     Board: [As Ks 6h]
	//   t: Turn (p: 1, d: 1, b: 1)
	//     0 Alice: [4h 5c 5d 7s]
	//     1 Bob: [Qs 4c 8d 9s]
	//     Discard: [8c 3h]
	//     Board: [As Ks 6h Ac]
	//   r: River (d: 1, b: 1)
	//     Discard: [8c 3h Js]
	//     Board: [As Ks 6h Ac 9h]
	// Run 1:
	//   r: River (d: 1, b: 1)
	//     Discard: [4s]
	//     Board: [As Ks 6h Ac 7d]
	// Run 2:
	//   r: River (d: 1, b: 1)
	//     Discard: [2h]
	//     Board: [As Ks 6h Ac 8s]
	// Showdown:
	//   Run 0:
	//     0: [Ac As 5c 5d Ks] [9h 7s 6h 4h] Two Pair, Aces over Fives, kicker King
	//        [] [] None
	//     1: [Ac As 9h 9s Qs] [Ks 8d 6h 4c] Two Pair, Aces over Nines, kicker Queen
	//        [] [] None
	//     Result: Bob scoops with Two Pair, Aces over Nines, kicker Queen
	//   Run 1:
	//     0: [Ac As 7d 7s 5c] [Ks 6h 5d 4h] Two Pair, Aces over Sevens, kicker Five
	//        [7d 6h 5c 4h As] [Ac Ks 7s 5d] Seven, Six, Five, Four, Ace-low
	//     1: [Ac As Ks Qs 9s] [8d 7d 6h 4c] Pair, Aces, kickers King, Queen, Nine
	//        [8d 7d 6h 4c As] [Ac Ks Qs 9s] Eight, Seven, Six, Four, Ace-low
	//     Result: Alice wins with Two Pair, Aces over Sevens, kicker Five
	//             Alice wins with Seven, Six, Five, Four, Ace-low
	//   Run 2:
	//     0: [Ac As 5c 5d Ks] [8s 7s 6h 4h] Two Pair, Aces over Fives, kicker King
	//        [8s 6h 5c 4h As] [Ac Ks 7s 5d] Eight, Six, Five, Four, Ace-low
	//     1: [As Ks Qs 9s 8s] [Ac 8d 6h 4c] Flush, Ace-high, kickers King, Queen, Nine, Eight
	//        [] [] None
	//     Result: Bob wins with Flush, Ace-high, kickers King, Queen, Nine, Eight
	//             Alice wins with Eight, Six, Five, Four, Ace-low
	// ------ FusionHiLo 2 ------
	// Deck:
	//   [2h 5s Ac Ts Kd 5h 6d Th]
	//   [2s 6s 7c 4h 8c 9h Ah 8s]
	//   [Kc 9d 5c 5d As 4d 3h 2c]
	//   [7s 8h 4c 7d 8d Qs 3c 7h]
	//   [Jc Jh 6c 3s Qd 9c 4s 3d]
	//   [Ks Ad Qc Td Tc Qh Js 6h]
	//   [2d 9s Jd Kh]
	// Run 0:
	//   p: Pre-Flop (p: 2)
	//     0 Alice: [2h 5h]
	//     1 Bob: [5s 6d]
	//     2 Carl: [Ac Th]
	//     3 Dave: [Ts 2s]
	//     4 Elizabeth: [Kd 6s]
	//   f: Flop (p: 1, d: 1, b: 3)
	//     0 Alice: [2h 5h 7c]
	//     1 Bob: [5s 6d 4h]
	//     2 Carl: [Ac Th 8c]
	//     3 Dave: [Ts 2s 9h]
	//     4 Elizabeth: [Kd 6s Ah]
	//     Discard: [8s]
	//     Board: [Kc 9d 5c]
	//   t: Turn (p: 1, d: 1, b: 1)
	//     0 Alice: [2h 5h 7c 5d]
	//     1 Bob: [5s 6d 4h As]
	//     2 Carl: [Ac Th 8c 4d]
	//     3 Dave: [Ts 2s 9h 3h]
	//     4 Elizabeth: [Kd 6s Ah 2c]
	//     Discard: [8s 7s]
	//     Board: [Kc 9d 5c 8h]
	//   r: River (d: 1, b: 1)
	//     Discard: [8s 7s 4c]
	//     Board: [Kc 9d 5c 8h 7d]
	// Run 1:
	//   t: Turn (p: 1, d: 1, b: 1)
	//     0 Alice: [2h 5h 7c 8d]
	//     1 Bob: [5s 6d 4h Qs]
	//     2 Carl: [Ac Th 8c 3c]
	//     3 Dave: [Ts 2s 9h 7h]
	//     4 Elizabeth: [Kd 6s Ah Jc]
	//     Discard: [Jh]
	//     Board: [Kc 9d 5c 6c]
	//   r: River (d: 1, b: 1)
	//     Discard: [Jh 3s]
	//     Board: [Kc 9d 5c 6c Qd]
	// Showdown:
	//   Run 0:
	//     0: inactive
	//     1: [9d 8h 7d 6d 5s] [As Kc 5c 4h] Straight, Nine-high
	//        [8h 7d 5c 4h As] [Kc 9d 6d 5s] Eight, Seven, Five, Four, Ace-low
	//     2: [8c 8h Ac Kc 9d] [Th 7d 5c 4d] Pair, Eights, kickers Ace, King, Nine
	//        [8h 7d 5c 4d Ac] [Kc Th 9d 8c] Eight, Seven, Five, Four, Ace-low
	//     3: inactive
	//     4: inactive
	//     Result: Bob wins with Straight, Nine-high
	//             Bob, Carl split with Eight, Seven, Five, Four, Ace-low
	//   Run 1:
	//     0: inactive
	//     1: [Qd Qs 6c 6d Kc] [9d 5c 5s 4h] Two Pair, Queens over Sixes, kicker King
	//        [] [] None
	//     2: [Ac Kc 8c 6c 5c] [Qd Th 9d 3c] Flush, Ace-high, kickers King, Eight, Six, Five
	//        [] [] None
	//     3: inactive
	//     4: inactive
	//     Result: Carl scoops with Flush, Ace-high, kickers King, Eight, Six, Five
	// ------ FusionHiLo 3 ------
	// Deck:
	//   [8h 5d 5c 3h Jc 6h Kd Td]
	//   [6s As 7c 6c 2c Jd 9h 8c]
	//   [7s 5s 8d Tc 3s Kc Qh Qd]
	//   [7d Ks Jh 4s 9s 4h Th Qc]
	//   [Ah 2d Ts 7h 4c Qs Kh 6d]
	//   [9d 2s Js 3d 5h 2h Ac Ad]
	//   [3c 8s 4d 9c]
	// Run 0:
	//   p: Pre-Flop (p: 2)
	//     0 Alice: [8h Kd]
	//     1 Bob: [5d Td]
	//     2 Carl: [5c 6s]
	//     3 Dave: [3h As]
	//     4 Elizabeth: [Jc 7c]
	//     5 Frank: [6h 6c]
	//   f: Flop (p: 1, d: 1, b: 3)
	//     0 Alice: [8h Kd 2c]
	//     1 Bob: [5d Td Jd]
	//     2 Carl: [5c 6s 9h]
	//     3 Dave: [3h As 8c]
	//     4 Elizabeth: [Jc 7c 7s]
	//     5 Frank: [6h 6c 5s]
	//     Discard: [8d]
	//     Board: [Tc 3s Kc]
	//   t: Turn (p: 1, d: 1, b: 1)
	//     0 Alice: [8h Kd 2c Qh]
	//     1 Bob: [5d Td Jd Qd]
	//     2 Carl: [5c 6s 9h 7d]
	//     3 Dave: [3h As 8c Ks]
	//     4 Elizabeth: [Jc 7c 7s Jh]
	//     5 Frank: [6h 6c 5s 4s]
	//     Discard: [8d 9s]
	//     Board: [Tc 3s Kc 4h]
	//   r: River (d: 1, b: 1)
	//     Discard: [8d 9s Th]
	//     Board: [Tc 3s Kc 4h Qc]
	// Run 1:
	//   r: River (d: 1, b: 1)
	//     Discard: [Ah]
	//     Board: [Tc 3s Kc 4h 2d]
	// Showdown:
	//   Run 0:
	//     0: inactive
	//     1: [Qc Qd Tc Td Kc] [Jd 5d 4h 3s] Two Pair, Queens over Tens, kicker King
	//        [] [] None
	//     2: [Kc Qc Tc 9h 7d] [6s 5c 4h 3s] King-high, kickers Queen, Ten, Nine, Seven
	//        [] [] None
	//     3: [Kc Ks 3h 3s Qc] [As Tc 8c 4h] Two Pair, Kings over Threes, kicker Queen
	//        [] [] None
	//     4: [Kc Qc Jc Tc 7c] [Jh 7s 4h 3s] Flush, King-high, kickers Queen, Jack, Ten, Seven
	//        [] [] None
	//     5: inactive
	//     Result: Elizabeth scoops with Flush, King-high, kickers Queen, Jack, Ten, Seven
	//   Run 1:
	//     0: inactive
	//     1: [Tc Td Kc Qd 4h] [Jd 5d 3s 2d] Pair, Tens, kickers King, Queen, Four
	//        [] [] None
	//     2: [6s 5c 4h 3s 2d] [Kc Tc 9h 7d] Straight, Six-high
	//        [6s 5c 4h 3s 2d] [Kc Tc 9h 7d] Six, Five, Four, Three, Two-low
	//     3: [Kc Ks 3h 3s Tc] [As 8c 4h 2d] Two Pair, Kings over Threes, kicker Ten
	//        [8c 4h 3s 2d As] [Kc Ks Tc 3h] Eight, Four, Three, Two, Ace-low
	//     4: [Jc Jh Kc Tc 4h] [7c 7s 3s 2d] Pair, Jacks, kickers King, Ten, Four
	//        [] [] None
	//     5: inactive
	//     Result: Carl wins with Straight, Six-high
	//             Carl wins with Six, Five, Four, Three, Two-low
	// ------ FusionHiLo 4 ------
	// Deck:
	//   [Qc 4h 2c 7c Kc 5c 9d 5h]
	//   [3c Tc 9c Qd As 4s 5d Jc]
	//   [4c Ad 9s 8s Qh 3h Td 7h]
	//   [7s Ks 6d Kd 7d Jh 2d Js]
	//   [4d 6h Th Ah Ac Ts 3d 6c]
	//   [Jd 2s 2h 9h 3s 5s 8d 8c]
	//   [Qs 8h Kh 6s]
	// Run 0:
	//   p: Pre-Flop (p: 2)
	//     0 Alice: [Qc 9d]
	//     1 Bob: [4h 5h]
	//     2 Carl: [2c 3c]
	//     3 Dave: [7c Tc]
	//     4 Elizabeth: [Kc 9c]
	//     5 Fenny: [5c Qd]
	//   f: Flop (p: 1, d: 1, b: 3)
	//     0 Alice: [Qc 9d As]
	//     1 Bob: [4h 5h 4s]
	//     2 Carl: [2c 3c 5d]
	//     3 Dave: [7c Tc Jc]
	//     4 Elizabeth: [Kc 9c 4c]
	//     5 Fenny: [5c Qd Ad]
	//     Discard: [9s]
	//     Board: [8s Qh 3h]
	//   t: Turn (p: 1, d: 1, b: 1)
	//     0 Alice: [Qc 9d As Td]
	//     1 Bob: [4h 5h 4s 7h]
	//     2 Carl: [2c 3c 5d 7s]
	//     3 Dave: [7c Tc Jc Ks]
	//     4 Elizabeth: [Kc 9c 4c 6d]
	//     5 Fenny: [5c Qd Ad Kd]
	//     Discard: [9s 7d]
	//     Board: [8s Qh 3h Jh]
	//   r: River (d: 1, b: 1)
	//     Discard: [9s 7d 2d]
	//     Board: [8s Qh 3h Jh Js]
	// Run 1:
	//   f: Flop (p: 1, d: 1, b: 3)
	//     0 Alice: [Qc 9d 4d]
	//     1 Bob: [4h 5h 6h]
	//     2 Carl: [2c 3c Th]
	//     3 Dave: [7c Tc Ah]
	//     4 Elizabeth: [Kc 9c Ac]
	//     5 Fenny: [5c Qd Ts]
	//     Discard: [3d]
	//     Board: [6c Jd 2s]
	//   t: Turn (p: 1, d: 1, b: 1)
	//     0 Alice: [Qc 9d 4d 2h]
	//     1 Bob: [4h 5h 6h 9h]
	//     2 Carl: [2c 3c Th 3s]
	//     3 Dave: [7c Tc Ah 5s]
	//     4 Elizabeth: [Kc 9c Ac 8d]
	//     5 Fenny: [5c Qd Ts 8c]
	//     Discard: [3d Qs]
	//     Board: [6c Jd 2s 8h]
	//   r: River (d: 1, b: 1)
	//     Discard: [3d Qs Kh]
	//     Board: [6c Jd 2s 8h 6s]
	// Showdown:
	//   Run 0:
	//     0: inactive
	//     1: inactive
	//     2: [Jh Js 3c 3h 7s] [Qh 8s 5d 2c] Two Pair, Jacks over Threes, kicker Seven
	//        [] [] None
	//     3: [Jc Jh Js Ks Qh] [Tc 8s 7c 3h] Three of a Kind, Jacks, kickers King, Queen
	//        [] [] None
	//     4: inactive
	//     5: [Qd Qh Jh Js Ad] [Kd 8s 5c 3h] Two Pair, Queens over Jacks, kicker Ace
	//        [] [] None
	//     Result: Dave scoops with Three of a Kind, Jacks, kickers King, Queen
	//   Run 1:
	//     0: inactive
	//     1: inactive
	//     2: [6c 6s 3c 3s Jd] [Th 8h 2c 2s] Two Pair, Sixes over Threes, kicker Jack
	//        [] [] None
	//     3: [6c 6s Ah Jd Tc] [8h 7c 5s 2s] Pair, Sixes, kickers Ace, Jack, Ten
	//        [8h 6c 5s 2s Ah] [Jd Tc 7c 6s] Eight, Six, Five, Two, Ace-low
	//     4: inactive
	//     5: [8c 8h 6c 6s Qd] [Jd Ts 5c 2s] Two Pair, Eights over Sixes, kicker Queen
	//        [] [] None
	//     Result: Fenny wins with Two Pair, Eights over Sixes, kicker Queen
	//             Dave wins with Eight, Six, Five, Two, Ace-low
}

func ExampleType_holdem() {
	for i, game := range []struct {
		seed    int64
		players int
	}{
		{3, 2},
		{278062, 2},
		{1928, 6},
		{6151, 6},
		{5680, 6},
		{23965, 2},
		{13959, 2},
		{23366, 6},
		{29555, 3},
		{472600, 3},
		{107, 10},
	} {
		// note: use a real random source
		r := rand.New(rand.NewSource(game.seed))
		pockets, board := cardrank.Holdem.Deal(r, 1, game.players)
		evs := cardrank.Holdem.EvalPockets(pockets, board)
		fmt.Printf("------ Holdem %d ------\n", i+1)
		fmt.Printf("Board:    %b\n", board)
		for j := range game.players {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := range pivot {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Holdem 1 ------
	// Board:    [J♠ T♠ 2♦ 2♠ Q♥]
	// Player 1: [6♦ 7♠] Pair, Twos, kickers Queen, Jack, Ten [2♦ 2♠ Q♥ J♠ T♠] [7♠ 6♦]
	// Player 2: [8♠ 4♣] Pair, Twos, kickers Queen, Jack, Ten [2♦ 2♠ Q♥ J♠ T♠] [8♠ 4♣]
	// Result:   Players 1, 2 push with Pair, Twos, kickers Queen, Jack, Ten
	// ------ Holdem 2 ------
	// Board:    [8♠ 9♠ J♠ 9♣ T♠]
	// Player 1: [7♠ 6♦] Straight Flush, Jack-high, Bronze Fist [J♠ T♠ 9♠ 8♠ 7♠] [9♣ 6♦]
	// Player 2: [T♣ Q♠] Straight Flush, Queen-high, Silver Tongue [Q♠ J♠ T♠ 9♠ 8♠] [T♣ 9♣]
	// Result:   Player 2 wins with Straight Flush, Queen-high, Silver Tongue
	// ------ Holdem 3 ------
	// Board:    [A♠ T♣ K♠ J♣ 6♥]
	// Player 1: [T♥ 5♦] Pair, Tens, kickers Ace, King, Jack [T♣ T♥ A♠ K♠ J♣] [6♥ 5♦]
	// Player 2: [2♠ K♦] Pair, Kings, kickers Ace, Jack, Ten [K♦ K♠ A♠ J♣ T♣] [6♥ 2♠]
	// Player 3: [Q♣ Q♥] Straight, Ace-high [A♠ K♠ Q♣ J♣ T♣] [Q♥ 6♥]
	// Player 4: [J♠ 7♣] Pair, Jacks, kickers Ace, King, Ten [J♣ J♠ A♠ K♠ T♣] [7♣ 6♥]
	// Player 5: [4♥ 6♠] Pair, Sixes, kickers Ace, King, Jack [6♥ 6♠ A♠ K♠ J♣] [T♣ 4♥]
	// Player 6: [Q♠ 3♣] Straight, Ace-high [A♠ K♠ Q♠ J♣ T♣] [6♥ 3♣]
	// Result:   Players 3, 6 push with Straight, Ace-high
	// ------ Holdem 4 ------
	// Board:    [9♦ J♣ A♥ 9♥ J♠]
	// Player 1: [K♠ 8♦] Two Pair, Jacks over Nines, kicker Ace [J♣ J♠ 9♦ 9♥ A♥] [K♠ 8♦]
	// Player 2: [7♦ 9♠] Full House, Nines full of Jacks [9♦ 9♥ 9♠ J♣ J♠] [A♥ 7♦]
	// Player 3: [A♦ 8♥] Two Pair, Aces over Jacks, kicker Nine [A♦ A♥ J♣ J♠ 9♦] [9♥ 8♥]
	// Player 4: [4♥ 6♣] Two Pair, Jacks over Nines, kicker Ace [J♣ J♠ 9♦ 9♥ A♥] [6♣ 4♥]
	// Player 5: [3♥ 5♥] Two Pair, Jacks over Nines, kicker Ace [J♣ J♠ 9♦ 9♥ A♥] [5♥ 3♥]
	// Player 6: [T♣ J♦] Full House, Jacks full of Nines [J♣ J♦ J♠ 9♦ 9♥] [A♥ T♣]
	// Result:   Player 6 wins with Full House, Jacks full of Nines
	// ------ Holdem 5 ------
	// Board:    [3♠ 9♥ A♦ 6♥ Q♦]
	// Player 1: [T♦ 4♥] Ace-high, kickers Queen, Ten, Nine, Six [A♦ Q♦ T♦ 9♥ 6♥] [4♥ 3♠]
	// Player 2: [8♦ 7♦] Ace-high, kickers Queen, Nine, Eight, Seven [A♦ Q♦ 9♥ 8♦ 7♦] [6♥ 3♠]
	// Player 3: [K♠ K♥] Pair, Kings, kickers Ace, Queen, Nine [K♥ K♠ A♦ Q♦ 9♥] [6♥ 3♠]
	// Player 4: [T♣ 5♦] Ace-high, kickers Queen, Ten, Nine, Six [A♦ Q♦ T♣ 9♥ 6♥] [5♦ 3♠]
	// Player 5: [7♥ T♥] Ace-high, kickers Queen, Ten, Nine, Seven [A♦ Q♦ T♥ 9♥ 7♥] [6♥ 3♠]
	// Player 6: [8♣ 5♣] Ace-high, kickers Queen, Nine, Eight, Six [A♦ Q♦ 9♥ 8♣ 6♥] [5♣ 3♠]
	// Result:   Player 3 wins with Pair, Kings, kickers Ace, Queen, Nine
	// ------ Holdem 6 ------
	// Board:    [T♥ 6♥ 7♥ 2♥ 7♣]
	// Player 1: [6♣ K♥] Flush, King-high, kickers Ten, Seven, Six, Two [K♥ T♥ 7♥ 6♥ 2♥] [7♣ 6♣]
	// Player 2: [6♠ 5♥] Flush, Ten-high, kickers Seven, Six, Five, Two [T♥ 7♥ 6♥ 5♥ 2♥] [7♣ 6♠]
	// Result:   Player 1 wins with Flush, King-high, kickers Ten, Seven, Six, Two
	// ------ Holdem 7 ------
	// Board:    [4♦ A♥ A♣ 4♠ A♦]
	// Player 1: [T♥ 9♣] Full House, Aces full of Fours [A♣ A♦ A♥ 4♦ 4♠] [T♥ 9♣]
	// Player 2: [T♠ A♠] Four of a Kind, Aces, kicker Ten [A♣ A♦ A♥ A♠ T♠] [4♦ 4♠]
	// Result:   Player 2 wins with Four of a Kind, Aces, kicker Ten
	// ------ Holdem 8 ------
	// Board:    [Q♥ T♥ T♠ J♥ K♥]
	// Player 1: [A♥ 8♥] Straight Flush, Ace-high, Royal [A♥ K♥ Q♥ J♥ T♥] [T♠ 8♥]
	// Player 2: [9♠ 8♦] Straight, King-high [K♥ Q♥ J♥ T♥ 9♠] [T♠ 8♦]
	// Player 3: [Q♣ 4♦] Two Pair, Queens over Tens, kicker King [Q♣ Q♥ T♥ T♠ K♥] [J♥ 4♦]
	// Player 4: [2♠ Q♦] Two Pair, Queens over Tens, kicker King [Q♦ Q♥ T♥ T♠ K♥] [J♥ 2♠]
	// Player 5: [6♥ A♦] Flush, King-high, kickers Queen, Jack, Ten, Six [K♥ Q♥ J♥ T♥ 6♥] [A♦ T♠]
	// Player 6: [3♦ T♣] Three of a Kind, Tens, kickers King, Queen [T♣ T♥ T♠ K♥ Q♥] [J♥ 3♦]
	// Result:   Player 1 wins with Straight Flush, Ace-high, Royal
	// ------ Holdem 9 ------
	// Board:    [A♣ 2♣ 4♣ 5♣ 9♥]
	// Player 1: [T♣ 6♠] Flush, Ace-high, kickers Ten, Five, Four, Two [A♣ T♣ 5♣ 4♣ 2♣] [9♥ 6♠]
	// Player 2: [J♦ 3♣] Straight Flush, Five-high, Steel Wheel [5♣ 4♣ 3♣ 2♣ A♣] [J♦ 9♥]
	// Player 3: [4♥ T♠] Pair, Fours, kickers Ace, Ten, Nine [4♣ 4♥ A♣ T♠ 9♥] [5♣ 2♣]
	// Result:   Player 2 wins with Straight Flush, Five-high, Steel Wheel
	// ------ Holdem 10 ------
	// Board:    [8♣ J♣ 8♥ 7♥ 9♥]
	// Player 1: [8♦ T♥] Straight, Jack-high [J♣ T♥ 9♥ 8♣ 7♥] [8♦ 8♥]
	// Player 2: [8♠ 3♣] Three of a Kind, Eights, kickers Jack, Nine [8♣ 8♥ 8♠ J♣ 9♥] [7♥ 3♣]
	// Player 3: [6♥ K♥] Flush, King-high, kickers Nine, Eight, Seven, Six [K♥ 9♥ 8♥ 7♥ 6♥] [J♣ 8♣]
	// Result:   Player 3 wins with Flush, King-high, kickers Nine, Eight, Seven, Six
	// ------ Holdem 11 ------
	// Board:    [5♥ 3♣ J♥ 6♦ 6♣]
	// Player 1: [8♥ T♥] Pair, Sixes, kickers Jack, Ten, Eight [6♣ 6♦ J♥ T♥ 8♥] [5♥ 3♣]
	// Player 2: [4♥ Q♣] Pair, Sixes, kickers Queen, Jack, Five [6♣ 6♦ Q♣ J♥ 5♥] [4♥ 3♣]
	// Player 3: [T♣ Q♠] Pair, Sixes, kickers Queen, Jack, Ten [6♣ 6♦ Q♠ J♥ T♣] [5♥ 3♣]
	// Player 4: [3♥ 5♦] Two Pair, Sixes over Fives, kicker Jack [6♣ 6♦ 5♦ 5♥ J♥] [3♣ 3♥]
	// Player 5: [A♠ T♠] Pair, Sixes, kickers Ace, Jack, Ten [6♣ 6♦ A♠ J♥ T♠] [5♥ 3♣]
	// Player 6: [6♠ 2♠] Three of a Kind, Sixes, kickers Jack, Five [6♣ 6♦ 6♠ J♥ 5♥] [3♣ 2♠]
	// Player 7: [J♠ 5♣] Two Pair, Jacks over Sixes, kicker Five [J♥ J♠ 6♣ 6♦ 5♣] [5♥ 3♣]
	// Player 8: [8♠ 9♦] Pair, Sixes, kickers Jack, Nine, Eight [6♣ 6♦ J♥ 9♦ 8♠] [5♥ 3♣]
	// Player 9: [6♥ J♣] Full House, Sixes full of Jacks [6♣ 6♦ 6♥ J♣ J♥] [5♥ 3♣]
	// Player 10: [2♣ A♣] Pair, Sixes, kickers Ace, Jack, Five [6♣ 6♦ A♣ J♥ 5♥] [3♣ 2♣]
	// Result:   Player 9 wins with Full House, Sixes full of Jacks
}

func ExampleType_short() {
	for i, game := range []struct {
		seed    int64
		players int
	}{
		{119, 2},
		{155, 4},
		{384, 8},
		{880, 4},
		{3453, 3},
		{5662, 3},
		{65481, 2},
		{27947, 4},
	} {
		// note: use a real random source
		r := rand.New(rand.NewSource(game.seed))
		pockets, board := cardrank.Short.Deal(r, 1, game.players)
		evs := cardrank.Short.EvalPockets(pockets, board)
		fmt.Printf("------ Short %d ------\n", i+1)
		fmt.Printf("Board:    %b\n", board)
		for j := range game.players {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := range pivot {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Short 1 ------
	// Board:    [9♥ A♦ A♥ 8♣ A♣]
	// Player 1: [8♥ A♠] Four of a Kind, Aces, kicker Nine [A♣ A♦ A♥ A♠ 9♥] [8♣ 8♥]
	// Player 2: [7♥ J♦] Three of a Kind, Aces, kickers Jack, Nine [A♣ A♦ A♥ J♦ 9♥] [8♣ 7♥]
	// Result:   Player 1 wins with Four of a Kind, Aces, kicker Nine
	// ------ Short 2 ------
	// Board:    [9♣ 6♦ A♠ J♠ 6♠]
	// Player 1: [T♥ A♣] Two Pair, Aces over Sixes, kicker Jack [A♣ A♠ 6♦ 6♠ J♠] [T♥ 9♣]
	// Player 2: [6♣ 7♣] Three of a Kind, Sixes, kickers Ace, Jack [6♣ 6♦ 6♠ A♠ J♠] [9♣ 7♣]
	// Player 3: [6♥ T♠] Three of a Kind, Sixes, kickers Ace, Jack [6♦ 6♥ 6♠ A♠ J♠] [T♠ 9♣]
	// Player 4: [9♥ K♠] Two Pair, Nines over Sixes, kicker Ace [9♣ 9♥ 6♦ 6♠ A♠] [K♠ J♠]
	// Result:   Players 2, 3 push with Three of a Kind, Sixes, kickers Ace, Jack
	// ------ Short 3 ------
	// Board:    [T♥ J♣ 7♥ 9♥ K♣]
	// Player 1: [8♥ T♣] Straight, Jack-high [J♣ T♣ 9♥ 8♥ 7♥] [K♣ T♥]
	// Player 2: [T♠ Q♠] Straight, King-high [K♣ Q♠ J♣ T♥ 9♥] [T♠ 7♥]
	// Player 3: [J♠ 7♣] Two Pair, Jacks over Sevens, kicker King [J♣ J♠ 7♣ 7♥ K♣] [T♥ 9♥]
	// Player 4: [6♣ Q♦] Straight, King-high [K♣ Q♦ J♣ T♥ 9♥] [7♥ 6♣]
	// Player 5: [7♦ 6♠] Pair, Sevens, kickers King, Jack, Ten [7♦ 7♥ K♣ J♣ T♥] [9♥ 6♠]
	// Player 6: [8♠ 8♦] Straight, Jack-high [J♣ T♥ 9♥ 8♦ 7♥] [K♣ 8♠]
	// Player 7: [9♣ K♥] Two Pair, Kings over Nines, kicker Jack [K♣ K♥ 9♣ 9♥ J♣] [T♥ 7♥]
	// Player 8: [A♥ K♦] Pair, Kings, kickers Ace, Jack, Ten [K♣ K♦ A♥ J♣ T♥] [9♥ 7♥]
	// Result:   Players 2, 4 push with Straight, King-high
	// ------ Short 4 ------
	// Board:    [T♦ 9♣ 9♦ Q♦ 8♦]
	// Player 1: [J♠ 9♥] Straight, Queen-high [Q♦ J♠ T♦ 9♣ 8♦] [9♦ 9♥]
	// Player 2: [T♥ 8♠] Two Pair, Tens over Nines, kicker Queen [T♦ T♥ 9♣ 9♦ Q♦] [8♦ 8♠]
	// Player 3: [6♣ J♦] Straight Flush, Queen-high, Silver Tongue [Q♦ J♦ T♦ 9♦ 8♦] [9♣ 6♣]
	// Player 4: [A♣ A♦] Flush, Ace-high, kickers Queen, Ten, Nine, Eight [A♦ Q♦ T♦ 9♦ 8♦] [A♣ 9♣]
	// Result:   Player 3 wins with Straight Flush, Queen-high, Silver Tongue
	// ------ Short 5 ------
	// Board:    [6♠ A♣ 7♦ A♠ 6♦]
	// Player 1: [9♣ T♦] Two Pair, Aces over Sixes, kicker Ten [A♣ A♠ 6♦ 6♠ T♦] [9♣ 7♦]
	// Player 2: [T♠ K♠] Two Pair, Aces over Sixes, kicker King [A♣ A♠ 6♦ 6♠ K♠] [T♠ 7♦]
	// Player 3: [J♥ A♥] Full House, Aces full of Sixes [A♣ A♥ A♠ 6♦ 6♠] [J♥ 7♦]
	// Result:   Player 3 wins with Full House, Aces full of Sixes
	// ------ Short 6 ------
	// Board:    [A♣ 6♣ 9♣ T♦ 8♣]
	// Player 1: [6♥ 9♠] Two Pair, Nines over Sixes, kicker Ace [9♣ 9♠ 6♣ 6♥ A♣] [T♦ 8♣]
	// Player 2: [7♣ J♥] Straight Flush, Nine-high, Iron Maiden [9♣ 8♣ 7♣ 6♣ A♣] [J♥ T♦]
	// Player 3: [6♠ Q♠] Pair, Sixes, kickers Ace, Queen, Ten [6♣ 6♠ A♣ Q♠ T♦] [9♣ 8♣]
	// Result:   Player 2 wins with Straight Flush, Nine-high, Iron Maiden
	// ------ Short 7 ------
	// Board:    [K♥ K♦ K♠ K♣ J♣]
	// Player 1: [7♦ 8♦] Four of a Kind, Kings, kicker Jack [K♣ K♦ K♥ K♠ J♣] [8♦ 7♦]
	// Player 2: [T♦ 6♥] Four of a Kind, Kings, kicker Jack [K♣ K♦ K♥ K♠ J♣] [T♦ 6♥]
	// Result:   Players 1, 2 push with Four of a Kind, Kings, kicker Jack
	// ------ Short 8 ------
	// Board:    [8♦ 8♥ 8♠ Q♠ T♦]
	// Player 1: [J♦ 9♣] Straight, Queen-high [Q♠ J♦ T♦ 9♣ 8♦] [8♥ 8♠]
	// Player 2: [T♣ J♣] Full House, Eights full of Tens [8♦ 8♥ 8♠ T♣ T♦] [Q♠ J♣]
	// Player 3: [K♠ T♥] Full House, Eights full of Tens [8♦ 8♥ 8♠ T♦ T♥] [K♠ Q♠]
	// Player 4: [T♠ 7♥] Full House, Eights full of Tens [8♦ 8♥ 8♠ T♦ T♠] [Q♠ 7♥]
	// Result:   Players 2, 3, 4 push with Full House, Eights full of Tens
}

func ExampleType_royal() {
	for i, game := range []struct {
		seed    int64
		players int
	}{
		{119, 2},
		{155, 3},
		{384, 4},
		{880, 5},
		{3453, 2},
		{5662, 3},
		{65481, 4},
		{27947, 5},
	} {
		// note: use a real random source
		r := rand.New(rand.NewSource(game.seed))
		pockets, board := cardrank.Royal.Deal(r, 1, game.players)
		evs := cardrank.Royal.EvalPockets(pockets, board)
		fmt.Printf("------ Royal %d ------\n", i+1)
		fmt.Printf("Board:    %b\n", board)
		for j := range game.players {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := range pivot {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Royal 1 ------
	// Board:    [K♦ A♦ T♥ T♣ J♠]
	// Player 1: [A♠ T♠] Full House, Tens full of Aces [T♣ T♥ T♠ A♦ A♠] [K♦ J♠]
	// Player 2: [A♥ K♠] Two Pair, Aces over Kings, kicker Jack [A♦ A♥ K♦ K♠ J♠] [T♣ T♥]
	// Result:   Player 1 wins with Full House, Tens full of Aces
	// ------ Royal 2 ------
	// Board:    [A♣ K♠ J♦ Q♣ J♣]
	// Player 1: [A♠ Q♠] Two Pair, Aces over Queens, kicker King [A♣ A♠ Q♣ Q♠ K♠] [J♣ J♦]
	// Player 2: [T♠ J♥] Straight, Ace-high [A♣ K♠ Q♣ J♣ T♠] [J♦ J♥]
	// Player 3: [K♣ T♥] Straight, Ace-high [A♣ K♣ Q♣ J♣ T♥] [K♠ J♦]
	// Result:   Players 2, 3 push with Straight, Ace-high
	// ------ Royal 3 ------
	// Board:    [K♠ T♦ T♣ Q♦ A♥]
	// Player 1: [T♠ T♥] Four of a Kind, Tens, kicker Ace [T♣ T♦ T♥ T♠ A♥] [K♠ Q♦]
	// Player 2: [J♣ Q♣] Straight, Ace-high [A♥ K♠ Q♣ J♣ T♣] [Q♦ T♦]
	// Player 3: [A♦ K♦] Two Pair, Aces over Kings, kicker Queen [A♦ A♥ K♦ K♠ Q♦] [T♣ T♦]
	// Player 4: [K♥ K♣] Full House, Kings full of Tens [K♣ K♥ K♠ T♣ T♦] [A♥ Q♦]
	// Result:   Player 1 wins with Four of a Kind, Tens, kicker Ace
	// ------ Royal 4 ------
	// Board:    [J♥ A♠ T♥ T♣ K♠]
	// Player 1: [Q♦ T♠] Straight, Ace-high [A♠ K♠ Q♦ J♥ T♣] [T♥ T♠]
	// Player 2: [K♥ T♦] Full House, Tens full of Kings [T♣ T♦ T♥ K♥ K♠] [A♠ J♥]
	// Player 3: [A♣ Q♠] Straight, Ace-high [A♣ K♠ Q♠ J♥ T♣] [A♠ T♥]
	// Player 4: [A♦ J♠] Two Pair, Aces over Jacks, kicker King [A♦ A♠ J♥ J♠ K♠] [T♣ T♥]
	// Player 5: [K♦ J♦] Two Pair, Kings over Jacks, kicker Ace [K♦ K♠ J♦ J♥ A♠] [T♣ T♥]
	// Result:   Player 2 wins with Full House, Tens full of Kings
	// ------ Royal 5 ------
	// Board:    [J♣ K♥ K♠ J♥ Q♣]
	// Player 1: [A♥ T♦] Straight, Ace-high [A♥ K♥ Q♣ J♣ T♦] [K♠ J♥]
	// Player 2: [J♦ Q♠] Full House, Jacks full of Kings [J♣ J♦ J♥ K♥ K♠] [Q♣ Q♠]
	// Result:   Player 2 wins with Full House, Jacks full of Kings
	// ------ Royal 6 ------
	// Board:    [K♥ A♠ K♦ K♠ A♣]
	// Player 1: [J♥ J♠] Full House, Kings full of Aces [K♦ K♥ K♠ A♣ A♠] [J♥ J♠]
	// Player 2: [Q♦ A♥] Full House, Aces full of Kings [A♣ A♥ A♠ K♦ K♥] [K♠ Q♦]
	// Player 3: [Q♠ T♣] Full House, Kings full of Aces [K♦ K♥ K♠ A♣ A♠] [Q♠ T♣]
	// Result:   Player 2 wins with Full House, Aces full of Kings
	// ------ Royal 7 ------
	// Board:    [J♥ T♦ Q♠ K♣ K♥]
	// Player 1: [K♦ J♣] Full House, Kings full of Jacks [K♣ K♦ K♥ J♣ J♥] [Q♠ T♦]
	// Player 2: [T♥ T♠] Full House, Tens full of Kings [T♦ T♥ T♠ K♣ K♥] [Q♠ J♥]
	// Player 3: [A♠ A♥] Straight, Ace-high [A♥ K♣ Q♠ J♥ T♦] [A♠ K♥]
	// Player 4: [Q♣ A♦] Straight, Ace-high [A♦ K♣ Q♣ J♥ T♦] [K♥ Q♠]
	// Result:   Player 1 wins with Full House, Kings full of Jacks
	// ------ Royal 8 ------
	// Board:    [A♠ K♦ Q♦ A♦ A♣]
	// Player 1: [Q♠ J♠] Full House, Aces full of Queens [A♣ A♦ A♠ Q♦ Q♠] [K♦ J♠]
	// Player 2: [T♦ A♥] Four of a Kind, Aces, kicker King [A♣ A♦ A♥ A♠ K♦] [Q♦ T♦]
	// Player 3: [J♥ K♠] Full House, Aces full of Kings [A♣ A♦ A♠ K♦ K♠] [Q♦ J♥]
	// Player 4: [Q♥ J♦] Full House, Aces full of Queens [A♣ A♦ A♠ Q♦ Q♥] [K♦ J♦]
	// Player 5: [K♣ T♥] Full House, Aces full of Kings [A♣ A♦ A♠ K♣ K♦] [Q♦ T♥]
	// Result:   Player 2 wins with Four of a Kind, Aces, kicker King
}

func ExampleType_omaha() {
	for i, game := range []struct {
		seed    int64
		players int
	}{
		{119, 2},
		{321, 5},
		{408, 6},
		{455, 6},
		{1113, 6},
	} {
		// note: use a real random source
		r := rand.New(rand.NewSource(game.seed))
		pockets, board := cardrank.Omaha.Deal(r, 1, game.players)
		evs := cardrank.Omaha.EvalPockets(pockets, board)
		fmt.Printf("------ Omaha %d ------\n", i+1)
		fmt.Printf("Board:    %b\n", board)
		for j := range game.players {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := range pivot {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Omaha 1 ------
	// Board:    [3♥ 5♥ 4♥ 7♥ K♣]
	// Player 1: [K♥ J♣ A♥ Q♠] Flush, Ace-high, kickers King, Seven, Five, Four [A♥ K♥ 7♥ 5♥ 4♥] [K♣ Q♠ J♣ 3♥]
	// Player 2: [7♣ 4♣ 5♠ 2♠] Two Pair, Sevens over Fives, kicker King [7♣ 7♥ 5♥ 5♠ K♣] [4♣ 4♥ 3♥ 2♠]
	// Result:   Player 1 wins with Flush, Ace-high, kickers King, Seven, Five, Four
	// ------ Omaha 2 ------
	// Board:    [3♥ 7♣ 3♣ 9♠ 9♣]
	// Player 1: [3♠ 3♦ T♠ Q♠] Four of a Kind, Threes, kicker Nine [3♣ 3♦ 3♥ 3♠ 9♠] [Q♠ T♠ 9♣ 7♣]
	// Player 2: [6♦ Q♣ 8♥ 6♣] Flush, Queen-high, kickers Nine, Seven, Six, Three [Q♣ 9♣ 7♣ 6♣ 3♣] [9♠ 8♥ 6♦ 3♥]
	// Player 3: [Q♦ K♠ 8♣ A♥] Pair, Nines, kickers Ace, King, Seven [9♣ 9♠ A♥ K♠ 7♣] [Q♦ 8♣ 3♣ 3♥]
	// Player 4: [K♦ T♦ 8♦ 4♥] Pair, Nines, kickers King, Ten, Seven [9♣ 9♠ K♦ T♦ 7♣] [8♦ 4♥ 3♣ 3♥]
	// Player 5: [J♦ 2♥ Q♥ 6♠] Pair, Nines, kickers Queen, Jack, Seven [9♣ 9♠ Q♥ J♦ 7♣] [6♠ 3♣ 3♥ 2♥]
	// Result:   Player 1 wins with Four of a Kind, Threes, kicker Nine
	// ------ Omaha 3 ------
	// Board:    [J♣ T♥ 4♥ K♣ Q♣]
	// Player 1: [K♠ Q♠ 4♣ J♦] Two Pair, Kings over Queens, kicker Jack [K♣ K♠ Q♣ Q♠ J♣] [J♦ T♥ 4♣ 4♥]
	// Player 2: [J♠ 3♣ 8♥ 2♠] Pair, Jacks, kickers King, Queen, Eight [J♣ J♠ K♣ Q♣ 8♥] [T♥ 4♥ 3♣ 2♠]
	// Player 3: [3♠ T♠ 2♣ Q♦] Two Pair, Queens over Tens, kicker King [Q♣ Q♦ T♥ T♠ K♣] [J♣ 4♥ 3♠ 2♣]
	// Player 4: [5♣ 5♥ T♦ 2♦] Pair, Tens, kickers King, Queen, Five [T♦ T♥ K♣ Q♣ 5♣] [J♣ 5♥ 4♥ 2♦]
	// Player 5: [7♠ 3♥ 6♠ A♣] Ace-high, kickers King, Queen, Jack, Seven [A♣ K♣ Q♣ J♣ 7♠] [T♥ 6♠ 4♥ 3♥]
	// Player 6: [4♠ 8♦ K♦ T♣] Two Pair, Kings over Tens, kicker Queen [K♣ K♦ T♣ T♥ Q♣] [J♣ 8♦ 4♥ 4♠]
	// Result:   Player 1 wins with Two Pair, Kings over Queens, kicker Jack
	// ------ Omaha 4 ------
	// Board:    [2♦ 6♦ 6♣ Q♣ 7♣]
	// Player 1: [6♠ K♥ A♣ 8♣] Flush, Ace-high, kickers Queen, Eight, Seven, Six [A♣ Q♣ 8♣ 7♣ 6♣] [K♥ 6♦ 6♠ 2♦]
	// Player 2: [Q♥ 4♥ J♣ 5♥] Two Pair, Queens over Sixes, kicker Jack [Q♣ Q♥ 6♣ 6♦ J♣] [7♣ 5♥ 4♥ 2♦]
	// Player 3: [2♣ 6♥ 5♣ Q♠] Full House, Sixes full of Queens [6♣ 6♦ 6♥ Q♣ Q♠] [7♣ 5♣ 2♣ 2♦]
	// Player 4: [9♠ J♥ K♠ J♠] Two Pair, Jacks over Sixes, kicker Queen [J♥ J♠ 6♣ 6♦ Q♣] [K♠ 9♠ 7♣ 2♦]
	// Player 5: [3♦ 4♦ K♣ 8♦] Pair, Sixes, kickers King, Queen, Eight [6♣ 6♦ K♣ Q♣ 8♦] [7♣ 4♦ 3♦ 2♦]
	// Player 6: [T♣ Q♦ A♠ 7♥] Two Pair, Queens over Sevens, kicker Six [Q♣ Q♦ 7♣ 7♥ 6♦] [A♠ T♣ 6♣ 2♦]
	// Result:   Player 3 wins with Full House, Sixes full of Queens
	// ------ Omaha 5 ------
	// Board:    [4♣ K♣ 6♦ 9♦ 5♠]
	// Player 1: [3♦ 4♦ 5♦ J♣] Two Pair, Fives over Fours, kicker King [5♦ 5♠ 4♣ 4♦ K♣] [J♣ 9♦ 6♦ 3♦]
	// Player 2: [T♥ J♠ K♠ 2♣] Pair, Kings, kickers Jack, Nine, Six [K♣ K♠ J♠ 9♦ 6♦] [T♥ 5♠ 4♣ 2♣]
	// Player 3: [A♣ 9♠ T♠ 3♠] Pair, Nines, kickers Ace, King, Six [9♦ 9♠ A♣ K♣ 6♦] [T♠ 5♠ 4♣ 3♠]
	// Player 4: [7♦ 3♣ 8♠ 7♣] Straight, Nine-high [9♦ 8♠ 7♦ 6♦ 5♠] [K♣ 7♣ 4♣ 3♣]
	// Player 5: [5♣ Q♠ J♥ 2♠] Pair, Fives, kickers King, Queen, Nine [5♣ 5♠ K♣ Q♠ 9♦] [J♥ 6♦ 4♣ 2♠]
	// Player 6: [6♠ 7♠ 7♥ 2♥] Pair, Sevens, kickers King, Nine, Six [7♥ 7♠ K♣ 9♦ 6♦] [6♠ 5♠ 4♣ 2♥]
	// Result:   Player 4 wins with Straight, Nine-high
}

func ExampleType_omahaHiLo() {
	for i, game := range []struct {
		seed    int64
		players int
	}{
		{119, 2},
		{321, 5},
		{408, 6},
		{455, 6},
		{1113, 6},
	} {
		// note: use a real random source
		r := rand.New(rand.NewSource(game.seed))
		pockets, board := cardrank.OmahaHiLo.Deal(r, 1, game.players)
		evs := cardrank.OmahaHiLo.EvalPockets(pockets, board)
		fmt.Printf("------ OmahaHiLo %d ------\n", i+1)
		fmt.Printf("Board: %b\n", board)
		for j := range game.players {
			hi, lo := evs[j].Desc(false), evs[j].Desc(true)
			fmt.Printf("Player %d: %b\n", j+1, pockets[j])
			fmt.Printf("  Hi: %s %b %b\n", hi, hi.Best, hi.Unused)
			fmt.Printf("  Lo: %s %b %b\n", lo, lo.Best, lo.Unused)
		}
		hiOrder, hiPivot := cardrank.Order(evs, false)
		loOrder, loPivot := cardrank.Order(evs, true)
		typ := "wins"
		if loPivot == 0 {
			typ = "scoops"
		}
		desc := evs[hiOrder[0]].Desc(false)
		if hiPivot == 1 {
			fmt.Printf("Result (Hi): Player %d %s with %s\n", hiOrder[0]+1, typ, desc)
		} else {
			var s []string
			for j := range hiPivot {
				s = append(s, strconv.Itoa(hiOrder[j]+1))
			}
			fmt.Printf("Result (Hi): Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
		if loPivot == 1 {
			desc := evs[loOrder[0]].Desc(true)
			fmt.Printf("Result (Lo): Player %d wins with %s\n", loOrder[0]+1, desc)
		} else if loPivot > 1 {
			var s []string
			for j := range loPivot {
				s = append(s, strconv.Itoa(loOrder[j]+1))
			}
			desc := evs[loOrder[0]].Desc(true)
			fmt.Printf("Result (Lo): Players %s push with %s\n", strings.Join(s, ", "), desc)
		} else {
			fmt.Printf("Result (Lo): no player made a lo\n")
		}
	}
	// Output:
	// ------ OmahaHiLo 1 ------
	// Board: [3♥ 5♥ 4♥ 7♥ K♣]
	// Player 1: [K♥ J♣ A♥ Q♠]
	//   Hi: Flush, Ace-high, kickers King, Seven, Five, Four [A♥ K♥ 7♥ 5♥ 4♥] [K♣ Q♠ J♣ 3♥]
	//   Lo: None [] []
	// Player 2: [7♣ 4♣ 5♠ 2♠]
	//   Hi: Two Pair, Sevens over Fives, kicker King [7♣ 7♥ 5♥ 5♠ K♣] [4♣ 4♥ 3♥ 2♠]
	//   Lo: Seven, Five, Four, Three, Two-low [7♣ 5♥ 4♥ 3♥ 2♠] [K♣ 7♥ 5♠ 4♣]
	// Result (Hi): Player 1 wins with Flush, Ace-high, kickers King, Seven, Five, Four
	// Result (Lo): Player 2 wins with Seven, Five, Four, Three, Two-low
	// ------ OmahaHiLo 2 ------
	// Board: [3♥ 7♣ 3♣ 9♠ 9♣]
	// Player 1: [3♠ 3♦ T♠ Q♠]
	//   Hi: Four of a Kind, Threes, kicker Nine [3♣ 3♦ 3♥ 3♠ 9♠] [Q♠ T♠ 9♣ 7♣]
	//   Lo: None [] []
	// Player 2: [6♦ Q♣ 8♥ 6♣]
	//   Hi: Flush, Queen-high, kickers Nine, Seven, Six, Three [Q♣ 9♣ 7♣ 6♣ 3♣] [9♠ 8♥ 6♦ 3♥]
	//   Lo: None [] []
	// Player 3: [Q♦ K♠ 8♣ A♥]
	//   Hi: Pair, Nines, kickers Ace, King, Seven [9♣ 9♠ A♥ K♠ 7♣] [Q♦ 8♣ 3♣ 3♥]
	//   Lo: None [] []
	// Player 4: [K♦ T♦ 8♦ 4♥]
	//   Hi: Pair, Nines, kickers King, Ten, Seven [9♣ 9♠ K♦ T♦ 7♣] [8♦ 4♥ 3♣ 3♥]
	//   Lo: None [] []
	// Player 5: [J♦ 2♥ Q♥ 6♠]
	//   Hi: Pair, Nines, kickers Queen, Jack, Seven [9♣ 9♠ Q♥ J♦ 7♣] [6♠ 3♣ 3♥ 2♥]
	//   Lo: None [] []
	// Result (Hi): Player 1 scoops with Four of a Kind, Threes, kicker Nine
	// Result (Lo): no player made a lo
	// ------ OmahaHiLo 3 ------
	// Board: [J♣ T♥ 4♥ K♣ Q♣]
	// Player 1: [K♠ Q♠ 4♣ J♦]
	//   Hi: Two Pair, Kings over Queens, kicker Jack [K♣ K♠ Q♣ Q♠ J♣] [J♦ T♥ 4♣ 4♥]
	//   Lo: None [] []
	// Player 2: [J♠ 3♣ 8♥ 2♠]
	//   Hi: Pair, Jacks, kickers King, Queen, Eight [J♣ J♠ K♣ Q♣ 8♥] [T♥ 4♥ 3♣ 2♠]
	//   Lo: None [] []
	// Player 3: [3♠ T♠ 2♣ Q♦]
	//   Hi: Two Pair, Queens over Tens, kicker King [Q♣ Q♦ T♥ T♠ K♣] [J♣ 4♥ 3♠ 2♣]
	//   Lo: None [] []
	// Player 4: [5♣ 5♥ T♦ 2♦]
	//   Hi: Pair, Tens, kickers King, Queen, Five [T♦ T♥ K♣ Q♣ 5♣] [J♣ 5♥ 4♥ 2♦]
	//   Lo: None [] []
	// Player 5: [7♠ 3♥ 6♠ A♣]
	//   Hi: Ace-high, kickers King, Queen, Jack, Seven [A♣ K♣ Q♣ J♣ 7♠] [T♥ 6♠ 4♥ 3♥]
	//   Lo: None [] []
	// Player 6: [4♠ 8♦ K♦ T♣]
	//   Hi: Two Pair, Kings over Tens, kicker Queen [K♣ K♦ T♣ T♥ Q♣] [J♣ 8♦ 4♥ 4♠]
	//   Lo: None [] []
	// Result (Hi): Player 1 scoops with Two Pair, Kings over Queens, kicker Jack
	// Result (Lo): no player made a lo
	// ------ OmahaHiLo 4 ------
	// Board: [2♦ 6♦ 6♣ Q♣ 7♣]
	// Player 1: [6♠ K♥ A♣ 8♣]
	//   Hi: Flush, Ace-high, kickers Queen, Eight, Seven, Six [A♣ Q♣ 8♣ 7♣ 6♣] [K♥ 6♦ 6♠ 2♦]
	//   Lo: Eight, Seven, Six, Two, Ace-low [8♣ 7♣ 6♦ 2♦ A♣] [K♥ Q♣ 6♣ 6♠]
	// Player 2: [Q♥ 4♥ J♣ 5♥]
	//   Hi: Two Pair, Queens over Sixes, kicker Jack [Q♣ Q♥ 6♣ 6♦ J♣] [7♣ 5♥ 4♥ 2♦]
	//   Lo: Seven, Six, Five, Four, Two-low [7♣ 6♦ 5♥ 4♥ 2♦] [Q♣ Q♥ J♣ 6♣]
	// Player 3: [2♣ 6♥ 5♣ Q♠]
	//   Hi: Full House, Sixes full of Queens [6♣ 6♦ 6♥ Q♣ Q♠] [7♣ 5♣ 2♣ 2♦]
	//   Lo: None [] []
	// Player 4: [9♠ J♥ K♠ J♠]
	//   Hi: Two Pair, Jacks over Sixes, kicker Queen [J♥ J♠ 6♣ 6♦ Q♣] [K♠ 9♠ 7♣ 2♦]
	//   Lo: None [] []
	// Player 5: [3♦ 4♦ K♣ 8♦]
	//   Hi: Pair, Sixes, kickers King, Queen, Eight [6♣ 6♦ K♣ Q♣ 8♦] [7♣ 4♦ 3♦ 2♦]
	//   Lo: Seven, Six, Four, Three, Two-low [7♣ 6♦ 4♦ 3♦ 2♦] [K♣ Q♣ 8♦ 6♣]
	// Player 6: [T♣ Q♦ A♠ 7♥]
	//   Hi: Two Pair, Queens over Sevens, kicker Six [Q♣ Q♦ 7♣ 7♥ 6♦] [A♠ T♣ 6♣ 2♦]
	//   Lo: None [] []
	// Result (Hi): Player 3 wins with Full House, Sixes full of Queens
	// Result (Lo): Player 5 wins with Seven, Six, Four, Three, Two-low
	// ------ OmahaHiLo 5 ------
	// Board: [4♣ K♣ 6♦ 9♦ 5♠]
	// Player 1: [3♦ 4♦ 5♦ J♣]
	//   Hi: Two Pair, Fives over Fours, kicker King [5♦ 5♠ 4♣ 4♦ K♣] [J♣ 9♦ 6♦ 3♦]
	//   Lo: None [] []
	// Player 2: [T♥ J♠ K♠ 2♣]
	//   Hi: Pair, Kings, kickers Jack, Nine, Six [K♣ K♠ J♠ 9♦ 6♦] [T♥ 5♠ 4♣ 2♣]
	//   Lo: None [] []
	// Player 3: [A♣ 9♠ T♠ 3♠]
	//   Hi: Pair, Nines, kickers Ace, King, Six [9♦ 9♠ A♣ K♣ 6♦] [T♠ 5♠ 4♣ 3♠]
	//   Lo: Six, Five, Four, Three, Ace-low [6♦ 5♠ 4♣ 3♠ A♣] [K♣ T♠ 9♦ 9♠]
	// Player 4: [7♦ 3♣ 8♠ 7♣]
	//   Hi: Straight, Nine-high [9♦ 8♠ 7♦ 6♦ 5♠] [K♣ 7♣ 4♣ 3♣]
	//   Lo: Seven, Six, Five, Four, Three-low [7♦ 6♦ 5♠ 4♣ 3♣] [K♣ 9♦ 8♠ 7♣]
	// Player 5: [5♣ Q♠ J♥ 2♠]
	//   Hi: Pair, Fives, kickers King, Queen, Nine [5♣ 5♠ K♣ Q♠ 9♦] [J♥ 6♦ 4♣ 2♠]
	//   Lo: None [] []
	// Player 6: [6♠ 7♠ 7♥ 2♥]
	//   Hi: Pair, Sevens, kickers King, Nine, Six [7♥ 7♠ K♣ 9♦ 6♦] [6♠ 5♠ 4♣ 2♥]
	//   Lo: Seven, Six, Five, Four, Two-low [7♠ 6♦ 5♠ 4♣ 2♥] [K♣ 9♦ 7♥ 6♠]
	// Result (Hi): Player 4 wins with Straight, Nine-high
	// Result (Lo): Player 3 wins with Six, Five, Four, Three, Ace-low
}

func ExampleType_stud() {
	for i, game := range []struct {
		seed    int64
		players int
	}{
		{119, 2},
		{321, 5},
		{408, 6},
		{455, 6},
		{1113, 6},
	} {
		// note: use a real random source
		r := rand.New(rand.NewSource(game.seed))
		pockets, _ := cardrank.Stud.Deal(r, 1, game.players)
		evs := cardrank.Stud.EvalPockets(pockets, nil)
		fmt.Printf("------ Stud %d ------\n", i+1)
		for j := range game.players {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := range pivot {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Stud 1 ------
	// Player 1: [K♥ J♣ A♥ Q♠ 6♣ 5♥ Q♦] Pair, Queens, kickers Ace, King, Jack [Q♦ Q♠ A♥ K♥ J♣] [6♣ 5♥]
	// Player 2: [7♣ 4♣ 5♠ 2♠ 3♥ 4♥ 7♥] Two Pair, Sevens over Fours, kicker Five [7♣ 7♥ 4♣ 4♥ 5♠] [3♥ 2♠]
	// Result:   Player 2 wins with Two Pair, Sevens over Fours, kicker Five
	// ------ Stud 2 ------
	// Player 1: [3♠ 3♦ T♠ Q♠ T♥ 9♠ K♥] Two Pair, Tens over Threes, kicker King [T♥ T♠ 3♦ 3♠ K♥] [Q♠ 9♠]
	// Player 2: [6♦ Q♣ 8♥ 6♣ 3♥ T♣ 7♥] Pair, Sixes, kickers Queen, Ten, Eight [6♣ 6♦ Q♣ T♣ 8♥] [7♥ 3♥]
	// Player 3: [Q♦ K♠ 8♣ A♥ 7♣ 9♣ 2♣] Ace-high, kickers King, Queen, Nine, Eight [A♥ K♠ Q♦ 9♣ 8♣] [7♣ 2♣]
	// Player 4: [K♦ T♦ 8♦ 4♥ 3♣ J♠ 2♦] King-high, kickers Jack, Ten, Eight, Four [K♦ J♠ T♦ 8♦ 4♥] [3♣ 2♦]
	// Player 5: [J♦ 2♥ Q♥ 6♠ 5♦ 7♠ A♦] Ace-high, kickers Queen, Jack, Seven, Six [A♦ Q♥ J♦ 7♠ 6♠] [5♦ 2♥]
	// Result:   Player 1 wins with Two Pair, Tens over Threes, kicker King
	// ------ Stud 3 ------
	// Player 1: [K♠ Q♠ 4♣ J♦ 7♥ 7♣ J♥] Two Pair, Jacks over Sevens, kicker King [J♦ J♥ 7♣ 7♥ K♠] [Q♠ 4♣]
	// Player 2: [J♠ 3♣ 8♥ 2♠ J♣ Q♣ 7♦] Pair, Jacks, kickers Queen, Eight, Seven [J♣ J♠ Q♣ 8♥ 7♦] [3♣ 2♠]
	// Player 3: [3♠ T♠ 2♣ Q♦ T♥ K♥ 3♦] Two Pair, Tens over Threes, kicker King [T♥ T♠ 3♦ 3♠ K♥] [Q♦ 2♣]
	// Player 4: [5♣ 5♥ T♦ 2♦ 4♥ 9♦ 2♥] Two Pair, Fives over Twos, kicker Ten [5♣ 5♥ 2♦ 2♥ T♦] [9♦ 4♥]
	// Player 5: [7♠ 3♥ 6♠ A♣ 8♠ 6♦ A♦] Two Pair, Aces over Sixes, kicker Eight [A♣ A♦ 6♦ 6♠ 8♠] [7♠ 3♥]
	// Player 6: [4♠ 8♦ K♦ T♣ K♣ 5♠ 9♣] Pair, Kings, kickers Ten, Nine, Eight [K♣ K♦ T♣ 9♣ 8♦] [5♠ 4♠]
	// Result:   Player 5 wins with Two Pair, Aces over Sixes, kicker Eight
	// ------ Stud 4 ------
	// Player 1: [6♠ K♥ A♣ 8♣ 2♠ 5♦ A♥] Pair, Aces, kickers King, Eight, Six [A♣ A♥ K♥ 8♣ 6♠] [5♦ 2♠]
	// Player 2: [Q♥ 4♥ J♣ 5♥ 2♦ 7♣ 3♠] Queen-high, kickers Jack, Seven, Five, Four [Q♥ J♣ 7♣ 5♥ 4♥] [3♠ 2♦]
	// Player 3: [2♣ 6♥ 5♣ Q♠ 6♦ 9♥ 3♣] Pair, Sixes, kickers Queen, Nine, Five [6♦ 6♥ Q♠ 9♥ 5♣] [3♣ 2♣]
	// Player 4: [9♠ J♥ K♠ J♠ 6♣ K♦ T♠] Two Pair, Kings over Jacks, kicker Ten [K♦ K♠ J♥ J♠ T♠] [9♠ 6♣]
	// Player 5: [3♦ 4♦ K♣ 8♦ 8♥ 9♣ T♥] Pair, Eights, kickers King, Ten, Nine [8♦ 8♥ K♣ T♥ 9♣] [4♦ 3♦]
	// Player 6: [T♣ Q♦ A♠ 7♥ Q♣ 7♦ 2♥] Two Pair, Queens over Sevens, kicker Ace [Q♣ Q♦ 7♦ 7♥ A♠] [T♣ 2♥]
	// Result:   Player 4 wins with Two Pair, Kings over Jacks, kicker Ten
	// ------ Stud 5 ------
	// Player 1: [3♦ 4♦ 5♦ J♣ 4♥ K♥ 8♣] Pair, Fours, kickers King, Jack, Eight [4♦ 4♥ K♥ J♣ 8♣] [5♦ 3♦]
	// Player 2: [T♥ J♠ K♠ 2♣ 4♣ 5♠ 2♦] Pair, Twos, kickers King, Jack, Ten [2♣ 2♦ K♠ J♠ T♥] [5♠ 4♣]
	// Player 3: [A♣ 9♠ T♠ 3♠ K♣ 8♦ A♥] Pair, Aces, kickers King, Ten, Nine [A♣ A♥ K♣ T♠ 9♠] [8♦ 3♠]
	// Player 4: [7♦ 3♣ 8♠ 7♣ 6♦ 6♥ 6♣] Full House, Sixes full of Sevens [6♣ 6♦ 6♥ 7♣ 7♦] [8♠ 3♣]
	// Player 5: [5♣ Q♠ J♥ 2♠ A♠ 8♥ 4♠] Ace-high, kickers Queen, Jack, Eight, Five [A♠ Q♠ J♥ 8♥ 5♣] [4♠ 2♠]
	// Player 6: [6♠ 7♠ 7♥ 2♥ 9♦ K♦ T♦] Pair, Sevens, kickers King, Ten, Nine [7♥ 7♠ K♦ T♦ 9♦] [6♠ 2♥]
	// Result:   Player 4 wins with Full House, Sixes full of Sevens
}

func ExampleType_studHiLo() {
	for i, game := range []struct {
		seed    int64
		players int
	}{
		{119, 2},
		{321, 5},
		{408, 6},
		{455, 6},
		{1113, 6},
	} {
		// note: use a real random source
		r := rand.New(rand.NewSource(game.seed))
		pockets, _ := cardrank.StudHiLo.Deal(r, 1, game.players)
		evs := cardrank.StudHiLo.EvalPockets(pockets, nil)
		fmt.Printf("------ StudHiLo %d ------\n", i+1)
		for j := range game.players {
			hi, lo := evs[j].Desc(false), evs[j].Desc(true)
			fmt.Printf("Player %d: %b\n", j+1, pockets[j])
			fmt.Printf("  Hi: %s %b %b\n", hi, hi.Best, hi.Unused)
			fmt.Printf("  Lo: %s %b %b\n", lo, lo.Best, lo.Unused)
		}
		hiOrder, hiPivot := cardrank.Order(evs, false)
		loOrder, loPivot := cardrank.Order(evs, true)
		typ := "wins"
		if loPivot == 0 {
			typ = "scoops"
		}
		desc := evs[hiOrder[0]].Desc(false)
		if hiPivot == 1 {
			fmt.Printf("Result (Hi): Player %d %s with %s\n", hiOrder[0]+1, typ, desc)
		} else {
			var s []string
			for j := range hiPivot {
				s = append(s, strconv.Itoa(hiOrder[j]+1))
			}
			fmt.Printf("Result (Hi): Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
		if loPivot == 1 {
			desc := evs[loOrder[0]].Desc(true)
			fmt.Printf("Result (Lo): Player %d wins with %s\n", loOrder[0]+1, desc)
		} else if loPivot > 1 {
			var s []string
			for j := range loPivot {
				s = append(s, strconv.Itoa(loOrder[j]+1))
			}
			desc := evs[loOrder[0]].Desc(true)
			fmt.Printf("Result (Lo): Players %s push with %s\n", strings.Join(s, ", "), desc)
		} else {
			fmt.Printf("Result (Lo): no player made a lo\n")
		}
	}
	// Output:
	// ------ StudHiLo 1 ------
	// Player 1: [K♥ J♣ A♥ Q♠ 6♣ 5♥ Q♦]
	//   Hi: Pair, Queens, kickers Ace, King, Jack [Q♦ Q♠ A♥ K♥ J♣] [6♣ 5♥]
	//   Lo: None [] []
	// Player 2: [7♣ 4♣ 5♠ 2♠ 3♥ 4♥ 7♥]
	//   Hi: Two Pair, Sevens over Fours, kicker Five [7♣ 7♥ 4♣ 4♥ 5♠] [3♥ 2♠]
	//   Lo: Seven, Five, Four, Three, Two-low [7♣ 5♠ 4♣ 3♥ 2♠] [7♥ 4♥]
	// Result (Hi): Player 2 wins with Two Pair, Sevens over Fours, kicker Five
	// Result (Lo): Player 2 wins with Seven, Five, Four, Three, Two-low
	// ------ StudHiLo 2 ------
	// Player 1: [3♠ 3♦ T♠ Q♠ T♥ 9♠ K♥]
	//   Hi: Two Pair, Tens over Threes, kicker King [T♥ T♠ 3♦ 3♠ K♥] [Q♠ 9♠]
	//   Lo: None [] []
	// Player 2: [6♦ Q♣ 8♥ 6♣ 3♥ T♣ 7♥]
	//   Hi: Pair, Sixes, kickers Queen, Ten, Eight [6♣ 6♦ Q♣ T♣ 8♥] [7♥ 3♥]
	//   Lo: None [] []
	// Player 3: [Q♦ K♠ 8♣ A♥ 7♣ 9♣ 2♣]
	//   Hi: Ace-high, kickers King, Queen, Nine, Eight [A♥ K♠ Q♦ 9♣ 8♣] [7♣ 2♣]
	//   Lo: None [] []
	// Player 4: [K♦ T♦ 8♦ 4♥ 3♣ J♠ 2♦]
	//   Hi: King-high, kickers Jack, Ten, Eight, Four [K♦ J♠ T♦ 8♦ 4♥] [3♣ 2♦]
	//   Lo: None [] []
	// Player 5: [J♦ 2♥ Q♥ 6♠ 5♦ 7♠ A♦]
	//   Hi: Ace-high, kickers Queen, Jack, Seven, Six [A♦ Q♥ J♦ 7♠ 6♠] [5♦ 2♥]
	//   Lo: Seven, Six, Five, Two, Ace-low [7♠ 6♠ 5♦ 2♥ A♦] [Q♥ J♦]
	// Result (Hi): Player 1 wins with Two Pair, Tens over Threes, kicker King
	// Result (Lo): Player 5 wins with Seven, Six, Five, Two, Ace-low
	// ------ StudHiLo 3 ------
	// Player 1: [K♠ Q♠ 4♣ J♦ 7♥ 7♣ J♥]
	//   Hi: Two Pair, Jacks over Sevens, kicker King [J♦ J♥ 7♣ 7♥ K♠] [Q♠ 4♣]
	//   Lo: None [] []
	// Player 2: [J♠ 3♣ 8♥ 2♠ J♣ Q♣ 7♦]
	//   Hi: Pair, Jacks, kickers Queen, Eight, Seven [J♣ J♠ Q♣ 8♥ 7♦] [3♣ 2♠]
	//   Lo: None [] []
	// Player 3: [3♠ T♠ 2♣ Q♦ T♥ K♥ 3♦]
	//   Hi: Two Pair, Tens over Threes, kicker King [T♥ T♠ 3♦ 3♠ K♥] [Q♦ 2♣]
	//   Lo: None [] []
	// Player 4: [5♣ 5♥ T♦ 2♦ 4♥ 9♦ 2♥]
	//   Hi: Two Pair, Fives over Twos, kicker Ten [5♣ 5♥ 2♦ 2♥ T♦] [9♦ 4♥]
	//   Lo: None [] []
	// Player 5: [7♠ 3♥ 6♠ A♣ 8♠ 6♦ A♦]
	//   Hi: Two Pair, Aces over Sixes, kicker Eight [A♣ A♦ 6♦ 6♠ 8♠] [7♠ 3♥]
	//   Lo: Eight, Seven, Six, Three, Ace-low [8♠ 7♠ 6♠ 3♥ A♣] [A♦ 6♦]
	// Player 6: [4♠ 8♦ K♦ T♣ K♣ 5♠ 9♣]
	//   Hi: Pair, Kings, kickers Ten, Nine, Eight [K♣ K♦ T♣ 9♣ 8♦] [5♠ 4♠]
	//   Lo: None [] []
	// Result (Hi): Player 5 wins with Two Pair, Aces over Sixes, kicker Eight
	// Result (Lo): Player 5 wins with Eight, Seven, Six, Three, Ace-low
	// ------ StudHiLo 4 ------
	// Player 1: [6♠ K♥ A♣ 8♣ 2♠ 5♦ A♥]
	//   Hi: Pair, Aces, kickers King, Eight, Six [A♣ A♥ K♥ 8♣ 6♠] [5♦ 2♠]
	//   Lo: Eight, Six, Five, Two, Ace-low [8♣ 6♠ 5♦ 2♠ A♣] [A♥ K♥]
	// Player 2: [Q♥ 4♥ J♣ 5♥ 2♦ 7♣ 3♠]
	//   Hi: Queen-high, kickers Jack, Seven, Five, Four [Q♥ J♣ 7♣ 5♥ 4♥] [3♠ 2♦]
	//   Lo: Seven, Five, Four, Three, Two-low [7♣ 5♥ 4♥ 3♠ 2♦] [Q♥ J♣]
	// Player 3: [2♣ 6♥ 5♣ Q♠ 6♦ 9♥ 3♣]
	//   Hi: Pair, Sixes, kickers Queen, Nine, Five [6♦ 6♥ Q♠ 9♥ 5♣] [3♣ 2♣]
	//   Lo: None [] []
	// Player 4: [9♠ J♥ K♠ J♠ 6♣ K♦ T♠]
	//   Hi: Two Pair, Kings over Jacks, kicker Ten [K♦ K♠ J♥ J♠ T♠] [9♠ 6♣]
	//   Lo: None [] []
	// Player 5: [3♦ 4♦ K♣ 8♦ 8♥ 9♣ T♥]
	//   Hi: Pair, Eights, kickers King, Ten, Nine [8♦ 8♥ K♣ T♥ 9♣] [4♦ 3♦]
	//   Lo: None [] []
	// Player 6: [T♣ Q♦ A♠ 7♥ Q♣ 7♦ 2♥]
	//   Hi: Two Pair, Queens over Sevens, kicker Ace [Q♣ Q♦ 7♦ 7♥ A♠] [T♣ 2♥]
	//   Lo: None [] []
	// Result (Hi): Player 4 wins with Two Pair, Kings over Jacks, kicker Ten
	// Result (Lo): Player 2 wins with Seven, Five, Four, Three, Two-low
	// ------ StudHiLo 5 ------
	// Player 1: [3♦ 4♦ 5♦ J♣ 4♥ K♥ 8♣]
	//   Hi: Pair, Fours, kickers King, Jack, Eight [4♦ 4♥ K♥ J♣ 8♣] [5♦ 3♦]
	//   Lo: None [] []
	// Player 2: [T♥ J♠ K♠ 2♣ 4♣ 5♠ 2♦]
	//   Hi: Pair, Twos, kickers King, Jack, Ten [2♣ 2♦ K♠ J♠ T♥] [5♠ 4♣]
	//   Lo: None [] []
	// Player 3: [A♣ 9♠ T♠ 3♠ K♣ 8♦ A♥]
	//   Hi: Pair, Aces, kickers King, Ten, Nine [A♣ A♥ K♣ T♠ 9♠] [8♦ 3♠]
	//   Lo: None [] []
	// Player 4: [7♦ 3♣ 8♠ 7♣ 6♦ 6♥ 6♣]
	//   Hi: Full House, Sixes full of Sevens [6♣ 6♦ 6♥ 7♣ 7♦] [8♠ 3♣]
	//   Lo: None [] []
	// Player 5: [5♣ Q♠ J♥ 2♠ A♠ 8♥ 4♠]
	//   Hi: Ace-high, kickers Queen, Jack, Eight, Five [A♠ Q♠ J♥ 8♥ 5♣] [4♠ 2♠]
	//   Lo: Eight, Five, Four, Two, Ace-low [8♥ 5♣ 4♠ 2♠ A♠] [Q♠ J♥]
	// Player 6: [6♠ 7♠ 7♥ 2♥ 9♦ K♦ T♦]
	//   Hi: Pair, Sevens, kickers King, Ten, Nine [7♥ 7♠ K♦ T♦ 9♦] [6♠ 2♥]
	//   Lo: None [] []
	// Result (Hi): Player 4 wins with Full House, Sixes full of Sevens
	// Result (Lo): Player 5 wins with Eight, Five, Four, Two, Ace-low
}

func ExampleType_razz() {
	for i, game := range []struct {
		seed    int64
		players int
	}{
		{119, 2},
		{321, 5},
		{408, 6},
		{455, 6},
		{1113, 6},
	} {
		// note: use a real random source
		r := rand.New(rand.NewSource(game.seed))
		pockets, _ := cardrank.Razz.Deal(r, 1, game.players)
		evs := cardrank.Razz.EvalPockets(pockets, nil)
		fmt.Printf("------ Razz %d ------\n", i+1)
		for j := range game.players {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := range pivot {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Razz 1 ------
	// Player 1: [K♥ J♣ A♥ Q♠ 6♣ 5♥ Q♦] Queen, Jack, Six, Five, Ace-low [Q♠ J♣ 6♣ 5♥ A♥] [K♥ Q♦]
	// Player 2: [7♣ 4♣ 5♠ 2♠ 3♥ 4♥ 7♥] Seven, Five, Four, Three, Two-low [7♣ 5♠ 4♣ 3♥ 2♠] [7♥ 4♥]
	// Result:   Player 2 wins with Seven, Five, Four, Three, Two-low
	// ------ Razz 2 ------
	// Player 1: [3♠ 3♦ T♠ Q♠ T♥ 9♠ K♥] King, Queen, Ten, Nine, Three-low [K♥ Q♠ T♠ 9♠ 3♠] [T♥ 3♦]
	// Player 2: [6♦ Q♣ 8♥ 6♣ 3♥ T♣ 7♥] Ten, Eight, Seven, Six, Three-low [T♣ 8♥ 7♥ 6♦ 3♥] [Q♣ 6♣]
	// Player 3: [Q♦ K♠ 8♣ A♥ 7♣ 9♣ 2♣] Nine, Eight, Seven, Two, Ace-low [9♣ 8♣ 7♣ 2♣ A♥] [K♠ Q♦]
	// Player 4: [K♦ T♦ 8♦ 4♥ 3♣ J♠ 2♦] Ten, Eight, Four, Three, Two-low [T♦ 8♦ 4♥ 3♣ 2♦] [K♦ J♠]
	// Player 5: [J♦ 2♥ Q♥ 6♠ 5♦ 7♠ A♦] Seven, Six, Five, Two, Ace-low [7♠ 6♠ 5♦ 2♥ A♦] [Q♥ J♦]
	// Result:   Player 5 wins with Seven, Six, Five, Two, Ace-low
	// ------ Razz 3 ------
	// Player 1: [K♠ Q♠ 4♣ J♦ 7♥ 7♣ J♥] King, Queen, Jack, Seven, Four-low [K♠ Q♠ J♦ 7♥ 4♣] [J♥ 7♣]
	// Player 2: [J♠ 3♣ 8♥ 2♠ J♣ Q♣ 7♦] Jack, Eight, Seven, Three, Two-low [J♠ 8♥ 7♦ 3♣ 2♠] [Q♣ J♣]
	// Player 3: [3♠ T♠ 2♣ Q♦ T♥ K♥ 3♦] King, Queen, Ten, Three, Two-low [K♥ Q♦ T♠ 3♠ 2♣] [T♥ 3♦]
	// Player 4: [5♣ 5♥ T♦ 2♦ 4♥ 9♦ 2♥] Ten, Nine, Five, Four, Two-low [T♦ 9♦ 5♣ 4♥ 2♦] [5♥ 2♥]
	// Player 5: [7♠ 3♥ 6♠ A♣ 8♠ 6♦ A♦] Eight, Seven, Six, Three, Ace-low [8♠ 7♠ 6♠ 3♥ A♣] [A♦ 6♦]
	// Player 6: [4♠ 8♦ K♦ T♣ K♣ 5♠ 9♣] Ten, Nine, Eight, Five, Four-low [T♣ 9♣ 8♦ 5♠ 4♠] [K♣ K♦]
	// Result:   Player 5 wins with Eight, Seven, Six, Three, Ace-low
	// ------ Razz 4 ------
	// Player 1: [6♠ K♥ A♣ 8♣ 2♠ 5♦ A♥] Eight, Six, Five, Two, Ace-low [8♣ 6♠ 5♦ 2♠ A♣] [A♥ K♥]
	// Player 2: [Q♥ 4♥ J♣ 5♥ 2♦ 7♣ 3♠] Seven, Five, Four, Three, Two-low [7♣ 5♥ 4♥ 3♠ 2♦] [Q♥ J♣]
	// Player 3: [2♣ 6♥ 5♣ Q♠ 6♦ 9♥ 3♣] Nine, Six, Five, Three, Two-low [9♥ 6♥ 5♣ 3♣ 2♣] [Q♠ 6♦]
	// Player 4: [9♠ J♥ K♠ J♠ 6♣ K♦ T♠] King, Jack, Ten, Nine, Six-low [K♠ J♥ T♠ 9♠ 6♣] [K♦ J♠]
	// Player 5: [3♦ 4♦ K♣ 8♦ 8♥ 9♣ T♥] Ten, Nine, Eight, Four, Three-low [T♥ 9♣ 8♦ 4♦ 3♦] [K♣ 8♥]
	// Player 6: [T♣ Q♦ A♠ 7♥ Q♣ 7♦ 2♥] Queen, Ten, Seven, Two, Ace-low [Q♦ T♣ 7♥ 2♥ A♠] [Q♣ 7♦]
	// Result:   Player 2 wins with Seven, Five, Four, Three, Two-low
	// ------ Razz 5 ------
	// Player 1: [3♦ 4♦ 5♦ J♣ 4♥ K♥ 8♣] Jack, Eight, Five, Four, Three-low [J♣ 8♣ 5♦ 4♦ 3♦] [K♥ 4♥]
	// Player 2: [T♥ J♠ K♠ 2♣ 4♣ 5♠ 2♦] Jack, Ten, Five, Four, Two-low [J♠ T♥ 5♠ 4♣ 2♣] [K♠ 2♦]
	// Player 3: [A♣ 9♠ T♠ 3♠ K♣ 8♦ A♥] Ten, Nine, Eight, Three, Ace-low [T♠ 9♠ 8♦ 3♠ A♣] [A♥ K♣]
	// Player 4: [7♦ 3♣ 8♠ 7♣ 6♦ 6♥ 6♣] Pair, Sixes, kickers Eight, Seven, Three [6♦ 6♥ 8♠ 7♦ 3♣] [7♣ 6♣]
	// Player 5: [5♣ Q♠ J♥ 2♠ A♠ 8♥ 4♠] Eight, Five, Four, Two, Ace-low [8♥ 5♣ 4♠ 2♠ A♠] [Q♠ J♥]
	// Player 6: [6♠ 7♠ 7♥ 2♥ 9♦ K♦ T♦] Ten, Nine, Seven, Six, Two-low [T♦ 9♦ 7♠ 6♠ 2♥] [K♦ 7♥]
	// Result:   Player 5 wins with Eight, Five, Four, Two, Ace-low
}

func ExampleType_badugi() {
	for i, game := range []struct {
		seed    int64
		players int
	}{
		{119, 2},
		{321, 5},
		{408, 6},
		{455, 6},
		{1113, 6},
	} {
		// note: use a real random source
		r := rand.New(rand.NewSource(game.seed))
		pockets, _ := cardrank.Badugi.Deal(r, 1, game.players)
		evs := cardrank.Badugi.EvalPockets(pockets, nil)
		fmt.Printf("------ Badugi %d ------\n", i+1)
		for j := range game.players {
			desc := evs[j].Desc(false)
			fmt.Printf("Player %d: %b %s %b %b\n", j+1, pockets[j], desc, desc.Best, desc.Unused)
		}
		order, pivot := cardrank.Order(evs, false)
		desc := evs[order[0]].Desc(false)
		if pivot == 1 {
			fmt.Printf("Result:   Player %d wins with %s\n", order[0]+1, desc)
		} else {
			var s []string
			for j := range pivot {
				s = append(s, strconv.Itoa(order[j]+1))
			}
			fmt.Printf("Result:   Players %s push with %s\n", strings.Join(s, ", "), desc)
		}
	}
	// Output:
	// ------ Badugi 1 ------
	// Player 1: [K♥ J♣ A♥ Q♠] Queen, Jack, Ace-low [Q♠ J♣ A♥] [K♥]
	// Player 2: [7♣ 4♣ 5♠ 2♠] Four, Two-low [4♣ 2♠] [7♣ 5♠]
	// Result:   Player 1 wins with Queen, Jack, Ace-low
	// ------ Badugi 2 ------
	// Player 1: [3♠ 3♦ T♠ Q♠] Ten, Three-low [T♠ 3♦] [Q♠ 3♠]
	// Player 2: [6♦ Q♣ 8♥ 6♣] Queen, Eight, Six-low [Q♣ 8♥ 6♦] [6♣]
	// Player 3: [Q♦ K♠ 8♣ A♥] King, Queen, Eight, Ace-low [K♠ Q♦ 8♣ A♥] []
	// Player 4: [K♦ T♦ 8♦ 4♥] Eight, Four-low [8♦ 4♥] [K♦ T♦]
	// Player 5: [J♦ 2♥ Q♥ 6♠] Jack, Six, Two-low [J♦ 6♠ 2♥] [Q♥]
	// Result:   Player 3 wins with King, Queen, Eight, Ace-low
	// ------ Badugi 3 ------
	// Player 1: [K♠ Q♠ 4♣ J♦] Queen, Jack, Four-low [Q♠ J♦ 4♣] [K♠]
	// Player 2: [J♠ 3♣ 8♥ 2♠] Eight, Three, Two-low [8♥ 3♣ 2♠] [J♠]
	// Player 3: [3♠ T♠ 2♣ Q♦] Queen, Three, Two-low [Q♦ 3♠ 2♣] [T♠]
	// Player 4: [5♣ 5♥ T♦ 2♦] Five, Two-low [5♥ 2♦] [T♦ 5♣]
	// Player 5: [7♠ 3♥ 6♠ A♣] Six, Three, Ace-low [6♠ 3♥ A♣] [7♠]
	// Player 6: [4♠ 8♦ K♦ T♣] Ten, Eight, Four-low [T♣ 8♦ 4♠] [K♦]
	// Result:   Player 5 wins with Six, Three, Ace-low
	// ------ Badugi 4 ------
	// Player 1: [6♠ K♥ A♣ 8♣] King, Six, Ace-low [K♥ 6♠ A♣] [8♣]
	// Player 2: [Q♥ 4♥ J♣ 5♥] Jack, Four-low [J♣ 4♥] [Q♥ 5♥]
	// Player 3: [2♣ 6♥ 5♣ Q♠] Queen, Six, Two-low [Q♠ 6♥ 2♣] [5♣]
	// Player 4: [9♠ J♥ K♠ J♠] Jack, Nine-low [J♥ 9♠] [K♠ J♠]
	// Player 5: [3♦ 4♦ K♣ 8♦] King, Three-low [K♣ 3♦] [8♦ 4♦]
	// Player 6: [T♣ Q♦ A♠ 7♥] Queen, Ten, Seven, Ace-low [Q♦ T♣ 7♥ A♠] []
	// Result:   Player 6 wins with Queen, Ten, Seven, Ace-low
	// ------ Badugi 5 ------
	// Player 1: [3♦ 4♦ 5♦ J♣] Jack, Three-low [J♣ 3♦] [5♦ 4♦]
	// Player 2: [T♥ J♠ K♠ 2♣] Jack, Ten, Two-low [J♠ T♥ 2♣] [K♠]
	// Player 3: [A♣ 9♠ T♠ 3♠] Three, Ace-low [3♠ A♣] [T♠ 9♠]
	// Player 4: [7♦ 3♣ 8♠ 7♣] Eight, Seven, Three-low [8♠ 7♦ 3♣] [7♣]
	// Player 5: [5♣ Q♠ J♥ 2♠] Jack, Five, Two-low [J♥ 5♣ 2♠] [Q♠]
	// Player 6: [6♠ 7♠ 7♥ 2♥] Six, Two-low [6♠ 2♥] [7♥ 7♠]
	// Result:   Player 4 wins with Eight, Seven, Three-low
}

func ExampleOddsCalc() {
	pockets := [][]cardrank.Card{
		cardrank.Must("Ah As Jc Qs"),
		cardrank.Must("3h 2h Ks Tc"),
	}
	board := cardrank.Must("6h 6s Jh")
	odds, _, ok := cardrank.Omaha.Odds(context.Background(), pockets, board)
	if !ok {
		panic("unable to calculate odds")
	}
	for i := range pockets {
		fmt.Printf("%d: %*v\n", i, i, odds)
	}
	// Output:
	// 0: 66.1% (542/820)
	// 1: 33.9% (278/820)
}

func ExampleExpValueCalc() {
	pocket, board := cardrank.Must("Kh 3h"), cardrank.Must("Ah 8h 3c")
	expv, ok := cardrank.Holdem.ExpValue(context.Background(), pocket, cardrank.WithBoard(board))
	if !ok {
		panic("unable to calculate expected value")
	}
	fmt.Println("expected value:", expv)
	// Output:
	// expected value: 75.6% (802371,13659/1070190)
}

func Example_computerHand() {
	pocket := cardrank.Must("Qh 7s")
	expv, ok := cardrank.Holdem.ExpValue(context.Background(), pocket)
	if !ok {
		panic("unable to calculate expected value")
	}
	fmt.Println("expected value:", expv)
	// Output:
	// expected value: 51.8% (1046780178,78084287/2097572400)
}

func Example_holdemPreflop() {
	pocket := cardrank.Must("2h 2d")
	ev := cardrank.EvalOf(cardrank.Holdem)
	ev.Eval(pocket, nil)
	fmt.Println("rank:", ev.HiRank)
	// Output:
	// rank: Pair
}

func Example_omahaPreflop() {
	pocket := cardrank.Must("Ah 7h Ad Jh")
	ev := cardrank.EvalOf(cardrank.Omaha)
	ev.Eval(pocket, nil)
	fmt.Println("rank:", ev.HiRank)
	// Output:
	// rank: Pair
}

/*
func ExampleType_calc() {
	board := cardrank.Must("Ah Ks 3c Qh")
	pockets := [][]cardrank.Card{
		cardrank.Must("Ac As"),
		cardrank.Must("Kd Kc"),
		cardrank.Must("3h Kh"),
	}
	odds, _, ok := cardrank.Holdem.CalcPockets(context.Background(), pockets, board)
	if !ok {
		panic("unable to finish calculating odds")
	}
	for i := range len(pockets) {
		fmt.Printf("%d: %*s - %*b\n", i, i, odds, i, odds)
	}
	// Output:
	// 0: 78.6% (33/42) - [Q♠ J♠ T♠ 9♠ 8♠ 7♠ 6♠ 5♠ 4♠ 3♠ 2♠ A♦ Q♦ J♦ T♦ 9♦ 8♦ 7♦ 6♦ 5♦ 4♦ 3♦ 2♦ Q♣ J♣ T♣ 9♣ 8♣ 7♣ 6♣ 5♣ 4♣ 2♣]
	// 1: 0.0% (0/42) - none
	// 2: 21.4% (9/42) - any ♥
}
*/
