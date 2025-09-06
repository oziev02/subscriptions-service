package postgres

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/oziev02/subscriptions-service/internal/domain"
	"github.com/oziev02/subscriptions-service/internal/usecase"
)

type SubscriptionRepo struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

func NewSubscriptionRepo(pool *pgxpool.Pool, log *zap.Logger) *SubscriptionRepo {
	return &SubscriptionRepo{pool: pool, log: log}
}

func (r *SubscriptionRepo) Create(ctx context.Context, s *domain.Subscription) error {
	const q = `INSERT INTO subscriptions
		(id, service_name, price, user_id, start_date, end_date, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.pool.Exec(ctx, q, s.ID, s.ServiceName, s.Price, s.UserID, s.Start.Time(), nullableYM(s.End), s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *SubscriptionRepo) Get(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	const q = `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions WHERE id=$1`
	row := r.pool.QueryRow(ctx, q, id)
	return scanSub(row)
}

func (r *SubscriptionRepo) Update(ctx context.Context, s *domain.Subscription) error {
	const q = `UPDATE subscriptions
		SET service_name=$2, price=$3, start_date=$4, end_date=$5, updated_at=$6
		WHERE id=$1`
	_, err := r.pool.Exec(ctx, q, s.ID, s.ServiceName, s.Price, s.Start.Time(), nullableYM(s.End), s.UpdatedAt)
	return err
}

func (r *SubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	cmd, err := r.pool.Exec(ctx, "DELETE FROM subscriptions WHERE id=$1", id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("not found")
	}
	return nil
}

func (r *SubscriptionRepo) List(ctx context.Context, f usecase.ListFilter) ([]*domain.Subscription, error) {
	var filters []string
	var args []any
	idx := 1
	if f.UserID != nil {
		filters = append(filters, "user_id = $"+itoa(idx))
		args = append(args, *f.UserID)
		idx++
	}
	if f.ServiceName != nil {
		filters = append(filters, "service_name ILIKE $"+itoa(idx))
		args = append(args, "%"+*f.ServiceName+"%")
		idx++
	}
	where := ""
	if len(filters) > 0 {
		where = "WHERE " + strings.Join(filters, " AND ")
	}
	limit := 100
	if f.Limit > 0 && f.Limit <= 100 {
		limit = f.Limit
	}
	offset := 0
	if f.Offset > 0 {
		offset = f.Offset
	}
	q := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions ` + where + ` ORDER BY created_at DESC LIMIT ` + itoa(limit) + ` OFFSET ` + itoa(offset)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*domain.Subscription
	for rows.Next() {
		s, err := scanSub(rows)
		if err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, rows.Err()
}

func (r *SubscriptionRepo) Summary(ctx context.Context, from, to domain.YearMonth, userID *uuid.UUID, serviceName *string) (int64, error) {
	// Суммируем price * количество месяцев пересечения в границах from..to (включительно).
	var filters []string
	var args []any
	idx := 1
	if userID != nil {
		filters = append(filters, "user_id = $"+itoa(idx))
		args = append(args, *userID)
		idx++
	}
	if serviceName != nil {
		filters = append(filters, "service_name ILIKE $"+itoa(idx))
		args = append(args, "%"+*serviceName+"%")
		idx++
	}
	where := ""
	if len(filters) > 0 {
		where = " AND " + strings.Join(filters, " AND ")
	}

	q := `
	WITH bounds AS (
		SELECT $` + itoa(idx) + `::date AS from_date, $` + itoa(idx+1) + `::date AS to_date
	),
	active AS (
		SELECT s.*, 
			GREATEST(date_trunc('month', s.start_date), date_trunc('month', (SELECT from_date FROM bounds))) AS eff_start,
			LEAST(
				date_trunc('month', COALESCE(s.end_date, (SELECT to_date FROM bounds))),
				date_trunc('month', (SELECT to_date FROM bounds))
			) AS eff_end
		FROM subscriptions s, bounds
		WHERE (s.end_date IS NULL OR s.end_date >= (SELECT from_date FROM bounds))
		  AND s.start_date <= (SELECT to_date FROM bounds)
		` + where + `
	),
	months AS (
		SELECT price, 1 + (date_part('year', eff_end) - date_part('year', eff_start)) * 12
			 + (date_part('month', eff_end) - date_part('month', eff_start)) AS months_overlap
		FROM active
		WHERE eff_end >= eff_start
	)
	SELECT COALESCE(SUM(price * months_overlap), 0) AS total FROM months;
	`
	args = append(args, from.Time(), to.Time())
	var total int64
	if err := r.pool.QueryRow(ctx, q, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func scanSub(row pgx.Row) (*domain.Subscription, error) {
	var s domain.Subscription
	var start, end *time.Time
	err := row.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &start, &end, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if start != nil {
		s.Start = domain.YearMonthFromTime(*start)
	}
	if end != nil {
		ym := domain.YearMonthFromTime(*end)
		s.End = &ym
	}
	return &s, nil
}

func nullableYM(ym *domain.YearMonth) any {
	if ym == nil {
		return nil
	}
	return ym.Time()
}

func itoa(i int) string { return strconv.Itoa(i) }
