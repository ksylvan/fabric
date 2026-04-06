package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/danielmiessler/fabric/internal/i18n"
	"github.com/jessevdk/go-flags"
)

func TestWriteHelpIncludesVisualFlagDescriptions(t *testing.T) {
	if _, err := i18n.Init("en"); err != nil {
		t.Fatalf("init i18n: %v", err)
	}

	parser := flags.NewParser(&Flags{}, flags.HelpFlag|flags.PassDoubleDash)
	parser.Name = "fabric"

	var output bytes.Buffer
	NewTranslatedHelpWriter(parser, &output).WriteHelp()

	helpText := output.String()
	expected := map[string]string{
		"--visual":             i18n.T("youtube_extract_visual_data_help"),
		"--visual-sensitivity": i18n.T("youtube_visual_sensitivity_help"),
		"--visual-fps":         i18n.T("youtube_visual_fps_help"),
	}

	for flagName, description := range expected {
		if !strings.Contains(helpText, flagName) {
			t.Fatalf("expected help output to include %s", flagName)
		}
		if !strings.Contains(helpText, description) {
			t.Fatalf("expected help output to include description %q for %s", description, flagName)
		}
	}
}
