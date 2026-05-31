package requestrecord

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryStartInsertsPendingRecord(t *testing.T) {
	db, mock := newRequestRecordSQLMock(t)
	repo := NewRepository(db)
	now := time.Date(2026, 5, 29, 14, 0, 0, 0, time.UTC)
	userID := int64(42)
	apiKeyID := int64(7)
	accountID := int64(9)
	groupID := int64(11)
	requestType := int16(2)

	mock.ExpectQuery(`INSERT INTO request_records`).
		WithArgs(
			now,
			"req-1",
			"creq-1",
			userID,
			apiKeyID,
			accountID,
			groupID,
			"sess-1",
			SessionSourceHeaderSessionID,
			"sess-1",
			"openai",
			"gpt-5.4-mini",
			"gpt-5.4-mini",
			"gpt-5.4",
			requestType,
			true,
			"/openai/v1/responses",
			"/v1/responses",
			"10.0.0.8",
			"codex-cli/1.0",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(int64(100), now, now))

	handle, err := repo.Start(context.Background(), &StartInput{
		Now:              now,
		RequestID:        " req-1 ",
		ClientRequestID:  "creq-1",
		UserID:           &userID,
		APIKeyID:         &apiKeyID,
		AccountID:        &accountID,
		GroupID:          &groupID,
		Session:          SessionIdentity{SessionID: "sess-1", Source: SessionSourceHeaderSessionID, ClientSessionID: "sess-1"},
		Platform:         "openai",
		Model:            "gpt-5.4-mini",
		RequestedModel:   "gpt-5.4-mini",
		UpstreamModel:    "gpt-5.4",
		RequestType:      &requestType,
		Stream:           true,
		InboundEndpoint:  "/openai/v1/responses",
		UpstreamEndpoint: "/v1/responses",
		IPAddress:        "10.0.0.8",
		UserAgent:        "codex-cli/1.0",
	})
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	if handle == nil || handle.ID != 100 || handle.RequestID != "req-1" {
		t.Fatalf("unexpected handle: %+v", handle)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositoryCompleteUpdatesRecord(t *testing.T) {
	db, mock := newRequestRecordSQLMock(t)
	repo := NewRepository(db)
	now := time.Date(2026, 5, 29, 14, 1, 0, 0, time.UTC)
	statusCode := 200
	durationMs := 1234
	firstTokenMs := 120
	inputTokens := 100
	outputTokens := 50
	totalCost := 0.0123
	actualCost := 0.011

	mock.ExpectExec(`UPDATE request_records`).
		WithArgs(
			now,
			now,
			OutcomeSuccess,
			statusCode,
			durationMs,
			firstTokenMs,
			true,
			inputTokens,
			outputTokens,
			totalCost,
			actualCost,
			nil,
			"req-1",
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Complete(context.Background(), &CompleteInput{
		Now:          now,
		RequestID:    "req-1",
		Outcome:      OutcomeSuccess,
		StatusCode:   &statusCode,
		DurationMs:   &durationMs,
		FirstTokenMs: &firstTokenMs,
		Billable:     true,
		InputTokens:  &inputTokens,
		OutputTokens: &outputTokens,
		TotalCost:    &totalCost,
		ActualCost:   &actualCost,
	})
	if err != nil {
		t.Fatalf("Complete() error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositoryListFiltersBySessionAndOutcome(t *testing.T) {
	db, mock := newRequestRecordSQLMock(t)
	repo := NewRepository(db)
	start := time.Date(2026, 5, 29, 13, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	createdAt := start.Add(10 * time.Minute)
	updatedAt := createdAt.Add(time.Second)
	statusCode := 503

	mock.ExpectQuery(`SELECT COUNT\(1\)\s+FROM request_records rr`).
		WithArgs(start, end, "sess-1", OutcomeError).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	rows := sqlmock.NewRows(requestRecordListColumns()).
		AddRow(
			int64(10),
			createdAt,
			updatedAt,
			nil,
			"req-1",
			"creq-1",
			int64(42),
			int64(7),
			int64(9),
			int64(11),
			"sess-1",
			SessionSourceHeaderSessionID,
			"sess-1",
			"openai",
			"gpt-5.4-mini",
			"gpt-5.4-mini",
			"gpt-5.4",
			int64(2),
			true,
			"/openai/v1/responses",
			"/v1/responses",
			OutcomeError,
			statusCode,
			int64(1234),
			int64(120),
			false,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			0.006,
			nil,
			"",
			"",
			"",
			"",
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			"",
			nil,
			nil,
			nil,
			"",
			"",
			"",
			"",
			"10.0.0.8",
			"codex-cli/1.0",
			"upstream unavailable",
			"",
			"user@example.com",
			"tester",
			"diagnostic-key",
			"openai-account",
			"default-group",
		)
	mock.ExpectQuery(`SELECT\s+rr\.id,`).
		WithArgs(start, end, "sess-1", OutcomeError, 50, 0).
		WillReturnRows(rows)

	items, total, err := repo.List(context.Background(), &Filter{
		StartTime: &start,
		EndTime:   &end,
		SessionID: "sess-1",
		Outcome:   OutcomeError,
		Page:      1,
		PageSize:  50,
	})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("total=%d len=%d, want 1/1", total, len(items))
	}
	if items[0].SessionID != "sess-1" || items[0].Outcome != OutcomeError || items[0].StatusCode == nil || *items[0].StatusCode != 503 {
		t.Fatalf("unexpected item: %+v", items[0])
	}
	if items[0].ActualCost != nil {
		t.Fatalf("non-billable error should not expose cost: %+v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositoryListClearsBillingFieldsForNonBillableRecords(t *testing.T) {
	db, mock := newRequestRecordSQLMock(t)
	repo := NewRepository(db)
	start := time.Date(2026, 5, 29, 13, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	createdAt := start.Add(10 * time.Minute)
	updatedAt := createdAt.Add(time.Second)

	mock.ExpectQuery(`SELECT COUNT\(1\)\s+FROM request_records rr`).
		WithArgs(start, end).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	rows := sqlmock.NewRows(requestRecordListColumns()).
		AddRow(
			int64(11),
			createdAt,
			updatedAt,
			updatedAt,
			"req-non-billable",
			"creq-non-billable",
			int64(42),
			int64(7),
			int64(9),
			int64(11),
			"sess-1",
			SessionSourceHeaderSessionID,
			"sess-1",
			"openai",
			"gpt-5.4-mini",
			"gpt-5.4-mini",
			"gpt-5.4",
			int64(1),
			false,
			"/v1/responses",
			"/v1/responses",
			OutcomeNonBillable,
			int64(200),
			int64(1234),
			int64(120),
			false,
			int64(100),
			int64(50),
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			"",
			"",
			"",
			"",
			nil,
			0.01,
			0.02,
			0.003,
			0.004,
			0.037,
			0.041,
			1.1,
			int64(0),
			"token",
			1.2,
			0.05,
			false,
			"",
			"",
			"priority",
			"xhigh",
			"10.0.0.8",
			"codex-cli/1.0",
			"",
			"",
			"user@example.com",
			"tester",
			"diagnostic-key",
			"openai-account",
			"default-group",
		)
	mock.ExpectQuery(`SELECT\s+rr\.id,`).
		WithArgs(start, end, 50, 0).
		WillReturnRows(rows)

	items, total, err := repo.List(context.Background(), &Filter{
		StartTime: &start,
		EndTime:   &end,
		Page:      1,
		PageSize:  50,
	})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("total=%d len=%d, want 1/1", total, len(items))
	}
	got := items[0]
	if got.Billable {
		t.Fatalf("record should be non-billable: %+v", got)
	}
	if got.ImageOutputCost != nil || got.TotalCost != nil || got.ActualCost != nil || got.InputCost != nil || got.OutputCost != nil || got.AccountStatsCost != nil || got.BillingMode != "" {
		t.Fatalf("non-billable record should not expose billing fields: %+v", got)
	}
	if got.InputTokens == nil || *got.InputTokens != 100 || got.OutputTokens == nil || *got.OutputTokens != 50 {
		t.Fatalf("non-billable record should keep token facts: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositoryListParsesUpstreamAttempts(t *testing.T) {
	db, mock := newRequestRecordSQLMock(t)
	repo := NewRepository(db)
	start := time.Date(2026, 5, 29, 13, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	createdAt := start.Add(10 * time.Minute)
	updatedAt := createdAt.Add(time.Second)
	statusCode := 200
	upstreamAttempts := `[{"at_unix_ms":1780000000000,"platform":"openai","account_id":20,"account_name":"L站福利1","upstream_status_code":401,"upstream_request_id":"up-1","upstream_url":"https://api.openai.com/v1/responses","kind":"http_error","message":"Invalid API key","detail":"token unavailable"}]`

	mock.ExpectQuery(`SELECT COUNT\(1\)\s+FROM request_records rr`).
		WithArgs(start, end).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	rows := sqlmock.NewRows(requestRecordListColumns()).
		AddRow(
			int64(12),
			createdAt,
			updatedAt,
			updatedAt,
			"req-success-with-recovered-upstream",
			"creq-success-with-recovered-upstream",
			int64(42),
			int64(7),
			int64(9),
			int64(11),
			"sess-1",
			SessionSourceHeaderSessionID,
			"sess-1",
			"openai",
			"gpt-5.4-mini",
			"gpt-5.4-mini",
			"gpt-5.4",
			int64(1),
			false,
			"/v1/responses",
			"/v1/responses",
			OutcomeSuccess,
			statusCode,
			int64(1234),
			int64(120),
			true,
			int64(100),
			int64(50),
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			"",
			"",
			"",
			"",
			nil,
			0.01,
			0.02,
			0.003,
			0.004,
			0.037,
			0.041,
			1.1,
			int64(0),
			"token",
			1.2,
			0.05,
			false,
			"",
			"",
			"priority",
			"xhigh",
			"10.0.0.8",
			"codex-cli/1.0",
			"",
			upstreamAttempts,
			"user@example.com",
			"tester",
			"diagnostic-key",
			"openai-account",
			"default-group",
		)
	mock.ExpectQuery(`SELECT\s+rr\.id,`).
		WithArgs(start, end, 50, 0).
		WillReturnRows(rows)

	items, total, err := repo.List(context.Background(), &Filter{
		StartTime: &start,
		EndTime:   &end,
		Page:      1,
		PageSize:  50,
	})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("total=%d len=%d, want 1/1", total, len(items))
	}
	got := items[0]
	if got.ErrorMessage != "" {
		t.Fatalf("successful request should not expose recovered upstream error as final error: %+v", got)
	}
	if len(got.UpstreamAttempts) != 1 {
		t.Fatalf("expected one upstream attempt, got %+v", got.UpstreamAttempts)
	}
	attempt := got.UpstreamAttempts[0]
	if attempt.UpstreamStatusCode != 401 || attempt.AccountID != 20 || attempt.Message != "Invalid API key" || attempt.UpstreamURL != "https://api.openai.com/v1/responses" {
		t.Fatalf("unexpected upstream attempt: %+v", attempt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRepositorySummaryCountsRequestOutcomes(t *testing.T) {
	db, mock := newRequestRecordSQLMock(t)
	repo := NewRepository(db)
	start := time.Date(2026, 5, 29, 13, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	mock.ExpectQuery(`SELECT\s+COUNT\(1\) AS total_requests,`).
		WithArgs(start, end).
		WillReturnRows(sqlmock.NewRows([]string{"total_requests", "success_requests", "error_requests"}).
			AddRow(int64(10), int64(7), int64(3)))

	summary, err := repo.Summary(context.Background(), &Filter{
		StartTime: &start,
		EndTime:   &end,
	})
	if err != nil {
		t.Fatalf("Summary() error: %v", err)
	}
	if summary.TotalRequests != 10 || summary.SuccessRequests != 7 || summary.ErrorRequests != 3 || summary.ErrorRate != 0.3 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRequestRecordListSelectJoinsOpsErrorsByClientRequestID(t *testing.T) {
	query := requestRecordListSelect()
	if !strings.Contains(query, "LEFT JOIN LATERAL") {
		t.Fatalf("request records should use a lateral ops_error_logs join to avoid duplicate rows: %s", query)
	}
	if !strings.Contains(query, "oe.client_request_id = rr.client_request_id") {
		t.Fatalf("request records should correlate ops_error_logs by client_request_id fallback: %s", query)
	}
}

func TestRequestRecordListSelectFallsBackToOpsErrorsByRequestFacts(t *testing.T) {
	query := requestRecordListSelect()
	for _, fragment := range []string{
		"oe.created_at >= rr.created_at",
		"oe.created_at < COALESCE(rr.completed_at, rr.updated_at, rr.created_at) + INTERVAL '5 minutes'",
		"(rr.api_key_id IS NULL OR oe.api_key_id = rr.api_key_id)",
		"(rr.account_id IS NULL OR oe.account_id = rr.account_id)",
		"rr.request_type IS NULL",
		"oe.request_path = rr.inbound_endpoint",
		"oe.upstream_status_code = rr.status_code",
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("request records should correlate ops_error_logs by request facts, missing %q in: %s", fragment, query)
		}
	}
}

func TestRequestRecordListSelectDoesNotTreatRecoveredUpstreamErrorAsFinalSuccessError(t *testing.T) {
	query := requestRecordListSelect()
	for _, fragment := range []string{
		"WHEN rr.outcome IN ('success', 'non_billable')",
		"THEN COALESCE(NULLIF(rr.error_message, ''), '')",
		"ELSE COALESCE(NULLIF(rr.error_message, ''), o.error_message, '')",
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("request records should keep final success error_message separate from recovered upstream errors, missing %q in: %s", fragment, query)
		}
	}
}

func TestRequestRecordListSelectExposesUpstreamAttemptsSeparately(t *testing.T) {
	query := requestRecordListSelect()
	if !strings.Contains(query, "COALESCE(o.upstream_errors::text, '')") {
		t.Fatalf("request records should expose upstream error attempts separately from final error_message: %s", query)
	}
}

func TestRequestRecordListSelectJoinsUsageLogsByStableBillingRequestIDs(t *testing.T) {
	query := requestRecordListSelect()
	if !strings.Contains(query, "ul.request_id = 'local:' || rr.request_id") {
		t.Fatalf("request records should correlate usage_logs by local request_id fallback: %s", query)
	}
	if !strings.Contains(query, "ul.request_id = 'client:' || rr.client_request_id") {
		t.Fatalf("request records should correlate usage_logs by client_request_id fallback: %s", query)
	}
}

func TestRequestRecordListSelectFallsBackToUsageLogsByRequestFacts(t *testing.T) {
	query := requestRecordListSelect()
	for _, fragment := range []string{
		"ul.created_at >= rr.created_at",
		"ul.created_at < COALESCE(rr.completed_at, rr.updated_at, rr.created_at) + INTERVAL '5 minutes'",
		"(rr.user_id IS NULL OR ul.user_id = rr.user_id)",
		"(rr.account_id IS NULL OR ul.account_id = rr.account_id)",
		"(rr.group_id IS NULL OR ul.group_id = rr.group_id)",
		"ul.stream = rr.stream",
		"rr.request_type IS NULL",
		"ul.inbound_endpoint = rr.inbound_endpoint",
		"ul.upstream_endpoint = rr.upstream_endpoint",
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("request records should correlate usage_logs by request facts, missing %q in: %s", fragment, query)
		}
	}
}

func newRequestRecordSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}
