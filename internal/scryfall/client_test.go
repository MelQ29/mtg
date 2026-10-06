package scryfall

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustRead(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParseHexingSingleFace(t *testing.T) {
	p, err := ParseCard(mustRead(t, "ecl-317.json"))
	if err != nil {
		t.Fatal(err)
	}
	if p.SetName != "Lorwyn Eclipsed" || p.Number != "317" || len(p.Faces) != 1 {
		t.Fatalf("%+v", p)
	}
	if p.PriceNonfoil != "29.59" || p.PriceFoil != "60.50" {
		t.Fatalf("prices %+v", p)
	}
}

func TestSearchPrintingsFallsBackToPartialName(t *testing.T) {
	card := mustRead(t, "ecl-317.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cards/named") || strings.Contains(r.URL.Query().Get("q"), `!"`) {
			http.Error(w, `{"object":"error","status":404,"details":"Your query didn’t match any cards."}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[` + string(card) + `]}`))
	}))
	defer srv.Close()
	got, err := (&Client{HTTP: srv.Client(), Base: srv.URL}).SearchPrintings("squelch")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Hexing Squelcher" {
		t.Fatalf("%+v", got)
	}
}

func TestFoldNameDropsNonLatin(t *testing.T) {
	if got := FoldName("Нексус Ликолесья (Maskwood Nexus)"); got != "maskwood nexus" {
		t.Fatalf("got %q", got)
	}
}

func TestParseAangTwoFaces(t *testing.T) {
	p, err := ParseCard(mustRead(t, "tla-298.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Faces) != 2 || p.Faces[0].Name != "Aang, Swift Savior" {
		t.Fatalf("%+v", p.Faces)
	}
	if p.Faces[1].ImageURL == "" || p.Faces[0].ImageURL == p.Faces[1].ImageURL {
		t.Fatal("each face needs its own image")
	}
	if p.ManaCost != "{1}{W}{U}" {
		t.Fatalf("mana %q", p.ManaCost)
	}
}
