package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MelQ29/mtg/internal/alloc"
	"github.com/MelQ29/mtg/internal/backup"
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
		exists, err := s.Has(body.Set, body.Number, body.Foil)
		if err != nil {
			fail(w, http.StatusInternalServerError, err)
			return
		}
		if !exists {
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
			if err := s.AddCopy(p.SetCode, p.Number, body.Foil, body.Qty); err != nil {
				fail(w, http.StatusInternalServerError, err)
				return
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
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		if body.Status == "" {
			body.Status = "draft"
		}
		id, err := s.CreateDeck(body.Name, body.Description, body.Status)
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
		writeJSON(w, http.StatusOK, map[string]any{"deck": d, "entries": entries})
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
			Name        *string `json:"name"`
			Description *string `json:"description"`
			Status      *string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			fail(w, http.StatusBadRequest, err)
			return
		}
		name, desc, status := current.Name, current.Description, current.Status
		if body.Name != nil {
			name = *body.Name
		}
		if body.Description != nil {
			desc = *body.Description
		}
		if body.Status != nil {
			status = *body.Status
		}
		if status == "built" && current.Status != "built" {
			if err := assertBuildable(s, id); err != nil {
				fail(w, http.StatusConflict, err)
				return
			}
		}
		if err := s.UpdateDeck(id, name, desc, status); err != nil {
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
		status, err := s.DeckStatus(id)
		if err != nil {
			fail(w, http.StatusNotFound, err)
			return
		}
		owned, err := s.Qty(body.Set, body.Number, body.Foil)
		if err != nil {
			fail(w, http.StatusInternalServerError, err)
			return
		}
		inThis, err := s.EntryQty(id, body.Set, body.Number, body.Foil)
		if err != nil {
			fail(w, http.StatusInternalServerError, err)
			return
		}
		other, _, err := s.OtherBuilt(body.Set, body.Number, body.Foil, id)
		if err != nil {
			fail(w, http.StatusInternalServerError, err)
			return
		}
		line := alloc.Line{Owned: owned, InThis: inThis, InOtherBuilt: other}
		if status == "built" {
			if !alloc.CanPlace(line, body.Qty) {
				fail(w, http.StatusConflict, errors.New("not enough free copies"))
				return
			}
		} else if inThis+body.Qty < 0 || (body.Qty > 0 && inThis+body.Qty > owned) {
			fail(w, http.StatusConflict, errors.New("not enough owned copies"))
			return
		}
		if err := s.SetEntry(id, body.Set, body.Number, body.Foil, inThis+body.Qty); err != nil {
			fail(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "qty": inThis + body.Qty})
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
		if err != nil {
			return nil, err
		}
		return []scryfall.Printing{p}, nil
	}
	return cards.Search(q)
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
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
