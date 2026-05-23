package scraper

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rhodneysimoes/ai-team/internal/sites"
)

const defaultPattern = `(?i)(promocao|promo|oferta|desconto|sale|off|cupom).{0,160}`

type Options struct {
	Concurrency int
}

type Promotion struct {
	Site         string    `json:"site"`
	URL          string    `json:"url"`
	Text         string    `json:"text"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	MatchedAt    time.Time `json:"matched_at"`
}

type SiteResult struct {
	Site        string      `json:"site"`
	URL         string      `json:"url"`
	StatusCode  int         `json:"status_code,omitempty"`
	Promotions  []Promotion `json:"promotions"`
	Blocked     bool        `json:"blocked,omitempty"`
	BlockReason string      `json:"block_reason,omitempty"`
	Error       string      `json:"error,omitempty"`
}

func Collect(ctx context.Context, client *http.Client, definitions []sites.Definition, options Options) ([]SiteResult, error) {
	if client == nil {
		return nil, fmt.Errorf("http client is required")
	}

	enabled := enabledDefinitions(definitions)
	if len(enabled) == 0 {
		return []SiteResult{}, nil
	}

	concurrency := options.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	if concurrency > len(enabled) {
		concurrency = len(enabled)
	}

	jobs := make(chan sites.Definition)
	results := make(chan SiteResult)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for definition := range jobs {
				results <- collectSite(ctx, client, definition)
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, definition := range enabled {
			select {
			case <-ctx.Done():
				return
			case jobs <- definition:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	collected := make([]SiteResult, 0, len(enabled))
	for result := range results {
		collected = append(collected, result)
	}

	if err := ctx.Err(); err != nil {
		return collected, fmt.Errorf("collect cancelled: %w", err)
	}
	return collected, nil
}

func collectSite(ctx context.Context, client *http.Client, definition sites.Definition) SiteResult {
	result := SiteResult{
		Site: definition.Name,
		URL:  definition.URL,
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, definition.URL, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	applyRequestHeaders(request, definition)

	response, err := client.Do(request)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer response.Body.Close()
	result.StatusCode = response.StatusCode

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 64<<10))
		result.Blocked, result.BlockReason = detectBlockedResponse(response, string(body))
		result.Error = fmt.Sprintf("unexpected status %s", response.Status)
		return result
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		result.Error = err.Error()
		return result
	}

	promotions, err := ExtractPromotions(definition, string(body), time.Now().UTC())
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Promotions = promotions
	return result
}

func ExtractPromotions(definition sites.Definition, html string, matchedAt time.Time) ([]Promotion, error) {
	pattern := strings.TrimSpace(definition.Pattern)
	if pattern == "" {
		pattern = defaultPattern
	}

	expression, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("compile pattern for %q: %w", definition.Name, err)
	}

	text := normalizeText(stripTags(html))
	matches := expression.FindAllString(text, -1)
	promotions := make([]Promotion, 0, len(matches))
	seen := map[string]struct{}{}
	thumbnailURL := extractThumbnailURL(html, definition.URL)

	for _, match := range matches {
		clean := normalizeText(match)
		if clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		promotions = append(promotions, Promotion{
			Site:         definition.Name,
			URL:          definition.URL,
			Text:         clean,
			ThumbnailURL: thumbnailURL,
			MatchedAt:    matchedAt,
		})
	}

	return promotions, nil
}

func enabledDefinitions(definitions []sites.Definition) []sites.Definition {
	enabled := make([]sites.Definition, 0, len(definitions))
	for _, definition := range definitions {
		if definition.Enabled {
			enabled = append(enabled, definition)
		}
	}
	return enabled
}

func applyRequestHeaders(request *http.Request, definition sites.Definition) {
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0 Safari/537.36")
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	request.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Pragma", "no-cache")
	request.Header.Set("Upgrade-Insecure-Requests", "1")
	request.Header.Set("Sec-Fetch-Dest", "document")
	request.Header.Set("Sec-Fetch-Mode", "navigate")
	request.Header.Set("Sec-Fetch-Site", "none")
	request.Header.Set("Sec-Fetch-User", "?1")

	if parsed, err := url.Parse(definition.URL); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		request.Header.Set("Referer", parsed.Scheme+"://"+parsed.Host+"/")
	}

	for name, value := range definition.Headers {
		request.Header.Set(name, value)
	}
}

func detectBlockedResponse(response *http.Response, body string) (bool, string) {
	lowerBody := strings.ToLower(body)
	server := strings.ToLower(response.Header.Get("Server"))

	switch {
	case response.StatusCode == http.StatusForbidden && strings.Contains(server, "cloudflare"):
		return true, "cloudflare returned 403"
	case strings.Contains(lowerBody, "enable javascript and cookies"):
		return true, "javascript or cookie challenge"
	case strings.Contains(lowerBody, "just a moment"):
		return true, "cloudflare challenge page"
	case strings.Contains(lowerBody, "cf-chl"):
		return true, "cloudflare challenge token"
	case response.StatusCode == http.StatusTooManyRequests:
		return true, "rate limited"
	default:
		return false, ""
	}
}

func stripTags(input string) string {
	replacer := strings.NewReplacer(
		"<br>", " ",
		"<br/>", " ",
		"<br />", " ",
		"</p>", " ",
		"</div>", " ",
		"&nbsp;", " ",
		"&amp;", "&",
		"&quot;", `"`,
		"&#39;", "'",
	)
	input = replacer.Replace(input)

	tags := regexp.MustCompile(`<[^>]+>`)
	return tags.ReplaceAllString(input, " ")
}

func normalizeText(input string) string {
	return strings.Join(strings.Fields(input), " ")
}

func extractThumbnailURL(input, pageURL string) string {
	for _, tag := range thumbnailCandidateTags(input) {
		for _, attribute := range []string{"src", "data-src", "data-lazy-src", "content"} {
			value := extractAttribute(tag, attribute)
			if value == "" {
				continue
			}
			return resolveURL(pageURL, html.UnescapeString(value))
		}
	}
	return ""
}

func thumbnailCandidateTags(input string) []string {
	tags := regexp.MustCompile(`(?is)<[^>]+>`).FindAllString(input, -1)
	candidates := make([]string, 0, len(tags))
	for _, tag := range tags {
		lower := strings.ToLower(tag)
		if strings.Contains(lower, "thumbnail") ||
			strings.Contains(lower, "tag descontos") ||
			strings.Contains(lower, `property="og:image"`) ||
			strings.Contains(lower, `property='og:image'`) ||
			strings.Contains(lower, `name="twitter:image"`) ||
			strings.Contains(lower, `name='twitter:image'`) {
			candidates = append(candidates, tag)
		}
	}
	return candidates
}

func extractAttribute(tag, name string) string {
	pattern := fmt.Sprintf(`(?is)\s%s\s*=\s*("([^"]*)"|'([^']*)'|([^\s>]+))`, regexp.QuoteMeta(name))
	matches := regexp.MustCompile(pattern).FindStringSubmatch(tag)
	if len(matches) == 0 {
		return ""
	}
	for _, match := range matches[2:] {
		if match != "" {
			return strings.TrimSpace(match)
		}
	}
	return ""
}

func resolveURL(baseURL, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}
	if parsed.IsAbs() {
		return parsed.String()
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return trimmed
	}
	return base.ResolveReference(parsed).String()
}
