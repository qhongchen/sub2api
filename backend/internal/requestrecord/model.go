package requestrecord

import "time"

func startOfLocalDay(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
}

const (
	SessionSourceHeaderSessionID       = "header_session_id"
	SessionSourceHeaderConversationID  = "header_conversation_id"
	SessionSourceHeaderXSessionID      = "header_x_session_id"
	SessionSourceHeaderXConversationID = "header_x_conversation_id"
	SessionSourcePromptCacheKey        = "prompt_cache_key"
	SessionSourceMetadataUserID        = "metadata_user_id"
	SessionSourceMetadataSessionID     = "metadata_session_id"
	SessionSourceGenerated             = "generated"
	SessionSourceUnknown               = "unknown"
)

const (
	OutcomePending     = "pending"
	OutcomeSuccess     = "success"
	OutcomeNonBillable = "non_billable"
	OutcomeError       = "error"
	OutcomeCancelled   = "cancelled"
	OutcomeUnknown     = "unknown"
)

type SessionIdentity struct {
	SessionID       string `json:"session_id,omitempty"`
	Source          string `json:"session_source,omitempty"`
	ClientSessionID string `json:"client_session_id,omitempty"`
}

type Record struct {
	ID              int64      `json:"id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	RequestID       string     `json:"request_id"`
	ClientRequestID string     `json:"client_request_id,omitempty"`

	UserID    *int64    `json:"user_id,omitempty"`
	APIKeyID  *int64    `json:"api_key_id,omitempty"`
	AccountID *int64    `json:"account_id,omitempty"`
	GroupID   *int64    `json:"group_id,omitempty"`
	User      *Identity `json:"user,omitempty"`
	APIKey    *Identity `json:"api_key,omitempty"`
	Account   *Identity `json:"account,omitempty"`
	Group     *Identity `json:"group,omitempty"`

	SessionID       string `json:"session_id,omitempty"`
	SessionSource   string `json:"session_source,omitempty"`
	ClientSessionID string `json:"client_session_id,omitempty"`

	Platform          string `json:"platform,omitempty"`
	Model             string `json:"model,omitempty"`
	RequestedModel    string `json:"requested_model,omitempty"`
	UpstreamModel     string `json:"upstream_model,omitempty"`
	BillingModel      string `json:"billing_model,omitempty"`
	ModelMappingChain string `json:"model_mapping_chain,omitempty"`
	BillingTier       string `json:"billing_tier,omitempty"`
	ServiceTier       string `json:"service_tier,omitempty"`
	ReasoningEffort   string `json:"reasoning_effort,omitempty"`
	RequestType       *int16 `json:"request_type,omitempty"`
	Stream            bool   `json:"stream"`
	InboundEndpoint   string `json:"inbound_endpoint,omitempty"`
	UpstreamEndpoint  string `json:"upstream_endpoint,omitempty"`

	Outcome      string `json:"outcome"`
	StatusCode   *int   `json:"status_code,omitempty"`
	DurationMs   *int   `json:"duration_ms,omitempty"`
	FirstTokenMs *int   `json:"first_token_ms,omitempty"`

	Billable              bool           `json:"billable"`
	InputTokens           *int           `json:"input_tokens,omitempty"`
	OutputTokens          *int           `json:"output_tokens,omitempty"`
	CacheCreationTokens   *int           `json:"cache_creation_tokens,omitempty"`
	CacheReadTokens       *int           `json:"cache_read_tokens,omitempty"`
	CacheCreation5mTokens *int           `json:"cache_creation_5m_tokens,omitempty"`
	CacheCreation1hTokens *int           `json:"cache_creation_1h_tokens,omitempty"`
	ImageOutputTokens     *int           `json:"image_output_tokens,omitempty"`
	ImageOutputCost       *float64       `json:"image_output_cost,omitempty"`
	ImageCount            *int           `json:"image_count,omitempty"`
	ImageSize             string         `json:"image_size,omitempty"`
	ImageInputSize        string         `json:"image_input_size,omitempty"`
	ImageOutputSize       string         `json:"image_output_size,omitempty"`
	ImageSizeSource       string         `json:"image_size_source,omitempty"`
	ImageSizeBreakdown    map[string]int `json:"image_size_breakdown,omitempty"`
	InputCost             *float64       `json:"input_cost,omitempty"`
	OutputCost            *float64       `json:"output_cost,omitempty"`
	CacheCreationCost     *float64       `json:"cache_creation_cost,omitempty"`
	CacheReadCost         *float64       `json:"cache_read_cost,omitempty"`
	TotalCost             *float64       `json:"total_cost,omitempty"`
	ActualCost            *float64       `json:"actual_cost,omitempty"`
	RateMultiplier        *float64       `json:"rate_multiplier,omitempty"`
	BillingType           *int8          `json:"billing_type,omitempty"`
	BillingMode           string         `json:"billing_mode,omitempty"`
	AccountRateMultiplier *float64       `json:"account_rate_multiplier,omitempty"`
	AccountStatsCost      *float64       `json:"account_stats_cost,omitempty"`
	CacheTTLOverridden    *bool          `json:"cache_ttl_overridden,omitempty"`

	IPAddress    string `json:"ip_address,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`

	UpstreamAttempts []UpstreamAttempt `json:"upstream_attempts,omitempty"`
}

type Identity struct {
	ID       int64  `json:"id"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
}

type UpstreamAttempt struct {
	AtUnixMs int64 `json:"at_unix_ms,omitempty"`

	Platform    string `json:"platform,omitempty"`
	AccountID   int64  `json:"account_id,omitempty"`
	AccountName string `json:"account_name,omitempty"`

	UpstreamStatusCode int    `json:"upstream_status_code,omitempty"`
	UpstreamRequestID  string `json:"upstream_request_id,omitempty"`
	UpstreamURL        string `json:"upstream_url,omitempty"`

	Kind    string `json:"kind,omitempty"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

type StartInput struct {
	Now             time.Time
	RequestID       string
	ClientRequestID string

	UserID    *int64
	APIKeyID  *int64
	AccountID *int64
	GroupID   *int64

	Session SessionIdentity

	Platform         string
	Model            string
	RequestedModel   string
	UpstreamModel    string
	RequestType      *int16
	Stream           bool
	InboundEndpoint  string
	UpstreamEndpoint string
	IPAddress        string
	UserAgent        string
}

type Handle struct {
	ID        int64
	RequestID string
	StartedAt time.Time
}

type CompleteInput struct {
	Now       time.Time
	RequestID string

	Outcome      string
	StatusCode   *int
	DurationMs   *int
	FirstTokenMs *int

	Billable     bool
	InputTokens  *int
	OutputTokens *int
	TotalCost    *float64
	ActualCost   *float64

	UpstreamEndpoint string
	ErrorMessage     string
}

type Filter struct {
	StartTime *time.Time
	EndTime   *time.Time

	Outcome  string
	Billable *bool

	StatusCodes      []int
	ExcludeStatus200 bool

	SessionID       string
	RequestID       string
	ClientRequestID string

	UserID    *int64
	APIKeyID  *int64
	AccountID *int64
	GroupID   *int64

	Platform string
	Model    string

	StatusCode  *int
	RequestType *int16
	Stream      *bool

	Query string
	Sort  string

	Page     int
	PageSize int
}

func (f *Filter) Normalize() (page, pageSize int, startTime, endTime time.Time) {
	page = 1
	pageSize = 50
	endTime = time.Now()
	startTime = startOfLocalDay(endTime)

	if f == nil {
		return page, pageSize, startTime, endTime
	}

	if f.Page > 0 {
		page = f.Page
	}
	if f.PageSize > 0 {
		pageSize = f.PageSize
	}
	if pageSize > 100 {
		pageSize = 100
	}

	if f.EndTime != nil {
		endTime = *f.EndTime
	}
	if f.StartTime != nil {
		startTime = *f.StartTime
	} else if f.EndTime != nil {
		startTime = endTime.Add(-1 * time.Hour)
	}
	if startTime.After(endTime) {
		startTime, endTime = endTime, startTime
	}
	return page, pageSize, startTime, endTime
}

type List struct {
	Items    []*Record `json:"items"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Pages    int       `json:"pages"`
	Summary  *Summary  `json:"summary,omitempty"`
}

type Summary struct {
	TotalRequests   int64   `json:"total_requests"`
	SuccessRequests int64   `json:"success_requests"`
	ErrorRequests   int64   `json:"error_requests"`
	ErrorRate       float64 `json:"error_rate"`
}
