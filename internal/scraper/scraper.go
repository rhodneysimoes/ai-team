package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/rhodneysimoes/ai-team/internal/sites"
)

const defaultPattern = `(?i)(promocao|promo|oferta|desconto|sale|off|cupom).{0,160}`

type Options struct {
	Concurrency int
}

type Promotion struct {
	Site          string    `json:"site"`
	URL           string    `json:"url"`
	Text          string    `json:"text"`
	ThumbnailURL  string    `json:"thumbnail_url,omitempty"`
	Price         float64   `json:"price,omitempty"`
	OriginalPrice float64   `json:"original_price,omitempty"`
	MatchedAt     time.Time `json:"matched_at"`
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

	htmlContent, promotions, statusCode, blocked, blockReason, err := fetchAndExtract(ctx, client, definition, definition.URL)
	result.StatusCode = statusCode
	result.Blocked = blocked
	result.BlockReason = blockReason
	if err != nil {
		result.Error = err.Error()
	}
	result.Promotions = promotions

	// If level 1 request was blocked or had an error, we do not proceed to level 2
	if blocked || err != nil {
		return result
	}

	// Extract links from level 1 HTML
	level2URLs := extractLinks(htmlContent, definition.URL)
	if len(level2URLs) == 0 {
		return result
	}

	// Deduplicate promotions using a map of their text
	seen := make(map[string]struct{})
	for _, p := range promotions {
		seen[p.Text] = struct{}{}
	}

	// Worker pool to fetch level 2 URLs concurrently
	type job struct {
		url string
	}
	type taskResult struct {
		promos []Promotion
		err    error
	}

	numWorkers := 5
	if len(level2URLs) < numWorkers {
		numWorkers = len(level2URLs)
	}

	jobsChan := make(chan job, len(level2URLs))
	resultsChan := make(chan taskResult, len(level2URLs))

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobsChan {
				// Create copy of definition with the level 2 URL
				defCopy := definition
				defCopy.URL = j.url

				_, pPromos, _, _, _, pErr := fetchAndExtract(ctx, client, defCopy, j.url)
				resultsChan <- taskResult{promos: pPromos, err: pErr}
			}
		}()
	}

	for _, u := range level2URLs {
		jobsChan <- job{url: u}
	}
	close(jobsChan)

	wg.Wait()
	close(resultsChan)

	// Collect promotions from level 2
	for res := range resultsChan {
		if res.err == nil && len(res.promos) > 0 {
			for _, p := range res.promos {
				if _, ok := seen[p.Text]; !ok {
					seen[p.Text] = struct{}{}
					result.Promotions = append(result.Promotions, p)
				}
			}
		}
	}

	return result
}

func fetchAndExtract(ctx context.Context, client *http.Client, definition sites.Definition, targetURL string) (string, []Promotion, int, bool, string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", nil, 0, false, "", err
	}
	applyRequestHeaders(request, definition)

	response, err := client.Do(request)
	if err != nil {
		// Try browser fallback on connection failure
		htmlContent, browserErr := fetchWithBrowser(ctx, targetURL)
		if browserErr != nil {
			return "", nil, 0, false, "", err
		}
		stillBlocked, stillBlockedReason := detectBlockedResponse(nil, htmlContent)
		if stillBlocked {
			return htmlContent, nil, http.StatusForbidden, true, stillBlockedReason, fmt.Errorf("browser fallback also blocked: %s", stillBlockedReason)
		}
		promotions, extractErr := ExtractPromotions(definition, htmlContent, time.Now().UTC())
		if extractErr != nil {
			return htmlContent, nil, http.StatusOK, false, "", extractErr
		}
		return htmlContent, promotions, http.StatusOK, false, "", nil
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 64<<10))
		blocked, blockReason := detectBlockedResponse(response, string(body))
		if blocked {
			htmlContent, browserErr := fetchWithBrowser(ctx, targetURL)
			if browserErr == nil {
				stillBlocked, stillBlockedReason := detectBlockedResponse(nil, htmlContent)
				if !stillBlocked {
					promotions, extractErr := ExtractPromotions(definition, htmlContent, time.Now().UTC())
					if extractErr == nil {
						return htmlContent, promotions, http.StatusOK, false, "", nil
					}
					return htmlContent, nil, http.StatusOK, false, "", extractErr
				}
				return htmlContent, nil, response.StatusCode, true, "browser fallback also blocked: " + stillBlockedReason, fmt.Errorf("unexpected status %s", response.Status)
			}
		}
		return string(body), nil, response.StatusCode, blocked, blockReason, fmt.Errorf("unexpected status %s", response.Status)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return "", nil, response.StatusCode, false, "", err
	}

	promos, err := ExtractPromotions(definition, string(body), time.Now().UTC())
	if err != nil {
		return string(body), nil, response.StatusCode, false, "", err
	}
	return string(body), promos, response.StatusCode, false, "", nil
}

func isStaticFile(path string) bool {
	path = strings.ToLower(path)
	extensions := []string{".png", ".jpg", ".jpeg", ".gif", ".svg", ".css", ".js", ".pdf", ".zip", ".mp4", ".mp3", ".ico", ".woff", ".woff2", ".ttf"}
	for _, ext := range extensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

func extractLinks(htmlContent, baseURL string) []string {
	var links []string
	re := regexp.MustCompile(`(?i)<a\s+[^>]*href=["']([^"']+)["']`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	baseParsed, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}

	seen := map[string]bool{}
	for _, match := range matches {
		href := strings.TrimSpace(match[1])
		if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
			continue
		}

		resolved := resolveURL(baseURL, href)
		parsedResolved, err := url.Parse(resolved)
		if err != nil {
			continue
		}

		// Only follow http/https
		if parsedResolved.Scheme != "http" && parsedResolved.Scheme != "https" {
			continue
		}

		// Only internal links (same host/domain)
		if parsedResolved.Host != baseParsed.Host {
			continue
		}

		// Skip static files
		if isStaticFile(parsedResolved.Path) {
			continue
		}

		// Normalize URL by removing fragment
		parsedResolved.Fragment = ""
		// Normalize trailing slash to avoid duplicate crawling
		path := parsedResolved.Path
		if len(path) > 1 && strings.HasSuffix(path, "/") {
			parsedResolved.Path = path[:len(path)-1]
		}
		normalized := parsedResolved.String()

		if !seen[normalized] {
			seen[normalized] = true
			links = append(links, normalized)
		}
	}
	return links
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

	// Try to extract from Next.js hydration payload if it exists
	if jsonPromos, ok := tryExtractNextData(definition, html, expression, definition.URL, matchedAt); ok {
		return jsonPromos, nil
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
		price, originalPrice := extractPricesFromText(clean)
		promotions = append(promotions, Promotion{
			Site:          definition.Name,
			URL:           definition.URL,
			Text:          clean,
			ThumbnailURL:  thumbnailURL,
			Price:         price,
			OriginalPrice: originalPrice,
			MatchedAt:     matchedAt,
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
	var server string
	var statusCode int
	if response != nil {
		server = strings.ToLower(response.Header.Get("Server"))
		statusCode = response.StatusCode
	}

	switch {
	case response != nil && statusCode == http.StatusForbidden && strings.Contains(server, "cloudflare"):
		return true, "cloudflare returned 403"
	case strings.Contains(lowerBody, "enable javascript and cookies"):
		return true, "javascript or cookie challenge"
	case strings.Contains(lowerBody, "just a moment"):
		return true, "cloudflare challenge page"
	case strings.Contains(lowerBody, "cf-chl"):
		return true, "cloudflare challenge token"
	case response != nil && statusCode == http.StatusTooManyRequests:
		return true, "rate limited"
	default:
		return false, ""
	}
}

func stripTags(input string) string {
	// Remove script blocks and style blocks including their contents
	scriptRegex := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	input = scriptRegex.ReplaceAllString(input, " ")

	styleRegex := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	input = styleRegex.ReplaceAllString(input, " ")

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
	tags := regexp.MustCompile(`(?is)<[^>]+>`).FindAllString(input, -1)

	var ogImages []string
	var twitterImages []string
	var tagDescontos []string
	var fallbackThumbnails []string

	for _, tag := range tags {
		lower := strings.ToLower(tag)
		// Exclude cookie consent and other unwanted tags from thumbnail processing
		if strings.Contains(lower, "cookielaw") || strings.Contains(lower, "onetrust") || strings.Contains(lower, "ot-sdk") || strings.Contains(lower, "cookie-") {
			continue
		}

		if strings.Contains(lower, `property="og:image"`) || strings.Contains(lower, `property='og:image'`) {
			ogImages = append(ogImages, tag)
		} else if strings.Contains(lower, `name="twitter:image"`) || strings.Contains(lower, `name='twitter:image'`) {
			twitterImages = append(twitterImages, tag)
		} else if strings.Contains(lower, "tag descontos") {
			tagDescontos = append(tagDescontos, tag)
		} else if strings.Contains(lower, "thumbnail") {
			fallbackThumbnails = append(fallbackThumbnails, tag)
		}
	}

	priorityGroups := [][]string{ogImages, twitterImages, tagDescontos, fallbackThumbnails}
	for _, group := range priorityGroups {
		for _, tag := range group {
			for _, attribute := range []string{"content", "src", "data-src", "data-lazy-src"} {
				value := extractAttribute(tag, attribute)
				if value == "" {
					continue
				}
				resolved := resolveURL(pageURL, html.UnescapeString(value))
				if resolved == "" {
					continue
				}
				lowerResolved := strings.ToLower(resolved)
				if strings.Contains(lowerResolved, "cookielaw") || strings.Contains(lowerResolved, "onetrust") || strings.Contains(lowerResolved, "cookie") {
					continue
				}
				return resolved
			}
		}
	}
	return ""
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

type nextDataPayload struct {
	Props struct {
		PageProps struct {
			Banners map[string]json.RawMessage `json:"banners"`
			Offers  struct {
				Products []struct {
					Name               string  `json:"name"`
					Thumbnail          string  `json:"thumbnail"`
					Link               string  `json:"link"`
					DiscountPercentage float64 `json:"discountPercentage"`
					Price              float64 `json:"price"`
					PriceWithDiscount  float64 `json:"priceWithDiscount"`
					Stamp              *struct {
						Title string `json:"title"`
						Type  string `json:"type"`
					} `json:"stamp"`
				} `json:"products"`
			} `json:"offers"`
		} `json:"pageProps"`
	} `json:"props"`
}

func tryExtractNextData(definition sites.Definition, htmlContent string, expression *regexp.Regexp, pageURL string, matchedAt time.Time) ([]Promotion, bool) {
	re := regexp.MustCompile(`(?s)<script id="__NEXT_DATA__"[^>]*>(.*?)</script>`)
	match := re.FindStringSubmatch(htmlContent)
	if len(match) < 2 {
		return nil, false
	}

	var data nextDataPayload
	if err := json.Unmarshal([]byte(match[1]), &data); err != nil {
		return nil, false
	}

	promotions := []Promotion{}
	seen := map[string]struct{}{}

	// Extract from Banners
	for key, rawList := range data.Props.PageProps.Banners {
		var list []struct {
			Title        string `json:"title"`
			Banner       string `json:"banner"`
			BannerMobile string `json:"bannerMobile"`
			Link         string `json:"link"`
			SubTitle     string `json:"subTitle"`
		}
		if err := json.Unmarshal(rawList, &list); err != nil {
			continue
		}

		for _, item := range list {
			matchText := item.Title
			if item.SubTitle != "" {
				matchText = fmt.Sprintf("%s - %s", item.Title, item.SubTitle)
			}
			matchText = fmt.Sprintf("Banner %s: %s", key, matchText)

			if expression.MatchString(matchText) || expression.MatchString(item.Link) {
				bannerURL := item.Banner
				if bannerURL == "" {
					bannerURL = item.BannerMobile
				}
				resolvedLink := resolveURL(pageURL, item.Link)
				text := item.Title
				if item.SubTitle != "" {
					text = fmt.Sprintf("%s - %s", item.Title, item.SubTitle)
				}
				text = normalizeText(text)
				if text == "" {
					text = fmt.Sprintf("Banner %s: %s", key, item.Title)
				}
				if _, ok := seen[text]; ok {
					continue
				}
				seen[text] = struct{}{}

				promotions = append(promotions, Promotion{
					Site:         definition.Name,
					URL:          resolvedLink,
					Text:         text,
					ThumbnailURL: resolveURL(pageURL, bannerURL),
					MatchedAt:    matchedAt,
				})
			}
		}
	}

	// Extract from Products/Offers
	for _, prod := range data.Props.PageProps.Offers.Products {
		stampTitle := ""
		if prod.Stamp != nil {
			stampTitle = prod.Stamp.Title
		}

		text := prod.Name
		if stampTitle != "" {
			text += fmt.Sprintf(" (Cupom: %s)", stampTitle)
		}
		if prod.DiscountPercentage > 0 {
			text += fmt.Sprintf(" - %.0f%% OFF", prod.DiscountPercentage)
		}
		if prod.PriceWithDiscount > 0 {
			text += fmt.Sprintf(" - De: R$ %.2f Por: R$ %.2f", prod.Price, prod.PriceWithDiscount)
		} else if prod.Price > 0 {
			text += fmt.Sprintf(" - R$ %.2f", prod.Price)
		}

		if expression.MatchString(text) || expression.MatchString(prod.Link) || (prod.Stamp != nil && expression.MatchString(prod.Stamp.Title)) {
			if _, ok := seen[text]; ok {
				continue
			}
			seen[text] = struct{}{}

			var price float64
			var originalPrice float64
			if prod.PriceWithDiscount > 0 {
				price = prod.PriceWithDiscount
				originalPrice = prod.Price
			} else if prod.Price > 0 {
				price = prod.Price
			}

			promotions = append(promotions, Promotion{
				Site:          definition.Name,
				URL:           resolveURL(pageURL, prod.Link),
				Text:          normalizeText(text),
				ThumbnailURL:  resolveURL(pageURL, prod.Thumbnail),
				Price:         price,
				OriginalPrice: originalPrice,
				MatchedAt:     matchedAt,
			})
		}
	}

	return promotions, len(promotions) > 0
}

func parsePrice(s string) float64 {
	s = strings.TrimSpace(s)
	// Remove non-numeric/non-punctuation characters except dots and commas
	var sb strings.Builder
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '.' || r == ',' {
			sb.WriteRune(r)
		}
	}
	cleaned := sb.String()
	if cleaned == "" {
		return 0
	}

	// Determine decimal separator
	lastDot := strings.LastIndex(cleaned, ".")
	lastComma := strings.LastIndex(cleaned, ",")

	if lastComma > lastDot {
		// Comma is the decimal separator (Brazilian format: e.g. 1.764,69 or 1764,69)
		// Remove all dots (thousands separator), then replace comma with dot
		cleaned = strings.ReplaceAll(cleaned, ".", "")
		cleaned = strings.ReplaceAll(cleaned, ",", ".")
	} else if lastDot > lastComma {
		// Dot is the decimal separator (English format: e.g. 1,764.69 or 1764.69)
		// Remove all commas (thousands separator)
		cleaned = strings.ReplaceAll(cleaned, ",", "")
	} else {
		// No separators or only one type which is at the end.
		// If there's only comma, replace with dot
		cleaned = strings.ReplaceAll(cleaned, ",", ".")
	}

	var val float64
	fmt.Sscanf(cleaned, "%f", &val)
	return val
}

func extractPricesFromText(text string) (price float64, originalPrice float64) {
	// Try to find "de: ... por: ..." patterns (with optional colons, optional spaces, optional R$)
	dePorRegex := regexp.MustCompile(`(?i)\bde:?\s*(?:r\$\s*)?([0-9]+(?:[.,][0-9]+)*)\s+por:?\s*(?:r\$\s*)?([0-9]+(?:[.,][0-9]+)*)`)
	matches := dePorRegex.FindStringSubmatch(text)
	if len(matches) >= 3 {
		originalPrice = parsePrice(matches[1])
		price = parsePrice(matches[2])
		return price, originalPrice
	}

	// Try to find any price preceded by R$ or "por R$" or "por "
	priceRegex := regexp.MustCompile(`(?i)(?:r\$\s*|por\s+r\$\s*|por\s+)([0-9]+(?:[.,][0-9]+)*)`)
	priceMatches := priceRegex.FindAllStringSubmatch(text, -1)
	if len(priceMatches) > 0 {
		price = parsePrice(priceMatches[0][1])
		return price, 0
	}

	return 0, 0
}

func fetchWithBrowser(ctx context.Context, urlStr string) (string, error) {
	// Create context with a timeout so it doesn't hang forever
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Configure chromedp options to run headlessly and bypass sandboxing
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("blink-settings", "imagesEnabled=false"), // Save bandwidth/speed up
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	chromeCtx, chromeCancel := chromedp.NewContext(allocCtx)
	defer chromeCancel()

	var htmlContent string
	err := chromedp.Run(chromeCtx,
		chromedp.Navigate(urlStr),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(5*time.Second), // Sleep to allow JS execution/Cloudflare solve
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		return "", err
	}

	return htmlContent, nil
}
