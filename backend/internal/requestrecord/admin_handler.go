package requestrecord

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

type ListService interface {
	ListRecords(ctx context.Context, filter *Filter) (*List, error)
}

type AdminHandler struct {
	service ListService
}

func NewAdminHandler(service ListService) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "Request record service not available")
		return
	}

	filter, err := parseListFilter(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	out, err := h.service.ListRecords(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list request records")
		return
	}
	response.Success(c, out)
}

func parseListFilter(c *gin.Context) (*Filter, error) {
	page, pageSize := response.ParsePagination(c)
	filter := &Filter{
		Page:            page,
		PageSize:        pageSize,
		Outcome:         firstNonEmpty(c.Query("outcome"), c.Query("kind")),
		SessionID:       strings.TrimSpace(c.Query("session_id")),
		RequestID:       strings.TrimSpace(c.Query("request_id")),
		ClientRequestID: strings.TrimSpace(c.Query("client_request_id")),
		Platform:        strings.TrimSpace(c.Query("platform")),
		Model:           strings.TrimSpace(c.Query("model")),
		Query:           strings.TrimSpace(c.Query("q")),
		Sort:            strings.TrimSpace(c.Query("sort")),
	}

	if start, ok, err := parseTimeQuery(c, "start_time"); err != nil {
		return nil, err
	} else if ok {
		filter.StartTime = &start
	}
	if end, ok, err := parseTimeQuery(c, "end_time"); err != nil {
		return nil, err
	} else if ok {
		filter.EndTime = &end
	}
	if filter.StartTime == nil && filter.EndTime == nil {
		if start, end, ok := parseRelativeTimeRange(c.Query("time_range")); ok {
			filter.StartTime = &start
			filter.EndTime = &end
		}
	}
	if v, ok, err := parseBoolQuery(c, "billable"); err != nil {
		return nil, err
	} else if ok {
		filter.Billable = &v
	}
	if err := parseStatusCodes(c, filter); err != nil {
		return nil, err
	}
	if v, ok, err := parseIntQuery(c, "status_code"); err != nil {
		return nil, err
	} else if ok {
		filter.StatusCode = &v
	}
	if raw := strings.TrimSpace(c.Query("request_type")); raw != "" {
		if v, err := parseRequestType(raw); err != nil {
			return nil, err
		} else {
			filter.RequestType = &v
		}
	}
	if v, ok, err := parseBoolQuery(c, "stream"); err != nil {
		return nil, err
	} else if ok {
		filter.Stream = &v
	}
	if v, ok, err := parseInt64Query(c, "user_id"); err != nil {
		return nil, err
	} else if ok {
		filter.UserID = &v
	}
	if v, ok, err := parseInt64Query(c, "api_key_id"); err != nil {
		return nil, err
	} else if ok {
		filter.APIKeyID = &v
	}
	if v, ok, err := parseInt64Query(c, "account_id"); err != nil {
		return nil, err
	} else if ok {
		filter.AccountID = &v
	}
	if v, ok, err := parseInt64Query(c, "group_id"); err != nil {
		return nil, err
	} else if ok {
		filter.GroupID = &v
	}
	return filter, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func parseRelativeTimeRange(raw string) (time.Time, time.Time, bool) {
	now := time.Now()
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "", "today":
		return startOfLocalDay(now), now, true
	case "this_week", "week":
		return startOfLocalWeek(now), now, true
	case "last_7_days", "7d":
		start := startOfLocalDay(now)
		start = start.AddDate(0, 0, -6)
		return start, now, true
	case "this_month", "month":
		return startOfLocalMonth(now), now, true
	case "last_30_days", "30d":
		start := startOfLocalDay(now)
		start = start.AddDate(0, 0, -29)
		return start, now, true
	case "5m":
		return now.Add(-5 * time.Minute), now, true
	case "30m":
		return now.Add(-30 * time.Minute), now, true
	case "1h":
		return now.Add(-1 * time.Hour), now, true
	case "6h":
		return now.Add(-6 * time.Hour), now, true
	case "24h":
		return now.Add(-24 * time.Hour), now, true
	default:
		return time.Time{}, time.Time{}, false
	}
}

func startOfLocalWeek(now time.Time) time.Time {
	start := startOfLocalDay(now)
	day := start.Weekday()
	diff := time.Monday - day
	if day == time.Sunday {
		diff = -6
	}
	start = start.AddDate(0, 0, int(diff))
	return start
}

func startOfLocalMonth(now time.Time) time.Time {
	start := startOfLocalDay(now)
	return time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
}

func parseStatusCodes(c *gin.Context, filter *Filter) error {
	if filter == nil {
		return nil
	}
	if raw := strings.TrimSpace(c.Query("status_codes_other")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		if parsed {
			filter.ExcludeStatus200 = true
		}
	}

	raw := strings.TrimSpace(c.Query("status_codes"))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(strings.ToLower(part))
		if p == "" {
			continue
		}
		switch p {
		case "!200":
			filter.ExcludeStatus200 = true
			continue
		case "4xx":
			for n := 400; n <= 499; n++ {
				out = append(out, n)
			}
			continue
		case "5xx":
			for n := 500; n <= 599; n++ {
				out = append(out, n)
			}
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			if err == nil {
				err = fmt.Errorf("invalid status_codes")
			}
			return err
		}
		out = append(out, n)
	}
	filter.StatusCodes = out
	return nil
}

func parseRequestType(raw string) (int16, error) {
	if parsed, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 16); err == nil {
		return int16(parsed), nil
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "unknown":
		return 0, nil
	case "sync":
		return 1, nil
	case "stream":
		return 2, nil
	case "ws_v2":
		return 3, nil
	default:
		return 0, fmt.Errorf("invalid request_type, allowed values: unknown, sync, stream, ws_v2")
	}
}

func parseTimeQuery(c *gin.Context, key string) (time.Time, bool, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return time.Time{}, false, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, false, err
	}
	return parsed, true, nil
}

func parseBoolQuery(c *gin.Context, key string) (bool, bool, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return false, false, nil
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false, false, err
	}
	return parsed, true, nil
}

func parseIntQuery(c *gin.Context, key string) (int, bool, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0, false, nil
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false, err
	}
	return parsed, true, nil
}

func parseInt16Query(c *gin.Context, key string) (int16, bool, error) {
	parsed, ok, err := parseIntQuery(c, key)
	if err != nil || !ok {
		return 0, ok, err
	}
	return int16(parsed), true, nil
}

func parseInt64Query(c *gin.Context, key string) (int64, bool, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0, false, nil
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return parsed, true, nil
}
