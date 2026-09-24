package images

import (
	"os"
	"path/filepath"

	"github.com/MelQ29/mtg/internal/scryfall"
)

// Save writes each face under dir/set/number. A second face is back.jpg.
// Returned paths are relative to dir.
func Save(dir, set, number string, faces []scryfall.Face, get func(url string) ([]byte, error)) (front, back string, err error) {
	folder := filepath.Join(dir, set, number)
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return "", "", err
	}
	write := func(name, url string) (string, error) {
		body, err := get(url)
		if err != nil {
			return "", err
		}
		rel := filepath.ToSlash(filepath.Join(set, number, name))
		if err := os.WriteFile(filepath.Join(dir, rel), body, 0o644); err != nil {
			return "", err
		}
		return rel, nil
	}
	if len(faces) == 0 {
		return "", "", os.ErrNotExist
	}
	front, err = write("front.jpg", faces[0].ImageURL)
	if err != nil {
		return "", "", err
	}
	if len(faces) > 1 {
		back, err = write("back.jpg", faces[1].ImageURL)
		if err != nil {
			return "", "", err
		}
	}
	return front, back, nil
}
