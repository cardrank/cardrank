package paycalc

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// Table is a tournament payout table.
type Table struct {
	top     float64
	name    string
	levels  []int
	entries []int
	amounts [][]float64
}

// New creates a new tournament payout table.
func New(top float64, name string, levels, entries []int, amounts [][]float64) (*Table, error) {
	rows, cols := len(levels), len(entries)
	if rows < 1 || cols < 1 {
		return nil, errors.New("invalid table")
	}
	// check levels
	for i, last := 0, 0; i < rows; i++ {
		switch n := len(amounts[i]); {
		case cols < n:
			return nil, fmt.Errorf("level %d has %d columns: more than %d", i, len(amounts[i]), cols)
		case levels[i] <= last:
			return nil, fmt.Errorf("levels are not ordered low to high: row %d value %d must be <= %d ", i, levels[i], last)
		case n < cols:
			amounts[i] = append(amounts[i], make([]float64, cols-n)...)
		}
		last = levels[i]
	}
	// check entries
	for i, last := 0, 9999; i < cols; i++ {
		if last <= entries[i] {
			return nil, fmt.Errorf("entries are not ordered high to low: col %d value %d must be <= %d", i, entries[i], last)
		}
		last = entries[i]
	}
	// calc sums
	sums := make([]float64, cols)
	for i := 0; i < rows; i++ {
		n := 1
		if 0 < i {
			n = levels[i] - levels[i-1]
		}
		for j := 0; j < cols; j++ {
			sums[j] += float64(n) * amounts[i][j]
		}
	}
	// check sums
	for i := 0; i < cols; i++ {
		if !EpsilonEqual(1.0, sums[i], 0.0000000001) {
			return nil, fmt.Errorf("entry column %d sum %f != 1.0", i, sums[i])
		}
	}
	return &Table{
		top:     top,
		name:    name,
		levels:  levels,
		entries: entries,
		amounts: amounts,
	}, nil
}

// LoadReader loads a CSV formatted tournament payout table from the reader.
func LoadReader(rdr io.Reader, top float64, name string) (*Table, error) {
	r := csv.NewReader(rdr)
	r.FieldsPerRecord = -1
	lines, err := r.ReadAll()
	switch {
	case err != nil:
		return nil, fmt.Errorf("unable to parse table: %w", err)
	case len(lines) < 2 || len(lines[0]) < 2:
		return nil, errors.New("invalid table")
	case lines[0][0] != reCell:
		return nil, fmt.Errorf("row 0 col 0 must be %q ", reCell)
	}
	rows, cols := len(lines)-1, len(lines[0])-1
	levels, entries, amounts := make([]int, rows), make([]int, cols), make([][]float64, rows)
	// parse levels
	for i := 1; i <= rows; i++ {
		switch n := len(lines[i]) - 1; {
		case cols < n:
			return nil, fmt.Errorf("row %d has %d columns: more than %d", i, n, cols)
		case n < cols:
			lines[i] = append(lines[i], make([]string, cols-n)...)
		}
		if levels[i-1], err = parseLevels(lines[i][0]); err != nil {
			return nil, fmt.Errorf("unable to parse row %d header %q: %w", i, lines[i][0], err)
		}
		amounts[i-1] = make([]float64, cols)
	}
	// parse entries
	for i := 1; i <= cols; i++ {
		if entries[i-1], err = parseEntries(lines[0][i]); err != nil {
			return nil, fmt.Errorf("unable to parse col %d header %q: %w", i, lines[0][i], err)
		}
	}
	// parse amounts
	for i := 1; i <= rows; i++ {
		for j := 1; j <= cols; j++ {
			cell := strings.TrimSpace(lines[i][j])
			if cell == "" {
				continue
			}
			if amounts[i-1][j-1], err = strconv.ParseFloat(cell, 64); err != nil {
				return nil, fmt.Errorf("unable to parse row %d col %d cell %q: %w", i, j, lines[i][j], err)
			}
			amounts[i-1][j-1] /= 100.0
		}
	}
	return New(top, name, levels, entries, amounts)
}

// LoadBytes loads CSV formatted tournament payout table from buf.
func LoadBytes(buf []byte, top float64, name string) (*Table, error) {
	return LoadReader(bytes.NewReader(buf), top, name)
}

// Top returns the proportion of paid rankings for the tournament payout table.
func (t *Table) Top() float64 {
	return t.top
}

// Name returns the name of the tournament payout table.
func (t *Table) Name() string {
	return t.name
}

// Title returns the title of the tournament payout table.
func (t *Table) Title() string {
	return fmt.Sprintf("Top %d%% Paid", int(t.Top()*100))
}

// Format satisfies the [fmt.Formatter] interface.
func (t *Table) Format(f fmt.State, verb rune) {
	switch verb {
	case 'd':
		fmt.Fprintf(f, "%d", int(t.Top()*100))
	case 'f':
		fmt.Fprintf(f, "%f", t.Top())
	case 't':
		fmt.Fprint(f, t.Title())
	case 'c':
		_ = t.WriteCSV(f)
	case 'm':
		_ = t.WriteMarkdown(f, 1)
	case 's', 'v':
		_ = t.WriteTable(f, true)
	default:
		fmt.Fprintf(f, "%%!%c(ERROR=unknown verb, Table)", verb)
	}
}

// EntriesMax returns the max entries for the tournament payout table.
func (t *Table) EntriesMax() int {
	return t.entries[0]
}

// LevelsMax returns the max levels for the tournament payout table for the
// specified entries.
func (t *Table) LevelsMax(entries int) int {
	col := t.Entries(entries)
	if col < 0 {
		return -1
	}
	for i := len(t.levels) - 1; 0 <= i; i-- {
		if t.amounts[i][col] != 0.0 {
			return t.levels[i]
		}
	}
	return -1
}

// RankingMax returns the max ranking for the tournament payout table.
func (t *Table) RankingMax() int {
	return t.levels[len(t.levels)-1]
}

// WriteTo writes the tournament payout table to w, formatting cells with f,
// breaking line output when early is true.
func (t *Table) WriteTo(w io.Writer, f func(interface{}, int, bool) string, divider func(int, int) string, early bool) error {
	if f == nil {
		f = tableFormat
	}
	maxlen := 7
	// generate headers
	v := make([]string, len(t.entries))
	for i, entries := range t.entries {
		s := t.EntriesTitle(entries)
		v[i], maxlen = s, max(maxlen, len(s))
	}
	// write 0,0 cell
	var err error
	if _, err = fmt.Fprint(w, f("", maxlen, false)); err != nil {
		return fmt.Errorf("unable to write table: %w", err)
	}
	// write headers
	for i := 0; i < len(t.entries); i++ {
		if _, err = fmt.Fprint(w, f(v[i], maxlen, i == len(t.entries)-1)); err != nil {
			return fmt.Errorf("unable to write table: %w", err)
		}
	}
	if _, err = fmt.Fprintln(w); err != nil {
		return fmt.Errorf("unable to write table: %w", err)
	}
	// write divider
	if divider != nil {
		if _, err = fmt.Fprintln(w, divider(len(t.entries)+1, maxlen)); err != nil {
			return fmt.Errorf("unable to write table: %w", err)
		}
	}
	// write levels/entries
	for i := 0; i < len(t.levels); i++ {
		// write ranking
		if _, err = fmt.Fprint(w, f(t.LevelsTitle(t.levels[i]), maxlen, false)); err != nil {
			return fmt.Errorf("unable to write table: %w", err)
		}
		// write entries
		for j := 0; j < len(t.entries); j++ {
			last := j == len(t.entries)-1 || (early && t.amounts[i][j+1] == 0)
			if _, err = fmt.Fprint(w, f(t.amounts[i][j], maxlen, last)); err != nil {
				return fmt.Errorf("unable to write table: %w", err)
			}
			if early && last {
				break
			}
		}
		if _, err = fmt.Fprintln(w); err != nil {
			return fmt.Errorf("unable to write table: %w", err)
		}
	}
	return nil
}

// WriteTable writes a plain text version of the tournament payout table, along
// with optional title, to w.
func (t *Table) WriteTable(w io.Writer, title bool) error {
	if title {
		if _, err := fmt.Fprintln(w, t.Title()+"\n"); err != nil {
			return fmt.Errorf("unable to write table: %w", err)
		}
	}
	return t.WriteTo(w, tableFormat, nil, true)
}

// WriteCSV writes a CSV version of the tournament payout table to w.
func (t *Table) WriteCSV(w io.Writer) error {
	return t.WriteTo(w, csvFormat, nil, true)
}

// WriteMarkdown writes a Markdown formatted version of the tournament payout
// table to w.
func (t *Table) WriteMarkdown(w io.Writer, header int) error {
	if 0 < header && header <= 6 {
		if _, err := fmt.Fprintln(w, strings.Repeat("#", header), t.Title()+"\n"); err != nil {
			return fmt.Errorf("unable to write markdown: %w", err)
		}
	}
	return t.WriteTo(w, markdownFormat, markdownDivider, false)
}

// Entries returns the entries column in the tournament payout table.
func (t *Table) Entries(entries int) int {
	if entries < 2 {
		return -1
	}
	for i := len(t.entries) - 1; 0 <= i; i-- {
		if entries <= t.entries[i] {
			return i
		}
	}
	return 0
}

// EntriesTitle returns the column title for the specified entries in the
// tournament payout table.
func (t *Table) EntriesTitle(entries int) string {
	col := t.Entries(entries)
	if col == -1 {
		return ""
	}
	s := strconv.Itoa(t.entries[col])
	if col != len(t.entries)-1 && t.entries[col]-t.entries[col+1] != 1 {
		s = strconv.Itoa(t.entries[col+1]+1) + "-" + s
	}
	if col == 0 {
		s += "+"
	}
	return s
}

// LevelsTitle returns the level (row) title for rank n in the tournament
// payout table.
func (t *Table) LevelsTitle(n int) string {
	for i, last := 0, 0; i < len(t.levels); i++ {
		if n <= t.levels[i] {
			return LevelTitle(last, t.levels[i])
		}
		last = t.levels[i]
	}
	return ""
}

// Levels returns the levels (rows) of the tournament table from [start, end).
func (t *Table) Levels(start, end int) []int {
	if start < 0 || end < 1 || end <= start {
		return nil
	}
	var i int
	for ; i < len(t.levels) && t.levels[i] < start; i++ {
	}
	var v []int
	for ; i < len(t.levels) && t.levels[i] < end; i++ {
		v = append(v, i)
	}
	if i != len(t.levels) {
		return append(v, i)
	}
	return v
}

// Rankings returns the levels for rankings from [start, end).
func (t *Table) Rankings(start, end int) []int {
	if start < 0 || end < 1 || end <= start {
		return nil
	}
	v, rows := t.Levels(start, end), make([]int, end-start)
	for i, pos := 0, 0; i < end-start; i++ {
		if t.levels[v[pos]] <= i+start {
			pos++
		}
		rows[i] = v[pos]
	}
	return rows
}

// Amounts returns the tournament payout amounts for positions for [start, end)
// as determined by the number of entries.
func (t *Table) Amounts(start, end, entries int) []float64 {
	rows, col := t.Rankings(start, end), t.Entries(entries)
	if len(rows) == 0 || col == -1 {
		return nil
	}
	v := make([]float64, end-start)
	for i := 0; i < len(rows); i++ {
		v[i] = t.amounts[rows[i]][col]
	}
	return v
}

// Payouts returns the tournament payouts for positions from [start, end) based
// on the number of entries.
func (t *Table) Payouts(start, end, entries int, buyin, guaranteed int64, rake float64) []int64 {
	rows, col := t.Rankings(start, end), t.Entries(entries)
	if len(rows) == 0 || col == -1 {
		return nil
	}
	v := make([]int64, end-start)
	for i := 0; i < len(rows); i++ {
		v[i] = Calc(t.amounts[rows[i]][col], entries, buyin, guaranteed, rake)
	}
	return v
}

// Stakes returns slice of the ranges of the paid levels in the form of [low,
// high), table amount, and calculated payouts for each position.
func (t *Table) Stakes(entries int, buyin, guaranteed int64, rake float64) ([][2]int, []float64, []int64) {
	maxLevel := t.LevelsMax(entries)
	rows, col := t.Levels(0, maxLevel), t.Entries(entries)
	n := len(rows)
	levels, amounts, payouts := make([][2]int, n), make([]float64, n), make([]int64, n)
	for i, last := 0, 0; i < n; i++ {
		level := t.levels[rows[i]]
		levels[i] = [2]int{last, level}
		amounts[i] = t.amounts[rows[i]][col]
		payouts[i] = Calc(amounts[i], entries, buyin, guaranteed, rake)
		last = level
	}
	return levels, amounts, payouts
}

// Amount returns the tournament payout amount for position n.
func (t *Table) Amount(n, entries int) float64 {
	if v := t.Amounts(n, n+1, entries); len(v) != 0 {
		return v[0]
	}
	return 0.0
}

// Payout returns the tournament payout for position n.
func (t *Table) Payout(n, entries int, buyin, guaranteed int64, rake float64) int64 {
	if v := t.Payouts(n, n+1, entries, buyin, guaranteed, rake); len(v) != 0 {
		return v[0]
	}
	return 0.0
}

// At returns the amount at row, col of the tournament payout table.
func (t *Table) At(row, col int) float64 {
	if 0 <= row && row < len(t.levels) && 0 <= col && col < len(t.entries) {
		return t.amounts[row][col]
	}
	return 0.0
}

// tableFormat formats v in table format.
func tableFormat(v any, n int, last bool) string {
	var s string
	switch x := v.(type) {
	case string:
		s = fmt.Sprintf("%*s", n, x)
	case float64:
		if x != 0.0 {
			s = fmt.Sprintf("%*.2f", n, x*100)
		}
	}
	if last {
		return s
	}
	return s + " "
}

// csvFormat formats v in CSV format.
func csvFormat(v any, n int, last bool) string {
	var s string
	switch x := v.(type) {
	case string:
		s = x
		if s == "" {
			s = reCell
		}
	case float64:
		if x != 0.0 {
			s = strconv.FormatFloat(math.Round(x*10000)/100, 'f', -1, 64)
		}
	}
	if last {
		return s
	}
	return s + ","
}

// markdownFormat formats v in Markdown format.
func markdownFormat(v any, n int, last bool) string {
	s := "| "
	switch x := v.(type) {
	case string:
		s += fmt.Sprintf("%*s", n, x)
	case float64:
		var v string
		if x != 0.0 {
			v = strconv.FormatFloat(math.Round(x*10000)/100, 'f', -1, 64)
		}
		s += fmt.Sprintf("%*s", n, v)
	}
	if last {
		return s + " |"
	}
	return s + " "
}

// markdownDivider returns a divider for Markdown tables.
func markdownDivider(count, n int) string {
	s, div := "", strings.Repeat("-", n+2)
	for i := 0; i < count; i++ {
		s += "|" + div
	}
	return s + "|"
}

// parseLevels parses levels title.
func parseLevels(s string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1st":
		return 1, nil
	case "2nd":
		return 2, nil
	case "3rd":
		return 3, nil
	case "4th":
		return 4, nil
	case "5th":
		return 5, nil
	case "6th":
		return 6, nil
	case "7th":
		return 7, nil
	case "8th":
		return 8, nil
	case "9th":
		return 9, nil
	case "10th":
		return 10, nil
	}
	return parseEntries(s)
}

// parseEntries parses the entries title.
func parseEntries(s string) (int, error) {
	v := strings.SplitN(s, "-", 2)
	i, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(v[len(v)-1]), "+"))
	if err != nil {
		return 0, fmt.Errorf("unable to parse col %q: %w", v[len(v)-1], err)
	}
	return i, nil
}

// reCell is the upper left cell value in a csv tournament pay table.
const reCell = "r/e"
