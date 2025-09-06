package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidPrice     = errors.New("price must be >= 0")
	ErrInvalidDateRange = errors.New("start_date must be <= end_date")
)

type YearMonth struct {
	time time.Time
}

func YearMonthFromTime(t time.Time) YearMonth {
	t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return YearMonth{time: t}
}

func MustYearMonth(mmYYYY string) YearMonth {
	ym, err := ParseYearMonth(mmYYYY)
	if err != nil {
		panic(err)
	}
	return ym
}

func ParseYearMonth(mmYYYY string) (YearMonth, error) {
	var t time.Time
	var err error
	if len(mmYYYY) == 7 && mmYYYY[2] == '-' {
		t, err = time.Parse("01-2006", mmYYYY)
	} else if len(mmYYYY) == 7 && mmYYYY[4] == '-' {
		t, err = time.Parse("2006-01", mmYYYY)
	} else {
		return YearMonth{}, errors.New("invalid year-month format; use MM-YYYY")
	}
	if err != nil {
		return YearMonth{}, err
	}
	t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return YearMonth{time: t}, nil
}

func (ym YearMonth) String() string  { return ym.time.Format("01-2006") }
func (ym YearMonth) Time() time.Time { return ym.time }
func (ym YearMonth) BeforeOrEqual(other YearMonth) bool {
	return !ym.time.After(other.time)
}
func (ym YearMonth) AfterOrEqual(other YearMonth) bool {
	return !ym.time.Before(other.time)
}
func (ym YearMonth) MonthsUntil(other YearMonth) int {
	y1, m1 := ym.time.Year(), ym.time.Month()
	y2, m2 := other.time.Year(), other.time.Month()
	return (y2-y1)*12 + int(m2-m1) + 1
}

type Subscription struct {
	ID          uuid.UUID
	ServiceName string
	Price       int
	UserID      uuid.UUID
	Start       YearMonth
	End         *YearMonth
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *Subscription) Validate() error {
	if s.Price < 0 {
		return ErrInvalidPrice
	}
	if s.End != nil && s.End.time.Before(s.Start.time) {
		return ErrInvalidDateRange
	}
	return nil
}

func (s *Subscription) OverlapMonths(from, to YearMonth) int {
	start := s.Start
	end := to
	if s.End != nil && s.End.BeforeOrEqual(to) {
		end = *s.End
	}
	if start.AfterOrEqual(to) && !start.BeforeOrEqual(to) {
		return 0
	}
	if end.BeforeOrEqual(from) && !end.AfterOrEqual(from) {
		return 0
	}
	if start.BeforeOrEqual(from) == false {
		from = start
	}
	if s.End != nil && end.AfterOrEqual(to) == false {
		to = end
	}
	if to.BeforeOrEqual(from) {
		if to.time.Equal(from.time) {
			return 1
		}
		return 0
	}
	return from.MonthsUntil(to)
}
