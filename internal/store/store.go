package store

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// Copy is one owned printing and finish.
type Copy struct {
	SetCode      string `json:"set"`
	Number       string `json:"number"`
	Foil         bool   `json:"foil"`
	Qty          int    `json:"owned"`
	Name         string `json:"name"`
	SetName      string `json:"set_name"`
	Rarity       string `json:"rarity"`
	ManaCost     string `json:"mana_cost"`
	TypeLine     string `json:"type_line"`
	Colors       string `json:"colors"`
	PriceUSD     string `json:"price_usd"`
	PriceOn      string `json:"price_on"`
	FrontImage   string `json:"front_image"`
	BackImage    string `json:"back_image"`
	OracleText   string `json:"oracle_text"`
	InThis       int    `json:"in_this"`
	InOtherBuilt int    `json:"in_other_built"`
	OtherDecks   string `json:"other_decks"`
}

// Deck is a named list. Status is draft or built.
type Deck struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Cards       int    `json:"cards"`
}

// Entry is one printing inside a deck.
type Entry struct {
	DeckID   int    `json:"deck_id"`
	SetCode  string `json:"set"`
	Number   string `json:"number"`
	Foil     bool   `json:"foil"`
	Qty      int    `json:"qty"`
	Name     string `json:"name"`
	ManaCost string `json:"mana_cost"`
	TypeLine string `json:"type_line"`
}

// Store is the SQLite collection.
type Store struct {
	db *sql.DB
}

// Open creates the file and schema.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close releases the database.
func (s *Store) Close() error { return s.db.Close() }

// AddCopy increases the quantity of a printing. Foil and nonfoil are separate.
func (s *Store) AddCopy(set, number string, foil bool, qty int) error {
	if qty <= 0 {
		return fmt.Errorf("qty must be positive")
	}
	_, err := s.db.Exec(`
		INSERT INTO copies(set_code, collector_number, foil, qty)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(set_code, collector_number, foil)
		DO UPDATE SET qty = qty + excluded.qty`, set, number, boolInt(foil), qty)
	return err
}

// SetQty replaces the owned count. Used by the markdown import.
func (s *Store) SetQty(set, number string, foil bool, qty int) error {
	if qty < 0 {
		return fmt.Errorf("qty must not be negative")
	}
	_, err := s.db.Exec(`
		INSERT INTO copies(set_code, collector_number, foil, qty)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(set_code, collector_number, foil)
		DO UPDATE SET qty = excluded.qty`, set, number, boolInt(foil), qty)
	return err
}

// Qty returns the owned count, or 0 when the printing is absent.
func (s *Store) Qty(set, number string, foil bool) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT qty FROM copies WHERE set_code=? AND collector_number=? AND foil=?`, set, number, boolInt(foil)).Scan(&n)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return n, err
}

// SetMeta fills the Scryfall fields of an existing row and does not change qty.
func (s *Store) SetMeta(c Copy) error {
	_, err := s.db.Exec(`
		UPDATE copies SET name=?, set_name=?, rarity=?, mana_cost=?, type_line=?, colors=?, price_usd=?, price_on=?, front_image=?, back_image=?, oracle_text=?
		WHERE set_code=? AND collector_number=? AND foil=?`,
		c.Name, c.SetName, c.Rarity, c.ManaCost, c.TypeLine, c.Colors, c.PriceUSD, c.PriceOn, c.FrontImage, c.BackImage, c.OracleText,
		c.SetCode, c.Number, boolInt(c.Foil))
	return err
}

// Get returns one printing. The bool is false when it is absent.
func (s *Store) Get(set, number string, foil bool) (Copy, bool, error) {
	rows, err := s.db.Query(`
		SELECT set_code, collector_number, foil, qty, name, set_name, rarity, mana_cost, type_line, colors, price_usd, price_on, front_image, back_image, oracle_text
		FROM copies WHERE set_code=? AND collector_number=? AND foil=?`, set, number, boolInt(foil))
	if err != nil {
		return Copy{}, false, err
	}
	defer rows.Close()
	list, err := scanCopies(rows)
	if err != nil || len(list) == 0 {
		return Copy{}, false, err
	}
	return list[0], true, nil
}

// Has reports whether the printing row exists.
func (s *Store) Has(set, number string, foil bool) (bool, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM copies WHERE set_code=? AND collector_number=? AND foil=?`, set, number, boolInt(foil)).Scan(&n)
	return n > 0, err
}

// List returns owned printings. q matches name, set name, set code, or collector number.
func (s *Store) List(q string) ([]Copy, error) {
	q = strings.TrimSpace(q)
	like := "%" + q + "%"
	rows, err := s.db.Query(`
		SELECT set_code, collector_number, foil, qty, name, set_name, rarity, mana_cost, type_line, colors, price_usd, price_on, front_image, back_image, oracle_text
		FROM copies
		WHERE ? = '' OR name LIKE ? OR set_name LIKE ? OR set_code LIKE ? OR collector_number LIKE ? OR (set_code || ' ' || collector_number) LIKE ?
		ORDER BY name COLLATE NOCASE, set_code, collector_number`,
		q, like, like, like, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCopies(rows)
}

// CreateDeck inserts a deck and returns its id.
func (s *Store) CreateDeck(name, description, status string) (int, error) {
	if status != "draft" && status != "built" {
		return 0, fmt.Errorf("status must be draft or built")
	}
	res, err := s.db.Exec(`INSERT INTO decks(name, description, status) VALUES(?, ?, ?)`, name, description, status)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

// UpdateDeck changes name, description, and status.
func (s *Store) UpdateDeck(id int, name, description, status string) error {
	if status != "draft" && status != "built" {
		return fmt.Errorf("status must be draft or built")
	}
	_, err := s.db.Exec(`UPDATE decks SET name=?, description=?, status=? WHERE id=?`, name, description, status, id)
	return err
}

// DeckStatus returns the status of a deck.
func (s *Store) DeckStatus(id int) (string, error) {
	var status string
	err := s.db.QueryRow(`SELECT status FROM decks WHERE id=?`, id).Scan(&status)
	return status, err
}

// GetDeck returns the deck and its name, description, and status.
func (s *Store) GetDeck(id int) (Deck, error) {
	var d Deck
	err := s.db.QueryRow(`
		SELECT d.id, d.name, d.description, d.status, COALESCE(SUM(e.qty), 0)
		FROM decks d LEFT JOIN entries e ON e.deck_id = d.id
		WHERE d.id=? GROUP BY d.id`, id).Scan(&d.ID, &d.Name, &d.Description, &d.Status, &d.Cards)
	return d, err
}

// ListDecks returns every deck with its card count.
func (s *Store) ListDecks() ([]Deck, error) {
	rows, err := s.db.Query(`
		SELECT d.id, d.name, d.description, d.status, COALESCE(SUM(e.qty), 0)
		FROM decks d LEFT JOIN entries e ON e.deck_id = d.id
		GROUP BY d.id ORDER BY d.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Deck
	for rows.Next() {
		var d Deck
		if err := rows.Scan(&d.ID, &d.Name, &d.Description, &d.Status, &d.Cards); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// Entries lists cards in a deck, joined to the collection for name and mana.
func (s *Store) Entries(deckID int) ([]Entry, error) {
	rows, err := s.db.Query(`
		SELECT e.set_code, e.collector_number, e.foil, e.qty, COALESCE(c.name, ''), COALESCE(c.mana_cost, ''), COALESCE(c.type_line, '')
		FROM entries e
		LEFT JOIN copies c ON c.set_code=e.set_code AND c.collector_number=e.collector_number AND c.foil=e.foil
		WHERE e.deck_id=?
		ORDER BY c.name COLLATE NOCASE`, deckID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		var e Entry
		var foil int
		if err := rows.Scan(&e.SetCode, &e.Number, &foil, &e.Qty, &e.Name, &e.ManaCost, &e.TypeLine); err != nil {
			return nil, err
		}
		e.Foil = foil == 1
		out = append(out, e)
	}
	return out, rows.Err()
}

// SetEntry sets the absolute quantity of a printing in a deck. Zero deletes the row.
func (s *Store) SetEntry(deckID int, set, number string, foil bool, qty int) error {
	if qty < 0 {
		return fmt.Errorf("qty must not be negative")
	}
	if qty == 0 {
		_, err := s.db.Exec(`DELETE FROM entries WHERE deck_id=? AND set_code=? AND collector_number=? AND foil=?`, deckID, set, number, boolInt(foil))
		return err
	}
	_, err := s.db.Exec(`
		INSERT INTO entries(deck_id, set_code, collector_number, foil, qty)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(deck_id, set_code, collector_number, foil)
		DO UPDATE SET qty = excluded.qty`, deckID, set, number, boolInt(foil), qty)
	return err
}

// EntryQty returns how many of a printing a deck holds.
func (s *Store) EntryQty(deckID int, set, number string, foil bool) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT qty FROM entries WHERE deck_id=? AND set_code=? AND collector_number=? AND foil=?`, deckID, set, number, boolInt(foil)).Scan(&n)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return n, err
}

// OtherBuilt reports how many copies other built decks hold, and their names.
func (s *Store) OtherBuilt(set, number string, foil bool, exceptDeck int) (int, string, error) {
	var n int
	err := s.db.QueryRow(`
		SELECT COALESCE(SUM(e.qty), 0)
		FROM entries e JOIN decks d ON d.id = e.deck_id
		WHERE d.status='built' AND d.id != ? AND e.set_code=? AND e.collector_number=? AND e.foil=?`,
		exceptDeck, set, number, boolInt(foil)).Scan(&n)
	if err != nil {
		return 0, "", err
	}
	rows, err := s.db.Query(`
		SELECT d.name FROM entries e JOIN decks d ON d.id = e.deck_id
		WHERE d.status='built' AND d.id != ? AND e.set_code=? AND e.collector_number=? AND e.foil=? AND e.qty > 0
		ORDER BY d.name`, exceptDeck, set, number, boolInt(foil))
	if err != nil {
		return 0, "", err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return 0, "", err
		}
		names = append(names, name)
	}
	return n, strings.Join(names, ", "), rows.Err()
}

// BuiltLines returns every printing used by built decks, for the status-change check.
func (s *Store) BuiltDemand(exceptDeck int) (map[string]int, error) {
	rows, err := s.db.Query(`
		SELECT e.set_code, e.collector_number, e.foil, SUM(e.qty)
		FROM entries e JOIN decks d ON d.id = e.deck_id
		WHERE d.status='built' AND d.id != ?
		GROUP BY e.set_code, e.collector_number, e.foil`, exceptDeck)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var set, number string
		var foil, qty int
		if err := rows.Scan(&set, &number, &foil, &qty); err != nil {
			return nil, err
		}
		out[key(set, number, foil == 1)] = qty
	}
	return out, rows.Err()
}

func key(set, number string, foil bool) string {
	f := "0"
	if foil {
		f = "1"
	}
	return set + "|" + number + "|" + f
}

func scanCopies(rows *sql.Rows) ([]Copy, error) {
	var out []Copy
	for rows.Next() {
		var c Copy
		var foil int
		if err := rows.Scan(&c.SetCode, &c.Number, &foil, &c.Qty, &c.Name, &c.SetName, &c.Rarity, &c.ManaCost, &c.TypeLine, &c.Colors, &c.PriceUSD, &c.PriceOn, &c.FrontImage, &c.BackImage, &c.OracleText); err != nil {
			return nil, err
		}
		c.Foil = foil == 1
		out = append(out, c)
	}
	if out == nil {
		out = []Copy{}
	}
	return out, rows.Err()
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
