package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/danielmiessler/fabric/internal/chat"

	"github.com/danielmiessler/fabric/internal/domain"
	"github.com/danielmiessler/fabric/internal/plugins/ai"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
	"github.com/danielmiessler/fabric/internal/plugins/strategy"
	"github.com/danielmiessler/fabric/internal/plugins/template"
)

const NoSessionPatternUserMessages = "no session, pattern or user messages provided"

type Chatter struct {
	db *fsdb.Db

	Stream bool
	DryRun bool

	model              string
	modelContextLength int
	vendor             ai.Vendor
	strategy           string
}

// Send processes a chat request and applies file changes for create_coding_feature pattern
func (o *Chatter) Send(request *domain.ChatRequest, opts *domain.ChatOptions) (session *fsdb.Session, err error) {
	// Use o.model (normalized) for NeedsRawMode check instead of opts.Model
	// This ensures case-insensitive model names work correctly (e.g., "GPT-5" → "gpt-5")
	if o.vendor.NeedsRawMode(o.model) {
		opts.Raw = true
	}
	if session, err = o.BuildSession(request, opts.Raw); err != nil {
		return
	}

	vendorMessages := session.GetVendorMessages()
	if len(vendorMessages) == 0 {
		if session.Name != "" {
			err = o.db.Sessions.SaveSession(session)
			if err != nil {
				return
			}
		}
		err = fmt.Errorf("cannot send chat request: no messages provided")
		return
	}

	// Always use the normalized model name from the Chatter
	// This handles cases where user provides "GPT-5" but we've normalized it to "gpt-5"
	opts.Model = o.model

	if opts.ModelContextLength == 0 {
		opts.ModelContextLength = o.modelContextLength
	}

	var message string
	var messageBuilder strings.Builder

	if o.Stream {
		responseChan := make(chan string)
		errChan := make(chan error, 1)
		done := make(chan struct{})
		printedStream := false

		go func() {
			defer close(done)
			if streamErr := o.vendor.SendStream(session.GetVendorMessages(), opts, responseChan); streamErr != nil {
				errChan <- streamErr
			}
		}()

		for response := range responseChan {
			messageBuilder.WriteString(response)
			if !opts.SuppressThink {
				fmt.Print(response)
				printedStream = true
			}
		}

		message = messageBuilder.String()
		if printedStream && !opts.SuppressThink && !strings.HasSuffix(message, "\n") {
			fmt.Println()
		}

		// Wait for goroutine to finish
		<-done

		// Check for errors in errChan
		select {
		case streamErr := <-errChan:
			if streamErr != nil {
				err = streamErr
				return
			}
		default:
			// No errors, continue
		}
	} else {
		if message, err = o.vendor.Send(context.Background(), session.GetVendorMessages(), opts); err != nil {
			return
		}
	}

	if opts.SuppressThink && !o.DryRun {
		message = domain.StripThinkBlocks(message, opts.ThinkStartTag, opts.ThinkEndTag)
	}

	if message == "" {
		session = nil
		err = fmt.Errorf("empty response from AI model")
		return
	}

	// Process file changes for create_coding_feature pattern
	if request.PatternName == "create_coding_feature" {
		summary, fileChanges, parseErr := domain.ParseFileChanges(message)
		if parseErr != nil {
			fmt.Printf("Warning: Failed to parse file changes: %v\n", parseErr)
		} else if len(fileChanges) > 0 {
			projectRoot, err := os.Getwd()
			if err != nil {
				fmt.Printf("Warning: Failed to get current directory: %v\n", err)
			} else {
				if applyErr := domain.ApplyFileChanges(projectRoot, fileChanges); applyErr != nil {
					fmt.Printf("Warning: Failed to apply file changes: %v\n", applyErr)
				} else {
					fmt.Println("Successfully applied file changes.")
					fmt.Printf("You can review the changes with 'git diff' if you're using git.\n\n")
				}
			}
		}
		message = summary
	}

	session.Append(&chat.ChatCompletionMessage{Role: chat.ChatMessageRoleAssistant, Content: message})

	if session.Name != "" {
		err = o.db.Sessions.SaveSession(session)
	}
	return
}

// BuildSession constructs a chat session from a ChatRequest, loading or creating the session,
// applying template variables, loading pattern and context content, and populating messages
// based on the mode (raw or normal). In raw mode, it builds multi-content messages with
// attachments. In normal mode, it processes user messages and applies pattern system prompts.
// Returns the fully populated session ready for AI interaction, or an error if any step fails.
func (o *Chatter) BuildSession(request *domain.ChatRequest, raw bool) (session *fsdb.Session, err error) {
	session, err = o.loadOrCreateSession(request)
	if err != nil {
		return
	}

	if request.Meta != "" {
		session.Append(&chat.ChatCompletionMessage{Role: domain.ChatMessageRoleMeta, Content: request.Meta})
	}

	contextContent, err := o.loadContextContent(request)
	if err != nil {
		return nil, err
	}

	if err = o.processMessageTemplateVariables(request); err != nil {
		return nil, err
	}

	patternContent, inputUsed, err := o.loadPatternContent(request)
	if err != nil {
		return nil, err
	}

	systemMessage, err := o.buildSystemMessage(request, contextContent, patternContent)
	if err != nil {
		return nil, err
	}

	o.populateSessionMessages(session, request, systemMessage, inputUsed, raw)

	if session.IsEmpty() {
		session = nil
		err = errors.New(NoSessionPatternUserMessages)
	}
	return
}

func (o *Chatter) loadOrCreateSession(request *domain.ChatRequest) (*fsdb.Session, error) {
	if request.SessionName != "" {
		sess, err := o.db.Sessions.Get(request.SessionName)
		if err != nil {
			return nil, fmt.Errorf("could not find session %s: %w", request.SessionName, err)
		}
		return sess, nil
	}
	return &fsdb.Session{}, nil
}

func (o *Chatter) loadContextContent(request *domain.ChatRequest) (string, error) {
	if request.ContextName == "" {
		return "", nil
	}

	ctx, err := o.db.Contexts.Get(request.ContextName)
	if err != nil {
		return "", fmt.Errorf("could not find context %s: %w", request.ContextName, err)
	}
	return ctx.Content, nil
}

func (o *Chatter) processMessageTemplateVariables(request *domain.ChatRequest) error {
	if request.Message == nil {
		request.Message = &chat.ChatCompletionMessage{
			Role:    chat.ChatMessageRoleUser,
			Content: "",
		}
	}

	if request.InputHasVars && !request.NoVariableReplacement {
		content, err := template.ApplyTemplate(request.Message.Content, request.PatternVariables, "")
		if err != nil {
			return err
		}
		request.Message.Content = content
	}

	return nil
}

func (o *Chatter) loadPatternContent(request *domain.ChatRequest) (string, bool, error) {
	if request.PatternName == "" {
		return "", false, nil
	}

	var pattern *fsdb.Pattern
	var err error

	if request.NoVariableReplacement {
		pattern, err = o.db.Patterns.GetWithoutVariables(request.PatternName, request.Message.Content)
	} else {
		pattern, err = o.db.Patterns.GetApplyVariables(request.PatternName, request.PatternVariables, request.Message.Content)
	}

	if err != nil {
		return "", false, fmt.Errorf("could not get pattern %s: %w", request.PatternName, err)
	}

	return pattern.Pattern, true, nil
}

func (o *Chatter) buildSystemMessage(request *domain.ChatRequest, contextContent, patternContent string) (string, error) {
	systemMessage := strings.TrimSpace(contextContent) + strings.TrimSpace(patternContent)

	if request.StrategyName != "" {
		strategy, err := strategy.LoadStrategy(request.StrategyName)
		if err != nil {
			return "", fmt.Errorf("could not load strategy %s: %w", request.StrategyName, err)
		}
		if strategy != nil && strategy.Prompt != "" {
			systemMessage = fmt.Sprintf("%s\n%s", strategy.Prompt, systemMessage)
		}
	}

	if request.Language != "" && request.Language != "en" {
		systemMessage = fmt.Sprintf("%s\n\nIMPORTANT: First, execute the instructions provided in this prompt using the user's input. Second, ensure your entire final response, including any section headers or titles generated as part of executing the instructions, is written ONLY in the %s language.", systemMessage, request.Language)
	}

	return systemMessage, nil
}

func (o *Chatter) populateSessionMessages(session *fsdb.Session, request *domain.ChatRequest, systemMessage string, inputUsed, raw bool) {
	if raw {
		o.populateRawModeMessages(session, request, systemMessage)
	} else {
		o.populateNormalModeMessages(session, request, systemMessage, inputUsed)
	}
}

func (o *Chatter) populateRawModeMessages(session *fsdb.Session, request *domain.ChatRequest, systemMessage string) {
	if systemMessage == "" {
		if request.Message != nil {
			session.Append(request.Message)
		}
		return
	}

	finalContent := o.buildRawModeContent(request, systemMessage)

	if len(request.Message.MultiContent) > 0 {
		request.Message = o.buildMultiContentMessage(request, finalContent)
	} else {
		request.Message = &chat.ChatCompletionMessage{
			Role:    chat.ChatMessageRoleUser,
			Content: finalContent,
		}
	}

	if request.Message != nil {
		session.Append(request.Message)
	}
}

func (o *Chatter) buildRawModeContent(request *domain.ChatRequest, systemMessage string) string {
	if request.PatternName != "" {
		return systemMessage
	}
	return fmt.Sprintf("%s\n\n%s", systemMessage, request.Message.Content)
}

func (o *Chatter) buildMultiContentMessage(request *domain.ChatRequest, finalContent string) *chat.ChatCompletionMessage {
	newMultiContent := []chat.ChatMessagePart{
		{
			Type: chat.ChatMessagePartTypeText,
			Text: finalContent,
		},
	}

	for _, part := range request.Message.MultiContent {
		if part.Type != chat.ChatMessagePartTypeText {
			newMultiContent = append(newMultiContent, part)
		}
	}

	return &chat.ChatCompletionMessage{
		Role:         chat.ChatMessageRoleUser,
		MultiContent: newMultiContent,
	}
}

func (o *Chatter) populateNormalModeMessages(session *fsdb.Session, request *domain.ChatRequest, systemMessage string, inputUsed bool) {
	if systemMessage != "" {
		session.Append(&chat.ChatCompletionMessage{Role: chat.ChatMessageRoleSystem, Content: systemMessage})
	}

	if len(request.Message.MultiContent) > 0 || (request.Message != nil && !inputUsed) {
		session.Append(request.Message)
	}
}
