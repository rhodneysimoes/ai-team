package sites

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Definition struct {
	Name    string            `json:"name"`
	URL     string            `json:"url"`
	Pattern string            `json:"pattern,omitempty"`
	Enabled bool              `json:"enabled"`
	Headers map[string]string `json:"headers,omitempty"`
}

// LoadMarkdown lê o arquivo markdown contendo a tabela de especificações dos e-commerces,
// extrai as linhas da tabela e as converte em definições de site utilizáveis pelo scraper.
func LoadMarkdown(path string) ([]Definition, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer file.Close()

	var definitions []Definition
	inCodeBlock := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}
		if !isTableDataLine(line) {
			continue
		}

		columns := splitMarkdownRow(line)
		if len(columns) < 4 || isHeaderRow(columns) || isSeparatorRow(columns) {
			continue
		}

		definition, err := parseDefinition(columns)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, definition)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %q: %w", path, err)
	}

	return definitions, nil
}

// parseDefinition converte as colunas extraídas de uma linha da tabela markdown
// em uma struct Definition de site, validando a obrigatoriedade do nome e a URL informada.
func parseDefinition(columns []string) (Definition, error) {
	definition := Definition{
		Name:    strings.TrimSpace(columns[0]),
		URL:     strings.TrimSpace(columns[1]),
		Pattern: strings.TrimSpace(columns[2]),
		Enabled: parseBool(columns[3]),
		Headers: map[string]string{},
	}
	if len(columns) >= 5 {
		definition.Headers = parseHeaders(columns[4])
	}

	if definition.Name == "" {
		return Definition{}, fmt.Errorf("site name is required")
	}
	if definition.URL == "" {
		return Definition{}, fmt.Errorf("site %q url is required", definition.Name)
	}
	parsed, err := url.ParseRequestURI(definition.URL)
	if err != nil {
		return Definition{}, fmt.Errorf("site %q has invalid url %q: %w", definition.Name, definition.URL, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return Definition{}, fmt.Errorf("site %q has invalid url %q: missing scheme or host", definition.Name, definition.URL)
	}

	return definition, nil
}

func isTableDataLine(line string) bool {
	return strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|")
}

func splitMarkdownRow(line string) []string {
	line = strings.Trim(line, "|")
	parts := make([]string, 0, 4)
	var current strings.Builder
	escaped := false

	for _, char := range line {
		switch {
		case escaped:
			if char != '|' {
				current.WriteRune('\\')
			}
			current.WriteRune(char)
			escaped = false
		case char == '\\':
			escaped = true
		case char == '|':
			parts = append(parts, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(char)
		}
	}
	if escaped {
		current.WriteRune('\\')
	}
	parts = append(parts, strings.TrimSpace(current.String()))
	return parts
}

func isHeaderRow(columns []string) bool {
	return strings.EqualFold(columns[0], "name")
}

func isSeparatorRow(columns []string) bool {
	for _, column := range columns {
		trimmed := strings.TrimSpace(column)
		if trimmed == "" {
			return false
		}
		for _, char := range trimmed {
			if char != '-' && char != ':' {
				return false
			}
		}
	}
	return true
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "yes", "y", "1", "sim":
		return true
	default:
		return false
	}
}

func parseHeaders(value string) map[string]string {
	headers := map[string]string{}
	for _, entry := range splitEscaped(value, ';') {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		name, headerValue, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		headerValue = strings.TrimSpace(headerValue)
		if name == "" || headerValue == "" {
			continue
		}
		headers[name] = headerValue
	}
	return headers
}

func splitEscaped(value string, separator rune) []string {
	parts := []string{}
	var current strings.Builder
	escaped := false

	for _, char := range value {
		switch {
		case escaped:
			if char != separator {
				current.WriteRune('\\')
			}
			current.WriteRune(char)
			escaped = false
		case char == '\\':
			escaped = true
		case char == separator:
			parts = append(parts, current.String())
			current.Reset()
		default:
			current.WriteRune(char)
		}
	}
	if escaped {
		current.WriteRune('\\')
	}
	parts = append(parts, current.String())
	return parts
}
