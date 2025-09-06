package httpapi

import (
	"time"

	"github.com/oziev02/subscriptions-service/internal/domain"
	"github.com/oziev02/subscriptions-service/internal/usecase"
)

type subDTO struct {
	ID          string  `json:"id"`
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func toDTO(s *domain.Subscription) subDTO {
	var end *string
	if s.End != nil {
		v := s.End.String()
		end = &v
	}
	return subDTO{
		ID:          s.ID.String(),
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID.String(),
		StartDate:   s.Start.String(),
		EndDate:     end,
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
	}
}

var _ = usecase.CreateInput{}
