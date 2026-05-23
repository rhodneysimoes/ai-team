package sites

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMarkdown(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sites.md")
	content := []byte(`# Sites

` + "```text" + `
| name | url | pattern | enabled |
| --- | --- | --- | --- |
| Example | https://example.com/promos | (?i)oferta | true |
` + "```" + `

| name | url | pattern | enabled | headers |
| --- | --- | --- | --- | --- |
| Store A | https://example.com/promos | (?i)(oferta\|desconto).{0,20} | true | Accept-Language=pt-BR,pt\;q=0.9;X-Test=yes |
| Store B | https://example.org |  | false | |
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	definitions, err := LoadMarkdown(path)
	if err != nil {
		t.Fatalf("LoadMarkdown() error = %v", err)
	}

	if len(definitions) != 2 {
		t.Fatalf("len(definitions) = %d, want 2", len(definitions))
	}
	if definitions[0].Name != "Store A" {
		t.Fatalf("definitions[0].Name = %q, want Store A", definitions[0].Name)
	}
	if definitions[0].Pattern != `(?i)(oferta|desconto).{0,20}` {
		t.Fatalf("definitions[0].Pattern = %q", definitions[0].Pattern)
	}
	if definitions[0].Headers["Accept-Language"] != "pt-BR,pt;q=0.9" {
		t.Fatalf("Accept-Language header = %q", definitions[0].Headers["Accept-Language"])
	}
	if definitions[0].Headers["X-Test"] != "yes" {
		t.Fatalf("X-Test header = %q", definitions[0].Headers["X-Test"])
	}
	if !definitions[0].Enabled {
		t.Fatalf("definitions[0].Enabled = false, want true")
	}
	if definitions[1].Enabled {
		t.Fatalf("definitions[1].Enabled = true, want false")
	}
}

func TestLoadMarkdownRejectsInvalidURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sites.md")
	content := []byte(`| name | url | pattern | enabled |
| --- | --- | --- | --- |
| Store A | not-a-url | oferta | true |
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadMarkdown(path); err == nil {
		t.Fatal("LoadMarkdown() error = nil, want error")
	}
}
