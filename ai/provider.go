// Package ai provides shared multi-provider LLM support across the
// missionctl suite, so a tool asking for an AI feature doesn't force every
// user into paying for an Anthropic API key.
//
// Provider auto-detection (first key found wins):
//
//	ANTHROPIC_API_KEY          → Claude Haiku
//	OPENAI_API_KEY             → GPT-4o mini
//	GEMINI_API_KEY             → Gemini (free tier available, no card required)
//	OLLAMA_HOST or default     → Ollama (free, fully local, no key)
//
// Override with <PREFIX>_PROVIDER=anthropic|openai|gemini|ollama, where
// PREFIX is the calling tool's name in caps (e.g. MAILCTL_PROVIDER). When
// on Ollama, pick a per-tool model with <PREFIX>_OLLAMA_MODEL (falls back
// to the shared OLLAMA_MODEL, then "llama3.2") — e.g. DIARYCTL_OLLAMA_MODEL
// for a code-aware model there while mailctl stays on a smaller, faster one.
//
// Deliberately not supported: reusing a claude.ai/chatgpt.com browser
// session in place of an API key. That means scraping or replaying a
// consumer web session server-side, which both providers' terms of service
// prohibit and which risks the user's own account getting flagged — not a
// trade worth making for a CLI feature.
package ai

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	anthropicopt "github.com/anthropics/anthropic-sdk-go/option"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// Provider identifies which LLM backend to use.
type Provider string

const (
	ProviderAnthropic Provider = "anthropic"
	ProviderOpenAI    Provider = "openai"
	ProviderGemini    Provider = "gemini"
	ProviderOllama    Provider = "ollama"
)

// ProviderInfo describes a detected provider.
type ProviderInfo struct {
	Name    Provider
	Model   string
	Display string // human-readable label for status lines
}

// Detect returns the active provider based on environment variables.
// envPrefix (e.g. "MAILCTL") makes <envPrefix>_PROVIDER override auto-detection.
func Detect(envPrefix string) (ProviderInfo, error) {
	override := os.Getenv(envPrefix + "_PROVIDER")

	check := func(p Provider) bool {
		return override == "" || Provider(override) == p
	}

	if check(ProviderAnthropic) && os.Getenv("ANTHROPIC_API_KEY") != "" {
		return ProviderInfo{ProviderAnthropic, "claude-haiku-4-5-20251001", "Claude Haiku (Anthropic)"}, nil
	}
	if check(ProviderOpenAI) && os.Getenv("OPENAI_API_KEY") != "" {
		return ProviderInfo{ProviderOpenAI, "gpt-4o-mini", "GPT-4o mini (OpenAI)"}, nil
	}
	if check(ProviderGemini) && os.Getenv("GEMINI_API_KEY") != "" {
		model := os.Getenv("GEMINI_MODEL")
		if model == "" {
			model = "gemini-flash-latest"
		}
		return ProviderInfo{ProviderGemini, model, "Gemini " + model + " (Google, free tier)"}, nil
	}
	if check(ProviderOllama) {
		// <PREFIX>_OLLAMA_MODEL lets each tool pin a different local model
		// (e.g. a code-aware model for diaryctl, a fast small one for
		// mailctl) — falls back to the shared OLLAMA_MODEL, then a
		// generalist default, so existing single-model setups are
		// unaffected.
		model := os.Getenv(envPrefix + "_OLLAMA_MODEL")
		if model == "" {
			model = os.Getenv("OLLAMA_MODEL")
		}
		if model == "" {
			model = "llama3.2"
		}
		return ProviderInfo{ProviderOllama, model, fmt.Sprintf("Ollama (%s, local)", model)}, nil
	}

	if override != "" {
		return ProviderInfo{}, fmt.Errorf("%s_PROVIDER=%q set but no matching credential found", envPrefix, override)
	}
	return ProviderInfo{}, fmt.Errorf(
		"no AI provider configured — set ANTHROPIC_API_KEY, OPENAI_API_KEY, a free GEMINI_API_KEY (aistudio.google.com/apikey), or run Ollama locally (ollama.com) with no key needed",
	)
}

// Call dispatches to the correct provider backend and returns the full
// response text. out, if non-nil, receives streamed chunks as they arrive.
func Call(ctx context.Context, info ProviderInfo, system, prompt string, out func(string)) (string, error) {
	switch info.Name {
	case ProviderAnthropic:
		return callAnthropic(ctx, info, system, prompt, out)
	default:
		return callOpenAICompat(ctx, info, system, prompt, out)
	}
}

func callAnthropic(ctx context.Context, info ProviderInfo, system, prompt string, out func(string)) (string, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	c := anthropic.NewClient(anthropicopt.WithAPIKey(key))

	stream := c.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(info.Model),
		MaxTokens: 4096,
		System:    []anthropic.TextBlockParam{{Text: system}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(prompt))},
	})

	var full strings.Builder
	for stream.Next() {
		event := stream.Current()
		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			if text := event.Delta.Text; text != "" {
				full.WriteString(text)
				if out != nil {
					out(text)
				}
			}
		}
	}
	if err := stream.Err(); err != nil {
		return "", fmt.Errorf("Anthropic: %w", err)
	}
	return full.String(), nil
}

// callOpenAICompat works for OpenAI, Gemini (via its OpenAI-compatible
// endpoint), and Ollama alike.
func callOpenAICompat(ctx context.Context, info ProviderInfo, system, prompt string, out func(string)) (string, error) {
	var opts []option.RequestOption

	switch info.Name {
	case ProviderOpenAI:
		opts = append(opts, option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))
	case ProviderGemini:
		opts = append(opts,
			option.WithAPIKey(os.Getenv("GEMINI_API_KEY")),
			option.WithBaseURL("https://generativelanguage.googleapis.com/v1beta/openai/"),
			option.WithMaxRetries(0), // free tier: no automatic retries — each retry burns quota
		)
	case ProviderOllama:
		host := os.Getenv("OLLAMA_HOST")
		if host == "" {
			host = "http://localhost:11434"
		}
		opts = append(opts,
			option.WithAPIKey("ollama"),
			option.WithBaseURL(host+"/v1/"),
		)
	}

	client := openai.NewClient(opts...)

	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model: info.Model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system),
			openai.UserMessage(prompt),
		},
		MaxTokens: openai.Int(4096),
	})

	var full strings.Builder
	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 {
			if text := chunk.Choices[0].Delta.Content; text != "" {
				full.WriteString(text)
				if out != nil {
					out(text)
				}
			}
		}
	}
	if err := stream.Err(); err != nil {
		return "", fmt.Errorf("%s: %w", info.Display, friendlyNetErr(err))
	}
	return full.String(), nil
}

// friendlyNetErr replaces low-level Go network errors and provider-specific
// status codes with readable, actionable messages.
func friendlyNetErr(err error) error {
	if code, ok := httpStatusCode(err); ok {
		return friendlyForStatus(code, err)
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "lookup") || strings.Contains(msg, "no such host"):
		return fmt.Errorf("DNS error — domain unreachable. Check VPN/proxy/firewall, or use Ollama (local)")
	case strings.Contains(msg, "connection refused"):
		return fmt.Errorf("connection refused — is Ollama running? (ollama serve)")
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded"):
		return fmt.Errorf("timeout — API server not responding")
	case strings.Contains(msg, "429") || strings.Contains(msg, "Too Many Requests") ||
		strings.Contains(msg, "RESOURCE_EXHAUSTED") || strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "quota") || strings.Contains(msg, "Quota"):
		return friendlyForStatus(429, err)
	case strings.Contains(msg, "404"):
		return fmt.Errorf("404 — model not found, check the model name")
	case strings.Contains(msg, "401") || strings.Contains(msg, "Unauthorized"):
		return fmt.Errorf("API key invalid")
	case strings.Contains(msg, "403") || strings.Contains(msg, "Forbidden"):
		return fmt.Errorf("access denied — API key has no permission for this resource")
	}
	return err
}

func httpStatusCode(err error) (int, bool) {
	if err == nil {
		return 0, false
	}
	v := reflect.ValueOf(err)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if f := v.FieldByName("StatusCode"); f.IsValid() && f.Kind() == reflect.Int {
		if code := int(f.Int()); code >= 400 {
			return code, true
		}
	}
	return 0, false
}

func friendlyForStatus(code int, orig error) error {
	switch code {
	case 429:
		raw := orig.Error()
		if strings.Contains(raw, "PerDay") || strings.Contains(raw, "perDay") ||
			strings.Contains(raw, "daily") || strings.Contains(raw, "Daily") {
			return fmt.Errorf("daily quota exhausted (free-tier limit). Resets at 00:00 PST, or switch provider")
		}
		return fmt.Errorf("rate limited — wait a minute and retry")
	case 404:
		return fmt.Errorf("404 — model not available, check the model name")
	case 401:
		return fmt.Errorf("401 — API key invalid")
	case 403:
		return fmt.Errorf("403 — access denied, API key has no permission for this model")
	}
	return orig
}
