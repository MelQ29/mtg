package httpapi

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MelQ29/mtg/internal/scryfall"
	"github.com/MelQ29/mtg/internal/store"
)

type fakeCards struct{}

func (fakeCards) Fetch(set, number string) (scryfall.Printing, error) {
	return scryfall.Printing{
		Name: "Hexing Squelcher", SetCode: set, Number: number, SetName: "Lorwyn Eclipsed",
		Rarity: "rare", ManaCost: "{1}{R}", TypeLine: "Creature — Goblin Sorcerer",
		Faces:        []scryfall.Face{{Name: "Hexing Squelcher", ImageURL: "http://img"}},
		PriceNonfoil: "29.59",
	}, nil
}
func (fakeCards) Search(string) ([]scryfall.Printing, error) { return nil, nil }
func (fakeCards) SaveImages(set, number string, _ []scryfall.Face) (string, string, error) {
	return set + "/" + number + "/front.jpg", "", nil
}

func handlerWithOneHexing(t *testing.T) http.Handler {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.AddCopy("ecl", "317", false, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.SetMeta(store.Copy{SetCode: "ecl", Number: "317", Name: "Hexing Squelcher"}); err != nil {
		t.Fatal(err)
	}
	return Handler(s, fakeCards{}, t.TempDir())
}

func postJSON(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestBuiltDeckRejectsSecondCopy(t *testing.T) {
	h := handlerWithOneHexing(t)
	if postJSON(t, h, "/api/decks", `{"name":"Built","description":"","status":"built"}`).Code != 200 {
		t.Fatal("create")
	}
	if postJSON(t, h, "/api/decks/1/entries", `{"set":"ecl","number":"317","foil":false,"qty":1}`).Code != 200 {
		t.Fatal("first copy")
	}
	if postJSON(t, h, "/api/decks/1/entries", `{"set":"ecl","number":"317","foil":false,"qty":1}`).Code != 409 {
		t.Fatal("second copy must be refused")
	}
}

func TestDraftDoesNotCountAsTaken(t *testing.T) {
	h := handlerWithOneHexing(t)
	postJSON(t, h, "/api/decks", `{"name":"Draft","description":"","status":"draft"}`)
	if postJSON(t, h, "/api/decks/1/entries", `{"set":"ecl","number":"317","foil":false,"qty":1}`).Code != 200 {
		t.Fatal("draft may list the only copy")
	}
	req := httptest.NewRequest(http.MethodGet, "/api/pool?deck_id=1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), `"in_other_built":0`) {
		t.Fatalf("pool %s", rec.Body.String())
	}
}
