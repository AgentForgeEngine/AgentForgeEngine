package main

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type ExtractedContent struct {
	URL         string            `json:"url"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	MainContent string            `json:"main_content"`
	Headings    []Heading         `json:"headings"`
	Links       []Link            `json:"links,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	TokenCount  int               `json:"token_count"`
	Truncated   bool              `json:"truncated"`
	WordCount   int               `json:"word_count"`
}

type Heading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
}

type Link struct {
	Text string `json:"text"`
	URL  string `json:"url"`
	Type string `json:"type"` // "internal", "external", "special"
}

func (wa *WebAgent) extractAndOptimizeContent(htmlContent, urlStr string, maxTokens int) map[string]interface{} {
	content := &ExtractedContent{
		URL:   urlStr,
		Links: []Link{},
		Metadata: map[string]string{
			"extraction_method": "html_parser",
			"agent_name":        wa.name,
		},
	}

	// Extract title
	content.Title = wa.extractTitle(htmlContent)

	// Extract meta description
	content.Description = wa.extractMetaDescription(htmlContent)

	// Extract headings
	content.Headings = wa.extractHeadings(htmlContent)

	// Extract main content (simplified approach)
	content.MainContent = wa.extractMainContent(htmlContent)

	// Extract links if enabled
	if wa.includeLinks {
		content.Links = wa.extractLinks(htmlContent, urlStr)
	}

	// Add metadata if enabled
	if wa.includeMetadata {
		content.Metadata["content_length"] = fmt.Sprintf("%d", len(htmlContent))
		content.Metadata["content_density"] = wa.calculateContentDensity(htmlContent)
	}

	// Count words and estimate tokens
	content.WordCount = wa.countWords(content.MainContent)
	content.TokenCount = wa.estimateTokens(content.MainContent)

	// Apply smart truncation if needed
	if content.TokenCount > maxTokens {
		content.MainContent = wa.smartTruncate(content.MainContent, maxTokens)
		content.TokenCount = wa.estimateTokens(content.MainContent)
		content.Truncated = true
	}

	// Convert to map for JSON output
	result := map[string]interface{}{
		"url":          content.URL,
		"title":        content.Title,
		"description":  content.Description,
		"main_content": content.MainContent,
		"headings":     content.Headings,
		"token_count":  content.TokenCount,
		"truncated":    content.Truncated,
		"word_count":   content.WordCount,
	}

	if len(content.Links) > 0 {
		result["links"] = content.Links
	}

	if len(content.Metadata) > 2 { // Only include if we added meaningful metadata
		result["metadata"] = content.Metadata
	}

	return result
}

func (wa *WebAgent) extractTitle(html string) string {
	// Simple regex to extract title
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return wa.cleanText(matches[1])
	}
	return ""
}

func (wa *WebAgent) extractMetaDescription(html string) string {
	// Extract meta description
	descRegex := regexp.MustCompile(`(?i)<meta[^>]+name=["\']description["\'][^>]+content=["\']([^"\']+)["\']`)
	matches := descRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return wa.cleanText(matches[1])
	}
	return ""
}

func (wa *WebAgent) extractHeadings(html string) []Heading {
	var headings []Heading

	// Match h1-h6 tags
	headingRegex := regexp.MustCompile(`(?i)<h([1-6])[^>]*>([^<]+)</h[1-6]>`)
	matches := headingRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 2 {
			level := 0
			fmt.Sscanf(match[1], "%d", &level)
			text := wa.cleanText(match[2])
			if text != "" && level > 0 {
				headings = append(headings, Heading{
					Level: level,
					Text:  text,
				})
			}
		}
	}

	return headings
}

func (wa *WebAgent) extractMainContent(html string) string {
	// Remove script, style, nav, header, footer elements
	replacements := []struct {
		pattern *regexp.Regexp
	}{
		{regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)},
		{regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)},
		{regexp.MustCompile(`(?i)<nav[^>]*>.*?</nav>`)},
		{regexp.MustCompile(`(?i)<header[^>]*>.*?</header>`)},
		{regexp.MustCompile(`(?i)<footer[^>]*>.*?</footer>`)},
		{regexp.MustCompile(`(?i)<aside[^>]*>.*?</aside>`)},
		{regexp.MustCompile(`(?i)<form[^>]*>.*?</form>`)},
	}

	cleaned := html
	for _, repl := range replacements {
		cleaned = repl.pattern.ReplaceAllString(cleaned, "")
	}

	// Look for common content containers
	contentSelectors := []string{
		`(?i)<main[^>]*>(.*?)</main>`,
		`(?i)<article[^>]*>(.*?)</article>`,
		`(?i)<div[^>]+class="[^"]*content[^"]*"[^>]*>(.*?)</div>`,
		`(?i)<div[^>]+id="[^"]*content[^"]*"[^>]*>(.*?)</div>`,
		`(?i)<div[^>]+class="[^"]*post[^"]*"[^>]*>(.*?)</div>`,
		`(?i)<div[^>]+class="[^"]*entry[^"]*"[^>]*>(.*?)</div>`,
	}

	for _, pattern := range contentSelectors {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(cleaned)
		if len(matches) > 1 {
			cleaned = matches[1]
			break
		}
	}

	// Remove all HTML tags
	htmlTagRegex := regexp.MustCompile(`<[^>]+>`)
	text := htmlTagRegex.ReplaceAllString(cleaned, " ")

	// Clean and normalize text
	return wa.cleanText(text)
}

func (wa *WebAgent) extractLinks(html string, baseUrl string) []Link {
	var links []Link

	// Extract links
	linkRegex := regexp.MustCompile(`(?i)<a[^>]+href=["\']([^"\']+)["\'][^>]*>([^<]*)</a>`)
	matches := linkRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 2 {
			url := match[1]
			text := wa.cleanText(match[2])

			// Skip empty or navigation links
			if text == "" || len(text) < 3 {
				continue
			}

			// Determine link type
			linkType := "external"
			if strings.HasPrefix(url, "#") || strings.HasPrefix(url, "/") {
				linkType = "internal"
			} else if strings.HasPrefix(url, "mailto:") || strings.HasPrefix(url, "tel:") {
				linkType = "special"
			}

			// Skip common navigation links
			if wa.isNavigationLink(text) {
				continue
			}

			links = append(links, Link{
				Text: text,
				URL:  url,
				Type: linkType,
			})
		}
	}

	// Limit to top links to save tokens
	if len(links) > 10 {
		links = links[:10]
	}

	return links
}

func (wa *WebAgent) cleanText(text string) string {
	// Remove extra whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// Remove common HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")

	return text
}

func (wa *WebAgent) countWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}

func (wa *WebAgent) estimateTokens(text string) int {
	// Rough estimation: 1 token ≈ 4 characters for English text
	// This is a simplification but works well for most content
	charCount := utf8.RuneCountInString(text)
	return charCount / 4
}

func (wa *WebAgent) smartTruncate(text string, maxTokens int) string {
	// Estimate target character count
	targetChars := maxTokens * 4

	if len(text) <= targetChars {
		return text
	}

	// Try to truncate at sentence boundaries
	sentences := regexp.MustCompile(`[.!?]+\s+`).Split(text, -1)
	var result string

	for _, sentence := range sentences {
		if len(result)+len(sentence) > targetChars {
			break
		}
		if result != "" {
			result += ". "
		}
		result += strings.TrimSpace(sentence)
	}

	// If we couldn't get a good sentence break, fall back to character truncation
	if len(result) == 0 {
		result = text[:targetChars]
		lastSpace := strings.LastIndex(result, " ")
		if lastSpace > 0 {
			result = result[:lastSpace]
		}
	}

	return result + " [...]"
}

func (wa *WebAgent) isNavigationLink(text string) bool {
	lowerText := strings.ToLower(text)
	navKeywords := []string{
		"home", "menu", "navigation", "nav", "skip", "search",
		"login", "register", "sign in", "sign up", "contact",
		"about", "privacy", "terms", "cookie", "settings",
	}

	for _, keyword := range navKeywords {
		if strings.Contains(lowerText, keyword) {
			return true
		}
	}

	return false
}

func (wa *WebAgent) calculateContentDensity(html string) string {
	// Simple content density: text length vs HTML length
	textRegex := regexp.MustCompile(`>[^<]+<`)
	textMatches := textRegex.FindAllString(html, -1)
	textLength := 0
	for _, match := range textMatches {
		textLength += len(strings.TrimSpace(match))
	}

	density := float64(textLength) / float64(len(html)) * 100

	if density > 50 {
		return "high"
	} else if density > 25 {
		return "medium"
	} else {
		return "low"
	}
}
