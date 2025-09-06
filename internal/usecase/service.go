package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/oziev02/subscriptions-service/internal/domain"
)

type SubscriptionRepo interface {
	Create(ctx context.Context, s *domain.Subscription) error
	Get(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	Update(ctx context.Context, s *domain.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ListFilter) ([]*domain.Subscription, error)
	Summary(ctx context.Context, from, to domain.YearMonth, userID *uuid.UUID, serviceName *string) (int64, error)
}

type ListFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	Limit       int
	Offset      int
}

type Service struct {
	repo SubscriptionRepo
}

func NewService(r SubscriptionRepo) *Service { return &Service{repo: r} }

func (s *Service) Create(ctx context.Context, in CreateInput) (*domain.Subscription, error) {
	start, err := domain.ParseYearMonth(in.StartDate)
	if err != nil {
		return nil, err
	}
	var endPtr *domain.YearMonth
	if in.EndDate != nil && *in.EndDate != "" {
		e, err := domain.ParseYearMonth(*in.EndDate)
		if err != nil {
			return nil, err
		}
		endPtr = &e
	}
	sub := &domain.Subscription{
		ID:          uuid.New(),
		ServiceName: in.ServiceName,
		Price:       in.Price,
		UserID:      in.UserID,
		Start:       start,
		End:         endPtr,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := sub.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpdateInput) (*domain.Subscription, error) {
	sub, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.ServiceName != nil {
		sub.ServiceName = *in.ServiceName
	}
	if in.Price != nil {
		sub.Price = *in.Price
	}
	if in.StartDate != nil {
		st, err := domain.ParseYearMonth(*in.StartDate)
		if err != nil {
			return nil, err
		}
		sub.Start = st
	}
	if in.EndDateSet {
		if in.EndDate == nil || *in.EndDate == "" {
			sub.End = nil
		} else {
			e, err := domain.ParseYearMonth(*in.EndDate)
			if err != nil {
				return nil, err
			}
			sub.End = &e
		}
	}
	sub.UpdatedAt = time.Now().UTC()
	if err := sub.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]*domain.Subscription, error) {
	return s.repo.List(ctx, f)
}

func (s *Service) Summary(ctx context.Context, fromStr, toStr string, userID *uuid.UUID, serviceName *string) (int64, error) {
	from, err := domain.ParseYearMonth(fromStr)
	if err != nil {
		return 0, err
	}
	to, err := domain.ParseYearMonth(toStr)
	if err != nil {
		return 0, err
	}
	return s.repo.Summary(ctx, from, to, userID, serviceName)
}

// DTOs

type CreateInput struct {
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   string  // MM-YYYY
	EndDate     *string // optional
}

type UpdateInput struct {
	ServiceName *string
	Price       *int
	StartDate   *string
	EndDate     *string
	EndDateSet  bool
}
