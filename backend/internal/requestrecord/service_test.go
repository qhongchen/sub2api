package requestrecord

import (
	"context"
	"testing"
	"time"
)

type fakeRequestRecordRepository struct {
	items   []*Record
	total   int64
	summary *Summary
	filter  *Filter
}

func (r *fakeRequestRecordRepository) Start(ctx context.Context, input *StartInput) (*Handle, error) {
	return nil, nil
}

func (r *fakeRequestRecordRepository) Complete(ctx context.Context, input *CompleteInput) error {
	return nil
}

func (r *fakeRequestRecordRepository) List(ctx context.Context, filter *Filter) ([]*Record, int64, error) {
	r.filter = filter
	return r.items, r.total, nil
}

func (r *fakeRequestRecordRepository) Summary(ctx context.Context, filter *Filter) (*Summary, error) {
	return r.summary, nil
}

func TestFilterNormalizeDefaultsAndClamp(t *testing.T) {
	end := time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC)
	start := end.Add(2 * time.Hour)
	filter := &Filter{
		StartTime: &start,
		EndTime:   &end,
		Page:      -1,
		PageSize:  999,
	}

	page, pageSize, gotStart, gotEnd := filter.Normalize()

	if page != 1 {
		t.Fatalf("page=%d, want 1", page)
	}
	if pageSize != 100 {
		t.Fatalf("pageSize=%d, want 100", pageSize)
	}
	if !gotStart.Equal(end) || !gotEnd.Equal(start) {
		t.Fatalf("time range not swapped: start=%v end=%v", gotStart, gotEnd)
	}
}

func TestFilterNormalizeDefaultsToToday(t *testing.T) {
	filter := &Filter{}

	page, pageSize, gotStart, gotEnd := filter.Normalize()

	if page != 1 || pageSize != 50 {
		t.Fatalf("pagination=(%d,%d), want (1,50)", page, pageSize)
	}
	now := time.Now()
	wantStart := startOfLocalDay(now)
	if !gotStart.Equal(wantStart) {
		t.Fatalf("start=%v, want local day start %v", gotStart, wantStart)
	}
	if gotEnd.Before(now.Add(-2*time.Second)) || gotEnd.After(now.Add(2*time.Second)) {
		t.Fatalf("end=%v, want around now %v", gotEnd, now)
	}
}

func TestListRecordsNilRepositoryReturnsEmpty(t *testing.T) {
	service := NewService(nil)

	out, err := service.ListRecords(nil, nil)
	if err != nil {
		t.Fatalf("ListRecords() error: %v", err)
	}
	if out == nil {
		t.Fatalf("ListRecords() returned nil")
	}
	if out.Page != 1 || out.PageSize != 50 || out.Total != 0 || out.Pages != 1 {
		t.Fatalf("unexpected pagination: %+v", out)
	}
	if len(out.Items) != 0 {
		t.Fatalf("items len=%d, want 0", len(out.Items))
	}
}

func TestServiceStartCompleteNilRepositoryAreNoop(t *testing.T) {
	service := NewService(nil)

	handle, err := service.Start(nil, &StartInput{RequestID: "req-1"})
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	if handle != nil {
		t.Fatalf("handle=%+v, want nil", handle)
	}
	if err := service.Complete(nil, &CompleteInput{RequestID: "req-1"}); err != nil {
		t.Fatalf("Complete() error: %v", err)
	}
}

func TestListRecordsIncludesRepositorySummary(t *testing.T) {
	repo := &fakeRequestRecordRepository{
		items: []*Record{{RequestID: "req-1", Outcome: OutcomeSuccess}},
		total: 2,
		summary: &Summary{
			TotalRequests:   2,
			SuccessRequests: 1,
			ErrorRequests:   1,
			ErrorRate:       0.5,
		},
	}
	service := NewService(repo)

	out, err := service.ListRecords(context.Background(), &Filter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListRecords() error: %v", err)
	}
	if out.Summary == nil || out.Summary.TotalRequests != 2 || out.Summary.ErrorRate != 0.5 {
		t.Fatalf("summary not propagated: %+v", out.Summary)
	}
	if repo.filter == nil || repo.filter.StartTime == nil || repo.filter.EndTime == nil {
		t.Fatalf("normalized filter not passed to repo: %+v", repo.filter)
	}
}
