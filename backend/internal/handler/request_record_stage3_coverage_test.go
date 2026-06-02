package handler

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRequestRecordStage3HooksCoverGatewayEntrypoints(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		fn        string
		wantAll   []string
		wantStart []string
	}{
		{
			name:      "openai responses",
			file:      "openai_gateway_handler.go",
			fn:        "func (h *OpenAIGatewayHandler) Responses(c *gin.Context)",
			wantAll:   []string{"defer completePendingRequestRecord", "completeRequestRecord("},
			wantStart: []string{"startRequestRecord(", "startOpenAIRequestRecord("},
		},
		{
			name:      "openai messages",
			file:      "openai_gateway_handler.go",
			fn:        "func (h *OpenAIGatewayHandler) Messages(c *gin.Context)",
			wantAll:   []string{"defer completePendingRequestRecord", "completeRequestRecord("},
			wantStart: []string{"startRequestRecord(", "startClaudeRequestRecord("},
		},
		{
			name:      "openai chat completions",
			file:      "openai_chat_completions.go",
			fn:        "func (h *OpenAIGatewayHandler) ChatCompletions(c *gin.Context)",
			wantAll:   []string{"defer completePendingRequestRecord", "completeRequestRecord("},
			wantStart: []string{"startRequestRecord(", "startOpenAIRequestRecord("},
		},
		{
			name:      "openai images",
			file:      "openai_images.go",
			fn:        "func (h *OpenAIGatewayHandler) Images(c *gin.Context)",
			wantAll:   []string{"defer completePendingRequestRecord", "completeRequestRecord("},
			wantStart: []string{"startRequestRecord(", "startOpenAIRequestRecord("},
		},
		{
			name:      "gateway messages",
			file:      "gateway_handler.go",
			fn:        "func (h *GatewayHandler) Messages(c *gin.Context)",
			wantAll:   []string{"defer completePendingRequestRecord", "completeRequestRecord("},
			wantStart: []string{"startRequestRecord(", "startClaudeRequestRecord("},
		},
		{
			name:      "gateway responses",
			file:      "gateway_handler_responses.go",
			fn:        "func (h *GatewayHandler) Responses(c *gin.Context)",
			wantAll:   []string{"defer completePendingRequestRecord", "completeRequestRecord("},
			wantStart: []string{"startRequestRecord(", "startOpenAIRequestRecord("},
		},
		{
			name:      "gateway chat completions",
			file:      "gateway_handler_chat_completions.go",
			fn:        "func (h *GatewayHandler) ChatCompletions(c *gin.Context)",
			wantAll:   []string{"defer completePendingRequestRecord", "completeRequestRecord("},
			wantStart: []string{"startRequestRecord(", "startOpenAIRequestRecord("},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := readHandlerSource(t, tt.file)
			fnBody := extractFunctionBody(t, body, tt.fn)
			for _, want := range tt.wantAll {
				if !strings.Contains(fnBody, want) {
					t.Fatalf("%s must contain %q", tt.fn, want)
				}
			}
			if !containsAny(fnBody, tt.wantStart) {
				t.Fatalf("%s must contain one of %v", tt.fn, tt.wantStart)
			}
		})
	}
}

func TestRequestRecordStage3UsageTasksPreserveStableRequestIDs(t *testing.T) {
	tests := []struct {
		name string
		file string
		fn   string
	}{
		{
			name: "openai responses",
			file: "openai_gateway_handler.go",
			fn:   "func (h *OpenAIGatewayHandler) Responses(c *gin.Context)",
		},
		{
			name: "openai messages",
			file: "openai_gateway_handler.go",
			fn:   "func (h *OpenAIGatewayHandler) Messages(c *gin.Context)",
		},
		{
			name: "openai chat completions",
			file: "openai_chat_completions.go",
			fn:   "func (h *OpenAIGatewayHandler) ChatCompletions(c *gin.Context)",
		},
		{
			name: "openai images",
			file: "openai_images.go",
			fn:   "func (h *OpenAIGatewayHandler) Images(c *gin.Context)",
		},
		{
			name: "gateway messages",
			file: "gateway_handler.go",
			fn:   "func (h *GatewayHandler) Messages(c *gin.Context)",
		},
		{
			name: "gateway responses",
			file: "gateway_handler_responses.go",
			fn:   "func (h *GatewayHandler) Responses(c *gin.Context)",
		},
		{
			name: "gateway chat completions",
			file: "gateway_handler_chat_completions.go",
			fn:   "func (h *GatewayHandler) ChatCompletions(c *gin.Context)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := readHandlerSource(t, tt.file)
			fnBody := extractFunctionBody(t, body, tt.fn)
			if !strings.Contains(fnBody, "c.Request.Context()") && !strings.Contains(fnBody, "submitOpenAIUsageRecordTask(ctx,") {
				t.Fatalf("%s must submit usage tasks with a parent request context", tt.fn)
			}
		})
	}
}

func TestGatewayMessagesGeminiBranchWritesRequestRecord(t *testing.T) {
	body := readHandlerSource(t, "gateway_handler.go")
	fnBody := extractFunctionBody(t, body, "func (h *GatewayHandler) Messages(c *gin.Context)")
	branch := extractBlockAfter(t, fnBody, "if platform == service.PlatformGemini {\n\t\tfs := NewFailoverState")

	writerPos := strings.Index(branch, "writerSizeBeforeForward := c.Writer.Size()")
	forwardPos := strings.Index(branch, "h.geminiCompatService.Forward(")
	if forwardPos < 0 {
		t.Fatalf("gemini branch forward call not found")
	}
	if writerPos < 0 || writerPos > forwardPos {
		t.Fatalf("gemini branch writer-size guard must be before forward call")
	}
	forwardPath := branch[writerPos:forwardPos]
	if !strings.Contains(forwardPath, "recordHandle := startClaudeRequestRecord(") {
		t.Fatalf("gemini branch must start request record before forwarding upstream")
	}
	if !strings.Contains(branch, "completeRequestRecord(c, h.requestRecordService, &requestrecord.CompleteInput{") {
		t.Fatalf("gemini branch must complete request record on terminal paths")
	}
	if !strings.Contains(branch, "Outcome:      requestrecord.OutcomeSuccess") {
		t.Fatalf("gemini branch must mark successful upstream requests as success")
	}
	if !strings.Contains(branch, "Outcome:      requestrecord.OutcomeError") {
		t.Fatalf("gemini branch must mark terminal upstream errors as error")
	}
}

func TestGatewayMessagesMockInterceptWritesNonBillableRequestRecord(t *testing.T) {
	body := readHandlerSource(t, "gateway_handler.go")
	fnBody := extractFunctionBody(t, body, "func (h *GatewayHandler) Messages(c *gin.Context)")
	branches := extractAllBlocksAfter(t, fnBody, "if interceptType != InterceptTypeNone")
	if len(branches) < 2 {
		t.Fatalf("expected gemini and standard mock intercept branches, got %d", len(branches))
	}

	for idx, branch := range branches {
		if !strings.Contains(branch, "recordHandle := startClaudeRequestRecord(") {
			t.Fatalf("mock intercept branch %d must start request record after account selection", idx)
		}
		if !strings.Contains(branch, "Outcome:") || !strings.Contains(branch, "requestrecord.OutcomeNonBillable") {
			t.Fatalf("mock intercept branch %d must complete request record as non_billable", idx)
		}
		if !strings.Contains(branch, "Billable:") || !strings.Contains(branch, "false") {
			t.Fatalf("mock intercept branch %d must explicitly remain non-billable", idx)
		}
	}
}

func readHandlerSource(t *testing.T, name string) string {
	t.Helper()
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	body, err := os.ReadFile(filepath.Join(filepath.Dir(current), name))
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func extractFunctionBody(t *testing.T, source, signature string) string {
	t.Helper()
	start := strings.Index(source, signature)
	if start < 0 {
		t.Fatalf("signature not found: %s", signature)
	}
	brace := strings.Index(source[start:], "{")
	if brace < 0 {
		t.Fatalf("function body not found: %s", signature)
	}
	pos := start + brace
	depth := 0
	for i := pos; i < len(source); i++ {
		switch source[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return source[pos : i+1]
			}
		}
	}
	t.Fatalf("function body not closed: %s", signature)
	return ""
}

func extractBlockAfter(t *testing.T, source, marker string) string {
	t.Helper()
	start := strings.Index(source, marker)
	if start < 0 {
		t.Fatalf("marker not found: %s", marker)
	}
	braceRel := strings.Index(source[start:], "{")
	if braceRel < 0 {
		t.Fatalf("block not found after marker: %s", marker)
	}
	pos := start + braceRel
	depth := 0
	for i := pos; i < len(source); i++ {
		switch source[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return source[pos : i+1]
			}
		}
	}
	t.Fatalf("block not closed after marker: %s", marker)
	return ""
}

func extractAllBlocksAfter(t *testing.T, source, marker string) []string {
	t.Helper()
	var out []string
	offset := 0
	for {
		idx := strings.Index(source[offset:], marker)
		if idx < 0 {
			break
		}
		start := offset + idx
		out = append(out, extractBlockAfter(t, source[start:], marker))
		offset = start + len(marker)
	}
	return out
}

func containsAny(body string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(body, needle) {
			return true
		}
	}
	return false
}
