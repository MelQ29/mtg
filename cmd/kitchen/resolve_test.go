package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MelQ29/mtg/internal/scryfall"
	"github.com/MelQ29/mtg/internal/store"
)

func TestLookupName(t *testing.T) {
	if got := lookupName("Wildvine Pummler"); got != "Wildvine Pummeler" {
		t.Fatalf("pummler %q", got)
	}
	if got := lookupName("Нексус Ликолесья (Maskwood Nexus)"); got != "Maskwood Nexus" {
		t.Fatalf("nexus %q", got)
	}
	if got := lookupName("Lightning Strike"); got != "Lightning Strike" {
		t.Fatalf("plain %q", got)
	}
}

func TestIndexByNameMatchesEitherFace(t *testing.T) {
	idx := indexByName([]store.Copy{{
		SetCode: "fin",
		Number:  "158",
		Name:    "Sidequest: Play Blitzball // World Champion, Celestial Weapon",
	}})
	got := idx[scryfall.FoldName("Sidequest: Play Blitzball")]
	if got.Number != "158" {
		t.Fatalf("front %+v", got)
	}
	if idx[scryfall.FoldName("World Champion, Celestial Weapon")].Number != "158" {
		t.Fatal("back face was not indexed")
	}
}

func TestLoadDeckFillsMissingFaceAndDoesNotDuplicate(t *testing.T) {
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "kitchen.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.AddCopy("fin", "158", false, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.SetMeta(store.Copy{
		SetCode: "fin", Number: "158",
		Name: "Sidequest: Play Blitzball // World Champion, Celestial Weapon",
	}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "deck.md")
	body := "- 1 Sidequest: Play Blitzball **[R]** — note\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := loadDeck(s, path, "Black-Red"); err != nil {
		t.Fatal(err)
	}
	if err := loadDeck(s, path, "Black-Red"); err != nil {
		t.Fatal(err)
	}
	decks, err := s.ListDecks()
	if err != nil {
		t.Fatal(err)
	}
	if len(decks) != 1 || decks[0].Cards != 1 || decks[0].Name != "Black-Red" {
		t.Fatalf("decks %+v", decks)
	}
}
