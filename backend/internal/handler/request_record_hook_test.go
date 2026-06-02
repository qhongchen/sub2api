package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/requestrecord"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type requestRecordRecorderStub struct {
	started      []*requestrecord.StartInput
	completed    []*requestrecord.CompleteInput
	completeErrs []error
}

func (r *requestRecordRecorderStub) Start(ctx context.Context, input *requestrecord.StartInput) (*requestrecord.Handle, error) {
	r.started = append(r.started, input)
	return &requestrecord.Handle{
		RequestID: input.RequestID,
		StartedAt: time.Now().Add(-1500 * time.Millisecond),
	}, nil
}

func (r *requestRecordRecorderStub) Complete(ctx context.Context, input *requestrecord.CompleteInput) error {
	copyInput := *input
	r.completed = append(r.completed, &copyInput)
	if len(r.completeErrs) > 0 {
		err := r.completeErrs[0]
		r.completeErrs = r.completeErrs[1:]
		return err
	}
	return nil
}

func newRequestRecordHookTestContext(t *testing.T) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	return c
}

func TestCompletePendingRequestRecordSkipsExplicitComplete(t *testing.T) {
	c := newRequestRecordHookTestContext(t)
	recorder := &requestRecordRecorderStub{}

	handle := startRequestRecord(c, recorder, &requestrecord.StartInput{RequestID: "req-explicit"})
	completeRequestRecord(c, recorder, &requestrecord.CompleteInput{
		RequestID: recordRequestID(handle, c),
		Outcome:   requestrecord.OutcomeSuccess,
	})
	completePendingRequestRecord(c, recorder)

	require.Len(t, recorder.completed, 1)
	require.Equal(t, requestrecord.OutcomeSuccess, recorder.completed[0].Outcome)
}

func TestCompletePendingRequestRecordMarksCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c := newRequestRecordHookTestContext(t)
	c.Request = c.Request.WithContext(ctx)
	recorder := &requestRecordRecorderStub{}

	startRequestRecord(c, recorder, &requestrecord.StartInput{RequestID: "req-cancelled"})
	completePendingRequestRecord(c, recorder)

	require.Len(t, recorder.completed, 1)
	require.Equal(t, requestrecord.OutcomeCancelled, recorder.completed[0].Outcome)
	require.NotNil(t, recorder.completed[0].StatusCode)
	require.Equal(t, 499, *recorder.completed[0].StatusCode)
	require.Contains(t, recorder.completed[0].ErrorMessage, "context canceled")
	require.NotNil(t, recorder.completed[0].DurationMs)
}

func TestCompletePendingRequestRecordMarksErrorStatus(t *testing.T) {
	c := newRequestRecordHookTestContext(t)
	c.Writer.WriteHeader(http.StatusBadGateway)
	recorder := &requestRecordRecorderStub{}

	startRequestRecord(c, recorder, &requestrecord.StartInput{RequestID: "req-error"})
	completePendingRequestRecord(c, recorder)

	require.Len(t, recorder.completed, 1)
	require.Equal(t, requestrecord.OutcomeError, recorder.completed[0].Outcome)
	require.NotNil(t, recorder.completed[0].StatusCode)
	require.Equal(t, http.StatusBadGateway, *recorder.completed[0].StatusCode)
	require.Contains(t, recorder.completed[0].ErrorMessage, "Bad Gateway")
	require.NotNil(t, recorder.completed[0].DurationMs)
}

func TestCompletePendingRequestRecordUsesCapturedErrorBody(t *testing.T) {
	c := newRequestRecordHookTestContext(t)
	w := acquireOpsCaptureWriter(c.Writer)
	defer releaseOpsCaptureWriter(w)
	c.Writer = w
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{
			"message": "upstream overloaded",
		},
	})
	recorder := &requestRecordRecorderStub{}

	startRequestRecord(c, recorder, &requestrecord.StartInput{RequestID: "req-captured-error"})
	completePendingRequestRecord(c, recorder)

	require.Len(t, recorder.completed, 1)
	require.Equal(t, requestrecord.OutcomeError, recorder.completed[0].Outcome)
	require.Equal(t, "upstream overloaded", recorder.completed[0].ErrorMessage)
}

func TestCompleteRequestRecordLeavesPendingFallbackWhenCompleteFails(t *testing.T) {
	c := newRequestRecordHookTestContext(t)
	c.Writer.WriteHeader(http.StatusServiceUnavailable)
	recorder := &requestRecordRecorderStub{
		completeErrs: []error{errors.New("temporary request record write failure")},
	}

	handle := startRequestRecord(c, recorder, &requestrecord.StartInput{RequestID: "req-retry"})
	completeRequestRecord(c, recorder, &requestrecord.CompleteInput{
		RequestID:    recordRequestID(handle, c),
		Outcome:      requestrecord.OutcomeError,
		StatusCode:   intPtr(http.StatusServiceUnavailable),
		ErrorMessage: "upstream unavailable",
	})
	completePendingRequestRecord(c, recorder)

	require.Len(t, recorder.completed, 2)
	require.Equal(t, requestrecord.OutcomeError, recorder.completed[0].Outcome)
	require.Equal(t, requestrecord.OutcomeError, recorder.completed[1].Outcome)
	require.NotNil(t, recorder.completed[1].StatusCode)
	require.Equal(t, http.StatusServiceUnavailable, *recorder.completed[1].StatusCode)
}

func TestGetRequestIDPrefersContextOverUpstreamResponseHeader(t *testing.T) {
	c := newRequestRecordHookTestContext(t)
	ctx := context.WithValue(c.Request.Context(), ctxkey.RequestID, "server-request-id")
	c.Request = c.Request.WithContext(ctx)
	c.Header("x-request-id", "upstream-request-id")

	require.Equal(t, "server-request-id", getRequestID(c))
}

func TestStartRequestRecordStoresStableIDsInRequestContext(t *testing.T) {
	c := newRequestRecordHookTestContext(t)
	recorder := &requestRecordRecorderStub{}

	startRequestRecord(c, recorder, &requestrecord.StartInput{
		RequestID:       " request-from-record ",
		ClientRequestID: " client-from-record ",
	})

	require.Equal(t, "request-from-record", c.Request.Context().Value(ctxkey.RequestID))
	require.Equal(t, "client-from-record", c.Request.Context().Value(ctxkey.ClientRequestID))
}
