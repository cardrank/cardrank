// Package paycalc contains a tournament payout tables.
package paycalc

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// tables are tournament payout tables.
var tables = make(map[Type]*Table)

// Register registers a tournament payout table.
func Register(typ Type, t *Table) error {
	if _, ok := tables[typ]; ok {
		return fmt.Errorf("type %d already registered", typ)
	}
	tables[typ] = t
	return nil
}

// RegisterReader registers a tournament payout table read as CSV from the
// reader.
func RegisterReader(typ Type, r io.Reader, top float64, name string) error {
	t, err := LoadReader(r, top, name)
	if err != nil {
		return err
	}
	return Register(typ, t)
}

// RegisterBytes registers a tournament payout table read as CSV from buf.
func RegisterBytes(typ Type, buf []byte, top float64, name string) error {
	return RegisterReader(typ, bytes.NewReader(buf), top, name)
}

// Init initializes the tournament payout tables.
func Init() error {
	if err := RegisterBytes(Top10, top10, 0.10, "top10"); err != nil {
		return err
	}
	if err := RegisterBytes(Top15, top15, 0.15, "top15"); err != nil {
		return err
	}
	if err := RegisterBytes(Top20, top20, 0.20, "top20"); err != nil {
		return err
	}
	return nil
}

// Type is a tournament payout table type.
type Type int

// Top tournament payout table types.
const (
	Top00 Type = iota
	Top10
	Top15
	Top20
)

// Top returns the proportion of paid rankings for the tournament payout table.
func (typ Type) Top() float64 {
	if t, ok := tables[typ]; ok {
		return t.Top()
	}
	return 0.0
}

// Name returns the name of the tournament payout table.
func (typ Type) Name() string {
	if t, ok := tables[typ]; ok {
		return t.Name()
	}
	return ""
}

// Title returns the title of the tournament payout table.
func (typ Type) Title() string {
	if t, ok := tables[typ]; ok {
		return t.Title()
	}
	return ""
}

// Format satisfies the [fmt.Formatter] interface.
func (typ Type) Format(f fmt.State, verb rune) {
	if t, ok := tables[typ]; ok {
		t.Format(f, verb)
	}
}

// EntriesMax returns the max entries for the tournament payout table.
func (typ Type) EntriesMax() int {
	if t, ok := tables[typ]; ok {
		return t.EntriesMax()
	}
	return 0
}

// LevelsMax returns the max levels for the tournament payout table for the
// specified entries.
func (typ Type) LevelsMax(entries int) int {
	if t, ok := tables[typ]; ok {
		return t.LevelsMax(entries)
	}
	return 0
}

// RankingMax returns the max ranking for the tournament payout table.
func (typ Type) RankingMax() int {
	if t, ok := tables[typ]; ok {
		return t.RankingMax()
	}
	return 0
}

// WriteTo writes the tournament payout table to w, formatting cells with f,
// breaking line output when early is true.
func (typ Type) WriteTo(w io.Writer, f func(interface{}, int, bool) string, divider func(int, int) string, early bool) error {
	if t, ok := tables[typ]; ok {
		return t.WriteTo(w, f, divider, early)
	}
	return nil
}

// WriteTable writes a plain text version of the tournament payout table, along
// with optional title, to w.
func (typ Type) WriteTable(w io.Writer, title bool) error {
	if t, ok := tables[typ]; ok {
		return t.WriteTable(w, title)
	}
	return nil
}

// WriteCSV writes a CSV version of the tournament payout table to w.
func (typ Type) WriteCSV(w io.Writer) error {
	if t, ok := tables[typ]; ok {
		return t.WriteCSV(w)
	}
	return nil
}

// WriteMarkdown writes a Markdown formatted version of the tournament payout
// table to w.
func (typ Type) WriteMarkdown(w io.Writer, header int) error {
	if t, ok := tables[typ]; ok {
		return t.WriteMarkdown(w, header)
	}
	return nil
}

// Entries returns the entries column in the tournament payout table.
func (typ Type) Entries(entries int) int {
	if t, ok := tables[typ]; ok {
		return t.Entries(entries)
	}
	return 0
}

// EntriesTitle returns the column title for the specified entries in the
// tournament payout table.
func (typ Type) EntriesTitle(entries int) string {
	if t, ok := tables[typ]; ok {
		return t.EntriesTitle(entries)
	}
	return ""
}

// LevelsTitle returns the level (row) title for rank n in the tournament
// payout table.
func (typ Type) LevelsTitle(n int) string {
	if t, ok := tables[typ]; ok {
		return t.LevelsTitle(n)
	}
	return ""
}

// Levels returns the levels (rows) of the tournament table from [start, end).
func (typ Type) Levels(start, end int) []int {
	if t, ok := tables[typ]; ok {
		return t.Levels(start, end)
	}
	return nil
}

// Rankings returns the levels for rankings from [start, end).
func (typ Type) Rankings(start, end int) []int {
	if t, ok := tables[typ]; ok {
		return t.Rankings(start, end)
	}
	return nil
}

// Amounts returns the tournament payout amounts for positions for [start, end)
// as determined by the number of entries.
func (typ Type) Amounts(start, end, entries int) []float64 {
	if t, ok := tables[typ]; ok {
		return t.Amounts(start, end, entries)
	}
	return nil
}

// Payouts returns the tournament payouts for positions from [start, end) based
// on the number of entries.
func (typ Type) Payouts(start, end, entries int, buyin, guaranteed int64, rake float64) []int64 {
	if t, ok := tables[typ]; ok {
		return t.Payouts(start, end, entries, buyin, guaranteed, rake)
	}
	return nil
}

// Stakes returns slice of the ranges of the paid levels in the form of [low,
// high), corresponding tournament payout table amount f, and calculated amount
// for each position.
func (typ Type) Stakes(entries int, buyin, guaranteed int64, rake float64) ([][2]int, []float64, []int64) {
	if t, ok := tables[typ]; ok {
		return t.Stakes(entries, buyin, guaranteed, rake)
	}
	return nil, nil, nil
}

// Amount returns the tournament payout value for position n.
func (typ Type) Amount(n, entries int) float64 {
	if t, ok := tables[typ]; ok {
		return t.Amount(n, entries)
	}
	return 0
}

// Payout returns the tournament payout for position n.
func (typ Type) Payout(n, entries int, buyin, guaranteed int64, rake float64) int64 {
	if t, ok := tables[typ]; ok {
		return t.Payout(n, entries, buyin, guaranteed, rake)
	}
	return 0
}

// At returns the amount at row, col of the tournament payout table.
func (typ Type) At(row, col int) float64 {
	if t, ok := tables[typ]; ok {
		return t.At(row, col)
	}
	return 0.0
}

// MarshalText satisfies the [encoding.TextMarshaler] interface.
func (typ Type) MarshalText() ([]byte, error) {
	return []byte(typ.Name()), nil
}

// UnmarshalText satisfies the [encoding.TextUnmarshaler] interface.
func (typ *Type) UnmarshalText(buf []byte) error {
	name := strings.ToLower(string(buf))
	for t, tbl := range tables {
		if strings.ToLower(tbl.name) == name {
			*typ = t
			return nil
		}
	}
	return fmt.Errorf("invalid type %q", string(buf))
}

// EntriesTitle formats the entries title of last, n.
func EntriesTitle(last, n int) string {
	switch {
	case last < 0 && n < 0:
		return ""
	case last == 0:
		return strconv.Itoa(n) + "+"
	case n-last < 1:
		return strconv.Itoa(last)
	}
	return strconv.Itoa(last+1) + "-" + strconv.Itoa(n)
}

// LevelsTitle formats the levels title of last, n.
func LevelsTitle(last, n int) string {
	if n-last < 2 {
		return strconv.Itoa(n) + ord(n)
	}
	return fmt.Sprintf("%d-%d", last+1, n)
}

// EpsilonEqual returns true when a and b are within epsilon.
func EpsilonEqual[R, S, T Ordered](a R, b S, epsilon T) bool {
	return math.Abs(float64(a)-float64(b)) <= float64(epsilon)
}

// Calc calculates a payout for the percentage f, based on entries, buyin,
// guaranteed amount, and rake. Uses [Round] to round to [Precision].
func Calc(f float64, entries int, buyin, guaranteed int64, rake float64) int64 {
	return int64(Round(f * float64(Total(entries, buyin, guaranteed, rake))))
}

// Total calculates the total payout based on entries, buyin, guaranteed
// amount, and rake. Uses [Round] to round to [Precision].
func Total(entries int, buyin, guaranteed int64, rake float64) int64 {
	amt := int64(entries) * buyin
	if 0.0 < rake && rake < 1.0 {
		amt = max(guaranteed, amt-int64(Round(rake*float64(amt))))
	}
	return amt
}

// Precision is the precision amount used by [Calc].
var Precision = int(math.Pow(10, 2))

// Round is the round implementation used by [Calc]. Rounds to [Precision].
var Round = func(f float64) float64 {
	return math.Round(f*float64(Precision)) / float64(Precision)
}

// ord returns the ordinal suffix for n.
func ord(n int) string {
	if 11 <= n && n <= 13 {
		return "th"
	}
	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	}
	return "th"
}

// tournament payout tables.
var (
	//go:embed top10.csv
	top10 []byte
	//go:embed top15.csv
	top15 []byte
	//go:embed top20.csv
	top20 []byte
)
