package backup

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/MelQ29/mtg/internal/store"
)

// Export writes a zip of collection.json and the image files it names.
// The archive is the portable copy. The git repo does not contain it.
func Export(s *store.Store, imageDir string) ([]byte, error) {
	snap, err := s.Dump()
	if err != nil {
		return nil, err
	}
	raw, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	entry, err := zw.Create("collection.json")
	if err != nil {
		return nil, err
	}
	if _, err := entry.Write(raw); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, c := range snap.Copies {
		for _, rel := range []string{c.FrontImage, c.BackImage} {
			if rel == "" || seen[rel] || strings.Contains(rel, "..") {
				continue
			}
			seen[rel] = true
			body, err := os.ReadFile(filepath.Join(imageDir, filepath.FromSlash(rel)))
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, err
			}
			img, err := zw.Create("images/" + rel)
			if err != nil {
				return nil, err
			}
			if _, err := img.Write(body); err != nil {
				return nil, err
			}
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Import replaces the local database and image files with an archive from Export.
func Import(s *store.Store, imageDir string, zipBytes []byte) error {
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return err
	}
	var snap store.Snapshot
	found := false
	for _, f := range zr.File {
		if f.Name != "collection.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		body, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return err
		}
		if err := json.Unmarshal(body, &snap); err != nil {
			return err
		}
		found = true
		break
	}
	if !found {
		return os.ErrNotExist
	}
	staging := imageDir + ".importing"
	if err := os.RemoveAll(staging); err != nil {
		return err
	}
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return err
	}
	for _, f := range zr.File {
		if !strings.HasPrefix(f.Name, "images/") || f.FileInfo().IsDir() {
			continue
		}
		rel := strings.TrimPrefix(f.Name, "images/")
		if rel == "" || strings.Contains(rel, "..") {
			continue
		}
		dest := filepath.Join(staging, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		body, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return err
		}
		if err := os.WriteFile(dest, body, 0o644); err != nil {
			return err
		}
	}
	if err := s.Restore(snap); err != nil {
		os.RemoveAll(staging)
		return err
	}
	if err := os.RemoveAll(imageDir); err != nil {
		return err
	}
	return os.Rename(staging, imageDir)
}
