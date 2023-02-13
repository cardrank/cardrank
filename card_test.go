package cardrank

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

func TestParse(t *testing.T) {
	tests := []struct {
		s   string
		exp []Card
		err error
	}{
		{"", nil, nil},
		{"z", nil, ErrInvalidCard},
		{"vs", nil, ErrInvalidCard},
		{"av", nil, ErrInvalidCard},
		{"AsKs", []Card{New(Ace, Spade), New(King, Spade)}, nil},
		{"As Ks", []Card{New(Ace, Spade), New(King, Spade)}, nil},
		{" 🂬   a♣  🃚  🂸  td ", []Card{New(Jack, Spade), New(Ace, Club), New(Ten, Club), New(Eight, Heart), New(Ten, Diamond)}, nil},
		{"10D 10C 10S 10h", []Card{New(Ten, Diamond), New(Ten, Club), New(10, Spade), New(10, Heart)}, nil},
	}
	for i, test := range tests {
		hand, err := Parse(test.s)
		switch {
		case test.err != nil && !errors.Is(err, test.err):
			t.Fatalf("test %d %q expected error %v, got: %v", i, test.s, test.err, err)
		case errors.Is(test.err, err):
			continue
		}
		if n, exp := len(hand), len(test.exp); n != exp {
			t.Fatalf("test %d %q expected len(hand) == %d, got: %d", i, test.s, exp, n)
		}
		for j, c := range hand {
			if exp := test.exp[j]; c != exp {
				t.Errorf("test %d %q card %d expected %s, got %s", i, test.s, j, exp, c)
			}
		}
	}
}

func TestCardUnmarshal(t *testing.T) {
	v := struct {
		Card Card
	}{
		Card: Must("Ah")[0],
	}
	switch err := json.Unmarshal([]byte(`{"card": "bz"}`), &v); {
	case err == nil || !errors.Is(err, ErrInvalidCard):
		t.Errorf("expected %v, got: %v", ErrInvalidCard, err)
	case v.Card != InvalidCard:
		t.Errorf("expected %d, got: %d", InvalidCard, v.Card)
	}
	var hand []Card
	if err := json.Unmarshal([]byte(`["Ah","Kh","Qh","Jh","Th"]`), &hand); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	for i, c := range Must("Ah Kh Qh Jh Th") {
		if !contains(hand, c) {
			t.Errorf("test %d hand does not contain %s", i, c)
		}
	}
}

func TestCardMarshal(t *testing.T) {
	if _, err := json.Marshal(struct{ Card Card }{
		Card: InvalidCard,
	}); !errors.Is(err, ErrInvalidCard) {
		t.Errorf("expected %v, got: %v", ErrInvalidCard, err)
	}
	buf, err := json.Marshal(Must("Ah Kh Qh Jh Th"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if s, exp := string(buf), `["Ah","Kh","Qh","Jh","Th"]`; s != exp {
		t.Errorf("expected %q == %q", exp, s)
	}
}

func TestCardIndex(t *testing.T) {
	i := 0
	for _, s := range []Suit{Spade, Heart, Diamond, Club} {
		for r := Two; r <= Ace; r++ {
			c := New(r, s)
			d := FromIndex(i)
			if d == InvalidCard {
				t.Fatalf("expected valid card")
			}
			if d != c {
				t.Errorf("expected %s to equal %s", c, d)
			}
			if n, exp := d.SuitIndex(), c.SuitIndex(); n != exp {
				t.Errorf("card %s expected suit index %d, got: %d", c, exp, n)
			}
			if n, exp := d.RankIndex(), c.RankIndex(); n != exp {
				t.Errorf("card %s expected rank index %d, got: %d", c, exp, n)
			}
			if n, exp := c.Index(), i; n != exp {
				t.Errorf("card %s expected index %d, got: %d", c, exp, n)
			}
			if unshuffledFrench[i] != c {
				t.Errorf("expected unshuffled[%d] == %s, has: %s", i, c, unshuffledFrench[i])
			}
			i++
		}
	}
}

func TestFromRune(t *testing.T) {
	tests := []struct {
		r   rune
		exp string
	}{
		{'🂱', "Ah"},
		{'🂮', "Ks"},
		{'🃝', "Qc"},
		{'🃌', "Jd"},
		{'🃋', "Jd"},
		{'🂺', "Th"},
		{'🂩', "9s"},
		{'🃘', "8c"},
		{'🃇', "7d"},
		{'🂶', "6h"},
		{'🂥', "5s"},
		{'🃔', "4c"},
		{'🃃', "3d"},
		{'🂲', "2h"},
	}
	for i, test := range tests {
		c := FromRune(test.r)
		if c == InvalidCard {
			t.Fatalf("test %d expected valid card", i)
		}
		if s := c.String(); s != test.exp {
			t.Errorf("test %d expected s == %s, got: %s", i, test.exp, s)
		}
	}
}

func TestCardFormat(t *testing.T) {
	tests := []struct {
		r   string
		s   string
		exp Card
		c   rune
		v   string
	}{
		{"A", "hH♥♡", 0x10002c29, '🂱', "Ace of Hearts"},
		{"K", "sS♠♤", 0x08001b25, '🂮', "King of Spades"},
		{"Q", "cC♣♧", 0x04008a1f, '🃝', "Queen of Clubs"},
		{"J", "dD♦♢", 0x0200491d, '🃋', "Jack of Diamonds"},
		{"T", "hH♥♡", 0x01002817, '🂺', "Ten of Hearts"},
		{"9", "sS♠♤", 0x00801713, '🂩', "Nine of Spades"},
		{"8", "cC♣♧", 0x00408611, '🃘', "Eight of Clubs"},
		{"7", "dD♦♢", 0x0020450d, '🃇', "Seven of Diamonds"},
		{"6", "hH♥♡", 0x0010240b, '🂶', "Six of Hearts"},
		{"5", "sS♠♤", 0x00081307, '🂥', "Five of Spades"},
		{"4", "cC♣♧", 0x00048205, '🃔', "Four of Clubs"},
		{"3", "dD♦♢", 0x00024103, '🃃', "Three of Diamonds"},
		{"2", "hH♥♡", 0x00012002, '🂲', "Two of Hearts"},
	}
	for i, test := range tests {
		z := []rune(test.r)
		if unicode.IsUpper(z[0]) {
			z = append(z, unicode.ToLower(z[0]))
		}
		for _, r := range z {
			for _, s := range test.s {
				v := string(r) + string(s)
				c := FromString(v)
				if c == InvalidCard {
					t.Fatalf("test %d expected valid card", i)
				}
				if c != test.exp {
					t.Errorf("test %d %q expected %d, got: %d", i, v, test.exp, c)
				}
			}
		}
		c := FromRune(test.c)
		if c == InvalidCard {
			t.Fatalf("test %d expected valid card", i)
		}
		if c != test.exp {
			t.Errorf("test %d expected %c to be %d, got: %d (%s)", i, test.c, test.exp, c, c)
		}
		if s, exp := fmt.Sprintf("%s", c), test.r[0:1]+test.s[0:1]; s != exp {
			t.Errorf("test %d expected %%s to be %q, got: %q", i, exp, s)
		}
		if s, exp := fmt.Sprintf("%S", c), test.r[0:1]+test.s[1:2]; s != exp {
			t.Errorf("test %d expected %%S to be %q, got: %q", i, exp, s)
		}
		if s, exp := fmt.Sprintf("%q", c), `"`+test.r[0:1]+test.s[0:1]+`"`; s != exp {
			t.Errorf("test %d expected %%q to be %q, got: %q", i, exp, s)
		}
		if s, exp := fmt.Sprintf("%b", c), test.r[0:1]+test.s[2:5]; s != exp {
			t.Errorf("test %d expected %%b to be %q, got: %q", i, exp, s)
		}
		if s, exp := fmt.Sprintf("%h", c), test.r[0:1]+test.s[5:8]; s != exp {
			t.Errorf("test %d expected %%h to be %q, got: %q", i, exp, s)
		}
		if s, exp := fmt.Sprintf("%r %u %B", c, c, c), test.r[0:1]+" "+test.s[0:1]+" "+test.s[2:5]; s != exp {
			t.Errorf("test %d expected %%r %%u %%B to be %q, got: %q", i, exp, s)
		}
		if s, exp := fmt.Sprintf("%c", c), string(test.c); s != exp {
			t.Errorf("test %d expected %%c to be %q, got: %q", i, exp, s)
		}
		if c.Rank() == Jack {
			if s, exp := fmt.Sprintf("%C", c), string(test.c+1); s != exp {
				t.Errorf("test %d expected %%C to be %q, got: %q", i, exp, s)
			}
		}
		if s, exp := fmt.Sprintf("%n of %l", c, c), strings.ToLower(test.v); s != exp {
			t.Errorf("test %d expected %%n of %%l to be %q, got: %q", i, exp, s)
		}
		if s, exp := fmt.Sprintf("%N of %L", c, c), test.v; s != exp {
			t.Errorf("test %d expected %%N of %%L to be %q, got: %q", i, exp, s)
		}
		if s, exp := fmt.Sprintf("%d", c), strconv.Itoa(int(c)); s != exp {
			t.Errorf("test %d expected %%d to be %q, got: %q", i, exp, s)
		}
	}
}

func contains(v []Card, c Card) bool {
	for i := 0; i < len(v); i++ {
		if v[i] == c {
			return true
		}
	}
	return false
}
