package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/MelQ29/mtg/internal/images"
	"github.com/MelQ29/mtg/internal/importmd"
	"github.com/MelQ29/mtg/internal/scryfall"
	"github.com/MelQ29/mtg/internal/store"
)

var nameAliases = map[string]string{
	"aquitect s defense": "Aquitect's Defenses",
	"sandbender s storm": "Sandbenders' Storm",
	"mouse mark iii":     "Mouser Mark III",
	"white mage s stuff": "White Mage's Staff",
	"wildvine pummler":   "Wildvine Pummeler",
}

// lookupName turns an import line into a Scryfall name.
// A folded alias wins. A latin name in parentheses is used when the
// rest of the line has no latin letters, as in a translated title.
func lookupName(raw string) string {
	if alias, ok := nameAliases[scryfall.FoldName(raw)]; ok {
		return alias
	}
	if inside := latinParenthetical(raw); inside != "" {
		if alias, ok := nameAliases[scryfall.FoldName(inside)]; ok {
			return alias
		}
		return inside
	}
	return raw
}

func latinParenthetical(raw string) string {
	i := strings.LastIndex(raw, "(")
	j := strings.LastIndex(raw, ")")
	if i < 0 || j <= i {
		return ""
	}
	inside := strings.TrimSpace(raw[i+1 : j])
	outside := strings.TrimSpace(raw[:i])
	if scryfall.FoldName(outside) != "" || scryfall.FoldName(inside) == "" {
		return ""
	}
	return inside
}

func resolveMissing(s *store.Store, client *scryfall.Client, imageDir string) {
	cards, err := s.List("")
	if err != nil {
		log.Fatal(err)
	}
	var failed []string
	done := 0
	for _, card := range cards {
		if card.FrontImage != "" {
			continue
		}
		lookup := card.Name
		if card.SetCode == "basic" {
			lookup = strings.TrimPrefix(card.Name, "Basic ")
		}
		lookup = lookupName(lookup)
		var printing scryfall.Printing
		if card.SetCode != "" && card.SetCode != "pending" && card.SetCode != "basic" {
			printing, err = client.Fetch(card.SetCode, card.Number)
		} else {
			printing, err = client.Named(lookup)
		}
		if err != nil {
			failed = append(failed, card.Name+": "+err.Error())
			continue
		}
		front, back, err := images.Save(imageDir, printing.SetCode, printing.Number, printing.Faces, httpGet)
		if err != nil {
			failed = append(failed, card.Name+": "+err.Error())
			continue
		}
		if card.SetCode != printing.SetCode || card.Number != printing.Number {
			qty := card.Qty
			if card.SetCode == "pending" {
				if taken, err := s.TakePending(card.Name); err == nil && taken > 0 {
					qty = taken
				}
			}
			if err := s.AddCopy(printing.SetCode, printing.Number, false, qty); err != nil {
				failed = append(failed, card.Name+": "+err.Error())
				continue
			}
			if card.SetCode != "pending" {
				if err := s.Delete(card.SetCode, card.Number, card.Foil); err != nil {
					failed = append(failed, card.Name+": "+err.Error())
					continue
				}
			}
		}
		price := printing.PriceNonfoil
		if err := s.SetMeta(store.Copy{
			SetCode: printing.SetCode, Number: printing.Number, Foil: false,
			Name: printing.Name, SetName: printing.SetName, Rarity: printing.Rarity,
			ManaCost: printing.ManaCost, TypeLine: printing.TypeLine, Colors: printing.Colors,
			PriceUSD: price, PriceOn: "resolved",
			FrontImage: front, BackImage: back, OracleText: printing.Oracle,
		}); err != nil {
			failed = append(failed, card.Name+": "+err.Error())
			continue
		}
		done++
		log.Printf("art %d %s -> %s/%s", done, card.Name, printing.SetCode, printing.Number)
	}
	log.Printf("pictures saved for %d cards", done)
	for _, line := range failed {
		log.Printf("no picture: %s", line)
	}
}

func loadDeck(s *store.Store, path, title string) error {
	body, err := readFile(path)
	if err != nil {
		return err
	}
	rows := importmd.ParseDeck(string(body))
	owned, err := s.List("")
	if err != nil {
		return err
	}
	byName := indexByName(owned)
	id, err := deckIDByName(s, title)
	if err != nil {
		return err
	}
	if id == 0 {
		id, err = s.CreateDeck(title, title+", 60 cards.", "built")
		if err != nil {
			return err
		}
	}
	for _, row := range rows {
		name := lookupName(row.Name)
		card, ok := byName[scryfall.FoldName(name)]
		if !ok {
			log.Printf("deck skip, not in collection: %s", row.Name)
			continue
		}
		have, err := s.EntryQty(id, card.SetCode, card.Number, card.Foil)
		if err != nil {
			return err
		}
		if have >= row.Qty {
			continue
		}
		if err := s.SetEntry(id, card.SetCode, card.Number, card.Foil, row.Qty); err != nil {
			return err
		}
	}
	log.Printf("deck %q id %d", title, id)
	return nil
}

// indexByName keys a printing by its full name and by each face before or after "//".
func indexByName(owned []store.Copy) map[string]store.Copy {
	byName := map[string]store.Copy{}
	put := func(name string, card store.Copy) {
		key := scryfall.FoldName(name)
		if key == "" {
			return
		}
		if _, exists := byName[key]; exists {
			return
		}
		byName[key] = card
	}
	for _, card := range owned {
		if card.SetCode == "pending" || card.Foil {
			continue
		}
		put(card.Name, card)
		if front, back, ok := strings.Cut(card.Name, " // "); ok {
			put(front, card)
			put(back, card)
		}
		if card.SetCode == "basic" || strings.HasPrefix(card.Name, "Basic ") {
			put(strings.TrimPrefix(card.Name, "Basic "), card)
		}
	}
	return byName
}

func deckIDByName(s *store.Store, title string) (int, error) {
	decks, err := s.ListDecks()
	if err != nil {
		return 0, err
	}
	for _, d := range decks {
		if d.Name == title {
			return d.ID, nil
		}
	}
	return 0, nil
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func httpGet(rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "kitchen-mtg/0.1")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errStatus(res.Status)
	}
	return body, nil
}

type errStatus string

func (e errStatus) Error() string { return string(e) }
