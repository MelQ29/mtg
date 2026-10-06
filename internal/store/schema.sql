CREATE TABLE IF NOT EXISTS copies (
    set_code TEXT NOT NULL,
    collector_number TEXT NOT NULL,
    foil INTEGER NOT NULL,
    qty INTEGER NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    set_name TEXT NOT NULL DEFAULT '',
    rarity TEXT NOT NULL DEFAULT '',
    mana_cost TEXT NOT NULL DEFAULT '',
    type_line TEXT NOT NULL DEFAULT '',
    colors TEXT NOT NULL DEFAULT '',
    price_usd TEXT NOT NULL DEFAULT '',
    price_on TEXT NOT NULL DEFAULT '',
    front_image TEXT NOT NULL DEFAULT '',
    back_image TEXT NOT NULL DEFAULT '',
    oracle_text TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (set_code, collector_number, foil)
);

CREATE TABLE IF NOT EXISTS decks (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('draft', 'built')),
    format TEXT NOT NULL DEFAULT 'kitchen',
    commander_set TEXT NOT NULL DEFAULT '',
    commander_number TEXT NOT NULL DEFAULT '',
    commander_foil INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS entries (
    deck_id INTEGER NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    set_code TEXT NOT NULL,
    collector_number TEXT NOT NULL,
    foil INTEGER NOT NULL,
    qty INTEGER NOT NULL,
    PRIMARY KEY (deck_id, set_code, collector_number, foil)
);
