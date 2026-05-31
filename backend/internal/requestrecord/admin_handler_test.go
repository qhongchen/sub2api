package requestrecord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeListService struct {
	gotFilter *Filter
	out       *List
	err       error
}

func (s *fakeListService) ListRecords(ctx context.Context, filter *Filter) (*List, error) {
	s.gotFilter = filter
	return s.out, s.err
}

func TestAdminHandlerListParsesFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	start := "2026-05-29T13:00:00Z"
	end := "2026-05-29T14:00:00Z"
	billable := false
	status := 503
	requestType := int16(2)
	stream := true
	svc := &fakeListService{
		out: &List{
			Items: []*Record{{
				ID:        1,
				CreatedAt: time.Date(2026, 5, 29, 13, 30, 0, 0, time.UTC),
				RequestID: "req-1",
				SessionID: "sess-1",
				Outcome:   OutcomeError,
			}},
			Total:    1,
			Page:     2,
			PageSize: 25,
			Pages:    1,
		},
	}
	handler := NewAdminHandler(svc)
	router := gin.New()
	router.GET("/records", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/records?start_time="+start+"&end_time="+end+"&page=2&page_size=25&session_id=sess-1&outcome=error&billable=false&request_id=req-1&client_request_id=creq-1&user_id=42&api_key_id=7&account_id=9&group_id=11&platform=openai&model=gpt&status_code=503&request_type=2&stream=true&q=timeout&sort=duration_desc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if svc.gotFilter == nil {
		t.Fatalf("expected filter")
	}
	if svc.gotFilter.SessionID != "sess-1" || svc.gotFilter.Outcome != OutcomeError || svc.gotFilter.Billable == nil || *svc.gotFilter.Billable != billable {
		t.Fatalf("basic filters not parsed: %+v", svc.gotFilter)
	}
	if svc.gotFilter.StatusCode == nil || *svc.gotFilter.StatusCode != status {
		t.Fatalf("status not parsed: %+v", svc.gotFilter.StatusCode)
	}
	if svc.gotFilter.RequestType == nil || *svc.gotFilter.RequestType != requestType {
		t.Fatalf("request_type not parsed: %+v", svc.gotFilter.RequestType)
	}
	if svc.gotFilter.Stream == nil || *svc.gotFilter.Stream != stream {
		t.Fatalf("stream not parsed: %+v", svc.gotFilter.Stream)
	}
	if svc.gotFilter.Page != 2 || svc.gotFilter.PageSize != 25 || svc.gotFilter.Sort != "duration_desc" {
		t.Fatalf("pagination/sort not parsed: %+v", svc.gotFilter)
	}

	var payload struct {
		Data List `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v body=%s", err, w.Body.String())
	}
	if payload.Data.Total != 1 || len(payload.Data.Items) != 1 {
		t.Fatalf("unexpected payload: %+v", payload.Data)
	}
}

func TestAdminHandlerListAcceptsPageCompatibilityParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAdminHandler(&fakeListService{out: &List{Items: []*Record{}, Total: 0, Page: 1, PageSize: 20, Pages: 1}})
	router := gin.New()
	router.GET("/records", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/records?kind=error&status_codes=5xx&status_codes_other=true&request_type=stream&time_range=30m", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	svc := handler.service.(*fakeListService)
	if svc.gotFilter == nil {
		t.Fatalf("expected filter")
	}
	if svc.gotFilter.Outcome != OutcomeError {
		t.Fatalf("kind compatibility not mapped: %+v", svc.gotFilter)
	}
	if len(svc.gotFilter.StatusCodes) != 100 || svc.gotFilter.StatusCodes[0] != 500 || svc.gotFilter.StatusCodes[99] != 599 {
		t.Fatalf("5xx status filter not expanded: %+v", svc.gotFilter.StatusCodes)
	}
	if !svc.gotFilter.ExcludeStatus200 {
		t.Fatalf("status_codes_other not parsed")
	}
	if svc.gotFilter.RequestType == nil || *svc.gotFilter.RequestType != 2 {
		t.Fatalf("string request_type not parsed: %+v", svc.gotFilter.RequestType)
	}
	if svc.gotFilter.StartTime == nil || svc.gotFilter.EndTime == nil {
		t.Fatalf("time_range not converted: %+v", svc.gotFilter)
	}
}

func TestAdminHandlerListRejectsInvalidBillable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAdminHandler(&fakeListService{})
	router := gin.New()
	router.GET("/records", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/records?billable=maybe", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
