package domain

import "testing"

func TestStripThinkBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		startTag string
		endTag   string
		want     string
	}{
		{
			name:     "standard think tags with newlines",
			input:    "<think>internal</think>\n\nresult",
			startTag: "<think>",
			endTag:   "</think>",
			want:     "result",
		},
		{
			name:     "custom tags",
			input:    "[[t]]hidden[[/t]] visible",
			startTag: "[[t]]",
			endTag:   "[[/t]]",
			want:     "visible",
		},
		{
			name:     "no think blocks",
			input:    "just visible text",
			startTag: "<think>",
			endTag:   "</think>",
			want:     "just visible text",
		},
		{
			name:     "multiple think blocks",
			input:    "<think>first</think> visible <think>second</think> more text",
			startTag: "<think>",
			endTag:   "</think>",
			want:     "visible more text",
		},
		{
			name:     "empty input",
			input:    "",
			startTag: "<think>",
			endTag:   "</think>",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripThinkBlocks(tt.input, tt.startTag, tt.endTag)
			if got != tt.want {
				t.Errorf("StripThinkBlocks() = %q, want %q", got, tt.want)
			}
		})
	}
}
