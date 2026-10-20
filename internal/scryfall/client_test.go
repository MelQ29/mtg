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

func TestParseMeldPointsAtTheCombinedCard(t *testing.T) {
	p, err := ParseCard([]byte(`{
		"name": "Vanille, Cheerful l'Cie",
		"set": "fin",
		"collector_number": "392",
		"layout": "meld",
		"image_uris": {"normal": "http://front"},
		"all_parts": [
			{"component": "meld_part", "name": "Vanille, Cheerful l'Cie", "uri": "http://vanille"},
			{"component": "meld_result", "name": "Ragnarok, Divine Deliverance", "uri": "http://ragnarok"}
		]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if p.MeldURI != "http://ragnarok" || len(p.Faces) != 1 {
		t.Fatalf("%+v", p)
	}
}

func TestParseMeldResultHasNoSecondURI(t *testing.T) {
	p, err := ParseCard([]byte(`{
		"name": "Ragnarok, Divine Deliverance",
		"set": "fin",
		"collector_number": "526b",
		"layout": "meld",
		"image_uris": {"normal": "http://back"},
		"all_parts": [
			{"component": "meld_result", "name": "Ragnarok, Divine Deliverance", "uri": "http://ragnarok"}
		]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if p.MeldURI != "" || len(p.Faces) != 1 {
		t.Fatalf("%+v", p)
	}
}

func TestFetchMeldAttachesBackFace(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/cards/fin/392") {
			_, _ = w.Write([]byte(`{
				"name": "Vanille, Cheerful l'Cie",
				"set": "fin",
				"collector_number": "392",
				"layout": "meld",
				"image_uris": {"normal": "http://front"},
				"all_parts": [
					{"component": "meld_result", "name": "Ragnarok, Divine Deliverance", "uri": "` + srv.URL + `/cards/ragnarok"}
				]
			}`))
			return
		}
		_, _ = w.Write([]byte(`{
			"name": "Ragnarok, Divine Deliverance",
			"set": "fin",
			"collector_number": "526b",
			"layout": "meld",
			"image_uris": {"normal": "http://back"}
		}`))
	}))
	defer srv.Close()
	p, err := (&Client{HTTP: srv.Client(), Base: srv.URL}).Fetch("fin", "392")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Faces) != 2 || p.Faces[0].ImageURL != "http://front" || p.Faces[1].Name != "Ragnarok, Divine Deliverance" || p.Faces[1].ImageURL != "http://back" {
		t.Fatalf("%+v", p.Faces)
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
