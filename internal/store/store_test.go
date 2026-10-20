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
	id, err := s.CreateDeck("priced", "", "draft")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetEntry(id, "ecl", "1", false, 3); err != nil {
		t.Fatal(err)
	}
	if err := s.SetEntry(id, "fin", "2", false, 1); err != nil {
		t.Fatal(err)
	}
	entries, err := s.Entries(id)
	if err != nil {
		t.Fatal(err)
	}
	prices := map[string]string{}
	for _, e := range entries {
		prices[e.SetCode] = e.PriceUSD
	}
	if prices["ecl"] != "1.25" || prices["fin"] != "" {
		t.Fatalf("entry prices %+v", entries)
	}
	count, err := s.OwnedCount()
	if err != nil {
		t.Fatal(err)
	}
	if count != 6 {
		t.Fatalf("count %d", count)
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

func TestDeckCoverUsesTheNamesake(t *testing.T) {
	s := openTemp(t)
	add := func(set, number, name, typeLine, image string, foil bool) {
		t.Helper()
		if err := s.AddCopy(set, number, foil, 1); err != nil {
			t.Fatal(err)
		}
		if err := s.SetMeta(Copy{SetCode: set, Number: number, Foil: foil, Name: name, TypeLine: typeLine, FrontImage: image}); err != nil {
			t.Fatal(err)
		}
	}
	add("fin", "3", "Adelbert Steiner", "Legendary Creature — Human Knight", "fin/3/front.jpg", false)
	add("fin", "552", "Cloud, Planet's Champion", "Legendary Creature — Human Soldier Mercenary", "fin/552/front.jpg", true)
	add("fin", "11", "Cloudbound Moogle", "Creature — Moogle", "fin/11/front.jpg", false)
	cloud, err := s.CreateDeck("Cloud — Final Fantasy Starter", "", "built")
	if err != nil {
		t.Fatal(err)
	}
	for _, number := range []string{"3", "11"} {
		if err := s.SetEntry(cloud, "fin", number, false, 1); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.SetEntry(cloud, "fin", "552", true, 1); err != nil {
		t.Fatal(err)
	}
	add("fin", "88", "Al Bhed Salvagers", "Creature — Human Rogue", "fin/88/front.jpg", false)
	add("fin", "116", "Sephiroth's Intervention", "Instant", "fin/116/front.jpg", false)
	add("fin", "553", "Sephiroth, Planet's Heir", "Legendary Creature — Human Avatar Soldier", "fin/553/front.jpg", true)
	seph, err := s.CreateDeck("Sephiroth — Final Fantasy Starter", "", "built")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetEntry(seph, "fin", "88", false, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.SetEntry(seph, "fin", "116", false, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.SetEntry(seph, "fin", "553", true, 1); err != nil {
		t.Fatal(err)
	}
	decks, err := s.ListDecks()
	if err != nil {
		t.Fatal(err)
	}
	if len(decks) != 2 || decks[0].Cover != "fin/552/front.jpg" || decks[1].Cover != "fin/553/front.jpg" {
		t.Fatalf("covers %+v", decks)
	}
}

func TestListMatchesRulesText(t *testing.T) {
	s := openTemp(t)
	add := func(number, name, oracle string) {
		t.Helper()
		if err := s.AddCopy("ecl", number, false, 1); err != nil {
			t.Fatal(err)
		}
		if err := s.SetMeta(Copy{SetCode: "ecl", Number: number, Name: name, OracleText: oracle}); err != nil {
			t.Fatal(err)
		}
	}
	add("1", "Overkill", "Destroy target creature.")
	add("2", "Shock", "Shock deals 2 damage to any target.")
	got, err := s.List("destroy target")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Overkill" {
		t.Fatalf("rules search %+v", got)
	}
	byName, err := s.List("shock")
	if err != nil {
		t.Fatal(err)
	}
	if len(byName) != 1 || byName[0].Name != "Shock" {
		t.Fatalf("name search %+v", byName)
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
