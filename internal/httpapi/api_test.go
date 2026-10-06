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

func TestCollectionValueIgnoresSearch(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.AddCopy("ecl", "317", false, 2); err != nil {
		t.Fatal(err)
	}
	if err := s.SetMeta(store.Copy{SetCode: "ecl", Number: "317", Name: "Hexing Squelcher", PriceUSD: "10.00"}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddCopy("fin", "1", false, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.SetMeta(store.Copy{SetCode: "fin", Number: "1", Name: "Blitzball"}); err != nil {
		t.Fatal(err)
	}
	h := Handler(s, fakeCards{}, t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/cards?q=nomatch", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	if rec.Body.String() != "[]\n" {
		t.Fatalf("search body %s", rec.Body.String())
	}
	if rec.Header().Get("X-Collection-USD") != "20.00" {
		t.Fatalf("value %q", rec.Header().Get("X-Collection-USD"))
	}
	if rec.Header().Get("X-Collection-Count") != "3" {
		t.Fatalf("count %q", rec.Header().Get("X-Collection-Count"))
	}
}

func TestCommanderRejectsASecondCopyAndAPrice(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.AddCopy("ecl", "1", false, 2); err != nil {
		t.Fatal(err)
	}
	if err := s.SetMeta(store.Copy{SetCode: "ecl", Number: "1", Name: "Overkill", TypeLine: "Instant", Colors: "B", PriceUSD: "0.30"}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddCopy("trk", "321", false, 5); err != nil {
		t.Fatal(err)
	}
	if err := s.SetMeta(store.Copy{SetCode: "trk", Number: "321", Name: "Swamp", TypeLine: "Basic Land — Swamp", Colors: "B"}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddCopy("ecl", "317", false, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.SetMeta(store.Copy{SetCode: "ecl", Number: "317", Name: "Hexing Squelcher", TypeLine: "Creature — Goblin", PriceUSD: "29.59"}); err != nil {
		t.Fatal(err)
	}
	h := Handler(s, fakeCards{}, t.TempDir())
	rec := postJSON(t, h, "/api/decks", `{"name":"Cmd","description":"","status":"draft","format":"commander"}`)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	if postJSON(t, h, "/api/decks/1/entries", `{"set":"ecl","number":"1","foil":false,"qty":1}`).Code != 200 {
		t.Fatal("first spell")
	}
	if postJSON(t, h, "/api/decks/1/entries", `{"set":"ecl","number":"1","foil":false,"qty":1}`).Code != 409 {
		t.Fatal("second copy")
	}
	if postJSON(t, h, "/api/decks/1/entries", `{"set":"trk","number":"321","foil":false,"qty":2}`).Code != 200 {
		t.Fatal("basics")
	}
	if postJSON(t, h, "/api/decks/1/entries", `{"set":"ecl","number":"317","foil":false,"qty":1}`).Code != 409 {
		t.Fatal("price")
	}
	del := httptest.NewRequest(http.MethodDelete, "/api/decks/1", nil)
	out := httptest.NewRecorder()
	h.ServeHTTP(out, del)
	if out.Code != 200 {
		t.Fatal(out.Body.String())
	}
	if _, err := s.GetDeck(1); err == nil {
		t.Fatal("deck still exists")
	}
}

func TestCollectionShowsABuiltDeck(t *testing.T) {
	h := handlerWithOneHexing(t)
	if postJSON(t, h, "/api/decks", `{"name":"White-Black","description":"","status":"built"}`).Code != 200 {
		t.Fatal("create")
	}
	if postJSON(t, h, "/api/decks/1/entries", `{"set":"ecl","number":"317","foil":false,"qty":1}`).Code != 200 {
		t.Fatal("entry")
	}
	req := httptest.NewRequest(http.MethodGet, "/api/cards?q=Hexing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), `"in_other_built":1`) || !strings.Contains(rec.Body.String(), `"other_decks":"White-Black"`) {
		t.Fatalf("cards %s", rec.Body.String())
	}
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
