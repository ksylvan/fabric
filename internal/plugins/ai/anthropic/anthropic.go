package anthropic

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/danielmiessler/fabric/internal/chat"
	"github.com/danielmiessler/fabric/internal/domain"
	debuglog "github.com/danielmiessler/fabric/internal/log"
	"github.com/danielmiessler/fabric/internal/plugins"
	"github.com/danielmiessler/fabric/internal/util"
)

const defaultBaseUrl = "https://api.anthropic.com/"

const webSearchToolName = "web_search"
const webSearchToolType = "web_search_20250305"
const sourcesHeader = "## Sources"

const authTokenIdentifier = "claude"

func NewClient() (ret *Client) {
	vendorName := "Anthropic"
	ret = &Client{}

	ret.PluginBase = &plugins.PluginBase{
		Name:            vendorName,
		EnvNamePrefix:   plugins.BuildEnvVariablePrefix(vendorName),
		ConfigureCustom: ret.configure,
	}

	ret.ApiBaseURL = ret.AddSetupQuestion("API Base URL", false)
	ret.ApiBaseURL.Value = defaultBaseUrl
	ret.UseOAuth = ret.AddSetupQuestionBool("Use OAuth login", false)
	ret.ApiKey = ret.PluginBase.AddSetupQuestion("API key", false)

	ret.maxTokens = 4096
	ret.defaultRequiredUserMessage = "Hi"
	ret.models = []string{
		string(anthropic.ModelClaude3_7SonnetLatest), string(anthropic.ModelClaude3_7Sonnet20250219),
		string(anthropic.ModelClaude3_5HaikuLatest), string(anthropic.ModelClaude3_5Haiku20241022),
		string(anthropic.ModelClaude3OpusLatest), string(anthropic.ModelClaude_3_Opus_20240229),
		string(anthropic.ModelClaude_3_Haiku_20240307),
		string(anthropic.ModelClaudeOpus4_20250514), string(anthropic.ModelClaudeSonnet4_20250514),
		string(anthropic.ModelClaudeOpus4_1_20250805),
		string(anthropic.ModelClaudeSonnet4_5),
		string(anthropic.ModelClaudeSonnet4_5_20250929),
		string(anthropic.ModelClaudeOpus4_5_20251101),
		string(anthropic.ModelClaudeOpus4_5),
		string(anthropic.ModelClaudeHaiku4_5),
		string(anthropic.ModelClaudeHaiku4_5_20251001),
	}

	ret.modelBetas = map[string][]string{
		string(anthropic.ModelClaudeSonnet4_20250514):   {"context-1m-2025-08-07"},
		string(anthropic.ModelClaudeSonnet4_5):          {"context-1m-2025-08-07"},
		string(anthropic.ModelClaudeSonnet4_5_20250929): {"context-1m-2025-08-07"},
	}

	return
}

// IsConfigured returns true if either the API key or OAuth is configured
func (an *Client) IsConfigured() bool {
	if an.ApiKey.Value != "" {
		return true
	}

	return an.checkOAuthConfiguration()
}

// checkOAuthConfiguration checks if OAuth is configured and has a valid token
func (an *Client) checkOAuthConfiguration() bool {
	if !plugins.ParseBoolElseFalse(an.UseOAuth.Value) {
		return false
	}

	storage, err := util.NewOAuthStorage()
	if err != nil {
		return false
	}

	// If valid token exists, we're configured
	if storage.HasValidToken(authTokenIdentifier, 5) {
		return true
	}

	// Try to authenticate via OAuth flow
	return an.runOAuthAuthentication(storage)
}

// runOAuthAuthentication runs the OAuth flow and validates the result
func (an *Client) runOAuthAuthentication(storage *util.OAuthStorage) bool {
	fmt.Println("OAuth enabled but no valid token found. Starting authentication...")

	_, err := RunOAuthFlow(authTokenIdentifier)
	if err != nil {
		fmt.Printf("OAuth authentication failed: %v\n", err)
		return false
	}

	return storage.HasValidToken(authTokenIdentifier, 5)
}

type Client struct {
	*plugins.PluginBase
	ApiBaseURL *plugins.SetupQuestion
	ApiKey     *plugins.SetupQuestion
	UseOAuth   *plugins.SetupQuestion

	maxTokens                  int
	defaultRequiredUserMessage string
	models                     []string
	modelBetas                 map[string][]string

	client anthropic.Client
}

func (an *Client) Setup() (err error) {
	if err = an.PluginBase.Ask(an.Name); err != nil {
		return
	}

	if plugins.ParseBoolElseFalse(an.UseOAuth.Value) {
		// Check if we have a valid stored token
		storage, err := util.NewOAuthStorage()
		if err != nil {
			return err
		}

		if !storage.HasValidToken(authTokenIdentifier, 5) {
			// No valid token, run OAuth flow
			if _, err = RunOAuthFlow(authTokenIdentifier); err != nil {
				return err
			}
		}
	}

	err = an.configure()
	return
}

func (an *Client) configure() (err error) {
	opts := []option.RequestOption{}

	if an.ApiBaseURL.Value != "" {
		opts = append(opts, option.WithBaseURL(an.ApiBaseURL.Value))
	}

	if plugins.ParseBoolElseFalse(an.UseOAuth.Value) {
		// For OAuth, use Bearer token with custom headers
		// Create custom HTTP client that adds OAuth Bearer token and beta header
		baseTransport := &http.Transport{}
		httpClient := &http.Client{
			Transport: NewOAuthTransport(an, baseTransport),
		}
		opts = append(opts, option.WithHTTPClient(httpClient))
	} else {
		opts = append(opts, option.WithAPIKey(an.ApiKey.Value))
	}

	an.client = anthropic.NewClient(opts...)
	return
}

func (an *Client) ListModels() (ret []string, err error) {
	return an.models, nil
}

func parseThinking(level domain.ThinkingLevel) (anthropic.ThinkingConfigParamUnion, bool) {
	lower := strings.ToLower(string(level))
	switch domain.ThinkingLevel(lower) {
	case domain.ThinkingOff:
		disabled := anthropic.NewThinkingConfigDisabledParam()
		return anthropic.ThinkingConfigParamUnion{OfDisabled: &disabled}, true
	case domain.ThinkingLow, domain.ThinkingMedium, domain.ThinkingHigh:
		if budget, ok := domain.ThinkingBudgets[domain.ThinkingLevel(lower)]; ok {
			return anthropic.ThinkingConfigParamOfEnabled(budget), true
		}
	default:
		if tokens, err := strconv.ParseInt(lower, 10, 64); err == nil {
			if tokens >= 1 && tokens <= 10000 {
				return anthropic.ThinkingConfigParamOfEnabled(tokens), true
			}
		}
	}
	return anthropic.ThinkingConfigParamUnion{}, false
}

func (an *Client) SendStream(
	msgs []*chat.ChatCompletionMessage, opts *domain.ChatOptions, channel chan string,
) (err error) {
	messages := an.toMessages(msgs)
	if len(messages) == 0 {
		close(channel)
		// No messages to send after normalization, consider this a non-error condition for streaming.
		return
	}

	ctx := context.Background()

	params := an.buildMessageParams(messages, opts)
	betas := an.modelBetas[opts.Model]
	var reqOpts []option.RequestOption
	if len(betas) > 0 {
		reqOpts = append(reqOpts, option.WithHeader(headerAnthropicBeta, strings.Join(betas, ",")))
	}
	stream := an.client.Messages.NewStreaming(ctx, params, reqOpts...)
	if stream.Err() != nil && len(betas) > 0 {
		debuglog.Debug(debuglog.Basic, "Anthropic beta feature %s failed: %v\n", strings.Join(betas, ","), stream.Err())
		stream = an.client.Messages.NewStreaming(ctx, params)
	}

	for stream.Next() {
		event := stream.Current()

		// directly send any non-empty delta text
		if event.Delta.Text != "" {
			channel <- event.Delta.Text
		}
	}

	if stream.Err() != nil {
		fmt.Fprintf(os.Stderr, "Messages stream error: %v\n", stream.Err())
	}
	close(channel)
	return
}

func (an *Client) buildMessageParams(msgs []anthropic.MessageParam, opts *domain.ChatOptions) (
	params anthropic.MessageNewParams) {

	params = anthropic.MessageNewParams{
		Model:     anthropic.Model(opts.Model),
		MaxTokens: int64(an.maxTokens),
		Messages:  msgs,
	}

	// Only set one of Temperature or TopP as some models don't allow both
	// Always set temperature to ensure consistent behavior (Anthropic default is 1.0, Fabric default is 0.7)
	if opts.TopP != domain.DefaultTopP {
		// User explicitly set TopP, so use that instead of temperature
		params.TopP = anthropic.Opt(opts.TopP)
	} else {
		// Use temperature (always set to ensure Fabric's default of 0.7, not Anthropic's 1.0)
		params.Temperature = anthropic.Opt(opts.Temperature)
	}

	// Add Claude Code spoofing system message for OAuth authentication
	if plugins.ParseBoolElseFalse(an.UseOAuth.Value) {
		params.System = []anthropic.TextBlockParam{
			{
				Type: "text",
				Text: "You are Claude Code, Anthropic's official CLI for Claude.",
			},
		}

	}

	if opts.Search {
		// Build the web-search tool definition:
		webTool := anthropic.WebSearchTool20250305Param{
			Name:         webSearchToolName,
			Type:         webSearchToolType,
			CacheControl: anthropic.NewCacheControlEphemeralParam(),
		}

		if opts.SearchLocation != "" {
			webTool.UserLocation.Type = "approximate"
			webTool.UserLocation.Timezone = anthropic.Opt(opts.SearchLocation)
		}

		// Wrap it in the union:
		params.Tools = []anthropic.ToolUnionParam{
			{OfWebSearchTool20250305: &webTool},
		}
	}

	if t, ok := parseThinking(opts.Thinking); ok {
		params.Thinking = t
	}

	return
}

func (an *Client) Send(ctx context.Context, msgs []*chat.ChatCompletionMessage, opts *domain.ChatOptions) (
	ret string, err error) {

	messages := an.toMessages(msgs)
	if len(messages) == 0 {
		// No messages to send after normalization, return empty string and no error.
		return
	}

	var message *anthropic.Message
	params := an.buildMessageParams(messages, opts)
	betas := an.modelBetas[opts.Model]

	message, err = an.sendMessageWithBetaFallback(ctx, params, betas)
	if err != nil {
		return
	}

	textParts, citations := an.extractTextAndCitations(message.Content)
	ret = an.formatResponseWithCitations(textParts, citations)

	return
}

// extractTextAndCitations extracts text content and citations from message blocks
func (an *Client) extractTextAndCitations(contentBlocks []anthropic.ContentBlockUnion) ([]string, []string) {
	var textParts []string
	var citations []string
	citationMap := make(map[string]bool)

	for _, block := range contentBlocks {
		if !an.isTextBlock(block) {
			continue
		}

		textParts = append(textParts, block.Text)
		citations = an.extractCitationsFromBlock(block, citationMap, citations)
	}

	return textParts, citations
}

// isTextBlock checks if a content block is a non-empty text block
func (an *Client) isTextBlock(block anthropic.ContentBlockUnion) bool {
	return block.Type == "text" && block.Text != ""
}

// extractCitationsFromBlock extracts unique citations from a content block
func (an *Client) extractCitationsFromBlock(block anthropic.ContentBlockUnion, citationMap map[string]bool, citations []string) []string {
	for _, citation := range block.Citations {
		if citation.Type != "web_search_result_location" {
			continue
		}

		citations = an.addUniqueCitation(citation, citationMap, citations)
	}
	return citations
}

// addUniqueCitation adds a citation if it hasn't been seen before
func (an *Client) addUniqueCitation(citation anthropic.TextCitationUnion, citationMap map[string]bool, citations []string) []string {
	citationKey := citation.URL + "|" + citation.Title
	if citationMap[citationKey] {
		return citations
	}

	citationMap[citationKey] = true
	citationText := fmt.Sprintf("- [%s](%s)", citation.Title, citation.URL)
	if citation.CitedText != "" {
		citationText += fmt.Sprintf(" - \"%s\"", citation.CitedText)
	}

	return append(citations, citationText)
}

// formatResponseWithCitations formats the response text with optional citations section
func (an *Client) formatResponseWithCitations(textParts []string, citations []string) string {
	var resultBuilder strings.Builder
	resultBuilder.WriteString(strings.Join(textParts, ""))

	if len(citations) > 0 {
		resultBuilder.WriteString("\n\n")
		resultBuilder.WriteString(sourcesHeader)
		resultBuilder.WriteString("\n\n")
		resultBuilder.WriteString(strings.Join(citations, "\n"))
	}

	return resultBuilder.String()
}

// sendMessageWithBetaFallback sends a message with beta features, falling back to standard API on failure
func (an *Client) sendMessageWithBetaFallback(ctx context.Context, params anthropic.MessageNewParams, betas []string) (*anthropic.Message, error) {
	// Try with beta features if available
	if len(betas) > 0 {
		reqOpts := []option.RequestOption{option.WithHeader(headerAnthropicBeta, strings.Join(betas, ","))}
		message, err := an.client.Messages.New(ctx, params, reqOpts...)
		if err != nil {
			debuglog.Debug(debuglog.Basic, "Anthropic beta feature %s failed: %v\n", strings.Join(betas, ","), err)
			// Fall back to standard API without beta features
			return an.client.Messages.New(ctx, params)
		}
		return message, nil
	}

	// No beta features, use standard API
	return an.client.Messages.New(ctx, params)
}

func (an *Client) toMessages(msgs []*chat.ChatCompletionMessage) (ret []anthropic.MessageParam) {
	// Custom normalization for Anthropic:
	// - System messages become the first part of the first user message.
	// - Messages must alternate user/assistant.
	// - Skip empty messages.

	var anthropicMessages []anthropic.MessageParam
	var systemContent string

	// Note: Claude Code spoofing is now handled in buildMessageParams

	isFirstUserMessage := true
	lastRoleWasUser := false

	for _, msg := range msgs {
		if strings.TrimSpace(msg.Content) == "" {
			continue // Skip empty messages
		}

		switch msg.Role {
		case chat.ChatMessageRoleSystem:
			// Accumulate system content. It will be prepended to the first user message.
			if systemContent != "" {
				systemContent += "\\n" + msg.Content
			} else {
				systemContent = msg.Content
			}
		case chat.ChatMessageRoleUser:
			userContent := msg.Content
			if isFirstUserMessage && systemContent != "" {
				userContent = systemContent + "\\n\\n" + userContent
				isFirstUserMessage = false // System content now consumed
			}
			if lastRoleWasUser {
				// Enforce alternation: add a minimal assistant message if two user messages are consecutive.
				// This shouldn't happen with current chatter.go logic but is a safeguard.
				anthropicMessages = append(anthropicMessages, anthropic.NewAssistantMessage(anthropic.NewTextBlock("Okay.")))
			}
			anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(userContent)))
			lastRoleWasUser = true
		case chat.ChatMessageRoleAssistant:
			// If the first message is an assistant message, and we have system content,
			// prepend a user message with the system content.
			if isFirstUserMessage && systemContent != "" {
				anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(systemContent)))
				lastRoleWasUser = true
				isFirstUserMessage = false // System content now consumed
			} else if !lastRoleWasUser && len(anthropicMessages) > 0 {
				// Enforce alternation: add a minimal user message if two assistant messages are consecutive
				// or if an assistant message is first without prior system prompt handling.
				anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(an.defaultRequiredUserMessage)))
				lastRoleWasUser = true
			}
			anthropicMessages = append(anthropicMessages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
			lastRoleWasUser = false
		default:
			// Other roles (like 'meta') are ignored for Anthropic's message structure.
			continue
		}
	}

	// If only system content was provided, create a user message with it.
	if len(anthropicMessages) == 0 && systemContent != "" {
		anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(systemContent)))
	}

	return anthropicMessages
}

func (an *Client) NeedsRawMode(modelName string) bool {
	return false
}
