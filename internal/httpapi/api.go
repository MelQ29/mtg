package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MelQ29/mtg/internal/alloc"
	"github.com/MelQ29/mtg/internal/backup"
	"github.com/MelQ29/mtg/internal/deckrules"
	"github.com/MelQ29/mtg/internal/images"
	"github.com/MelQ29/mtg/internal/scryfall"
	"github.com/MelQ29/mtg/internal/store"
)

// Cards talks to Scryfall and the image folder. Tests substitute a fake.
type Cards interface {
	Fetch(set, number string) (scryfall.Printing, error)
	Search(name string) ([]scryfall.Printing, error)
	SaveImages(set, number string, faces []scryfall.Face) (string, string, error)
}

// Handler serves the collection API.
func Handler(s *store.Store, cards Cards, imageDir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/cards", func(w http.ResponseWriter, r *http.Request) {
		list, err := s.List(r.URL.Query().Get("q"))
		if err != nil {
			fail(w, http.StatusInternalServerError, err)
			return
		}
		for i := range list {
			other, names, err := s.OtherBuilt(list[i].SetCode, list[i].Number, list[i].Foil, 0)
			if err != nil {
				fail(w, http.StatusInternalServerError, err)
				return
			}
			list[i].InOtherBuilt = other
			list[i].OtherDecks = names
		}
		noteValue(w, s)
		writeJSON(w, http.StatusOK, list)
	})
	mux.HandleFunc("GET /api/lookup", func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		prints, err := lookup(cards, q)
		if err != nil {
			fail(w, http.StatusBadGateway, err)
			return
		}
		if prints == nil {
			prints = []scryfall.Printing{}
		}
		writeJSON(w, http.StatusOK, prints)
	})
	mux.HandleFunc("POST /api/cards", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Set    string `json:"set"`
			Number string `json:"number"`
			Foil   bool   `json:"foil"`
			Qty    int    `json:"qty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		if body.Qty <= 0 {
			body.Qty = 1
		}
		existing, exists, err := s.Get(body.Set, body.Number, body.Foil)
		if err != nil {
			fail(w, http.StatusInternalServerError, err)
			return
		}
		if !exists || existing.FrontImage == "" {
			p, err := cards.Fetch(body.Set, body.Number)
			if err != nil {
				fail(w, http.StatusBadGateway, err)
				return
			}
			front, back, err := cards.SaveImages(p.SetCode, p.Number, p.Faces)
			if err != nil {
				fail(w, http.StatusBadGateway, err)
				return
			}
			price := p.PriceNonfoil
			if body.Foil {
				price = p.PriceFoil
			}
			pending, err := s.TakePending(p.Name)
			if err != nil {
				fail(w, http.StatusInternalServerError, err)
				return
			}
			add := body.Qty
			if !exists && pending > add {
				add = pending
			}
			if exists && existing.FrontImage == "" {
				add = 0
			}
			if add > 0 {
				if err := s.AddCopy(p.SetCode, p.Number, body.Foil, add); err != nil {
					fail(w, http.StatusInternalServerError, err)
					return
				}
			}
			if err := s.SetMeta(store.Copy{
				SetCode: p.SetCode, Number: p.Number, Foil: body.Foil,
				Name: p.Name, SetName: p.SetName, Rarity: p.Rarity,
				ManaCost: p.ManaCost, TypeLine: p.TypeLine, Colors: p.Colors,
				PriceUSD: price, PriceOn: time.Now().Format("2006-01-02"),
				FrontImage: front, BackImage: back, OracleText: p.Oracle,
			}); err != nil {
				fail(w, http.StatusInternalServerError, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
		if err := s.AddCopy(body.Set, body.Number, body.Foil, body.Qty); err != nil {
			fail(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("GET /api/decks", func(w http.ResponseWriter, r *http.Request) {
		list, err := s.ListDecks()
		if err != nil {
			fail(w, http.StatusInternalServerError, err)
			return
		}
		if list == nil {
			list = []store.Deck{}
		}
		writeJSON(w, http.StatusOK, list)
	})
	mux.HandleFunc("POST /api/decks", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Status      string `json:"status"`
			Format      string `json:"format"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		if body.Status == "" {
			body.Status = "draft"
		}
		format := body.Format
		if format == "" {
			format = "kitchen"
		}
		id, err := s.CreateDeckFormat(body.Name, body.Description, body.Status, format)
		if err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"id": id})
	})
	mux.HandleFunc("GET /api/decks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		d, err := s.GetDeck(id)
		if err != nil {
			fail(w, http.StatusNotFound, err)
			return
		}
		entries, err := s.Entries(id)
		if err != nil {
			fail(w, http.StatusInternalServerError, err)
			return
		}
		if entries == nil {
			entries = []store.Entry{}
		}
		writeJSON(w, http.StatusOK, map[string]any{"deck": d, "entries": entries, "problems": commanderProblems(s, d, entries)})
	})
	mux.HandleFunc("DELETE /api/decks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		if err := s.DeleteDeck(id); err != nil {
			fail(w, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("PATCH /api/decks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		current, err := s.GetDeck(id)
		if err != nil {
			fail(w, http.StatusNotFound, err)
			return
		}
		var body struct {
			Name           *string `json:"name"`
			Description    *string `json:"description"`
			Status         *string `json:"status"`
			Format         *string `json:"format"`
			ClearCommander *bool   `json:"clear_commander"`
			Commander      *struct {
				Set    string `json:"set"`
				Number string `json:"number"`
				Foil   bool   `json:"foil"`
			} `json:"commander"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		if body.Name != nil {
			current.Name = *body.Name
		}
		if body.Description != nil {
			current.Description = *body.Description
		}
		if body.Status != nil {
			current.Status = *body.Status
		}
		if body.Format != nil {
			current.Format = *body.Format
		}
		if body.ClearCommander != nil && *body.ClearCommander {
			current.CommanderSet, current.CommanderNumber, current.CommanderFoil = "", "", false
		}
		if body.Commander != nil {
			if err := setCommander(s, &current, body.Commander.Set, body.Commander.Number, body.Commander.Foil); err != nil {
				fail(w, http.StatusConflict, err)
				return
			}
		}
		if current.Status == "built" {
			fresh, err := s.GetDeck(id)
			if err != nil {
				fail(w, http.StatusNotFound, err)
				return
			}
			if fresh.Status != "built" {
				if err := assertBuildable(s, id); err != nil {
					fail(w, http.StatusConflict, err)
					return
				}
			}
		}
		if err := s.UpdateDeck(current); err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("POST /api/decks/{id}/entries", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		var body struct {
			Set    string `json:"set"`
			Number string `json:"number"`
			Foil   bool   `json:"foil"`
			Qty    int    `json:"qty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		if body.Qty == 0 {
			body.Qty = 1
		}
		deck, err := s.GetDeck(id)
		if err != nil {
			fail(w, http.StatusNotFound, err)
			return
		}
		if body.Qty > 0 {
			if err := checkAdd(s, deck, body.Set, body.Number, body.Foil, body.Qty); err != nil {
				fail(w, http.StatusConflict, err)
				return
			}
		}
		next, err := s.ApplyEntryDelta(id, body.Set, body.Number, body.Foil, body.Qty)
		if err != nil {
			if err.Error() == "not enough free copies" || err.Error() == "not enough owned copies" {
				fail(w, http.StatusConflict, err)
				return
			}
			if strings.Contains(err.Error(), "no rows") {
				fail(w, http.StatusNotFound, err)
				return
			}
			fail(w, http.StatusBadRequest, err)
			return
		}
		if next == 0 && deck.CommanderSet == body.Set && deck.CommanderNumber == body.Number && deck.CommanderFoil == body.Foil {
			deck.CommanderSet, deck.CommanderNumber, deck.CommanderFoil = "", "", false
			if err := s.UpdateDeck(deck); err != nil {
				fail(w, http.StatusInternalServerError, err)
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "qty": next})
	})
	mux.HandleFunc("GET /api/pool", func(w http.ResponseWriter, r *http.Request) {
		deckID, _ := strconv.Atoi(r.URL.Query().Get("deck_id"))
		list, err := s.List(r.URL.Query().Get("q"))
		if err != nil {
			fail(w, http.StatusInternalServerError, err)
			return
		}
		for i := range list {
			other, names, err := s.OtherBuilt(list[i].SetCode, list[i].Number, list[i].Foil, deckID)
			if err != nil {
				fail(w, http.StatusInternalServerError, err)
				return
			}
			inThis, err := s.EntryQty(deckID, list[i].SetCode, list[i].Number, list[i].Foil)
			if err != nil {
				fail(w, http.StatusInternalServerError, err)
				return
			}
			list[i].InOtherBuilt = other
			list[i].InThis = inThis
			list[i].OtherDecks = names
		}
		writeJSON(w, http.StatusOK, list)
	})
	mux.HandleFunc("GET /api/export", func(w http.ResponseWriter, r *http.Request) {
		blob, err := backup.Export(s, imageDir)
		if err != nil {
			fail(w, http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="kitchen-collection.zip"`)
		w.Write(blob)
	})
	mux.HandleFunc("POST /api/import", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		defer file.Close()
		blob, err := io.ReadAll(file)
		if err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		if err := backup.Import(s, imageDir, blob); err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	return mux
}

func lookup(cards Cards, q string) ([]scryfall.Printing, error) {
	parts := strings.Fields(q)
	if len(parts) == 2 && isNumber(parts[1]) {
		p, err := cards.Fetch(strings.ToLower(parts[0]), parts[1])
		if err == nil {
			return []scryfall.Printing{p}, nil
		}
		if !scryfall.NotFound(err) {
			return nil, err
		}
	}
	prints, err := cards.Search(q)
	if scryfall.NotFound(err) {
		return []scryfall.Printing{}, nil
	}
	return prints, err
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	digits := false
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
			digits = true
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '-', r == '★':
		default:
			return false
		}
	}
	return digits
}

func ruleCards(entries []store.Entry) []deckrules.Card {
	out := make([]deckrules.Card, len(entries))
	for i, e := range entries {
		out[i] = deckrules.Card{Name: e.Name, TypeLine: e.TypeLine, Colors: e.Colors, Oracle: e.Oracle, Qty: e.Qty}
	}
	return out
}

func tooExpensive(price string) bool {
	if price == "" {
		return false
	}
	n, err := strconv.ParseFloat(price, 64)
	return err == nil && n > 6
}

func checkAdd(s *store.Store, deck store.Deck, set, number string, foil bool, qty int) error {
	card, ok, err := s.Get(set, number, foil)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("card is not in the collection")
	}
	if tooExpensive(card.PriceUSD) {
		return fmt.Errorf("cards over $6 stay in the collection")
	}
	entries, err := s.Entries(deck.ID)
	if err != nil {
		return err
	}
	var leader deckrules.Card
	has := deck.CommanderSet != ""
	if has {
		c, ok, err := s.Get(deck.CommanderSet, deck.CommanderNumber, deck.CommanderFoil)
		if err != nil {
			return err
		}
		if ok {
			leader = deckrules.Card{Name: c.Name, TypeLine: c.TypeLine, Colors: c.Colors, Oracle: c.OracleText}
		}
	}
	return deckrules.CheckAdd(deck.Format, leader, has, ruleCards(entries), deckrules.Card{
		Name: card.Name, TypeLine: card.TypeLine, Colors: card.Colors, Oracle: card.OracleText,
	}, qty)
}

func setCommander(s *store.Store, deck *store.Deck, set, number string, foil bool) error {
	n, err := s.EntryQty(deck.ID, set, number, foil)
	if err != nil {
		return err
	}
	if n <= 0 {
		return fmt.Errorf("the commander has to be in the deck")
	}
	card, ok, err := s.Get(set, number, foil)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("card is not in the collection")
	}
	entries, err := s.Entries(deck.ID)
	if err != nil {
		return err
	}
	leader := deckrules.Card{Name: card.Name, TypeLine: card.TypeLine, Colors: card.Colors, Oracle: card.OracleText}
	if err := deckrules.CheckCommander(leader, ruleCards(entries)); err != nil {
		return err
	}
	deck.CommanderSet, deck.CommanderNumber, deck.CommanderFoil = set, number, foil
	return nil
}

func commanderProblems(s *store.Store, deck store.Deck, entries []store.Entry) []string {
	var leader deckrules.Card
	has := deck.CommanderSet != ""
	if has {
		if c, ok, err := s.Get(deck.CommanderSet, deck.CommanderNumber, deck.CommanderFoil); err == nil && ok {
			leader = deckrules.Card{Name: c.Name, TypeLine: c.TypeLine, Colors: c.Colors, Oracle: c.OracleText}
		}
	}
	out := deckrules.Problems(deck.Format, leader, has, ruleCards(entries))
	if out == nil {
		return []string{}
	}
	return out
}

func assertBuildable(s *store.Store, deckID int) error {
	entries, err := s.Entries(deckID)
	if err != nil {
		return err
	}
	for _, e := range entries {
		owned, err := s.Qty(e.SetCode, e.Number, e.Foil)
		if err != nil {
			return err
		}
		other, _, err := s.OtherBuilt(e.SetCode, e.Number, e.Foil, deckID)
		if err != nil {
			return err
		}
		if !alloc.CanPlace(alloc.Line{Owned: owned, InThis: 0, InOtherBuilt: other}, e.Qty) {
			return errors.New("not enough free copies to mark this deck built")
		}
	}
	return nil
}

func noteValue(w http.ResponseWriter, s *store.Store) {
	v, err := s.ValueUSD()
	if err != nil {
		return
	}
	w.Header().Set("X-Collection-USD", strconv.FormatFloat(v, 'f', 2, 64))
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]string{"error": err.Error()})
}

// LiveCards is the production Scryfall and image saver.
type LiveCards struct {
	Client   *scryfall.Client
	ImageDir string
	Get      func(url string) ([]byte, error)
}

func (l LiveCards) Fetch(set, number string) (scryfall.Printing, error) {
	return l.Client.Fetch(set, number)
}

func (l LiveCards) Search(name string) ([]scryfall.Printing, error) {
	return l.Client.SearchPrintings(name)
}

func (l LiveCards) SaveImages(set, number string, faces []scryfall.Face) (string, string, error) {
	get := l.Get
	if get == nil {
		get = httpGet
	}
	return images.Save(l.ImageDir, set, number, faces, get)
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
		return nil, errors.New(res.Status)
	}
	return body, nil
}
