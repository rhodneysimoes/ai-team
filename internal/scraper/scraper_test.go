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

func TestExtractPromotionsStripsScriptsAndStyles(t *testing.T) {
	definition := sites.Definition{
		Name:    "Store A",
		URL:     "https://example.com",
		Pattern: `(?i)oferta.{0,40}`,
		Enabled: true,
	}
	html := `<html>
<head>
	<script>var oferta_internal = "ignored oferta js";</script>
	<style>body { content: "oferta css"; }</style>
</head>
<body>
	<h1>Oferta especial real</h1>
</body>
</html>`

	promotions, err := ExtractPromotions(definition, html, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("ExtractPromotions() error = %v", err)
	}
	if len(promotions) != 1 {
		t.Fatalf("len(promotions) = %d, want 1", len(promotions))
	}
	if promotions[0].Text != "Oferta especial real" {
		t.Fatalf("promotion text = %q, want 'Oferta especial real'", promotions[0].Text)
	}
}

func TestExtractThumbnailPriorityAndCookieExclusion(t *testing.T) {
	definition := sites.Definition{
		Name:    "Store A",
		URL:     "https://example.com",
		Pattern: `(?i)oferta.{0,40}`,
		Enabled: true,
	}
	// Has cookielaw thumbnail (should be ignored), generic thumbnail, and og:image
	html := `<html>
<head>
	<meta property="og:image" content="https://cdn.example.com/real-og.jpg">
</head>
<body>
	<img class="ot-sdk-thumbnail" src="https://cdn.cookielaw.org/logo.png">
	<img class="product-thumbnail" src="https://cdn.example.com/generic-thumb.jpg">
	<h1>Oferta real</h1>
</body>
</html>`

	promotions, err := ExtractPromotions(definition, html, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("ExtractPromotions() error = %v", err)
	}
	if len(promotions) != 1 {
		t.Fatalf("len(promotions) = %d, want 1", len(promotions))
	}
	// og:image should be selected as it has higher priority, and cookie logo should be skipped
	if promotions[0].ThumbnailURL != "https://cdn.example.com/real-og.jpg" {
		t.Fatalf("thumbnail url = %q, want 'https://cdn.example.com/real-og.jpg'", promotions[0].ThumbnailURL)
	}
}

func TestExtractPromotionsFromNextData(t *testing.T) {
	definition := sites.Definition{
		Name:    "Kabum",
		URL:     "https://www.kabum.com.br",
		Pattern: `(?i)(oferta|cupom).{0,40}`,
		Enabled: true,
	}

	html := `<html>
<body>
	<script id="__NEXT_DATA__" type="application/json">
	{
		"props": {
			"pageProps": {
				"banners": {
					"mainBanner": [
						{
							"title": "Cupom de Desconto Especial",
							"banner": "/banners/megamaio.png",
							"link": "/ofertas/megamaio"
						}
					]
				},
				"offers": {
					"products": [
						{
							"name": "Monitor Gamer Boreal 34",
							"thumbnail": "/images/monitor.jpg",
							"link": "/produto/monitor",
							"price": 2000.0,
							"priceWithDiscount": 1800.0,
							"discountPercentage": 10,
							"stamp": {
								"title": "CUPOM MEGA100"
							}
						},
						{
							"name": "Teclado normal",
							"thumbnail": "/images/teclado.jpg",
							"link": "/produto/teclado",
							"price": 100.0
						}
					]
				}
			}
		}
	}
	</script>
</body>
</html>`

	promotions, err := ExtractPromotions(definition, html, time.Unix(0, 0).UTC())
	if err != nil {
		t.Fatalf("ExtractPromotions() error = %v", err)
	}

	// We expect 2 matching promotions: 1 from mainBanner (matches "Cupom"), 1 from product "Monitor Gamer Boreal 34" (matches "CUPOM MEGA100" or the stamp).
	if len(promotions) != 2 {
		t.Fatalf("len(promotions) = %d, want 2", len(promotions))
	}

	// Check Banner Promotion
	if promotions[0].Text != "Cupom de Desconto Especial" {
		t.Errorf("promotions[0].Text = %q, want 'Cupom de Desconto Especial'", promotions[0].Text)
	}
	if promotions[0].ThumbnailURL != "https://www.kabum.com.br/banners/megamaio.png" {
		t.Errorf("promotions[0].ThumbnailURL = %q, want 'https://www.kabum.com.br/banners/megamaio.png'", promotions[0].ThumbnailURL)
	}
	if promotions[0].URL != "https://www.kabum.com.br/ofertas/megamaio" {
		t.Errorf("promotions[0].URL = %q", promotions[0].URL)
	}

	// Check Product Promotion
	expectedProdText := "Monitor Gamer Boreal 34 (Cupom: CUPOM MEGA100) - 10% OFF - De: R$ 2000.00 Por: R$ 1800.00"
	if promotions[1].Text != expectedProdText {
		t.Errorf("promotions[1].Text = %q, want %q", promotions[1].Text, expectedProdText)
	}
	if promotions[1].ThumbnailURL != "https://www.kabum.com.br/images/monitor.jpg" {
		t.Errorf("promotions[1].ThumbnailURL = %q", promotions[1].ThumbnailURL)
	}
	if promotions[1].URL != "https://www.kabum.com.br/produto/monitor" {
		t.Errorf("promotions[1].URL = %q", promotions[1].URL)
	}
	if promotions[1].Price != 1800.00 {
		t.Errorf("promotions[1].Price = %f, want 1800.00", promotions[1].Price)
	}
	if promotions[1].OriginalPrice != 2000.00 {
		t.Errorf("promotions[1].OriginalPrice = %f, want 2000.00", promotions[1].OriginalPrice)
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

func TestExtractPricesFromText(t *testing.T) {
	tests := []struct {
		text          string
		wantPrice     float64
		wantOrigPrice float64
	}{
		{
			text:          "Placa de Vídeo PCyes NVIDIA RTX 2060 - De: R$ 1764.69 Por: R$ 1499.99",
			wantPrice:     1499.99,
			wantOrigPrice: 1764.69,
		},
		{
			text:          "Gabinete Gamer Liketec Cygnus Dark - De: R$ 366.66 Por: R$ 329.99",
			wantPrice:     329.99,
			wantOrigPrice: 366.66,
		},
		{
			text:          "Smart TV LED 50 - de R$2.500,00 por R$ 1.899,90",
			wantPrice:     1899.90,
			wantOrigPrice: 2500.00,
		},
		{
			text:          "Smartphone Motorola Moto G - por R$ 899",
			wantPrice:     899.0,
			wantOrigPrice: 0.0,
		},
		{
			text:          "Fone de Ouvido Bluetooth JBL - R$ 199,90",
			wantPrice:     199.90,
			wantOrigPrice: 0.0,
		},
		{
			text:          "Sem preço definido no texto",
			wantPrice:     0.0,
			wantOrigPrice: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			gotPrice, gotOrigPrice := extractPricesFromText(tt.text)
			if gotPrice != tt.wantPrice {
				t.Errorf("extractPricesFromText() gotPrice = %v, want %v", gotPrice, tt.wantPrice)
			}
			if gotOrigPrice != tt.wantOrigPrice {
				t.Errorf("extractPricesFromText() gotOrigPrice = %v, want %v", gotOrigPrice, tt.wantOrigPrice)
			}
		})
	}
}

func TestCollectLevel2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/" {
			_, _ = w.Write([]byte(`<html><body><a href="/promo-page">Link to Promo</a></body></html>`))
		} else if request.URL.Path == "/promo-page" {
			_, _ = w.Write([]byte(`<html><body><p>Oferta relampago com frete gratis</p></body></html>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	results, err := Collect(context.Background(), server.Client(), []sites.Definition{
		{
			Name:    "Store B",
			URL:     server.URL,
			Pattern: `(?i)oferta.{0,40}`,
			Enabled: true,
		},
	}, Options{Concurrency: 1})

	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if len(results[0].Promotions) != 1 {
		t.Fatalf("len(results[0].Promotions) = %d, want 1, got %d", len(results[0].Promotions), len(results[0].Promotions))
	}
	if results[0].Promotions[0].Text != "Oferta relampago com frete gratis" {
		t.Fatalf("promotion text = %q, want 'Oferta relampago com frete gratis'", results[0].Promotions[0].Text)
	}
	// The URL should point to the promo-page
	if !strings.HasSuffix(results[0].Promotions[0].URL, "/promo-page") {
		t.Fatalf("promotion url = %q, want suffix '/promo-page'", results[0].Promotions[0].URL)
	}
}

func TestIsProductURL(t *testing.T) {
	tests := []struct {
		urlStr string
		want   bool
	}{
		// Kabum
		{"https://www.kabum.com.br/produto/12345/mouse-gamer", true},
		{"https://www.kabum.com.br/hardware/coolers", false},
		{"https://www.kabum.com.br/promocao/maisvendidos", false},
		// Terabyte
		{"https://www.terabyteshop.com.br/produto/54321/teclado-mecanico", true},
		{"https://www.terabyteshop.com.br/categoria/hardware", false},
		// Pichau
		{"https://www.pichau.com.br/suporte-para-monitor-zinnia-tms-90-17-pol-a-32-pol-preto-zno-tms90-bl01", true},
		{"https://www.pichau.com.br/mesa-gamer-brateck-120cm-preto-brateck-gmd15-19", true},
		{"https://www.pichau.com.br/monitores", false},
		{"https://www.pichau.com.br/cadeiras/gamer", false},
		{"https://www.pichau.com.br/search?q=mouse", false},
		// Mocks / Others
		{"https://example.com/some-path", true},
		{"http://127.0.0.1:8080/test", true},
	}

	for _, tt := range tests {
		t.Run(tt.urlStr, func(t *testing.T) {
			got := isProductURL(tt.urlStr)
			if got != tt.want {
				t.Errorf("isProductURL(%q) = %v, want %v", tt.urlStr, got, tt.want)
			}
		})
	}
}

func TestCollectDeduplicatesByURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/" {
			_, _ = w.Write([]byte(`<html><body>
				<a href="/produto/item1">Item 1 First Link</a>
				<a href="/produto/item1">Item 1 Second Link</a>
			</body></html>`))
		} else if request.URL.Path == "/produto/item1" {
			_, _ = w.Write([]byte(`<html><body><p>Oferta do item 1 com frete gratis</p></body></html>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	results, err := Collect(context.Background(), server.Client(), []sites.Definition{
		{
			Name:    "Store C",
			URL:     server.URL,
			Pattern: `(?i)oferta.{0,40}`,
			Enabled: true,
		},
	}, Options{Concurrency: 1})

	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	// Mesmo que o link apareça duas vezes, esperamos exatamente 1 promoção no resultado.
	if len(results[0].Promotions) != 1 {
		t.Fatalf("len(results[0].Promotions) = %d, want 1 (deduplicated by URL), got %d", len(results[0].Promotions), len(results[0].Promotions))
	}
}



