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

// Client calls the Scryfall API with a polite pause between requests.
type Client struct {
	HTTP *http.Client
	mu   sync.Mutex
	last time.Time
}

// Fetch loads one printing by set code and collector number.
func (c *Client) Fetch(set, number string) (Printing, error) {
	body, err := c.get("https://api.scryfall.com/cards/" + url.PathEscape(set) + "/" + url.PathEscape(number))
	if err != nil {
		return Printing{}, err
	}
	return ParseCard(body)
}

// SearchPrintings lists every printing of an exact card name.
func (c *Client) SearchPrintings(name string) ([]Printing, error) {
	q := url.QueryEscape(`!"` + name + `" unique:prints`)
	body, err := c.get("https://api.scryfall.com/cards/search?q=" + q)
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
