package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MelQ29/mtg/internal/store"
)

func TestExportImportRoundTrip(t *testing.T) {
	dir := t.TempDir()
	src, err := store.Open(filepath.Join(dir, "a.db"))
	if err != nil {
		t.Fatal(err)
	}
	img := filepath.Join(dir, "images")
	if err := os.MkdirAll(filepath.Join(img, "ecl", "317"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(img, "ecl", "317", "front.jpg"), []byte("jpeg"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := src.AddCopy("ecl", "317", false, 1); err != nil {
		t.Fatal(err)
	}
	if err := src.SetMeta(store.Copy{SetCode: "ecl", Number: "317", Name: "Hexing Squelcher", FrontImage: "ecl/317/front.jpg", Qty: 1}); err != nil {
		t.Fatal(err)
	}
	id, err := src.CreateDeck("Black-Red", "notes", "built")
	if err != nil {
		t.Fatal(err)
	}
	if err := src.SetEntry(id, "ecl", "317", false, 1); err != nil {
		t.Fatal(err)
	}
	blob, err := Export(src, img)
	src.Close()
	if err != nil {
		t.Fatal(err)
	}

	dstDir := filepath.Join(dir, "restored")
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dst, err := store.Open(filepath.Join(dstDir, "b.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer dst.Close()
	images := filepath.Join(dstDir, "images")
	if err := Import(dst, images, blob); err != nil {
		t.Fatal(err)
	}
	qty, err := dst.Qty("ecl", "317", false)
	if err != nil || qty != 1 {
		t.Fatalf("qty %d %v", qty, err)
	}
	body, err := os.ReadFile(filepath.Join(images, "ecl", "317", "front.jpg"))
	if err != nil || string(body) != "jpeg" {
		t.Fatalf("image %q %v", body, err)
	}
	decks, err := dst.ListDecks()
	if err != nil || len(decks) != 1 || decks[0].Name != "Black-Red" || decks[0].Cards != 1 {
		t.Fatalf("%+v %v", decks, err)
	}
}
