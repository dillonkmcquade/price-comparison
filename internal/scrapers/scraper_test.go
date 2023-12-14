package scrapers

import (
	"net/url"
	"testing"
)

func TestNewScraper(t *testing.T) {
	scr := NewIgaScraper(nil, nil, "")
	if scr == nil {
		t.Error("Scraper should not be nil")
	}
	scr2 := NewMetroScraper(nil, nil, "")
	if scr2 == nil {
		t.Error("Scraper should not be nil")
	}
}

func TestSetQuery(t *testing.T) {
	url, err := url.Parse("http://localhost:3001/api/products?search=carrots")
	if err != nil {
		t.Error("failed to parse url")
	}

	SetQuery(url, "page", "0")
	query := url.Query()

	if p := query.Get("page"); p == "" {
		t.Error("failed to set query")
	}
}

func TestStrToFloat(t *testing.T) {
	str := "$59.99"

	flt, err := strToFloat(str)
	if err != nil {
		t.Error("error parsing float")
	}

	if flt != 59.99 {
		t.Error("Error running strToFloat")
	}
}

/* func TestVisit(t *testing.T) {
	l := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	db := database.NewDatabase(":memory:")
	scr := NewIgaScraper(l, db, "")
	err := scr.Visit("carrots")
	if err != nil {
		t.Error("error visiting:", err)
	}
	scr.Wait()
} */
