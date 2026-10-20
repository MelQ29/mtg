package scryfall

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Face is one side of a card. Single-faced cards have one.
type Face struct {
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

// Printing is one Scryfall card object, including both faces when present.
type Printing struct {
	Name         string `json:"name"`
	SetCode      string `json:"set"`
	SetName      string `json:"set_name"`
	Number       string `json:"number"`
	Rarity       string `json:"rarity"`
	Oracle       string `json:"oracle_text"`
	ManaCost     string `json:"mana_cost"`
	TypeLine     string `json:"type_line"`
	Colors       string `json:"colors"`
	Faces        []Face `json:"faces"`
	PriceNonfoil string `json:"price_nonfoil"`
	PriceFoil    string `json:"price_foil"`
	// MeldURI is the combined back of a meld card. Scryfall keeps that
	// picture on a separate object, not in this card's faces.
	MeldURI string `json:"meld_uri,omitempty"`
}

type rawCard struct {
	Name            string     `json:"name"`
	Set             string     `json:"set"`
	SetName         string     `json:"set_name"`
	CollectorNumber string     `json:"collector_number"`
	Rarity          string     `json:"rarity"`
	OracleText      string     `json:"oracle_text"`
	ManaCost        string     `json:"mana_cost"`
	TypeLine        string     `json:"type_line"`
	ColorIdentity   []string   `json:"color_identity"`
	ImageURIs       *rawImages `json:"image_uris"`
	Prices          rawPrices  `json:"prices"`
	CardFaces       []rawFace  `json:"card_faces"`
	Layout          string     `json:"layout"`
	AllParts        []rawPart  `json:"all_parts"`
}

type rawPart struct {
	Name      string `json:"name"`
	Component string `json:"component"`
	URI       string `json:"uri"`
}

type rawFace struct {
	Name       string     `json:"name"`
	OracleText string     `json:"oracle_text"`
	ManaCost   string     `json:"mana_cost"`
	TypeLine   string     `json:"type_line"`
	ImageURIs  *rawImages `json:"image_uris"`
}

type rawImages struct {
	Normal string `json:"normal"`
}

type rawPrices struct {
	USD     *string `json:"usd"`
	USDFoil *string `json:"usd_foil"`
}

type rawList struct {
	Data []json.RawMessage `json:"data"`
}

// ParseCard reads one Scryfall card object.
func ParseCard(body []byte) (Printing, error) {
	var raw rawCard
	if err := json.Unmarshal(body, &raw); err != nil {
		return Printing{}, err
	}
	if raw.Name == "" || raw.Set == "" || raw.CollectorNumber == "" {
		return Printing{}, fmt.Errorf("scryfall card is missing name, set, or collector number")
	}
	p := Printing{
		Name:         raw.Name,
		SetCode:      raw.Set,
		SetName:      raw.SetName,
		Number:       raw.CollectorNumber,
		Rarity:       raw.Rarity,
		Oracle:       raw.OracleText,
		ManaCost:     raw.ManaCost,
		TypeLine:     raw.TypeLine,
		Colors:       strings.Join(raw.ColorIdentity, ""),
		PriceNonfoil: deref(raw.Prices.USD),
		PriceFoil:    deref(raw.Prices.USDFoil),
		MeldURI:      meldResultURI(raw),
	}
	if raw.ImageURIs != nil && raw.ImageURIs.Normal != "" {
		p.Faces = []Face{{Name: raw.Name, ImageURL: raw.ImageURIs.Normal}}
	} else {
		var oracle []string
		for _, face := range raw.CardFaces {
			if face.ImageURIs == nil || face.ImageURIs.Normal == "" {
				continue
			}
			p.Faces = append(p.Faces, Face{Name: face.Name, ImageURL: face.ImageURIs.Normal})
			if face.OracleText != "" {
				oracle = append(oracle, face.OracleText)
			}
		}
		if p.Oracle == "" {
			p.Oracle = strings.Join(oracle, "\n")
		}
		if p.ManaCost == "" && len(raw.CardFaces) > 0 {
			p.ManaCost = raw.CardFaces[0].ManaCost
		}
		if p.TypeLine == "" && len(raw.CardFaces) > 0 {
			p.TypeLine = raw.CardFaces[0].TypeLine
		}
	}
	if len(p.Faces) == 0 {
		return Printing{}, fmt.Errorf("scryfall card %s has no image", raw.Name)
	}
	return p, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// meldResultURI is the Scryfall object for the card these two halves become.
// The result card lists itself, so a matching name means there is no other back.
func meldResultURI(raw rawCard) string {
	if raw.Layout != "meld" {
		return ""
	}
	for _, part := range raw.AllParts {
		if part.Component == "meld_result" && part.URI != "" && part.Name != raw.Name {
			return part.URI
		}
	}
	return ""
}

// Client calls the Scryfall API with a polite pause between requests.
type Client struct {
	HTTP *http.Client
	Base string
	mu   sync.Mutex
	last time.Time
}

func (c *Client) endpoint(path string) string {
	base := c.Base
	if base == "" {
		base = "https://api.scryfall.com"
	}
	return base + path
}

// NotFound reports a Scryfall search or card request that matched nothing.
func NotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "404")
}

// Named loads the canonical printing for a card name.
// Exact match is tried first. A fuzzy hit is kept only when the names still agree
// once case and punctuation are ignored, so a wrong card is not stored.
func (c *Client) Named(name string) (Printing, error) {
	body, err := c.get(c.endpoint("/cards/named?exact=" + url.QueryEscape(name)))
	if err != nil {
		body, err = c.get(c.endpoint("/cards/named?fuzzy=" + url.QueryEscape(name)))
		if err != nil {
			return Printing{}, err
		}
		p, err := ParseCard(body)
		if err != nil {
			return Printing{}, err
		}
		if FoldName(p.Name) != FoldName(name) {
			return Printing{}, fmt.Errorf("scryfall fuzzy %q is %q", name, p.Name)
		}
		return c.attachMeld(p)
	}
	p, err := ParseCard(body)
	if err != nil {
		return Printing{}, err
	}
	return c.attachMeld(p)
}

// FoldName compares card names without case or punctuation.
func FoldName(s string) string {
	var b strings.Builder
	space := false
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			space = false
			continue
		}
		if !space {
			b.WriteByte(' ')
			space = true
		}
	}
	return strings.TrimSpace(b.String())
}

// Fetch loads one printing by set code and collector number.
func (c *Client) Fetch(set, number string) (Printing, error) {
	body, err := c.get(c.endpoint("/cards/" + url.PathEscape(set) + "/" + url.PathEscape(number)))
	if err != nil {
		return Printing{}, err
	}
	p, err := ParseCard(body)
	if err != nil {
		return Printing{}, err
	}
	return c.attachMeld(p)
}

// attachMeld adds the combined meld card as a second face.
func (c *Client) attachMeld(p Printing) (Printing, error) {
	if p.MeldURI == "" || len(p.Faces) == 0 {
		return p, nil
	}
	body, err := c.get(p.MeldURI)
	if err != nil {
		return Printing{}, err
	}
	back, err := ParseCard(body)
	if err != nil {
		return Printing{}, err
	}
	if len(back.Faces) == 0 || back.Faces[0].ImageURL == "" || back.Faces[0].ImageURL == p.Faces[0].ImageURL {
		return p, nil
	}
	p.Faces = append(p.Faces, Face{Name: back.Name, ImageURL: back.Faces[0].ImageURL})
	return p, nil
}

// SearchPrintings lists printings for a name. An exact name is tried first.
// A partial name is used when that matches nothing, and no match is an empty list.
func (c *Client) SearchPrintings(name string) ([]Printing, error) {
	clean := strings.ReplaceAll(strings.TrimSpace(name), `"`, "")
	out, err := c.search(`!"` + clean + `" unique:prints`)
	if err == nil && len(out) > 0 {
		return out, nil
	}
	if err != nil && !NotFound(err) {
		return nil, err
	}
	if p, namedErr := c.Named(clean); namedErr == nil {
		out, err = c.search(`!"` + strings.ReplaceAll(p.Name, `"`, "") + `" unique:prints`)
		if err == nil && len(out) > 0 {
			return out, nil
		}
		if err != nil && !NotFound(err) {
			return nil, err
		}
	}
	out, err = c.search(`name:"` + clean + `" unique:prints`)
	if NotFound(err) {
		return []Printing{}, nil
	}
	return out, err
}

func (c *Client) search(query string) ([]Printing, error) {
	body, err := c.get(c.endpoint("/cards/search?q=" + url.QueryEscape(query)))
	if err != nil {
		return nil, err
	}
	var list rawList
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, err
	}
	out := make([]Printing, 0, len(list.Data))
	for _, item := range list.Data {
		p, err := ParseCard(item)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (c *Client) get(rawURL string) ([]byte, error) {
	c.mu.Lock()
	if wait := 100*time.Millisecond - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
	c.mu.Unlock()

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "kitchen-mtg/0.1")
	req.Header.Set("Accept", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scryfall %s: %s", res.Status, strings.TrimSpace(string(body)))
	}
	return body, nil
}
