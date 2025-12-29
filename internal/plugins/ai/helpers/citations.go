package helpers

import (
	"fmt"
	"strings"
)

// Citation represents a web citation with URL and title.
type Citation struct {
	URL       string
	Title     string
	CitedText string // Optional quoted text from the citation
}

// CitationDeduplicator helps deduplicate and format citations.
type CitationDeduplicator struct {
	seen map[string]bool
}

// NewCitationDeduplicator creates a new citation deduplicator.
func NewCitationDeduplicator() *CitationDeduplicator {
	return &CitationDeduplicator{
		seen: make(map[string]bool),
	}
}

// Add adds a citation if it hasn't been seen before.
// Returns true if the citation was added, false if it was a duplicate.
func (d *CitationDeduplicator) Add(citation Citation) bool {
	// Create unique key from URL and title
	key := citation.URL + "|" + citation.Title

	if d.seen[key] {
		return false
	}

	d.seen[key] = true
	return true
}

// AddAndFormat adds a citation and returns its markdown formatted string if it's new.
// Returns an empty string if the citation was a duplicate.
func (d *CitationDeduplicator) AddAndFormat(citation Citation) string {
	if !d.Add(citation) {
		return ""
	}
	return FormatCitation(citation)
}

// FormatCitation formats a single citation as a markdown list item.
func FormatCitation(citation Citation) string {
	citationText := fmt.Sprintf("- [%s](%s)", citation.Title, citation.URL)
	if citation.CitedText != "" {
		citationText += fmt.Sprintf(" - \"%s\"", citation.CitedText)
	}
	return citationText
}

// FormatCitations formats a list of citations as markdown list items.
func FormatCitations(citations []Citation) []string {
	result := make([]string, 0, len(citations))
	for _, citation := range citations {
		result = append(result, FormatCitation(citation))
	}
	return result
}

// FormatCitationsSection creates a complete "Sources" section with the citations.
// Returns an empty string if there are no citations.
func FormatCitationsSection(citations []string) string {
	if len(citations) == 0 {
		return ""
	}
	return "\n\n## Sources\n\n" + strings.Join(citations, "\n")
}

// DeduplicateAndFormat takes a list of citations, deduplicates them, and returns
// formatted markdown strings.
func DeduplicateAndFormat(citations []Citation) []string {
	deduplicator := NewCitationDeduplicator()
	result := make([]string, 0, len(citations))

	for _, citation := range citations {
		if formatted := deduplicator.AddAndFormat(citation); formatted != "" {
			result = append(result, formatted)
		}
	}

	return result
}
