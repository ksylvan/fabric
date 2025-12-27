package converter

import (
	"fmt"
	"strings"

	"github.com/go-shiori/go-readability"
)

// HtmlReadability converts HTML input into clean, readable text content.
// It extracts the main article content from a web page, removing navigation,
// ads, and other non-essential elements using the go-readability library.
//
// Parameters:
//   - html: Full HTML content of a web page
//
// Returns:
//   - string: Extracted text content of the main article
//   - error: Parser error if HTML is malformed or cannot be processed
func HtmlReadability(html string) (ret string, err error) {
	var article readability.Article
	if article, err = readability.FromReader(strings.NewReader(html), nil); err != nil {
		err = fmt.Errorf("failed to parse HTML: %w", err)
		return
	}
	ret = article.TextContent
	return
}
