package images

import (
	"testing"

	"github.com/MelQ29/mtg/internal/scryfall"
)

func TestSaveWritesBothFaces(t *testing.T) {
	dir := t.TempDir()
	front, back, err := Save(dir, "tla", "298", []scryfall.Face{
		{Name: "Aang, Swift Savior", ImageURL: "http://front"},
		{Name: "Aang and La, Ocean's Fury", ImageURL: "http://back"},
	}, func(url string) ([]byte, error) {
		return []byte(url), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if front == "" || back == "" {
		t.Fatalf("front %q back %q", front, back)
	}
}
