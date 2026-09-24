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
