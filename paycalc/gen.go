//go:build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	urlstr := flag.String("url", "https://www.wsop.com/how-to-play-poker/mtt-tournament-payouts/", "url")
	flag.Parse()
	if err := run(*urlstr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(urlstr string) error {
	buf, err := grab(urlstr)
	if err != nil {
		return err
	}
	cleanRE := regexp.MustCompile(`\s`)
	tables := make(map[int]*Table)
	for _, tablem := range tableRE.FindAllSubmatchIndex(buf, -1) {
		typ, err := strconv.ParseInt(string(buf[tablem[2]:tablem[3]]), 10, 64)
		if err != nil {
			return err
		}
		log.Printf("type: %d", typ)
		start := tablem[0]
		end := start + bytes.Index(buf[start:], []byte(`</table>`))
		i := start + bytes.Index(buf[start:end], []byte(`</thead>`))
		thead := buf[start:i]
		var entries []string
		for _, th := range headRE.FindAllStringSubmatch(string(thead), -1) {
			entries = append(entries, cleanRE.ReplaceAllString(th[1], ""))
		}
		log.Printf("entries: %v", entries)
		t, ok := tables[int(typ)]
		if !ok {
			t = new(Table)
		}
		t.Entries = append(entries, t.Entries...)
		count := bytes.Count(buf[i:end], []byte("<tr"))
		log.Printf("rows: %d", count)
		switch {
		case t.Ranking != nil && count != len(t.Ranking):
			return fmt.Errorf("invalid row count: expected %d, got: %d", len(t.Ranking), count)
		case t.Amounts != nil && count != len(t.Amounts):
		case t.Ranking == nil:
			t.Ranking = make([]string, count)
			t.Amounts = make([][]string, count)
		}
		var v []string
		for n := 0; i < end && bytes.Contains(buf[i:end], []byte("<tr")); n++ {
			i, v = parseRow(buf, i)
			switch {
			case len(v)-1 != len(entries):
				return fmt.Errorf("expected %d-1 cells, got: %d\n%q", len(entries), len(v), v)
			case t.Ranking[n] != "" && t.Ranking[n] != v[0]:
				return fmt.Errorf("row %d defined as %q, expected: %q", n, t.Ranking[n], v[0])
			case t.Ranking[n] == "":
				t.Ranking[n] = v[0]
			}
			log.Printf("  %v", v)
			t.Amounts[n] = append(v[1:], t.Amounts[n]...)
		}
		tables[int(typ)] = t
	}
	for typ, t := range tables {
		if err := t.WriteTo(fmt.Sprintf("top%d.csv", typ)); err != nil {
			return err
		}
	}
	return nil
}

type Table struct {
	Ranking []string
	Entries []string
	Amounts [][]string
}

func (t *Table) WriteTo(s string) error {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "r/e,%s\n", strings.Join(t.Entries, ","))
	for i := range len(t.Amounts) {
		fmt.Fprintf(buf, "%s,%s\n", t.Ranking[i], strings.Join(t.Amounts[i], ","))
	}
	re := regexp.MustCompile(`(?m),+$`)
	b := re.ReplaceAll(buf.Bytes(), nil)
	return os.WriteFile(s, b, 0644)
}

func grab(urlstr string) ([]byte, error) {
	if _, err := os.Stat("table.html"); err == nil || !strings.HasPrefix(urlstr, "https://") {
		return os.ReadFile("table.html")
	}
	res, err := http.Get(urlstr)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile("table.html", body, 0644); err != nil {
		return nil, err
	}
	return body, nil
}

func parseRow(buf []byte, i int) (int, []string) {
	end := i + bytes.Index(buf[i:], []byte(`</tr>`)) + 5
	var v []string
	for _, m := range cellRE.FindAllStringSubmatch(string(buf[i:end]), -1) {
		v = append(v, m[1])
	}
	return end, v
}

var (
	tableRE = regexp.MustCompile(`<table class="payoutstructure" summary="[^0-9]+([0-9]{2})%`)
	headRE  = regexp.MustCompile(`<th.+?>(.+?)</th>`) // conveniently ignores the empty cell
	cellRE  = regexp.MustCompile(`<td.+?>([^<]*)</td>`)
)
