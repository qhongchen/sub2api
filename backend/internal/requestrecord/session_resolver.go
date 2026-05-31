package requestrecord

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

type SessionResolver struct{}

func (r SessionResolver) ResolveOpenAI(req *http.Request, body []byte) SessionIdentity {
	if req != nil {
		if v := strings.TrimSpace(req.Header.Get("session_id")); v != "" {
			return SessionIdentity{SessionID: v, Source: SessionSourceHeaderSessionID, ClientSessionID: v}
		}
		if v := strings.TrimSpace(req.Header.Get("conversation_id")); v != "" {
			return SessionIdentity{SessionID: v, Source: SessionSourceHeaderConversationID, ClientSessionID: v}
		}
	}
	if v := strings.TrimSpace(gjson.GetBytes(body, "prompt_cache_key").String()); v != "" {
		return SessionIdentity{SessionID: v, Source: SessionSourcePromptCacheKey, ClientSessionID: v}
	}
	return SessionIdentity{Source: SessionSourceUnknown}
}

func (r SessionResolver) ResolveClaude(body []byte) SessionIdentity {
	metadataUserID := strings.TrimSpace(gjson.GetBytes(body, "metadata.user_id").String())
	if metadataUserID != "" {
		if sessionID := extractSessionIDFromMetadataUserID(metadataUserID); sessionID != "" {
			return SessionIdentity{SessionID: sessionID, Source: SessionSourceMetadataUserID, ClientSessionID: sessionID}
		}
	}
	if v := strings.TrimSpace(gjson.GetBytes(body, "metadata.session_id").String()); v != "" {
		return SessionIdentity{SessionID: v, Source: SessionSourceMetadataSessionID, ClientSessionID: v}
	}
	return SessionIdentity{Source: SessionSourceUnknown}
}

func extractSessionIDFromMetadataUserID(raw string) string {
	if raw == "" {
		return ""
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return ""
	}
	if v, ok := obj["session_id"].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}
