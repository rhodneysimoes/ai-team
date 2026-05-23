package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/rhodneysimoes/ai-team/internal/scraper"
)

func main() {
	var (
		inputPath  = flag.String("input", "promotions.json", "path to input promotions JSON file")
		minOutPath = flag.String("min-output", "menor_preco.json", "path to output lowest price JSON file")
		maxOutPath = flag.String("max-output", "maior_preco.json", "path to output highest price JSON file")
	)
	flag.Parse()

	if err := run(*inputPath, *minOutPath, *maxOutPath); err != nil {
		log.Fatal(err)
	}
}

func run(inputPath, minOutPath, maxOutPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read input file %q: %w", inputPath, err)
	}

	var results []scraper.SiteResult
	if err := json.Unmarshal(data, &results); err != nil {
		return fmt.Errorf("decode input file: %w", err)
	}

	// Group promotions by canonical product key
	groups := make(map[string][]scraper.Promotion)
	for _, res := range results {
		for _, promo := range res.Promotions {
			key := GenerateMatchKey(promo.Text)
			if key == "" {
				continue
			}
			groups[key] = append(groups[key], promo)
		}
	}

	var minPromos []scraper.Promotion
	var maxPromos []scraper.Promotion

	// Collect keys and sort them to make output deterministic
	var keys []string
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		promos := groups[key]
		if len(promos) == 0 {
			continue
		}

		// Find min and max price promotions
		minPromo := promos[0]
		maxPromo := promos[0]

		for _, p := range promos[1:] {
			if p.Price < minPromo.Price {
				minPromo = p
			}
			if p.Price > maxPromo.Price {
				maxPromo = p
			}
		}

		minPromos = append(minPromos, minPromo)
		if maxPromo.Price > minPromo.Price {
			maxPromos = append(maxPromos, maxPromo)
		}
	}

	// Write menor_preco.json
	minData, err := json.MarshalIndent(minPromos, "", "  ")
	if err != nil {
		return fmt.Errorf("encode lowest price results: %w", err)
	}
	minData = append(minData, '\n')
	if err := os.WriteFile(minOutPath, minData, 0o600); err != nil {
		return fmt.Errorf("write lowest price file %q: %w", minOutPath, err)
	}

	// Write maior_preco.json
	maxData, err := json.MarshalIndent(maxPromos, "", "  ")
	if err != nil {
		return fmt.Errorf("encode highest price results: %w", err)
	}
	maxData = append(maxData, '\n')
	if err := os.WriteFile(maxOutPath, maxData, 0o600); err != nil {
		return fmt.Errorf("write highest price file %q: %w", maxOutPath, err)
	}

	return nil
}

func CleanProductText(text string) string {
	text = strings.TrimSpace(text)

	// Remove prefixos conhecidos da Pichau
	pichauPrefixRegex := regexp.MustCompile(`(?i)^(?:desconto no PIX|Ofertas em Destaque|Na Pichau Tem!|Links Relacionados)\s*(?:R\$\s*[0-9.,\s]+)?\s*(?:Em até\s*[0-9]+\s*x\s*de\s*R\$\s*[0-9.,\s]+\s*(?:Sem juros)?\s*(?:no cartão)?)?\s*(?:[0-9]+%\s*OFF)?\s*(?:[0-9]+\s*UNID)?\s*(?:Frete Grátis:\s*[^\s]+(?:\s+e\s+[^\s]+)?)?\s*(?:15%\s*Desconto\s+em\s+Até\s+[0-9]+x)?\s*(?:Montado\s+e\s+Certificado)?\s*`)
	text = pichauPrefixRegex.ReplaceAllString(text, "")

	// Remove prefixos conhecidos da Terabyte
	tbPrefixRegex := regexp.MustCompile(`(?i)^(?:CUPOM\s+)?(?:[0-9]+º\s+)?(?:Mais vendido|Lançamento|Frete grátis)?\s*`)
	text = tbPrefixRegex.ReplaceAllString(text, "")

	// Remove prefixos conhecidos da Kabum
	kbPrefixRegex := regexp.MustCompile(`(?i)^(?:OFERTA LENDÁRIA|CUPOM\s+[A-Z0-9_]+|Frete grátis|OpenBox|Prime Ninja)\s*`)
	text = kbPrefixRegex.ReplaceAllString(text, "")

	// Remove sufixos comuns de De: ... Por: ...
	dePorSuffixRegex := regexp.MustCompile(`(?i)\b(?:de|por):?\s*(?:r\$\s*)?[0-9.,\s]+(?:\s+(?:por|de):?\s*(?:r\$\s*)?[0-9.,\s]+).*$`)
	text = dePorSuffixRegex.ReplaceAllString(text, "")

	priceSuffixRegex := regexp.MustCompile(`(?i)\b(?:r\$\s*|por\s+r\$\s*|por\s+)[0-9.,\s]+.*$`)
	text = priceSuffixRegex.ReplaceAllString(text, "")

	// Remove sufixo Cupom
	cupomSuffixRegex := regexp.MustCompile(`(?i)\(Cupom:.*?\).*$`)
	text = cupomSuffixRegex.ReplaceAllString(text, "")

	// Remove sufixo generic OFF/discount
	offSuffixRegex := regexp.MustCompile(`(?i)\b[0-9]+%\s*OFF.*$`)
	text = offSuffixRegex.ReplaceAllString(text, "")

	// Remove colchetes ou parênteses específicos do Terabyte
	tbBracketsRegex := regexp.MustCompile(`\([0-9]+\)\s*\([0-9]+\s*cupons restantes\).*$`)
	text = tbBracketsRegex.ReplaceAllString(text, "")

	// Remove espaços extras
	text = strings.Join(strings.Fields(text), " ")
	return text
}

func GenerateMatchKey(text string) string {
	cleaned := CleanProductText(text)
	cleaned = strings.ToLower(cleaned)

	var sb strings.Builder
	for _, r := range cleaned {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		} else {
			sb.WriteRune(' ')
		}
	}

	words := strings.Fields(sb.String())
	sort.Strings(words)
	return strings.Join(words, " ")
}
