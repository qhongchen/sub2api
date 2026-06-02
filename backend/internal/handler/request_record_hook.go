package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/requestrecord"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const requestRecordCompleteTimeout = 2 * time.Second

const (
	requestRecordHandleContextKey    = "request_record_handle"
	requestRecordCompletedContextKey = "request_record_completed"
)

type requestRecordRecorder interface {
	Start(ctx context.Context, input *requestrecord.StartInput) (*requestrecord.Handle, error)
	Complete(ctx context.Context, input *requestrecord.CompleteInput) error
}

func getRequestID(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if rid, ok := c.Request.Context().Value(ctxkey.RequestID).(string); ok {
		return strings.TrimSpace(rid)
	}
	return strings.TrimSpace(c.Writer.Header().Get("X-Request-ID"))
}

func getClientRequestID(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if rid, ok := c.Request.Context().Value(ctxkey.ClientRequestID).(string); ok {
		return strings.TrimSpace(rid)
	}
	return ""
}

func ensureRequestRecordIDsInContext(c *gin.Context, requestID, clientRequestID string) {
	if c == nil || c.Request == nil {
		return
	}
	ctx := withUsageRecordRequestIDs(c.Request.Context(), requestID, clientRequestID)
	c.Request = c.Request.WithContext(ctx)
}

func startRequestRecord(
	c *gin.Context,
	recorder requestRecordRecorder,
	input *requestrecord.StartInput,
) *requestrecord.Handle {
	if c == nil || c.Request == nil || recorder == nil || input == nil {
		return nil
	}
	if strings.TrimSpace(input.RequestID) == "" {
		input.RequestID = getRequestID(c)
	}
	if strings.TrimSpace(input.ClientRequestID) == "" {
		input.ClientRequestID = getClientRequestID(c)
	}
	ensureRequestRecordIDsInContext(c, input.RequestID, input.ClientRequestID)
	if strings.TrimSpace(input.IPAddress) == "" {
		input.IPAddress = ip.GetClientIP(c)
	}
	if strings.TrimSpace(input.UserAgent) == "" {
		input.UserAgent = c.GetHeader("User-Agent")
	}
	if input.Now.IsZero() {
		input.Now = getRequestStartedAt(c)
	}
	handle, err := recorder.Start(c.Request.Context(), input)
	if err != nil {
		logger.FromContext(c.Request.Context()).Warn("request_record.start_failed", zap.Error(err))
		return nil
	}
	c.Set(requestRecordHandleContextKey, handle)
	c.Set(requestRecordCompletedContextKey, false)
	return handle
}

func startOpenAIRequestRecord(
	c *gin.Context,
	recorder requestRecordRecorder,
	subject middleware2.AuthSubject,
	apiKey *service.APIKey,
	account *service.Account,
	body []byte,
	model string,
	requestedModel string,
	upstreamModel string,
	requestType service.RequestType,
	stream bool,
	upstreamEndpoint string,
) *requestrecord.Handle {
	return startProviderRequestRecord(c, recorder, subject, apiKey, account, requestrecord.SessionResolver{}.ResolveOpenAI(c.Request, body), model, requestedModel, upstreamModel, requestType, stream, upstreamEndpoint)
}

func startClaudeRequestRecord(
	c *gin.Context,
	recorder requestRecordRecorder,
	subject middleware2.AuthSubject,
	apiKey *service.APIKey,
	account *service.Account,
	body []byte,
	model string,
	requestedModel string,
	upstreamModel string,
	requestType service.RequestType,
	stream bool,
	upstreamEndpoint string,
) *requestrecord.Handle {
	return startProviderRequestRecord(c, recorder, subject, apiKey, account, requestrecord.SessionResolver{}.ResolveClaude(body), model, requestedModel, upstreamModel, requestType, stream, upstreamEndpoint)
}

func startProviderRequestRecord(
	c *gin.Context,
	recorder requestRecordRecorder,
	subject middleware2.AuthSubject,
	apiKey *service.APIKey,
	account *service.Account,
	session requestrecord.SessionIdentity,
	model string,
	requestedModel string,
	upstreamModel string,
	requestType service.RequestType,
	stream bool,
	upstreamEndpoint string,
) *requestrecord.Handle {
	var (
		userID    = subject.UserID
		apiKeyID  *int64
		accountID *int64
		groupID   *int64
		platform  string
	)
	if apiKey != nil {
		apiKeyID = &apiKey.ID
		groupID = apiKey.GroupID
	}
	if account != nil {
		accountID = &account.ID
		platform = account.Platform
		if strings.TrimSpace(upstreamEndpoint) == "" {
			upstreamEndpoint = GetUpstreamEndpoint(c, account.Platform)
		}
	}
	return startRequestRecord(c, recorder, &requestrecord.StartInput{
		RequestID:        getRequestID(c),
		ClientRequestID:  getClientRequestID(c),
		UserID:           &userID,
		APIKeyID:         apiKeyID,
		AccountID:        accountID,
		GroupID:          groupID,
		Session:          session,
		Platform:         platform,
		Model:            model,
		RequestedModel:   requestedModel,
		UpstreamModel:    upstreamModel,
		RequestType:      requestTypePtr(requestType),
		Stream:           stream,
		InboundEndpoint:  GetInboundEndpoint(c),
		UpstreamEndpoint: upstreamEndpoint,
		IPAddress:        ip.GetClientIP(c),
		UserAgent:        c.GetHeader("User-Agent"),
	})
}

func completeRequestRecord(
	c *gin.Context,
	recorder requestRecordRecorder,
	input *requestrecord.CompleteInput,
) {
	if c == nil || c.Request == nil || recorder == nil || input == nil {
		return
	}
	if strings.TrimSpace(input.RequestID) == "" {
		input.RequestID = getRequestID(c)
	}
	if input.Now.IsZero() {
		input.Now = time.Now()
	}
	if strings.TrimSpace(input.Outcome) == "" {
		input.Outcome = requestrecord.OutcomeUnknown
	}
	ctx, cancel := context.WithTimeout(context.Background(), requestRecordCompleteTimeout)
	defer cancel()
	if err := recorder.Complete(ctx, input); err != nil {
		logger.FromContext(c.Request.Context()).Warn("request_record.complete_failed", zap.Error(err))
		return
	}
	c.Set(requestRecordCompletedContextKey, true)
}

func completePendingRequestRecord(c *gin.Context, recorder requestRecordRecorder) {
	if c == nil || c.Request == nil || recorder == nil {
		return
	}
	if completed, ok := c.Get(requestRecordCompletedContextKey); ok {
		if done, _ := completed.(bool); done {
			return
		}
	}
	rawHandle, ok := c.Get(requestRecordHandleContextKey)
	if !ok {
		return
	}
	handle, ok := rawHandle.(*requestrecord.Handle)
	if !ok || handle == nil || strings.TrimSpace(handle.RequestID) == "" {
		return
	}

	status := c.Writer.Status()
	outcome := requestrecord.OutcomeUnknown
	errorMessage := ""
	if err := c.Request.Context().Err(); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			status = 499
			outcome = requestrecord.OutcomeCancelled
			errorMessage = err.Error()
		}
	} else if status >= http.StatusBadRequest {
		outcome = requestrecord.OutcomeError
		errorMessage = http.StatusText(status)
		if parsedMessage := capturedRequestRecordErrorMessage(c); parsedMessage != "" {
			errorMessage = parsedMessage
		}
	}

	var durationMs *int
	if !handle.StartedAt.IsZero() {
		durationMs = intPtrFromDuration(time.Since(handle.StartedAt))
	}
	completeRequestRecord(c, recorder, &requestrecord.CompleteInput{
		RequestID:    handle.RequestID,
		Outcome:      outcome,
		StatusCode:   &status,
		DurationMs:   durationMs,
		ErrorMessage: errorMessage,
	})
}

func capturedRequestRecordErrorMessage(c *gin.Context) string {
	if c == nil || c.Writer == nil {
		return ""
	}
	w, ok := c.Writer.(*opsCaptureWriter)
	if !ok || w == nil || w.buf.Len() == 0 {
		return ""
	}
	parsed := parseOpsErrorResponse(w.buf.Bytes())
	return strings.TrimSpace(parsed.Message)
}

func withUsageRecordRequestIDs(ctx context.Context, requestID, clientRequestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if v := strings.TrimSpace(requestID); v != "" {
		ctx = context.WithValue(ctx, ctxkey.RequestID, v)
	}
	if v := strings.TrimSpace(clientRequestID); v != "" {
		ctx = context.WithValue(ctx, ctxkey.ClientRequestID, v)
	}
	return ctx
}

func getRequestStartedAt(c *gin.Context) time.Time {
	if c == nil || c.Request == nil {
		return time.Now()
	}
	if startedAt, ok := c.Request.Context().Value(ctxkey.RequestStartedAt).(time.Time); ok && !startedAt.IsZero() {
		return startedAt
	}
	return time.Now()
}

func requestTypePtr(v service.RequestType) *int16 {
	n := int16(v)
	return &n
}

func recordRequestID(handle *requestrecord.Handle, c *gin.Context) string {
	if handle != nil && strings.TrimSpace(handle.RequestID) != "" {
		return handle.RequestID
	}
	return getRequestID(c)
}

func recordDurationMs(handle *requestrecord.Handle) *int {
	if handle == nil || handle.StartedAt.IsZero() {
		return nil
	}
	return intPtrFromDuration(time.Since(handle.StartedAt))
}

func intPtr(v int) *int {
	return &v
}

func statusCodePtr(v int) *int {
	if v < http.StatusBadRequest {
		v = http.StatusInternalServerError
	}
	return &v
}

func intPtrFromInt64(v int64) *int {
	n := int(v)
	return &n
}

func intPtrFromDuration(v time.Duration) *int {
	n := int(v.Milliseconds())
	return &n
}

func openAIFirstTokenPtr(result *service.OpenAIForwardResult) *int {
	if result == nil {
		return nil
	}
	return result.FirstTokenMs
}

func forwardResultFirstTokenMs(result *service.ForwardResult) *int {
	if result == nil {
		return nil
	}
	return result.FirstTokenMs
}

func forwardResultDurationMs(result *service.ForwardResult) *int {
	if result == nil {
		return nil
	}
	return intPtrFromDuration(result.Duration)
}
