package requestrecord

import (
	"context"
	"math"
)

type Repository interface {
	Start(ctx context.Context, input *StartInput) (*Handle, error)
	Complete(ctx context.Context, input *CompleteInput) error
	List(ctx context.Context, filter *Filter) ([]*Record, int64, error)
	Summary(ctx context.Context, filter *Filter) (*Summary, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Start(ctx context.Context, input *StartInput) (*Handle, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	return s.repo.Start(ctx, input)
}

func (s *Service) Complete(ctx context.Context, input *CompleteInput) error {
	if s == nil || s.repo == nil {
		return nil
	}
	return s.repo.Complete(ctx, input)
}

func (s *Service) ListRecords(ctx context.Context, filter *Filter) (*List, error) {
	page, pageSize, startTime, endTime := filter.Normalize()
	if s == nil || s.repo == nil {
		return &List{
			Items:    []*Record{},
			Total:    0,
			Page:     page,
			PageSize: pageSize,
			Pages:    1,
			Summary:  &Summary{},
		}, nil
	}

	filterCopy := &Filter{}
	if filter != nil {
		*filterCopy = *filter
	}
	filterCopy.Page = page
	filterCopy.PageSize = pageSize
	filterCopy.StartTime = &startTime
	filterCopy.EndTime = &endTime

	items, total, err := s.repo.List(ctx, filterCopy)
	if err != nil {
		return nil, err
	}
	summary, err := s.repo.Summary(ctx, filterCopy)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []*Record{}
	}
	if summary == nil {
		summary = &Summary{}
	}
	pages := 1
	if pageSize > 0 && total > 0 {
		pages = int(math.Ceil(float64(total) / float64(pageSize)))
	}
	return &List{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
		Summary:  summary,
	}, nil
}
