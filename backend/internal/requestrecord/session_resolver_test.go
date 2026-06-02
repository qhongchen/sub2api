package requestrecord

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionResolverOpenAIExplicitPriority(t *testing.T) {
	resolver := SessionResolver{}
	req := httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	req.Header.Set("session_id", " sess-header ")
	req.Header.Set("conversation_id", "conv-header")

	got := resolver.ResolveOpenAI(req, []byte(`{"prompt_cache_key":"body-cache"}`))

	if got.SessionID != "sess-header" {
		t.Fatalf("SessionID=%q, want sess-header", got.SessionID)
	}
	if got.Source != SessionSourceHeaderSessionID {
		t.Fatalf("Source=%q, want %q", got.Source, SessionSourceHeaderSessionID)
	}
	if got.ClientSessionID != "sess-header" {
		t.Fatalf("ClientSessionID=%q, want sess-header", got.ClientSessionID)
	}
}

func TestSessionResolverOpenAIConversationAndPromptCacheFallback(t *testing.T) {
	resolver := SessionResolver{}

	reqWithConversation := httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	reqWithConversation.Header.Set("conversation_id", "conv-123")
	gotConversation := resolver.ResolveOpenAI(reqWithConversation, []byte(`{"prompt_cache_key":"body-cache"}`))
	if gotConversation.SessionID != "conv-123" || gotConversation.Source != SessionSourceHeaderConversationID {
		t.Fatalf("conversation fallback=%+v", gotConversation)
	}

	reqWithBody := httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	gotBody := resolver.ResolveOpenAI(reqWithBody, []byte(`{"prompt_cache_key":" body-cache "}`))
	if gotBody.SessionID != "body-cache" || gotBody.Source != SessionSourcePromptCacheKey {
		t.Fatalf("prompt cache fallback=%+v", gotBody)
	}
}

func TestSessionResolverOpenAIXHeaderFallback(t *testing.T) {
	resolver := SessionResolver{}

	reqWithSession := httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	reqWithSession.Header.Set("x-session-id", "x-session-123")
	gotSession := resolver.ResolveOpenAI(reqWithSession, nil)
	if gotSession.SessionID != "x-session-123" || gotSession.Source != SessionSourceHeaderXSessionID {
		t.Fatalf("x-session-id fallback=%+v", gotSession)
	}

	reqWithConversation := httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	reqWithConversation.Header.Set("x-conversation-id", "x-conv-123")
	gotConversation := resolver.ResolveOpenAI(reqWithConversation, nil)
	if gotConversation.SessionID != "x-conv-123" || gotConversation.Source != SessionSourceHeaderXConversationID {
		t.Fatalf("x-conversation-id fallback=%+v", gotConversation)
	}
}

func TestSessionResolverOpenAIMetadataFallback(t *testing.T) {
	resolver := SessionResolver{}
	req := httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)

	gotSession := resolver.ResolveOpenAI(req, []byte(`{"metadata":{"session_id":"metadata-session"}}`))
	if gotSession.SessionID != "metadata-session" || gotSession.Source != SessionSourceMetadataSessionID {
		t.Fatalf("metadata.session_id fallback=%+v", gotSession)
	}

	gotUserID := resolver.ResolveOpenAI(req, []byte(`{"metadata":{"user_id":"{\"device_id\":\"abc\",\"session_id\":\"metadata-user-session\"}"}}`))
	if gotUserID.SessionID != "metadata-user-session" || gotUserID.Source != SessionSourceMetadataUserID {
		t.Fatalf("metadata.user_id fallback=%+v", gotUserID)
	}
}

func TestSessionResolverOpenAIUnknownWhenNoExplicitSignal(t *testing.T) {
	resolver := SessionResolver{}
	req := httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)

	got := resolver.ResolveOpenAI(req, []byte(`{"model":"gpt-5"}`))

	if got.SessionID != "" {
		t.Fatalf("SessionID=%q, want empty", got.SessionID)
	}
	if got.Source != SessionSourceUnknown {
		t.Fatalf("Source=%q, want %q", got.Source, SessionSourceUnknown)
	}
}

func TestSessionResolverClaudeMetadataUserID(t *testing.T) {
	resolver := SessionResolver{}
	body := []byte(`{
		"model": "claude-sonnet-4-5",
		"metadata": {
			"user_id": "{\"device_id\":\"abc\",\"session_id\":\"claude-session\"}"
		}
	}`)

	got := resolver.ResolveClaude(body)

	if got.SessionID != "claude-session" {
		t.Fatalf("SessionID=%q, want claude-session", got.SessionID)
	}
	if got.Source != SessionSourceMetadataUserID {
		t.Fatalf("Source=%q, want %q", got.Source, SessionSourceMetadataUserID)
	}
	if got.ClientSessionID != "claude-session" {
		t.Fatalf("ClientSessionID=%q, want claude-session", got.ClientSessionID)
	}
}

func TestSessionResolverClaudeMetadataSessionID(t *testing.T) {
	resolver := SessionResolver{}
	body := []byte(`{
		"model": "claude-sonnet-4-5",
		"metadata": {
			"session_id": "metadata-session"
		}
	}`)

	got := resolver.ResolveClaude(body)

	if got.SessionID != "metadata-session" {
		t.Fatalf("SessionID=%q, want metadata-session", got.SessionID)
	}
	if got.Source != SessionSourceMetadataSessionID {
		t.Fatalf("Source=%q, want %q", got.Source, SessionSourceMetadataSessionID)
	}
}
