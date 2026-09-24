package importmd

import (
	"bufio"
	"regexp"
	"strings"
)

// Row is one collection line.
type Row struct {
	Name    string
	Set     string
	Number  string
	Foil    bool
	Qty     int
	Price   string
	Rarity  string
	PriceOn string
}

var lineRe = regexp.MustCompile(`^- (.+?) ×(\d+)(?: — (.+))?$`)
var deckLineRe = regexp.MustCompile(`^- (\d+) (.+?)(?: \*\*| —|$)`)

// ParseDeck reads a deck markdown list: "- 2 Card Name **[R]**".
func ParseDeck(markdown string) []Row {
	var out []Row
	sc := bufio.NewScanner(strings.NewReader(markdown))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		m := deckLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		out = append(out, Row{Name: strings.TrimSpace(m[2]), Qty: atoi(m[1])})
	}
	return out
}

// Parse reads the kitchen collection markdown.
func Parse(markdown string) ([]Row, error) {
	var out []Row
	sc := bufio.NewScanner(strings.NewReader(markdown))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "- ") {
			continue
		}
		line = stripColorTag(line)
		m := lineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		qty := atoi(m[2])
		row := Row{Name: strings.TrimSpace(m[1]), Qty: qty}
		if m[3] != "" {
			applyTail(&row, m[3])
		}
		applyBasic(&row)
		out = append(out, row)
	}
	return out, sc.Err()
}

func stripColorTag(line string) string {
	if i := strings.LastIndex(line, " `"); i > 0 && strings.HasSuffix(line, "`") {
		return strings.TrimSpace(line[:i])
	}
	return line
}

func applyTail(row *Row, tail string) {
	parts := strings.Split(tail, " · ")
	if len(parts) > 0 {
		row.Rarity = strings.TrimSpace(parts[0])
	}
	if len(parts) > 1 && strings.EqualFold(strings.TrimSpace(parts[1]), "foil") {
		row.Foil = true
	}
	if len(parts) > 2 {
		row.Price = strings.TrimPrefix(strings.TrimSpace(parts[2]), "$")
	}
	if len(parts) > 3 {
		setNum := strings.TrimSpace(parts[3])
		set, num, ok := strings.Cut(setNum, " #")
		if ok {
			row.Set = strings.TrimSpace(set)
			row.Number = strings.TrimSpace(num)
		}
	}
	if len(parts) > 4 {
		row.PriceOn = strings.TrimSpace(parts[4])
	}
}

func applyBasic(row *Row) {
	lower := strings.ToLower(row.Name)
	if !strings.HasPrefix(lower, "basic ") {
		return
	}
	row.Set = "basic"
	row.Number = strings.TrimSpace(lower[len("basic "):])
	row.Name = "Basic " + title(row.Number)
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func atoi(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}
