package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rhodneysimoes/ai-team/internal/sites"
)

func TestExtractPromotions(t *testing.T) {
	definition := sites.Definition{
		Name:    "Store A",
		URL:     "https://example.com",
		Pattern: `(?i)oferta.{0,40}`,
		Enabled: true,
	}
	html := `<html><body><img class="product-thumbnail" src="/img/notebook.jpg"><h1>Oferta especial: notebook com 20% OFF</h1></body></html>`

	promotions, err := ExtractPromotions(definition, html, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("ExtractPromotions() error = %v", err)
	}
	if len(promotions) != 1 {
		t.Fatalf("len(promotions) = %d, want 1", len(promotions))
	}
	if promotions[0].Text != "Oferta especial: notebook com 20% OFF" {
		t.Fatalf("promotion text = %q", promotions[0].Text)
	}
	if promotions[0].ThumbnailURL != "https://example.com/img/notebook.jpg" {
		t.Fatalf("thumbnail url = %q", promotions[0].ThumbnailURL)
	}
}

func TestExtractPromotionsUsesOpenGraphImageAsThumbnail(t *testing.T) {
	definition := sites.Definition{
		Name:    "Store A",
		URL:     "https://example.com/promos",
		Pattern: `(?i)oferta.{0,40}`,
		Enabled: true,
	}
	html := `<html><head><meta property="og:image" content="https://cdn.example.com/thumb.jpg"></head><body>Oferta relampago</body></html>`

	promotions, err := ExtractPromotions(definition, html, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("ExtractPromotions() error = %v", err)
	}
	if len(promotions) != 1 {
		t.Fatalf("len(promotions) = %d, want 1", len(promotions))
	}
	if promotions[0].ThumbnailURL != "https://cdn.example.com/thumb.jpg" {
		t.Fatalf("thumbnail url = %q", promotions[0].ThumbnailURL)
	}
}

func TestExtractPromotionsUsesTagDescontosImageAsThumbnail(t *testing.T) {
	definition := sites.Definition{
		Name:    "Terabyte Shop",
		URL:     "https://www.terabyteshop.com.br",
		Pattern: `(?i)[0-9]{1,2}%\s*OFF`,
		Enabled: true,
	}
	html := `<html><body><img alt="Tag descontos" src="/descontos.png"><p>37% OFF</p></body></html>`

	promotions, err := ExtractPromotions(definition, html, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("ExtractPromotions() error = %v", err)
	}
	if len(promotions) != 1 {
		t.Fatalf("len(promotions) = %d, want 1", len(promotions))
	}
	if promotions[0].ThumbnailURL != "https://www.terabyteshop.com.br/descontos.png" {
		t.Fatalf("thumbnail url = %q", promotions[0].ThumbnailURL)
	}
}

func TestCollect(t *testing.T) {
	var userAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		userAgent = request.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`<p>Oferta relampago com frete gratis</p>`))
	}))
	defer server.Close()

	results, err := Collect(context.Background(), server.Client(), []sites.Definition{
		{
			Name:    "Store A",
			URL:     server.URL,
			Pattern: `(?i)oferta.{0,40}`,
			Enabled: true,
		},
		{
			Name:    "Disabled",
			URL:     server.URL,
			Enabled: false,
		},
	}, Options{Concurrency: 2})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if len(results[0].Promotions) != 1 {
		t.Fatalf("len(results[0].Promotions) = %d, want 1", len(results[0].Promotions))
	}
	if !strings.Contains(userAgent, "Mozilla/5.0") {
		t.Fatalf("User-Agent = %q, want browser-like value", userAgent)
	}
}

func TestCollectReportsBlockedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Server", "cloudflare")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Just a moment... Enable JavaScript and cookies to continue"))
	}))
	defer server.Close()

	results, err := Collect(context.Background(), server.Client(), []sites.Definition{
		{
			Name:    "Blocked Store",
			URL:     server.URL,
			Enabled: true,
		},
	}, Options{Concurrency: 1})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if !results[0].Blocked {
		t.Fatal("results[0].Blocked = false, want true")
	}
	if results[0].StatusCode != http.StatusForbidden {
		t.Fatalf("StatusCode = %d, want %d", results[0].StatusCode, http.StatusForbidden)
	}
}
