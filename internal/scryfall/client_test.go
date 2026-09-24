package scryfall

import (
	"os"
	"path/filepath"
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
