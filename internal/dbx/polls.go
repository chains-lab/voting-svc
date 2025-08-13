package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const pollsTable = "polls"

type PollModel struct {
	ID          uuid.UUID `db:"id"`
	CityID      uuid.UUID `db:"city_id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Status      string    `db:"status"` // poll_status
	InitiatorID uuid.UUID `db:"initiator_id"`
	EndDate     time.Time `db:"end_date"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`

	// Гео, прочитанное как lat/lng; если location NULL — поля будут nil
	Lat *float64 `db:"lat"`
	Lng *float64 `db:"lng"`
}

type PollsQ struct {
	db       *sql.DB
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewPollsQ(db *sql.DB) PollsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	selectCols := []string{
		"id",
		"city_id",
		"title",
		"description",
		"status",
		"initiator_id",
		"end_date",
		"created_at",
		"updated_at",
		"ST_Y(location) AS lat",
		"ST_X(location) AS lng",
	}

	return PollsQ{
		db:       db,
		selector: builder.Select(selectCols...).From(pollsTable),
		inserter: builder.Insert(pollsTable),
		updater:  builder.Update(pollsTable),
		deleter:  builder.Delete(pollsTable),
		counter:  builder.Select("COUNT(*) AS count").From(pollsTable),
	}
}

func (q PollsQ) New() PollsQ { return NewPollsQ(q.db) }

// ---------- Commands

type InsertPollInput struct {
	ID          uuid.UUID
	CityID      uuid.UUID
	Title       string
	Description string
	Status      string
	InitiatorID uuid.UUID
	EndDate     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Location    *GeoPoint // nil => NULL
}

func (q PollsQ) Insert(ctx context.Context, in InsertPollInput) error {
	values := map[string]interface{}{
		"id":           in.ID,
		"city_id":      in.CityID,
		"title":        in.Title,
		"description":  in.Description,
		"status":       in.Status,
		"initiator_id": in.InitiatorID,
		"end_date":     in.EndDate,
		"created_at":   in.CreatedAt,
		"updated_at":   in.UpdatedAt,
	}
	if in.Location != nil {
		values["location"] = sq.Expr("ST_SetSRID(ST_MakePoint(?, ?), 4326)", in.Location.Lng, in.Location.Lat)
	} else {
		values["location"] = nil
	}

	query, args, err := q.inserter.SetMap(values).ToSql()
	if err != nil {
		return fmt.Errorf("building inserter query for table %s: %w", pollsTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

type UpdatePollInput struct {
	Title       *string
	Description *string
	Status      *string
	EndDate     *time.Time
	UpdatedAt   *time.Time
	Location    *GeoPoint // nil => не менять
}

func (q PollsQ) Update(ctx context.Context, in UpdatePollInput) error {
	updates := map[string]interface{}{}

	if in.Title != nil {
		updates["title"] = *in.Title
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}
	if in.Status != nil {
		updates["status"] = *in.Status
	}
	if in.EndDate != nil {
		updates["end_date"] = *in.EndDate
	}
	if in.UpdatedAt != nil {
		updates["updated_at"] = *in.UpdatedAt
	}
	if in.Location != nil {
		updates["location"] = sq.Expr("ST_SetSRID(ST_MakePoint(?, ?), 4326)", in.Location.Lng, in.Location.Lat)
	}

	if len(updates) == 0 {
		return nil
	}

	query, args, err := q.updater.SetMap(updates).ToSql()
	if err != nil {
		return fmt.Errorf("building updater query for table %s: %w", pollsTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

func (q PollsQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building deleter query for table %s: %w", pollsTable, err)
	}
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

// ---------- Queries

func (q PollsQ) Get(ctx context.Context) (PollModel, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return PollModel{}, fmt.Errorf("building selector query for table %s: %w", pollsTable, err)
	}

	var m PollModel
	var row *sql.Row
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		row = tx.QueryRowContext(ctx, query, args...)
	} else {
		row = q.db.QueryRowContext(ctx, query, args...)
	}

	err = row.Scan(
		&m.ID,
		&m.CityID,
		&m.Title,
		&m.Description,
		&m.Status,
		&m.InitiatorID,
		&m.EndDate,
		&m.CreatedAt,
		&m.UpdatedAt,
		&m.Lat,
		&m.Lng,
	)
	return m, err
}

func (q PollsQ) Select(ctx context.Context) ([]PollModel, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building selector query for table %s: %w", pollsTable, err)
	}

	var rows *sql.Rows
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		rows, err = tx.QueryContext(ctx, query, args...)
	} else {
		rows, err = q.db.QueryContext(ctx, query, args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PollModel
	for rows.Next() {
		var m PollModel
		if err := rows.Scan(
			&m.ID,
			&m.CityID,
			&m.Title,
			&m.Description,
			&m.Status,
			&m.InitiatorID,
			&m.EndDate,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.Lat,
			&m.Lng,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

// ---------- Filters

func (q PollsQ) FilterID(id uuid.UUID) PollsQ {
	q.selector = q.selector.Where(sq.Eq{"id": id})
	q.counter = q.counter.Where(sq.Eq{"id": id})
	q.updater = q.updater.Where(sq.Eq{"id": id})
	q.deleter = q.deleter.Where(sq.Eq{"id": id})
	return q
}

func (q PollsQ) FilterCityID(cityID uuid.UUID) PollsQ {
	q.selector = q.selector.Where(sq.Eq{"city_id": cityID})
	q.counter = q.counter.Where(sq.Eq{"city_id": cityID})
	q.updater = q.updater.Where(sq.Eq{"city_id": cityID})
	q.deleter = q.deleter.Where(sq.Eq{"city_id": cityID})
	return q
}

func (q PollsQ) FilterInitiatorID(initiatorID uuid.UUID) PollsQ {
	q.selector = q.selector.Where(sq.Eq{"initiator_id": initiatorID})
	q.counter = q.counter.Where(sq.Eq{"initiator_id": initiatorID})
	q.updater = q.updater.Where(sq.Eq{"initiator_id": initiatorID})
	q.deleter = q.deleter.Where(sq.Eq{"initiator_id": initiatorID})
	return q
}

func (q PollsQ) FilterStatus(status string) PollsQ {
	q.selector = q.selector.Where(sq.Eq{"status": status})
	q.counter = q.counter.Where(sq.Eq{"status": status})
	q.updater = q.updater.Where(sq.Eq{"status": status})
	q.deleter = q.deleter.Where(sq.Eq{"status": status})
	return q
}

func (q PollsQ) TitleLike(s string) PollsQ {
	p := fmt.Sprintf("%%%s%%", s)
	q.selector = q.selector.Where("title ILIKE ?", p)
	q.counter = q.counter.Where("title ILIKE ?", p)
	return q
}

// Гео-фильтры (как в petitions)

func (q PollsQ) BBox(minLng, minLat, maxLng, maxLat float64) PollsQ {
	env := sq.Expr("ST_MakeEnvelope(?, ?, ?, ?, 4326)", minLng, minLat, maxLng, maxLat)
	q.selector = q.selector.Where(sq.Expr("location IS NOT NULL AND ST_Intersects(location, ?)", env))
	q.counter = q.counter.Where(sq.Expr("location IS NOT NULL AND ST_Intersects(location, ?)", env))
	return q
}

func (q PollsQ) WithinRadius(lng, lat, radiusMeters float64) PollsQ {
	pt := sq.Expr("ST_SetSRID(ST_MakePoint(?, ?), 4326)", lng, lat)
	q.selector = q.selector.Where(sq.Expr("location IS NOT NULL AND ST_DWithin(location::geography, (? )::geography, ?)", pt, radiusMeters))
	q.counter = q.counter.Where(sq.Expr("location IS NOT NULL AND ST_DWithin(location::geography, (? )::geography, ?)", pt, radiusMeters))
	return q
}

// ---------- Pagination

func (q PollsQ) Count(ctx context.Context) (uint64, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for table %s: %w", pollsTable, err)
	}
	var c uint64
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		err = tx.QueryRowContext(ctx, query, args...).Scan(&c)
	} else {
		err = q.db.QueryRowContext(ctx, query, args...).Scan(&c)
	}
	return c, err
}

func (q PollsQ) Page(limit, offset uint64) PollsQ {
	q.selector = q.selector.Limit(limit).Offset(offset)
	q.counter = q.counter.Limit(limit).Offset(offset)
	return q
}
