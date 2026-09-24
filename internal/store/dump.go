package store

// Snapshot is the portable collection. It is not stored in git.
type Snapshot struct {
	Copies  []Copy  `json:"copies"`
	Decks   []Deck  `json:"decks"`
	Entries []Entry `json:"entries"`
}

// Dump reads every owned printing, deck, and deck entry.
func (s *Store) Dump() (Snapshot, error) {
	copies, err := s.List("")
	if err != nil {
		return Snapshot{}, err
	}
	decks, err := s.ListDecks()
	if err != nil {
		return Snapshot{}, err
	}
	var entries []Entry
	for _, d := range decks {
		rows, err := s.rawEntries(d.ID)
		if err != nil {
			return Snapshot{}, err
		}
		entries = append(entries, rows...)
	}
	if entries == nil {
		entries = []Entry{}
	}
	return Snapshot{Copies: copies, Decks: decks, Entries: entries}, nil
}

func (s *Store) rawEntries(deckID int) ([]Entry, error) {
	rows, err := s.db.Query(`SELECT deck_id, set_code, collector_number, foil, qty FROM entries WHERE deck_id=?`, deckID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		var e Entry
		var foil int
		if err := rows.Scan(&e.DeckID, &e.SetCode, &e.Number, &foil, &e.Qty); err != nil {
			return nil, err
		}
		e.Foil = foil == 1
		out = append(out, e)
	}
	return out, rows.Err()
}

// Restore replaces the local collection with a snapshot.
func (s *Store) Restore(snap Snapshot) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM entries`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM decks`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM copies`); err != nil {
		return err
	}
	for _, c := range snap.Copies {
		if _, err := tx.Exec(`
			INSERT INTO copies(set_code, collector_number, foil, qty, name, set_name, rarity, mana_cost, type_line, colors, price_usd, price_on, front_image, back_image, oracle_text)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			c.SetCode, c.Number, boolInt(c.Foil), c.Qty, c.Name, c.SetName, c.Rarity, c.ManaCost, c.TypeLine, c.Colors, c.PriceUSD, c.PriceOn, c.FrontImage, c.BackImage, c.OracleText); err != nil {
			return err
		}
	}
	for _, d := range snap.Decks {
		if _, err := tx.Exec(`INSERT INTO decks(id, name, description, status) VALUES(?, ?, ?, ?)`, d.ID, d.Name, d.Description, d.Status); err != nil {
			return err
		}
	}
	for _, e := range snap.Entries {
		if _, err := tx.Exec(`INSERT INTO entries(deck_id, set_code, collector_number, foil, qty) VALUES(?, ?, ?, ?, ?)`, e.DeckID, e.SetCode, e.Number, boolInt(e.Foil), e.Qty); err != nil {
			return err
		}
	}
	return tx.Commit()
}
