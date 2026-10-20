package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MelQ29/mtg/internal/backup"
	"github.com/MelQ29/mtg/internal/store"
)

func TestImportArchiveLoadsOnce(t *testing.T) {
	dir := t.TempDir()
	src, err := store.Open(filepath.Join(dir, "src.db"))
	if err != nil {
		t.Fatal(err)
	}
	img := filepath.Join(dir, "src-images")
	if err := os.MkdirAll(filepath.Join(img, "ecl", "1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(img, "ecl", "1", "front.jpg"), []byte("jpeg"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := src.AddCopy("ecl", "1", false, 2); err != nil {
		t.Fatal(err)
	}
	if err := src.SetMeta(store.Copy{SetCode: "ecl", Number: "1", Name: "Island", FrontImage: "ecl/1/front.jpg"}); err != nil {
		t.Fatal(err)
	}
	deckID, err := src.CreateDeck("Соня", "", "built")
	if err != nil {
		t.Fatal(err)
	}
	if err := src.SetEntry(deckID, "ecl", "1", false, 1); err != nil {
		t.Fatal(err)
	}
	blob, err := backup.Export(src, img)
	src.Close()
	if err != nil {
		t.Fatal(err)
	}
	zipPath := filepath.Join(dir, "kitchen-collection.zip")
	if err := os.WriteFile(zipPath, blob, 0o644); err != nil {
		t.Fatal(err)
	}

	dstDir := filepath.Join(dir, "dst")
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dst, err := store.Open(filepath.Join(dstDir, "kitchen.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer dst.Close()
	images := filepath.Join(dstDir, "images")
	if err := importArchive(dst, images, zipPath, false); err != nil {
		t.Fatal(err)
	}
	n, err := dst.OwnedCount()
	if err != nil || n != 2 {
		t.Fatalf("imported count %d %v", n, err)
	}
	decks, err := dst.ListDecks()
	if err != nil || len(decks) != 1 || decks[0].Name != "Соня" {
		t.Fatalf("imported decks %+v %v", decks, err)
	}
	decks[0].Name = "Edited"
	if err := dst.UpdateDeck(decks[0]); err != nil {
		t.Fatal(err)
	}
	if err := dst.AddCopy("ecl", "1", false, 1); err != nil {
		t.Fatal(err)
	}
	if err := importArchive(dst, images, zipPath, false); err != nil {
		t.Fatal(err)
	}
	n, err = dst.OwnedCount()
	if err != nil || n != 3 {
		t.Fatalf("second launch should keep edits, count %d %v", n, err)
	}
	decks, err = dst.ListDecks()
	if err != nil || len(decks) != 1 || decks[0].Name != "Edited" {
		t.Fatalf("second launch reimported decks %+v %v", decks, err)
	}
	if err := importArchive(dst, images, zipPath, true); err != nil {
		t.Fatal(err)
	}
	n, err = dst.OwnedCount()
	if err != nil || n != 2 {
		t.Fatalf("reimport count %d %v", n, err)
	}
}
