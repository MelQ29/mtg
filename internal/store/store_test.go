package store

import (
	"path/filepath"
	"testing"
)

func openTemp(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestValueUSDMultipliesQuantity(t *testing.T) {
	s := openTemp(t)
	if err := s.AddCopy("ecl", "1", false, 2); err != nil {
		t.Fatal(err)
	}
	if err := s.SetMeta(Copy{SetCode: "ecl", Number: "1", PriceUSD: "1.25"}); err != nil {
		t.Fatal(err)
	}
	if err := s.AddCopy("fin", "2", false, 4); err != nil {
		t.Fatal(err)
	}
	if err := s.SetMeta(Copy{SetCode: "fin", Number: "2", PriceUSD: ""}); err != nil {
		t.Fatal(err)
	}
	got, err := s.ValueUSD()
	if err != nil {
		t.Fatal(err)
	}
	if got != 2.5 {
		t.Fatalf("value %v", got)
	}
}

func TestDeckCoverPrefersACreature(t *testing.T) {
	s := openTemp(t)
	add := func(set, number, name, typeLine, image string) {
		t.Helper()
		if err := s.AddCopy(set, number, false, 1); err != nil {
			t.Fatal(err)
		}
		if err := s.SetMeta(Copy{SetCode: set, Number: number, Name: name, TypeLine: typeLine, FrontImage: image}); err != nil {
			t.Fatal(err)
		}
	}
	add("tla", "1", "Aaa Land", "Land", "tla/1/front.jpg")
	add("tla", "2", "Aab Spell", "Instant", "tla/2/front.jpg")
	add("tla", "3", "Zzz Beast", "Creature — Beast", "tla/3/front.jpg")
	id, err := s.CreateDeck("With creature", "", "built")
	if err != nil {
		t.Fatal(err)
	}
	for _, number := range []string{"1", "2", "3"} {
		if err := s.SetEntry(id, "tla", number, false, 1); err != nil {
			t.Fatal(err)
		}
	}
	spellOnly, err := s.CreateDeck("No creature", "", "draft")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetEntry(spellOnly, "tla", "1", false, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.SetEntry(spellOnly, "tla", "2", false, 1); err != nil {
		t.Fatal(err)
	}
	decks, err := s.ListDecks()
	if err != nil {
		t.Fatal(err)
	}
	if len(decks) != 2 || decks[0].Cover != "tla/3/front.jpg" || decks[1].Cover != "tla/2/front.jpg" {
		t.Fatalf("covers %+v", decks)
	}
	entries, err := s.Entries(id)
	if err != nil {
		t.Fatal(err)
	}
	var land Entry
	for _, e := range entries {
		if e.Number == "1" {
			land = e
		}
	}
	if land.FrontImage != "tla/1/front.jpg" {
		t.Fatalf("entry image %+v", land)
	}
}

func TestAddCopyIncreasesSamePrinting(t *testing.T) {
	s := openTemp(t)
	if err := s.AddCopy("ecl", "317", false, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.AddCopy("ecl", "317", false, 1); err != nil {
		t.Fatal(err)
	}
	got, err := s.Qty("ecl", "317", false)
	if err != nil || got != 2 {
		t.Fatalf("qty=%d err=%v", got, err)
	}
	foil, err := s.Qty("ecl", "317", true)
	if err != nil || foil != 0 {
		t.Fatalf("foil qty=%d err=%v", foil, err)
	}
}
