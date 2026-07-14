package requestrecord

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func (r *repository) Start(ctx context.Context, input *StartInput) (*Handle, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil requestrecord repository")
	}
	if input == nil {
		return nil, fmt.Errorf("nil start input")
	}
	requestID := strings.TrimSpace(input.RequestID)
	if requestID == "" {
		return nil, fmt.Errorf("request_id is required")
	}
	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}

	var (
		id        int64
		createdAt time.Time
		updatedAt time.Time
	)
	err := r.db.QueryRowContext(ctx, `
INSERT INTO request_records (
  created_at,
  updated_at,
  request_id,
  client_request_id,
  user_id,
  api_key_id,
  account_id,
  group_id,
  session_id,
  session_source,
  client_session_id,
  platform,
  model,
  requested_model,
  upstream_model,
  request_type,
  stream,
  inbound_endpoint,
  upstream_endpoint,
  ip_address,
  user_agent,
  outcome,
  billable
) VALUES (
  $1,$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,'pending',false
)
ON CONFLICT (request_id) DO UPDATE SET
  updated_at = EXCLUDED.updated_at,
  client_request_id = COALESCE(EXCLUDED.client_request_id, request_records.client_request_id),
  user_id = COALESCE(EXCLUDED.user_id, request_records.user_id),
  api_key_id = COALESCE(EXCLUDED.api_key_id, request_records.api_key_id),
  account_id = COALESCE(EXCLUDED.account_id, request_records.account_id),
  group_id = COALESCE(EXCLUDED.group_id, request_records.group_id),
  session_id = COALESCE(EXCLUDED.session_id, request_records.session_id),
  session_source = COALESCE(EXCLUDED.session_source, request_records.session_source),
  client_session_id = COALESCE(EXCLUDED.client_session_id, request_records.client_session_id),
  platform = COALESCE(EXCLUDED.platform, request_records.platform),
  model = COALESCE(EXCLUDED.model, request_records.model),
  requested_model = COALESCE(EXCLUDED.requested_model, request_records.requested_model),
  upstream_model = COALESCE(EXCLUDED.upstream_model, request_records.upstream_model),
  request_type = COALESCE(EXCLUDED.request_type, request_records.request_type),
  stream = EXCLUDED.stream,
  inbound_endpoint = COALESCE(EXCLUDED.inbound_endpoint, request_records.inbound_endpoint),
  upstream_endpoint = COALESCE(EXCLUDED.upstream_endpoint, request_records.upstream_endpoint),
  ip_address = COALESCE(EXCLUDED.ip_address, request_records.ip_address),
  user_agent = COALESCE(EXCLUDED.user_agent, request_records.user_agent)
RETURNING id, created_at, updated_at
`,
		now,
		requestID,
		nullString(input.ClientRequestID),
		nullInt64(input.UserID),
		nullInt64(input.APIKeyID),
		nullInt64(input.AccountID),
		nullInt64(input.GroupID),
		nullString(input.Session.SessionID),
		nullString(input.Session.Source),
		nullString(input.Session.ClientSessionID),
		nullString(input.Platform),
		nullString(input.Model),
		nullString(input.RequestedModel),
		nullString(input.UpstreamModel),
		nullInt16(input.RequestType),
		input.Stream,
		nullString(input.InboundEndpoint),
		nullString(input.UpstreamEndpoint),
		nullString(input.IPAddress),
		nullString(input.UserAgent),
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	_ = updatedAt
	return &Handle{ID: id, RequestID: requestID, StartedAt: createdAt}, nil
}

func (r *repository) Complete(ctx context.Context, input *CompleteInput) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("nil requestrecord repository")
	}
	if input == nil {
		return fmt.Errorf("nil complete input")
	}
	requestID := strings.TrimSpace(input.RequestID)
	if requestID == "" {
		return fmt.Errorf("request_id is required")
	}
	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	outcome := strings.TrimSpace(input.Outcome)
	if outcome == "" {
		outcome = OutcomeUnknown
	}

	_, err := r.db.ExecContext(ctx, `
UPDATE request_records
SET
  completed_at = $1,
  updated_at = $2,
  outcome = $3,
  status_code = $4,
  duration_ms = $5,
  first_token_ms = $6,
  billable = $7,
  input_tokens = $8,
  output_tokens = $9,
  total_cost = $10,
  actual_cost = $11,
  error_message = $12,
  upstream_endpoint = COALESCE(NULLIF($13, ''), upstream_endpoint)
WHERE request_id = $14
`,
		now,
		now,
		outcome,
		nullInt(input.StatusCode),
		nullInt(input.DurationMs),
		nullInt(input.FirstTokenMs),
		input.Billable,
		nullInt(input.InputTokens),
		nullInt(input.OutputTokens),
		nullFloat64(input.TotalCost),
		nullFloat64(input.ActualCost),
		nullString(input.ErrorMessage),
		strings.TrimSpace(input.UpstreamEndpoint),
		requestID,
	)
	return err
}

func (r *repository) List(ctx context.Context, filter *Filter) ([]*Record, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("nil requestrecord repository")
	}
	page, pageSize, startTime, endTime := filter.Normalize()
	offset := (page - 1) * pageSize
	where, args := buildWhere(filter, startTime, endTime)

	countQuery := "SELECT COUNT(1) FROM request_records rr " + where
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sort := "ORDER BY rr.created_at DESC, rr.id DESC"
	if filter != nil && strings.TrimSpace(filter.Sort) == "duration_desc" {
		sort = "ORDER BY rr.duration_ms DESC NULLS LAST, rr.created_at DESC, rr.id DESC"
	}

	query := requestRecordListSelect() + `
` + where + `
` + sort + `
LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)

	listArgs := append(append([]any{}, args...), pageSize, offset)
	rows, err := r.db.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]*Record, 0, pageSize)
	for rows.Next() {
		item, err := scanRecord(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *repository) Summary(ctx context.Context, filter *Filter) (*Summary, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil requestrecord repository")
	}
	_, _, startTime, endTime := filter.Normalize()
	where, args := buildWhere(filter, startTime, endTime)
	query := `
SELECT
  COUNT(1) AS total_requests,
  COUNT(1) FILTER (WHERE rr.outcome IN ('success', 'non_billable')) AS success_requests,
  COUNT(1) FILTER (WHERE rr.outcome IN ('error', 'cancelled')) AS error_requests
FROM request_records rr
` + where

	out := &Summary{}
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&out.TotalRequests,
		&out.SuccessRequests,
		&out.ErrorRequests,
	); err != nil {
		return nil, err
	}
	if out.TotalRequests > 0 {
		out.ErrorRate = float64(out.ErrorRequests) / float64(out.TotalRequests)
	}
	return out, nil
}

func requestRecordListSelect() string {
	return `
SELECT
  rr.id,
  rr.created_at,
  rr.updated_at,
  rr.completed_at,
  rr.request_id,
  COALESCE(rr.client_request_id, ''),
  rr.user_id,
  rr.api_key_id,
  rr.account_id,
  rr.group_id,
  COALESCE(rr.session_id, ''),
  COALESCE(rr.session_source, ''),
  COALESCE(rr.client_session_id, ''),
  COALESCE(rr.platform, ''),
	  COALESCE(rr.model, ''),
	  COALESCE(NULLIF(ul.requested_model, ''), rr.requested_model, ''),
	  COALESCE(NULLIF(ul.upstream_model, ''), rr.upstream_model, ''),
	  COALESCE(NULLIF(ul.upstream_model, ''), NULLIF(rr.upstream_model, ''), NULLIF(ul.model, ''), NULLIF(rr.model, ''), ''),
	  rr.request_type,
  rr.stream,
  COALESCE(rr.inbound_endpoint, ul.inbound_endpoint, o.inbound_endpoint, ''),
  COALESCE(rr.upstream_endpoint, ul.upstream_endpoint, o.upstream_endpoint, ''),
  rr.outcome,
  rr.status_code,
  rr.duration_ms,
  rr.first_token_ms,
  rr.billable,
  COALESCE(ul.input_tokens, rr.input_tokens),
  COALESCE(ul.output_tokens, rr.output_tokens),
  ul.cache_creation_tokens,
  ul.cache_read_tokens,
  ul.cache_creation_5m_tokens,
  ul.cache_creation_1h_tokens,
  ul.image_output_tokens,
  ul.image_output_cost,
  ul.image_count,
  COALESCE(ul.image_size, ''),
  COALESCE(ul.image_input_size, ''),
  COALESCE(ul.image_output_size, ''),
  COALESCE(ul.image_size_source, ''),
  ul.image_size_breakdown,
  ul.input_cost,
  ul.output_cost,
  ul.cache_creation_cost,
  ul.cache_read_cost,
  COALESCE(ul.total_cost, rr.total_cost),
  COALESCE(ul.actual_cost, rr.actual_cost),
  ul.rate_multiplier,
  ul.billing_type,
  COALESCE(ul.billing_mode, ''),
  ul.account_rate_multiplier,
  ul.account_stats_cost,
  ul.cache_ttl_overridden,
  COALESCE(ul.model_mapping_chain, ''),
  COALESCE(ul.billing_tier, ''),
  COALESCE(ul.service_tier, ''),
  COALESCE(ul.reasoning_effort, ''),
  COALESCE(rr.ip_address, CASE WHEN o.client_ip IS NULL THEN NULL ELSE o.client_ip::TEXT END, ul.ip_address, ''),
  COALESCE(rr.user_agent, ul.user_agent, o.user_agent, ''),
  CASE
    WHEN rr.outcome IN ('success', 'non_billable') OR COALESCE(rr.status_code, 0) BETWEEN 200 AND 299
    THEN COALESCE(NULLIF(rr.error_message, ''), '')
    ELSE COALESCE(NULLIF(rr.error_message, ''), o.error_message, '')
  END,
  COALESCE(o.upstream_errors::text, ''),
  COALESCE(u.email, ''),
  COALESCE(u.username, ''),
  COALESCE(ak.name, ''),
  COALESCE(a.name, ''),
  COALESCE(g.name, '')
FROM request_records rr
LEFT JOIN LATERAL (
  SELECT ul.*
  FROM usage_logs ul
  WHERE (rr.api_key_id IS NULL OR ul.api_key_id = rr.api_key_id)
    AND (
      ul.request_id = rr.request_id
      OR ul.request_id = 'local:' || rr.request_id
      OR (
        COALESCE(rr.client_request_id, '') <> ''
        AND ul.request_id = 'client:' || rr.client_request_id
      )
      OR (
        rr.completed_at IS NOT NULL
        AND rr.outcome IN ('success', 'non_billable')
        AND ul.created_at >= rr.created_at
        AND ul.created_at < COALESCE(rr.completed_at, rr.updated_at, rr.created_at) + INTERVAL '5 minutes'
        AND (rr.user_id IS NULL OR ul.user_id = rr.user_id)
        AND (rr.account_id IS NULL OR ul.account_id = rr.account_id)
        AND (rr.group_id IS NULL OR ul.group_id = rr.group_id)
        AND ul.stream = rr.stream
        AND (
          rr.request_type IS NULL
          OR ul.request_type = rr.request_type
          OR (ul.request_type = 0 AND rr.request_type IN (1, 2))
        )
        AND (
          COALESCE(NULLIF(rr.model, ''), NULLIF(rr.requested_model, ''), NULLIF(rr.upstream_model, ''), '') = ''
          OR LOWER(COALESCE(ul.model, '')) = LOWER(COALESCE(rr.model, ''))
          OR LOWER(COALESCE(ul.model, '')) = LOWER(COALESCE(rr.requested_model, ''))
          OR LOWER(COALESCE(ul.requested_model, '')) = LOWER(COALESCE(rr.model, ''))
          OR LOWER(COALESCE(ul.requested_model, '')) = LOWER(COALESCE(rr.requested_model, ''))
          OR LOWER(COALESCE(ul.upstream_model, '')) = LOWER(COALESCE(rr.upstream_model, ''))
        )
        AND (
          COALESCE(rr.inbound_endpoint, '') = ''
          OR COALESCE(ul.inbound_endpoint, '') = ''
          OR ul.inbound_endpoint = rr.inbound_endpoint
        )
        AND (
          COALESCE(rr.upstream_endpoint, '') = ''
          OR COALESCE(ul.upstream_endpoint, '') = ''
          OR ul.upstream_endpoint = rr.upstream_endpoint
        )
      )
    )
  ORDER BY
    CASE
      WHEN ul.request_id = rr.request_id THEN 0
      WHEN ul.request_id = 'local:' || rr.request_id THEN 1
      WHEN COALESCE(rr.client_request_id, '') <> '' AND ul.request_id = 'client:' || rr.client_request_id THEN 2
      ELSE 3
    END,
    CASE
      WHEN rr.output_tokens IS NOT NULL AND ul.output_tokens = rr.output_tokens THEN 0
      ELSE 1
    END,
    ABS(EXTRACT(EPOCH FROM (ul.created_at - COALESCE(rr.completed_at, rr.updated_at, rr.created_at)))),
    ul.created_at DESC,
    ul.id DESC
  LIMIT 1
) ul ON true
LEFT JOIN LATERAL (
  SELECT oe.*
  FROM ops_error_logs oe
  WHERE oe.request_id = rr.request_id
    OR (
      COALESCE(rr.client_request_id, '') <> ''
      AND oe.client_request_id = rr.client_request_id
    )
    OR (
      rr.completed_at IS NOT NULL
      AND rr.outcome IN ('error', 'cancelled')
      AND oe.created_at >= rr.created_at
      AND oe.created_at < COALESCE(rr.completed_at, rr.updated_at, rr.created_at) + INTERVAL '5 minutes'
      AND (rr.user_id IS NULL OR oe.user_id = rr.user_id)
      AND (rr.api_key_id IS NULL OR oe.api_key_id = rr.api_key_id)
      AND (rr.account_id IS NULL OR oe.account_id = rr.account_id)
      AND (rr.group_id IS NULL OR oe.group_id = rr.group_id)
      AND oe.stream = rr.stream
      AND (
        rr.request_type IS NULL
        OR oe.request_type = rr.request_type
        OR oe.request_type IS NULL
      )
      AND (
        COALESCE(NULLIF(rr.platform, ''), '') = ''
        OR LOWER(COALESCE(oe.platform, '')) = LOWER(COALESCE(rr.platform, ''))
      )
      AND (
        COALESCE(NULLIF(rr.model, ''), NULLIF(rr.requested_model, ''), NULLIF(rr.upstream_model, ''), '') = ''
        OR LOWER(COALESCE(oe.model, '')) = LOWER(COALESCE(rr.model, ''))
        OR LOWER(COALESCE(oe.model, '')) = LOWER(COALESCE(rr.requested_model, ''))
        OR LOWER(COALESCE(oe.requested_model, '')) = LOWER(COALESCE(rr.model, ''))
        OR LOWER(COALESCE(oe.requested_model, '')) = LOWER(COALESCE(rr.requested_model, ''))
        OR LOWER(COALESCE(oe.upstream_model, '')) = LOWER(COALESCE(rr.upstream_model, ''))
      )
      AND (
        COALESCE(rr.inbound_endpoint, '') = ''
        OR COALESCE(oe.inbound_endpoint, '') = ''
        OR oe.inbound_endpoint = rr.inbound_endpoint
        OR oe.request_path = rr.inbound_endpoint
      )
      AND (
        rr.status_code IS NULL
        OR oe.status_code = rr.status_code
        OR oe.upstream_status_code = rr.status_code
      )
    )
  ORDER BY
    CASE
      WHEN oe.request_id = rr.request_id THEN 0
      WHEN COALESCE(rr.client_request_id, '') <> '' AND oe.client_request_id = rr.client_request_id THEN 1
      ELSE 2
    END,
    CASE
      WHEN rr.status_code IS NOT NULL AND COALESCE(oe.upstream_status_code, oe.status_code) = rr.status_code THEN 0
      ELSE 1
    END,
    ABS(EXTRACT(EPOCH FROM (oe.created_at - COALESCE(rr.completed_at, rr.updated_at, rr.created_at)))),
    oe.created_at DESC,
    oe.id DESC
  LIMIT 1
) o ON true
LEFT JOIN users u ON u.id = rr.user_id
LEFT JOIN api_keys ak ON ak.id = rr.api_key_id
LEFT JOIN accounts a ON a.id = rr.account_id
LEFT JOIN groups g ON g.id = rr.group_id`
}

func buildWhere(filter *Filter, startTime, endTime time.Time) (string, []any) {
	conditions := []string{"rr.created_at >= $1", "rr.created_at < $2"}
	args := []any{startTime, endTime}
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if filter != nil {
		if v := strings.TrimSpace(filter.SessionID); v != "" {
			add("rr.session_id = $%d", v)
		}
		if v := strings.TrimSpace(filter.Outcome); v != "" && strings.ToLower(v) != "all" {
			add("rr.outcome = $%d", v)
		}
		if filter.Billable != nil {
			add("rr.billable = $%d", *filter.Billable)
		}
		if len(filter.StatusCodes) > 0 {
			add("rr.status_code = ANY($%d)", pq.Array(filter.StatusCodes))
		}
		if filter.ExcludeStatus200 {
			conditions = append(conditions, "COALESCE(rr.status_code, 0) <> 200")
		}
		if v := strings.TrimSpace(filter.RequestID); v != "" {
			add(`(
  rr.request_id = $%[1]d OR
  rr.client_request_id = $%[1]d OR
  ('client:' || COALESCE(rr.client_request_id, '')) = $%[1]d OR
  ('local:' || rr.request_id) = $%[1]d
)`, v)
		}
		if v := strings.TrimSpace(filter.ClientRequestID); v != "" {
			add("rr.client_request_id = $%d", v)
		}
		if filter.UserID != nil {
			add("rr.user_id = $%d", *filter.UserID)
		}
		if filter.APIKeyID != nil {
			add("rr.api_key_id = $%d", *filter.APIKeyID)
		}
		if filter.AccountID != nil {
			add("rr.account_id = $%d", *filter.AccountID)
		}
		if filter.GroupID != nil {
			add("rr.group_id = $%d", *filter.GroupID)
		}
		if v := strings.TrimSpace(filter.Platform); v != "" {
			add("LOWER(COALESCE(rr.platform,'')) = LOWER($%d)", v)
		}
		if v := strings.TrimSpace(filter.Model); v != "" {
			add("LOWER(COALESCE(rr.model,'')) LIKE LOWER($%d)", "%"+v+"%")
		}
		if filter.StatusCode != nil {
			add("rr.status_code = $%d", *filter.StatusCode)
		}
		if filter.RequestType != nil {
			add("rr.request_type = $%d", *filter.RequestType)
		}
		if filter.Stream != nil {
			add("rr.stream = $%d", *filter.Stream)
		}
		if v := strings.TrimSpace(filter.Query); v != "" {
			args = append(args, "%"+strings.ToLower(v)+"%")
			n := len(args)
			conditions = append(conditions, fmt.Sprintf(`(
  LOWER(COALESCE(rr.request_id,'')) LIKE $%d OR
  LOWER(COALESCE(rr.client_request_id,'')) LIKE $%d OR
  LOWER(COALESCE(rr.session_id,'')) LIKE $%d OR
  LOWER(COALESCE(rr.model,'')) LIKE $%d OR
  LOWER(COALESCE(rr.error_message,'')) LIKE $%d OR
  LOWER(COALESCE(rr.user_agent,'')) LIKE $%d
)`, n, n, n, n, n, n))
		}
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

type recordScanner interface {
	Scan(dest ...any) error
}

func scanRecord(rows recordScanner) (*Record, error) {
	var item Record
	var completedAt sql.NullTime
	var userID, apiKeyID, accountID, groupID sql.NullInt64
	var requestType, statusCode, durationMs, firstTokenMs sql.NullInt64
	var inputTokens, outputTokens, cacheCreationTokens, cacheReadTokens, cacheCreation5mTokens, cacheCreation1hTokens sql.NullInt64
	var imageOutputTokens, imageCount, billingType sql.NullInt64
	var imageOutputCost, inputCost, outputCost, cacheCreationCost, cacheReadCost, totalCost, actualCost, rateMultiplier, accountRateMultiplier, accountStatsCost sql.NullFloat64
	var cacheTTLOverridden sql.NullBool
	var imageSizeBreakdown sql.NullString
	var upstreamAttempts sql.NullString
	var userEmail, userUsername, apiKeyName, accountName, groupName sql.NullString
	if err := rows.Scan(
		&item.ID,
		&item.CreatedAt,
		&item.UpdatedAt,
		&completedAt,
		&item.RequestID,
		&item.ClientRequestID,
		&userID,
		&apiKeyID,
		&accountID,
		&groupID,
		&item.SessionID,
		&item.SessionSource,
		&item.ClientSessionID,
		&item.Platform,
		&item.Model,
		&item.RequestedModel,
		&item.UpstreamModel,
		&item.BillingModel,
		&requestType,
		&item.Stream,
		&item.InboundEndpoint,
		&item.UpstreamEndpoint,
		&item.Outcome,
		&statusCode,
		&durationMs,
		&firstTokenMs,
		&item.Billable,
		&inputTokens,
		&outputTokens,
		&cacheCreationTokens,
		&cacheReadTokens,
		&cacheCreation5mTokens,
		&cacheCreation1hTokens,
		&imageOutputTokens,
		&imageOutputCost,
		&imageCount,
		&item.ImageSize,
		&item.ImageInputSize,
		&item.ImageOutputSize,
		&item.ImageSizeSource,
		&imageSizeBreakdown,
		&inputCost,
		&outputCost,
		&cacheCreationCost,
		&cacheReadCost,
		&totalCost,
		&actualCost,
		&rateMultiplier,
		&billingType,
		&item.BillingMode,
		&accountRateMultiplier,
		&accountStatsCost,
		&cacheTTLOverridden,
		&item.ModelMappingChain,
		&item.BillingTier,
		&item.ServiceTier,
		&item.ReasoningEffort,
		&item.IPAddress,
		&item.UserAgent,
		&item.ErrorMessage,
		&upstreamAttempts,
		&userEmail,
		&userUsername,
		&apiKeyName,
		&accountName,
		&groupName,
	); err != nil {
		return nil, err
	}
	if completedAt.Valid {
		item.CompletedAt = &completedAt.Time
	}
	item.UserID = nullInt64Ptr(userID)
	item.APIKeyID = nullInt64Ptr(apiKeyID)
	item.AccountID = nullInt64Ptr(accountID)
	item.GroupID = nullInt64Ptr(groupID)
	item.User = identityPtr(item.UserID, "", userEmail.String, userUsername.String)
	item.APIKey = identityPtr(item.APIKeyID, apiKeyName.String, "", "")
	item.Account = identityPtr(item.AccountID, accountName.String, "", "")
	item.Group = identityPtr(item.GroupID, groupName.String, "", "")
	if requestType.Valid {
		v := int16(requestType.Int64)
		item.RequestType = &v
	}
	item.StatusCode = nullIntPtr(statusCode)
	item.DurationMs = nullIntPtr(durationMs)
	item.FirstTokenMs = nullIntPtr(firstTokenMs)
	item.InputTokens = nullIntPtr(inputTokens)
	item.OutputTokens = nullIntPtr(outputTokens)
	item.CacheCreationTokens = nullIntPtr(cacheCreationTokens)
	item.CacheReadTokens = nullIntPtr(cacheReadTokens)
	item.CacheCreation5mTokens = nullIntPtr(cacheCreation5mTokens)
	item.CacheCreation1hTokens = nullIntPtr(cacheCreation1hTokens)
	item.ImageOutputTokens = nullIntPtr(imageOutputTokens)
	item.ImageOutputCost = nullFloat64Ptr(imageOutputCost)
	item.ImageCount = nullIntPtr(imageCount)
	item.ImageSizeBreakdown = imageSizeBreakdownMap(imageSizeBreakdown)
	item.InputCost = nullFloat64Ptr(inputCost)
	item.OutputCost = nullFloat64Ptr(outputCost)
	item.CacheCreationCost = nullFloat64Ptr(cacheCreationCost)
	item.CacheReadCost = nullFloat64Ptr(cacheReadCost)
	item.TotalCost = nullFloat64Ptr(totalCost)
	item.ActualCost = nullFloat64Ptr(actualCost)
	item.RateMultiplier = nullFloat64Ptr(rateMultiplier)
	if billingType.Valid {
		v := int8(billingType.Int64)
		item.BillingType = &v
	}
	item.AccountRateMultiplier = nullFloat64Ptr(accountRateMultiplier)
	item.AccountStatsCost = nullFloat64Ptr(accountStatsCost)
	item.CacheTTLOverridden = nullBoolPtr(cacheTTLOverridden)
	item.UpstreamAttempts = parseUpstreamAttempts(upstreamAttempts.String)
	if !item.Billable {
		clearBillingDisplayFields(&item)
	}
	return &item, nil
}

func clearBillingDisplayFields(item *Record) {
	item.ImageOutputCost = nil
	item.InputCost = nil
	item.OutputCost = nil
	item.CacheCreationCost = nil
	item.CacheReadCost = nil
	item.TotalCost = nil
	item.ActualCost = nil
	item.RateMultiplier = nil
	item.AccountRateMultiplier = nil
	item.AccountStatsCost = nil
	item.BillingType = nil
	item.BillingMode = ""
}

func requestRecordListColumns() []string {
	return []string{
		"id", "created_at", "updated_at", "completed_at", "request_id", "client_request_id",
		"user_id", "api_key_id", "account_id", "group_id", "session_id", "session_source",
		"client_session_id", "platform", "model", "requested_model", "upstream_model", "billing_model",
		"request_type", "stream", "inbound_endpoint", "upstream_endpoint", "outcome",
		"status_code", "duration_ms", "first_token_ms", "billable", "input_tokens",
		"output_tokens", "cache_creation_tokens", "cache_read_tokens", "cache_creation_5m_tokens",
		"cache_creation_1h_tokens", "image_output_tokens", "image_output_cost", "image_count",
		"image_size", "image_input_size", "image_output_size", "image_size_source", "image_size_breakdown",
		"input_cost", "output_cost", "cache_creation_cost", "cache_read_cost", "total_cost", "actual_cost",
		"rate_multiplier", "billing_type", "billing_mode", "account_rate_multiplier", "account_stats_cost",
		"cache_ttl_overridden", "model_mapping_chain", "billing_tier", "service_tier", "reasoning_effort",
		"ip_address", "user_agent", "error_message", "upstream_attempts", "user_email", "user_username", "api_key_name", "account_name", "group_name",
	}
}

func parseUpstreamAttempts(raw string) []UpstreamAttempt {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" || raw == "[]" {
		return nil
	}

	var attempts []UpstreamAttempt
	if err := json.Unmarshal([]byte(raw), &attempts); err != nil || len(attempts) == 0 {
		return nil
	}

	out := attempts[:0]
	for _, attempt := range attempts {
		attempt.Platform = strings.TrimSpace(attempt.Platform)
		attempt.AccountName = strings.TrimSpace(attempt.AccountName)
		attempt.UpstreamRequestID = strings.TrimSpace(attempt.UpstreamRequestID)
		attempt.UpstreamURL = strings.TrimSpace(attempt.UpstreamURL)
		attempt.Kind = strings.TrimSpace(attempt.Kind)
		attempt.Message = strings.TrimSpace(attempt.Message)
		attempt.Detail = strings.TrimSpace(attempt.Detail)
		if attempt.UpstreamStatusCode <= 0 && attempt.Message == "" && attempt.Detail == "" && attempt.AccountID <= 0 && attempt.AccountName == "" {
			continue
		}
		out = append(out, attempt)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func nullString(v string) any {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return v
}

func nullInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullInt16(v *int16) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullInt(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullFloat64(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func nullIntPtr(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	out := int(v.Int64)
	return &out
}

func nullFloat64Ptr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	return &v.Float64
}

func nullBoolPtr(v sql.NullBool) *bool {
	if !v.Valid {
		return nil
	}
	return &v.Bool
}

func identityPtr(id *int64, name, email, username string) *Identity {
	if id == nil {
		return nil
	}
	return &Identity{
		ID:       *id,
		Name:     strings.TrimSpace(name),
		Email:    strings.TrimSpace(email),
		Username: strings.TrimSpace(username),
	}
}

func imageSizeBreakdownMap(v sql.NullString) map[string]int {
	if !v.Valid || strings.TrimSpace(v.String) == "" {
		return nil
	}
	out := map[string]int{}
	if err := json.Unmarshal([]byte(v.String), &out); err != nil {
		return nil
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
