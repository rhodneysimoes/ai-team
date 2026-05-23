package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rhodneysimoes/ai-team/internal/scraper"
)

func TestCleanProductText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Kabum normal",
			input:    "Placa de Vídeo PCyes NVIDIA RTX 2060 (Cupom: CUPOM MEGAMAIO) - 9% OFF - De: R$ 1764.69 Por: R$ 1499.99",
			expected: "Placa de Vídeo PCyes NVIDIA RTX 2060",
		},
		{
			name:     "Terabyte normal",
			input:    "CUPOM 2º Mais vendido Memória DDR4 Kingston Fury Beast, 8GB (154) (64 cupons restantes) De: R$ 959,90 por: R$ 589,90 à vista",
			expected: "Memória DDR4 Kingston Fury Beast, 8GB",
		},
		{
			name:     "Pichau normal",
			input:    "desconto no PIX R$ 305,87 Em até 12x de R$ 25,49 Sem juros no cartão 28% OFF 6 UNID Poltrona Eletrica Reclinavel Pichau Oblivion, Suede",
			expected: "Poltrona Eletrica Reclinavel Pichau Oblivion, Suede",
		},
		{
			name:     "Pichau de/por suffix",
			input:    "desconto no PIX R$ 270,58 Em até 12x de R$ 22,55 Sem juros no cartão 7% OFF 49 UNID Raquete de Beach Tennis Pichau Venice Beach 18K Carbon 28 Furos, PES-RKT18-VNB de R$ 399,99 por R$ 270,58",
			expected: "Raquete de Beach Tennis Pichau Venice Beach 18K Carbon 28 Furos, PES-RKT18-VNB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanProductText(tt.input)
			if got != tt.expected {
				t.Errorf("CleanProductText() got = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGenerateMatchKey(t *testing.T) {
	tests := []struct {
		name     string
		input1   string
		input2   string
		expected bool
	}{
		{
			name:     "Case and spacing",
			input1:   "Memoria DDR4 Kingston 8GB",
			input2:   "memoria   ddr4  kingston  8gb",
			expected: true,
		},
		{
			name:     "Word order",
			input1:   "Memoria DDR4 Kingston 8GB",
			input2:   "Kingston Memoria 8GB DDR4",
			expected: true,
		},
		{
			name:     "Special characters",
			input1:   "Placa-Mãe Gigabyte B550M",
			input2:   "placa mãe gigabyte b550m",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1 := GenerateMatchKey(tt.input1)
			key2 := GenerateMatchKey(tt.input2)
			got := key1 == key2
			if got != tt.expected {
				t.Errorf("GenerateMatchKey match got = %v, want %v (key1=%q, key2=%q)", got, tt.expected, key1, key2)
			}
		})
	}
}

func TestRun(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "input.json")
	minOutPath := filepath.Join(tempDir, "menor.json")
	maxOutPath := filepath.Join(tempDir, "maior.json")

	// Create mock input promotions data
	mockData := []scraper.SiteResult{
		{
			Site:       "Store A",
			StatusCode: 200,
			Promotions: []scraper.Promotion{
				{
					Site:      "Store A",
					URL:       "https://storea.com/prod1",
					Text:      "Teclado Gamer Mecanico RGB De: R$ 200 Por: R$ 150",
					Price:     150,
					MatchedAt: time.Now().UTC(),
				},
				{
					Site:      "Store A",
					URL:       "https://storea.com/prod2",
					Text:      "Mouse Gamer RGB De: R$ 100 Por: R$ 80",
					Price:     80,
					MatchedAt: time.Now().UTC(),
				},
				{
					Site:      "Store A",
					URL:       "https://storea.com/prod3",
					Text:      "Monitor UltraWide De: R$ 3000 Por: R$ 2500", // Unique product
					Price:     2500,
					MatchedAt: time.Now().UTC(),
				},
			},
		},
		{
			Site:       "Store B",
			StatusCode: 200,
			Promotions: []scraper.Promotion{
				{
					Site:      "Store B",
					URL:       "https://storeb.com/prod1",
					Text:      "Teclado Mecanico RGB Gamer De: R$ 220 Por: R$ 130", // Same keyboard, cheaper
					Price:     130,
					MatchedAt: time.Now().UTC(),
				},
				{
					Site:      "Store B",
					URL:       "https://storeb.com/prod2",
					Text:      "Mouse RGB Gamer De: R$ 120 Por: R$ 90", // Same mouse, more expensive
					Price:     90,
					MatchedAt: time.Now().UTC(),
				},
			},
		},
	}

	bytes, err := json.Marshal(mockData)
	if err != nil {
		t.Fatalf("marshal mock data: %v", err)
	}
	if err := os.WriteFile(inputPath, bytes, 0o600); err != nil {
		t.Fatalf("write mock input: %v", err)
	}

	// Run comparator
	if err := run(inputPath, minOutPath, maxOutPath); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Verify menor.json
	minBytes, err := os.ReadFile(minOutPath)
	if err != nil {
		t.Fatalf("read menor.json: %v", err)
	}
	var minPromos []scraper.Promotion
	if err := json.Unmarshal(minBytes, &minPromos); err != nil {
		t.Fatalf("unmarshal menor.json: %v", err)
	}
	if len(minPromos) != 3 {
		t.Fatalf("len(minPromos) = %d, want 3", len(minPromos))
	}
	// Sorted by canonical product name:
	// "gamer mecanico rgb teclado" (keyboard) -> Min is 130 (Store B)
	// "gamer mouse rgb" (mouse) -> Min is 80 (Store A)
	// "monitor ultrawide" (monitor, unique) -> Min is 2500 (Store A)
	if minPromos[0].Price != 130 {
		t.Errorf("expected keyboard min price to be 130, got %f", minPromos[0].Price)
	}
	if minPromos[0].Site != "Store B" {
		t.Errorf("expected keyboard min site to be Store B, got %s", minPromos[0].Site)
	}
	if minPromos[1].Price != 80 {
		t.Errorf("expected mouse min price to be 80, got %f", minPromos[1].Price)
	}
	if minPromos[1].Site != "Store A" {
		t.Errorf("expected mouse min site to be Store A, got %s", minPromos[1].Site)
	}
	if minPromos[2].Price != 2500 {
		t.Errorf("expected monitor min price to be 2500, got %f", minPromos[2].Price)
	}
	if minPromos[2].Site != "Store A" {
		t.Errorf("expected monitor min site to be Store A, got %s", minPromos[2].Site)
	}

	// Verify maior.json
	maxBytes, err := os.ReadFile(maxOutPath)
	if err != nil {
		t.Fatalf("read maior.json: %v", err)
	}
	var maxPromos []scraper.Promotion
	if err := json.Unmarshal(maxBytes, &maxPromos); err != nil {
		t.Fatalf("unmarshal maior.json: %v", err)
	}
	// Should only have 2 items (monitor is unique, so excluded)
	if len(maxPromos) != 2 {
		t.Fatalf("len(maxPromos) = %d, want 2", len(maxPromos))
	}
	// Max prices:
	// Keyboard -> Max is 150 (Store A)
	// Mouse -> Max is 90 (Store B)
	if maxPromos[0].Price != 150 {
		t.Errorf("expected keyboard max price to be 150, got %f", maxPromos[0].Price)
	}
	if maxPromos[0].Site != "Store A" {
		t.Errorf("expected keyboard max site to be Store A, got %s", maxPromos[0].Site)
	}
	if maxPromos[1].Price != 90 {
		t.Errorf("expected mouse max price to be 90, got %f", maxPromos[1].Price)
	}
	if maxPromos[1].Site != "Store B" {
		t.Errorf("expected mouse max site to be Store B, got %s", maxPromos[1].Site)
	}
}
