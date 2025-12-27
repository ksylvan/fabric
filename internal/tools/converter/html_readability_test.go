package converter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHtmlReadability(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Empty HTML",
			html:     "",
			expected: "",
		},
		{
			name:     "HTML with text",
			html:     "<p>Hello World</p>",
			expected: "Hello World",
		},
		{
			name:     "HTML with nested tags",
			html:     "<div><p>Hello</p><p>World</p></div>",
			expected: "HelloWorld",
		},
		{
			name:     "HTML missing tags",
			html:     "<div><p>Hello</p><p>World</div>",
			expected: "HelloWorld",
		},
		{
			name: "Real web page with navigation and ads",
			html: `<html><body>
				<nav>Site Navigation</nav>
				<aside class="ad">Advertisement</aside>
				<article>
					<h1>Main Article Title</h1>
					<p>This is the main content.</p>
				</article>
				<footer>Copyright 2025</footer>
			</body></html>`,
			expected: "Main Article Title",
		},
		{
			name:     "HTML with special characters",
			html:     "<p>Hello © 世界 🌍</p>",
			expected: "Hello © 世界 🌍",
		},
		{
			name:     "HTML with only scripts",
			html:     "<script>alert('xss')</script>",
			expected: "",
		},
		{
			name:     "HTML with multiple paragraphs",
			html:     "<html><body><article><p>First paragraph.</p><p>Second paragraph.</p></article></body></html>",
			expected: "First paragraph",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := HtmlReadability(tc.html)

			assert.NoError(t, err)
			assert.Contains(t, result, tc.expected)
		})
	}
}

func TestHtmlReadability_LargeInput(t *testing.T) {
	// Stress test with large HTML (simulates a large web page)
	largeHTML := "<html><body><article>" + strings.Repeat("<p>Content paragraph here.</p>", 10000) + "</article></body></html>"

	result, err := HtmlReadability(largeHTML)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "Content paragraph here")
}
