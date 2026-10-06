package deckrules

import (
	"fmt"
	"strings"

	"github.com/MelQ29/mtg/internal/scryfall"
)

// Card is the part of a printing Commander construction looks at.
type Card struct {
	Name     string
	TypeLine string
	Colors   string
	Oracle   string
	Qty      int
}

// BasicLand reports a land with the basic supertype. Snow-covered basics count.
// Other lands do not, even when they have a basic land type.
func BasicLand(typeLine string) bool {
	t := strings.ToLower(typeLine)
	return strings.Contains(t, "basic") && strings.Contains(t, "land")
}

// AnyNumber reports an ability that lets a deck contain any number of that name.
func AnyNumber(oracle string) bool {
	return strings.Contains(strings.ToLower(oracle), "a deck can have any number of cards named")
}

// CanBeCommander reports a legendary creature, Vehicle, or Spacecraft,
// or a card that says it can be your commander.
func CanBeCommander(typeLine, oracle string) bool {
	if strings.Contains(strings.ToLower(oracle), "can be your commander") {
		return true
	}
	tl := strings.ToLower(typeLine)
	if !strings.Contains(tl, "legendary") {
		return false
	}
	return strings.Contains(tl, "creature") || strings.Contains(tl, "vehicle") || strings.Contains(tl, "spacecraft")
}

// IdentityOK reports whether every color in card is in commander.
func IdentityOK(card, commander string) bool {
	for _, r := range card {
		if !strings.ContainsRune(commander, r) {
			return false
		}
	}
	return true
}

// LandColors returns the colors of mana a card's basic land types can produce.
func LandColors(typeLine string) string {
	var b strings.Builder
	for _, item := range []struct{ word, color string }{
		{"Plains", "W"},
		{"Island", "U"},
		{"Swamp", "B"},
		{"Mountain", "R"},
		{"Forest", "G"},
	} {
		if hasWord(typeLine, item.word) {
			b.WriteString(item.color)
		}
	}
	return b.String()
}

func hasWord(typeLine, word string) bool {
	parts := strings.FieldsFunc(typeLine, func(r rune) bool {
		return r == ' ' || r == '—' || r == '-' || r == '/'
	})
	for _, part := range parts {
		if strings.EqualFold(part, word) {
			return true
		}
	}
	return false
}

func sameName(a, b string) bool {
	return scryfall.FoldName(a) == scryfall.FoldName(b)
}

// CheckAdd reports why this card cannot be added to a Commander deck.
// Kitchen decks are not checked here. addQty is the number of copies being added.
func CheckAdd(format string, commander Card, hasCommander bool, entries []Card, adding Card, addQty int) error {
	if format != "commander" || addQty <= 0 {
		return nil
	}
	total := addQty
	nameQty := addQty
	for _, e := range entries {
		total += e.Qty
		if sameName(e.Name, adding.Name) {
			nameQty += e.Qty
		}
	}
	if total > 100 {
		return fmt.Errorf("a Commander deck is 100 cards")
	}
	if !BasicLand(adding.TypeLine) && !AnyNumber(adding.Oracle) && nameQty > 1 {
		return fmt.Errorf("Commander allows one copy of %s", adding.Name)
	}
	if !hasCommander {
		return nil
	}
	if !IdentityOK(adding.Colors, commander.Colors) {
		return fmt.Errorf("%s is outside the commander's colors", adding.Name)
	}
	if !IdentityOK(LandColors(adding.TypeLine), commander.Colors) {
		return fmt.Errorf("%s needs colors the commander does not have", adding.Name)
	}
	return nil
}

// CheckCommander reports whether this card can lead the deck that already exists.
func CheckCommander(card Card, entries []Card) error {
	if !CanBeCommander(card.TypeLine, card.Oracle) {
		return fmt.Errorf("%s cannot be a commander", card.Name)
	}
	var outside []string
	for _, e := range entries {
		if !IdentityOK(e.Colors, card.Colors) || !IdentityOK(LandColors(e.TypeLine), card.Colors) {
			outside = append(outside, e.Name)
		}
	}
	if len(outside) > 0 {
		return fmt.Errorf("outside this commander's colors: %s", strings.Join(outside, ", "))
	}
	return nil
}

// Problems lists what still keeps a Commander deck from the format rules.
func Problems(format string, commander Card, hasCommander bool, entries []Card) []string {
	if format != "commander" {
		return nil
	}
	var out []string
	if !hasCommander {
		out = append(out, "Choose a commander.")
	}
	total := 0
	names := map[string]int{}
	labels := map[string]string{}
	basic := map[string]bool{}
	for _, e := range entries {
		total += e.Qty
		key := scryfall.FoldName(e.Name)
		names[key] += e.Qty
		labels[key] = e.Name
		if BasicLand(e.TypeLine) || AnyNumber(e.Oracle) {
			basic[key] = true
		}
		if hasCommander && (!IdentityOK(e.Colors, commander.Colors) || !IdentityOK(LandColors(e.TypeLine), commander.Colors)) {
			out = append(out, e.Name+" is outside the commander's colors.")
		}
	}
	for key, n := range names {
		if n > 1 && !basic[key] {
			out = append(out, "Commander allows one copy of "+labels[key]+".")
		}
	}
	if total > 100 {
		out = append(out, "A Commander deck is 100 cards.")
	}
	return out
}
