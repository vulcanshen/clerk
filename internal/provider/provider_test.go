package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vulcanshen/clerk/internal/config"
)

// --- Preset tests ---

func TestFindPreset(t *testing.T) {
	tests := []struct {
		name  string
		found bool
	}{
		{"groq", true},
		{"gemini", true},
		{"openai", true},
		{"ollama", true},
		{"claude", true},
		{"unknown", false},
	}
	for _, tt := range tests {
		p := FindPreset(tt.name)
		if tt.found && p == nil {
			t.Errorf("FindPreset(%q) = nil, want preset", tt.name)
		}
		if !tt.found && p != nil {
			t.Errorf("FindPreset(%q) = %v, want nil", tt.name, p)
		}
	}
}

func TestPresetFields(t *testing.T) {
	groq := FindPreset("groq")
	if groq.Endpoint == "" || groq.DefaultModel == "" {
		t.Error("groq preset should have endpoint and default model")
	}
	claude := FindPreset("claude")
	if claude.Endpoint != "" {
		t.Error("claude preset should have empty endpoint (uses CLI)")
	}
}

// --- Resolve tests ---

func TestIsClaudeProvider(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"claude", true},
		{"Claude", true},
		{"CLAUDE", true},
		{"groq", false},
		{"openai", false},
	}
	for _, tt := range tests {
		got := isClaudeProvider(tt.input)
		if got != tt.expected {
			t.Errorf("isClaudeProvider(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestApplyPreset(t *testing.T) {
	// Known provider with empty fields → preset fills in
	pc := applyPreset(config.ProviderConfig{Name: "groq"})
	if pc.Endpoint == "" {
		t.Error("applyPreset should fill endpoint for groq")
	}
	if pc.Model == "" {
		t.Error("applyPreset should fill model for groq")
	}

	// User-set values take priority
	pc = applyPreset(config.ProviderConfig{
		Name:     "groq",
		Model:    "custom-model",
		Endpoint: "https://custom.endpoint",
	})
	if pc.Model != "custom-model" {
		t.Errorf("user model should take priority, got %q", pc.Model)
	}
	if pc.Endpoint != "https://custom.endpoint" {
		t.Errorf("user endpoint should take priority, got %q", pc.Endpoint)
	}

	// Unknown provider → no change
	pc = applyPreset(config.ProviderConfig{Name: "unknown"})
	if pc.Endpoint != "" || pc.Model != "" {
		t.Error("unknown provider should not get preset values")
	}
}

func TestResolveForSummary(t *testing.T) {
	// Claude (default)
	cfg := config.Config{}
	p := ResolveForSummary(cfg)
	if _, ok := p.(*ClaudeCliProvider); !ok {
		t.Error("empty provider should resolve to ClaudeCliProvider")
	}

	// Groq
	cfg.Summary.Provider.Name = "groq"
	cfg.Summary.Provider.APIKey = "test-key"
	p = ResolveForSummary(cfg)
	op, ok := p.(*OpenAIProvider)
	if !ok {
		t.Fatal("groq should resolve to OpenAIProvider")
	}
	if op.Endpoint == "" {
		t.Error("groq should have preset endpoint")
	}
	if op.APIKey != "test-key" {
		t.Errorf("api key should be test-key, got %q", op.APIKey)
	}
}

func TestResolveForReport(t *testing.T) {
	// Report with no config → fallback to summary
	cfg := config.Config{}
	cfg.Summary.Provider.Name = "groq"
	cfg.Summary.Provider.APIKey = "summary-key"

	p := ResolveForReport(cfg)
	op, ok := p.(*OpenAIProvider)
	if !ok {
		t.Fatal("should fallback to summary's groq provider")
	}
	if op.APIKey != "summary-key" {
		t.Errorf("should use summary api key, got %q", op.APIKey)
	}

	// Report with own config → overrides
	cfg.Report.Provider.Name = "openai"
	cfg.Report.Provider.APIKey = "report-key"
	p = ResolveForReport(cfg)
	op, ok = p.(*OpenAIProvider)
	if !ok {
		t.Fatal("should resolve to OpenAIProvider")
	}
	if op.APIKey != "report-key" {
		t.Errorf("should use report api key, got %q", op.APIKey)
	}
}

func TestMergeProvider(t *testing.T) {
	primary := config.ProviderConfig{Name: "groq", APIKey: "key1"}
	secondary := config.ProviderConfig{Name: "openai", Model: "gpt-4", Endpoint: "https://api.openai.com", APIKey: "key2"}

	merged := mergeProvider(primary, secondary)
	if merged.Name != "groq" {
		t.Errorf("name should be primary, got %q", merged.Name)
	}
	if merged.Model != "gpt-4" {
		t.Errorf("model should fallback to secondary, got %q", merged.Model)
	}
	if merged.Endpoint != "https://api.openai.com" {
		t.Errorf("endpoint should fallback to secondary, got %q", merged.Endpoint)
	}
	if merged.APIKey != "key1" {
		t.Errorf("api key should be primary, got %q", merged.APIKey)
	}
}

// --- OpenAI provider tests (with mock server) ---

func TestOpenAIProviderComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer test-key, got %s", r.Header.Get("Authorization"))
		}

		var req chatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "test-model" {
			t.Errorf("expected model test-model, got %s", req.Model)
		}
		if len(req.Messages) != 2 {
			t.Errorf("expected 2 messages (system + user), got %d", len(req.Messages))
		}

		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "mock response"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &OpenAIProvider{
		Endpoint: server.URL,
		Model:    "test-model",
		APIKey:   "test-key",
	}

	out, err := p.Complete(context.Background(), "hello", "be helpful")
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if out != "mock response" {
		t.Errorf("expected 'mock response', got %q", out)
	}
}

func TestOpenAIProviderNoSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if len(req.Messages) != 1 {
			t.Errorf("expected 1 message (user only), got %d", len(req.Messages))
		}
		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "ok"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &OpenAIProvider{Endpoint: server.URL, Model: "m"}
	out, err := p.Complete(context.Background(), "test", "")
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if out != "ok" {
		t.Errorf("expected 'ok', got %q", out)
	}
}

func TestOpenAIProviderAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer server.Close()

	p := &OpenAIProvider{Endpoint: server.URL, Model: "m", APIKey: "k", InitBackoff: time.Millisecond}
	_, err := p.Complete(context.Background(), "test", "")
	if err == nil {
		t.Error("expected error for 429 response")
	}
}

func TestOpenAIProviderEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatResponse{})
	}))
	defer server.Close()

	p := &OpenAIProvider{Endpoint: server.URL, Model: "m"}
	_, err := p.Complete(context.Background(), "test", "")
	if err == nil {
		t.Error("expected error for empty choices")
	}
}

func TestOpenAIProviderTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	p := &OpenAIProvider{Endpoint: server.URL, Model: "m"}
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()

	_, err := p.Complete(ctx, "test", "")
	if err == nil {
		t.Error("expected timeout error")
	}
}

// --- ListModels tests ---

func TestListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Errorf("expected /models path, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(modelsResponse{
			Data: []ModelInfo{
				{ID: "model-b"},
				{ID: "model-a"},
			},
		})
	}))
	defer server.Close()

	models, err := ListModels(context.Background(), server.URL, "key")
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
	// Should be sorted
	if models[0] != "model-a" || models[1] != "model-b" {
		t.Errorf("models should be sorted, got %v", models)
	}
}

func TestOpenAIProviderRetryOn429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(429)
			w.Write([]byte(`{"error":{"message":"rate limited"}}`))
			return
		}
		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "success after retry"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &OpenAIProvider{Endpoint: server.URL, Model: "m", APIKey: "k", InitBackoff: time.Millisecond}
	out, err := p.Complete(context.Background(), "test", "")
	if err != nil {
		t.Fatalf("expected success after retry, got error: %v", err)
	}
	if out != "success after retry" {
		t.Errorf("expected 'success after retry', got %q", out)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestOpenAIProviderRetryExhausted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer server.Close()

	p := &OpenAIProvider{Endpoint: server.URL, Model: "m", APIKey: "k", InitBackoff: time.Millisecond}
	_, err := p.Complete(context.Background(), "test", "")
	if err == nil {
		t.Error("expected error after all retries exhausted")
	}
	if !strings.Contains(err.Error(), "rate limited") || !strings.Contains(err.Error(), "retries") {
		t.Errorf("error should mention rate limiting and retries, got: %v", err)
	}
}

func TestOpenAIProviderRetryOn500(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 1 {
			w.WriteHeader(500)
			w.Write([]byte("internal server error"))
			return
		}
		resp := chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "recovered"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &OpenAIProvider{Endpoint: server.URL, Model: "m", APIKey: "k", InitBackoff: time.Millisecond}
	out, err := p.Complete(context.Background(), "test", "")
	if err != nil {
		t.Fatalf("expected success after 500 retry, got: %v", err)
	}
	if out != "recovered" {
		t.Errorf("expected 'recovered', got %q", out)
	}
}

// --- fallback tests ---

func TestFallback(t *testing.T) {
	if fallback("a", "b") != "a" {
		t.Error("should return primary when non-empty")
	}
	if fallback("", "b") != "b" {
		t.Error("should return fallback when primary empty")
	}
	if fallback("", "") != "" {
		t.Error("should return empty when both empty")
	}
}
