package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/MelQ29/mtg/internal/backup"
	"github.com/MelQ29/mtg/internal/httpapi"
	"github.com/MelQ29/mtg/internal/importmd"
	"github.com/MelQ29/mtg/internal/scryfall"
	"github.com/MelQ29/mtg/internal/store"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8787", "listen address")
	dbPath := flag.String("db", "data/kitchen.db", "sqlite file, kept out of git")
	imageDir := flag.String("images", "data/images", "downloaded card images, kept out of git")
	importPath := flag.String("import", "", "markdown collection to load once into the local database")
	archivePath := flag.String("archive", "", "kitchen collection zip to import before serving")
	reimport := flag.Bool("reimport", false, "replace an existing collection with -archive")
	resolve := flag.Bool("resolve", false, "download Scryfall art for cards that have none")
	deckPath := flag.String("deck", "", "markdown deck to add as a built deck")
	deckName := flag.String("deck-name", "Black-Red", "name for -deck")
	webDir := flag.String("web", "web", "frontend files")
	flag.Parse()

	if err := os.MkdirAll(filepath.Dir(*dbPath), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(*imageDir, 0o755); err != nil {
		log.Fatal(err)
	}
	s, err := store.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	if err := importArchive(s, *imageDir, *archivePath, *reimport); err != nil {
		log.Fatal(err)
	}
	if *importPath != "" {
		if err := loadMarkdown(s, *importPath); err != nil {
			log.Fatal(err)
		}
		log.Printf("imported %s", *importPath)
	}
	if *resolve {
		resolveMissing(s, &scryfall.Client{}, *imageDir)
	}
	if *deckPath != "" {
		if err := loadDeck(s, *deckPath, *deckName); err != nil {
			log.Fatal(err)
		}
	}

	host, port, err := net.SplitHostPort(*addr)
	if err != nil {
		log.Fatal(err)
	}
	if host != "127.0.0.1" && host != "localhost" && host != "::1" {
		log.Fatal("refusing to listen outside localhost")
	}
	_ = port

	api := httpapi.Handler(s, httpapi.LiveCards{Client: &scryfall.Client{}, ImageDir: *imageDir}, *imageDir)
	mux := http.NewServeMux()
	mux.Handle("/api/", api)
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(*imageDir))))
	mux.Handle("/", http.FileServer(http.Dir(*webDir)))

	log.Printf("kitchen listening on http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

// importArchive loads a zip from Export. An existing collection is left
// alone unless reimport is set, so a second launch does not wipe edits.
func importArchive(s *store.Store, imageDir, path string, reimport bool) error {
	if path == "" {
		return nil
	}
	n, err := s.OwnedCount()
	if err != nil {
		return err
	}
	if n > 0 && !reimport {
		log.Printf("collection already has %d cards; not importing %s", n, path)
		return nil
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := backup.Import(s, imageDir, blob); err != nil {
		return err
	}
	log.Printf("imported collection archive %s", path)
	return nil
}

func loadMarkdown(s *store.Store, path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	rows, err := importmd.Parse(string(body))
	if err != nil {
		return err
	}
	for _, row := range rows {
		set, number := row.Set, row.Number
		if set == "" {
			set = "pending"
			number = row.Name
		}
		if err := s.SetQty(set, number, row.Foil, row.Qty); err != nil {
			return err
		}
		if err := s.SetMeta(store.Copy{
			SetCode: set, Number: number, Foil: row.Foil,
			Name: row.Name, Rarity: row.Rarity, PriceUSD: row.Price, PriceOn: row.PriceOn,
		}); err != nil {
			return err
		}
	}
	return nil
}
