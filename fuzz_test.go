package cardrank

import (
	"slices"
	"testing"
)

func FuzzParse(f *testing.F) {
	for _, s := range []string{
		"Ah Kh Qh Jh Th",
		"10s 10c",
		"J♤ K♠ q♡",
		"🂡🂱🃁🃑",
		"🂬",
		"1",
		"Ax",
		"",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, err := Parse(s)
		if err != nil {
			return
		}
		for _, c := range v {
			if c == InvalidCard || c.Rank() > Ace {
				t.Fatalf("Parse(%q) returned invalid card %d", s, c)
			}
		}
		// round trip through the canonical string representation
		u, err := Parse(joinCards(v))
		if err != nil {
			t.Fatalf("expected no error re-parsing %v, got: %v", v, err)
		}
		if !slices.Equal(v, u) {
			t.Fatalf("expected %v == %v", v, u)
		}
	})
}

func FuzzFromString(f *testing.F) {
	for _, s := range []string{"Ah", "10d", "tc", "K♠", "🃑", "", "1", "10", "Zz"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		c := FromString(s)
		var u Card
		err := u.UnmarshalText([]byte(s))
		switch {
		case c == InvalidCard && err == nil:
			t.Fatalf("expected UnmarshalText(%q) error for invalid card", s)
		case c != InvalidCard && err != nil:
			t.Fatalf("expected UnmarshalText(%q) to succeed, got: %v", s, err)
		case c == InvalidCard:
			return
		case c != u:
			t.Fatalf("expected FromString(%q) == UnmarshalText, %s != %s", s, c, u)
		}
		buf, err := c.MarshalText()
		if err != nil {
			t.Fatalf("expected no error marshaling %s, got: %v", c, err)
		}
		if d := FromString(string(buf)); d != c {
			t.Fatalf("expected round trip of %q to be %s, got: %s", s, c, d)
		}
	})
}

func FuzzFromIndex(f *testing.F) {
	for _, i := range []int{-53, -14, -1, 0, 12, 13, 51, 52, 1 << 40} {
		f.Add(i)
	}
	f.Fuzz(func(t *testing.T, i int) {
		c := FromIndex(i)
		switch valid := 0 <= i && i < 52; {
		case !valid && c != InvalidCard:
			t.Fatalf("expected FromIndex(%d) to be invalid, got: %s", i, c)
		case valid && c.Index() != i:
			t.Fatalf("expected FromIndex(%d).Index() == %d, got: %d", i, i, c.Index())
		}
	})
}

// joinCards returns the cards as a space separated string.
func joinCards(v []Card) string {
	var b []byte
	for i, c := range v {
		if i != 0 {
			b = append(b, ' ')
		}
		b = append(b, c.RankByte(), c.SuitByte())
	}
	return string(b)
}
