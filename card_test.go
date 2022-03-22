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
		{"vs", nil, ErrInvalidCardRank},
		{"av", nil, ErrInvalidCardSuit},
		{"AsKs", []Card{MustCard("As"), MustCard("Ks")}, nil},
		{"As Ks", []Card{MustCard("As"), MustCard("Ks")}, nil},
		{" 🂬   a♣  🃚  🂸  td ", []Card{MustCard("Js"), MustCard("As"), MustCard("Ts"), MustCard("8h"), MustCard("Td")}, nil},
		{"10D 10C 10S 10h", []Card{MustCard("10d"), MustCard("10c"), MustCard("10s"), MustCard("10h")}, nil},
	}
	for i, test := range tests {
		hand, err := Parse(test.s)
		switch {
		case test.err != nil && !errors.Is(test.err, err):
			t.Fatalf("test %d Parse(%q) expected error %v, got: %v", i, test.s, test.err, err)
		case errors.Is(test.err, err):
			continue
		}
		if n, exp := len(hand), len(test.exp); n != exp {
			t.Fatalf("test %d expected len(hand) == %d, got: %d", i, exp, n)
		}
		for j, c := range hand {
			if exp := test.exp[j]; c != exp {
				t.Errorf("test %d card %d expected %s, got %s", i, j, exp, c)
			}
		}
	}
}

func TestCardBitRank(t *testing.T) {
	hand := Must("Ks")
	if r, exp := hand[0].BitRank(), uint32(0x800); r != exp {
		t.Errorf("expected %d == %d", exp, r)
	}
}

func TestCardUnmarshal(t *testing.T) {
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
	buf, err := json.Marshal(Must("Ah Kh Qh Jh Th"))
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if s, exp := string(buf), `["Ah","Kh","Qh","Jh","Th"]`; s != exp {
		t.Errorf("expected %q == %q", exp, s)
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
		c, err := FromRune(test.r)
		if err != nil {
			t.Fatalf("test %d expected no error, got: %v", i, err)
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
				c, err := FromString(v)
				if err != nil {
					t.Fatalf("test %d %q expected no error, got: %v", i, v, err)
				}
				if c != test.exp {
					t.Errorf("test %d %q expected %d, got: %d", i, v, test.exp, c)
				}
			}
		}
		c, err := FromRune(test.c)
		if err != nil {
			t.Errorf("test %d expected no error, got: %v", i, err)
		}
		if c != test.exp {
			t.Errorf("test %d expected %c to be %d, got: %d (%s)", i, test.c, test.exp, c, c)
		}
		if s, exp := fmt.Sprintf("%s", c), string(test.r[0])+string(test.s[0]); s != exp {
			t.Errorf("test %d expected %%s to be %q, got: %q", i, exp, s)
		}
		if s, exp := fmt.Sprintf("%S", c), string(test.r[0])+string(test.s[1]); s != exp {
			t.Errorf("test %d expected %%S to be %q, got: %q", i, exp, s)
		}
		if s, exp := fmt.Sprintf("%q", c), `"`+string(test.r[0])+string(test.s[0])+`"`; s != exp {
			t.Errorf("test %d expected %%q to be %q, got: %q", i, exp, s)
		}
		if s, exp := fmt.Sprintf("%b", c), string(test.r[0])+string(test.s[2:5]); s != exp {
			t.Errorf("test %d expected %%b to be %q, got: %q", i, exp, s)
		}
		if s, exp := fmt.Sprintf("%h", c), string(test.r[0])+string(test.s[5:8]); s != exp {
			t.Errorf("test %d expected %%h to be %q, got: %q", i, exp, s)
		}
		if s, exp := fmt.Sprintf("%r %u %B", c, c, c), string(test.r[0])+" "+string(test.s[0])+" "+string(test.s[2:5]); s != exp {
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
		if s, exp := fmt.Sprintf("%n of %l", c, c), strings.ToLower(string(test.v)); s != exp {
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
