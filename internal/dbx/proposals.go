package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const proposalsTable = "proposals"

// GeoPoint уже есть у тебя рядом с Petitions; если нет — раскомментируй.
// type GeoPoint struct {
// 	Lat float64
// 	Lng float64
// }

type Proposal struct {
	ID           uuid.UUID  `db:"id"`
	CityID       uuid.UUID  `db:"city_id"`
	Title        string     `db:"title"`
	Description  string     `db:"description"`
	Status       string     `db:"status"` // proposal_status
	InitiatorID  uuid.UUID  `db:"initiator_id"`
	AddressToID  *uuid.UUID `db:"address_to_id"` // NULLable
	AgreedNum    int        `db:"agreed_num"`
	DisagreedNum int        `db:"disagreed_num"`
	EndDate      time.Time  `db:"end_date"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`

	// извлекаемыe координаты; если location NULL — будут nil
	Lat *float64 `db:"lat"`
	Lng *float64 `db:"lng"`
}

type ProposalsQ struct {
	db       *sql.DB
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewProposalsQ(db *sql.DB) ProposalsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	selectCols := []string{
		"id",
		"city_id",
		"title",
		"description",
		"status",
		"initiator_id",
		"address_to_id",
		"agreed_num",
		"disagreed_num",
		"end_date",
		"created_at",
		"updated_at",
		"ST_Y(location) AS lat",
		"ST_X(location) AS lng",
	}

	return ProposalsQ{
		db:       db,
		selector: builder.Select(selectCols...).From(proposalsTable),
		inserter: builder.Insert(proposalsTable),
		updater:  builder.Update(proposalsTable),
		deleter:  builder.Delete(proposalsTable),
		counter:  builder.Select("COUNT(*) AS count").From(proposalsTable),
	}
}

func (q ProposalsQ) New() ProposalsQ {
	return NewProposalsQ(q.db)
}

// -------- Insert

type InsertProposalInput struct {
	ID          uuid.UUID
	CityID      uuid.UUID
	Title       string
	Description string
	Status      string
	InitiatorID uuid.UUID
	AddressToID *uuid.UUID
	EndDate     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Location    *GeoPoint // nil -> NULL
	// agreed/disagreed не передаём — по умолчанию 0; считаются триггерами голосов
}

func (q ProposalsQ) Insert(ctx context.Context, in InsertProposalInput) error {
	values := map[string]interface{}{
		"id":            in.ID,
		"city_id":       in.CityID,
		"title":         in.Title,
		"description":   in.Description,
		"status":        in.Status,
		"initiator_id":  in.InitiatorID,
		"address_to_id": in.AddressToID,
		"end_date":      in.EndDate,
		"created_at":    in.CreatedAt,
		"updated_at":    in.UpdatedAt,
	}

	if in.Location != nil {
		values["location"] = sq.Expr("ST_SetSRID(ST_MakePoint(?, ?), 4326)", in.Location.Lng, in.Location.Lat)
	} else {
		values["location"] = nil
	}

	query, args, err := q.inserter.SetMap(values).ToSql()
	if err != nil {
		return fmt.Errorf("building inserter query for table %s: %w", proposalsTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

// -------- Read

func (q ProposalsQ) Get(ctx context.Context) (Proposal, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return Proposal{}, fmt.Errorf("building selector query for table %s: %w", proposalsTable, err)
	}

	var m Proposal
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
		&m.AddressToID,
		&m.AgreedNum,
		&m.DisagreedNum,
		&m.EndDate,
		&m.CreatedAt,
		&m.UpdatedAt,
		&m.Lat,
		&m.Lng,
	)
	return m, err
}

func (q ProposalsQ) Select(ctx context.Context) ([]Proposal, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building selector query for table %s: %w", proposalsTable, err)
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

	var out []Proposal
	for rows.Next() {
		var m Proposal
		if err := rows.Scan(
			&m.ID,
			&m.CityID,
			&m.Title,
			&m.Description,
			&m.Status,
			&m.InitiatorID,
			&m.AddressToID,
			&m.AgreedNum,
			&m.DisagreedNum,
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

// -------- Update (без изменения agreed/disagreed)

type UpdateProposalInput struct {
	Title       *string
	Description *string
	Status      *string
	AddressToID **uuid.UUID // отличаем "NULL" от "не менять"
	EndDate     *time.Time
	UpdatedAt   *time.Time
	Location    *GeoPoint // nil -> не менять
}

func (q ProposalsQ) Update(ctx context.Context, in UpdateProposalInput) error {
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
	if in.AddressToID != nil {
		updates["address_to_id"] = *in.AddressToID // может быть nil => NULL
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
		return fmt.Errorf("building updater query for table %s: %w", proposalsTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

// -------- Delete

func (q ProposalsQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building deleter query for table %s: %w", proposalsTable, err)
	}
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

// -------- Фильтры

func (q ProposalsQ) FilterID(id uuid.UUID) ProposalsQ {
	q.selector = q.selector.Where(sq.Eq{"id": id})
	q.counter = q.counter.Where(sq.Eq{"id": id})
	q.updater = q.updater.Where(sq.Eq{"id": id})
	q.deleter = q.deleter.Where(sq.Eq{"id": id})
	return q
}

func (q ProposalsQ) FilterCityID(cityID uuid.UUID) ProposalsQ {
	q.selector = q.selector.Where(sq.Eq{"city_id": cityID})
	q.counter = q.counter.Where(sq.Eq{"city_id": cityID})
	q.updater = q.updater.Where(sq.Eq{"city_id": cityID})
	q.deleter = q.deleter.Where(sq.Eq{"city_id": cityID})
	return q
}

func (q ProposalsQ) FilterInitiatorID(initiatorID uuid.UUID) ProposalsQ {
	q.selector = q.selector.Where(sq.Eq{"initiator_id": initiatorID})
	q.counter = q.counter.Where(sq.Eq{"initiator_id": initiatorID})
	q.updater = q.updater.Where(sq.Eq{"initiator_id": initiatorID})
	q.deleter = q.deleter.Where(sq.Eq{"initiator_id": initiatorID})
	return q
}

func (q ProposalsQ) FilterStatus(status string) ProposalsQ {
	q.selector = q.selector.Where(sq.Eq{"status": status})
	q.counter = q.counter.Where(sq.Eq{"status": status})
	q.updater = q.updater.Where(sq.Eq{"status": status})
	q.deleter = q.deleter.Where(sq.Eq{"status": status})
	return q
}

func (q ProposalsQ) FilterAddressedTo(addressToID uuid.UUID) ProposalsQ {
	q.selector = q.selector.Where(sq.Eq{"address_to_id": addressToID})
	q.counter = q.counter.Where(sq.Eq{"address_to_id": addressToID})
	q.updater = q.updater.Where(sq.Eq{"address_to_id": addressToID})
	q.deleter = q.deleter.Where(sq.Eq{"address_to_id": addressToID})
	return q
}

func (q ProposalsQ) FilterAddressedToCityGov() ProposalsQ {
	// address_to_id IS NULL
	q.selector = q.selector.Where("address_to_id IS NULL")
	q.counter = q.counter.Where("address_to_id IS NULL")
	q.updater = q.updater.Where("address_to_id IS NULL")
	q.deleter = q.deleter.Where("address_to_id IS NULL")
	return q
}

func (q ProposalsQ) TitleLike(s string) ProposalsQ {
	p := fmt.Sprintf("%%%s%%", s)
	q.selector = q.selector.Where("title ILIKE ?", p)
	q.counter = q.counter.Where("title ILIKE ?", p)
	return q
}

// -------- Геофильтры

func (q ProposalsQ) BBox(minLng, minLat, maxLng, maxLat float64) ProposalsQ {
	env := sq.Expr("ST_MakeEnvelope(?, ?, ?, ?, 4326)", minLng, minLat, maxLng, maxLat)
	q.selector = q.selector.Where(sq.Expr("location IS NOT NULL AND ST_Intersects(location, ?)", env))
	q.counter = q.counter.Where(sq.Expr("location IS NOT NULL AND ST_Intersects(location, ?)", env))
	return q
}

func (q ProposalsQ) WithinRadius(lng, lat, radiusMeters float64) ProposalsQ {
	pt := sq.Expr("ST_SetSRID(ST_MakePoint(?, ?), 4326)", lng, lat)
	q.selector = q.selector.Where(sq.Expr("location IS NOT NULL AND ST_DWithin(location::geography, (? )::geography, ?)", pt, radiusMeters))
	q.counter = q.counter.Where(sq.Expr("location IS NOT NULL AND ST_DWithin(location::geography, (? )::geography, ?)", pt, radiusMeters))
	return q
}

// -------- Сортировки

func (q ProposalsQ) OrderByCreatedAsc() ProposalsQ {
	q.selector = q.selector.OrderBy("created_at ASC")
	return q
}

func (q ProposalsQ) OrderByCreatedDesc() ProposalsQ {
	q.selector = q.selector.OrderBy("created_at DESC")
	return q
}

func (q ProposalsQ) OrderByAgreedDesc() ProposalsQ {
	q.selector = q.selector.OrderBy("agreed_num DESC")
	return q
}

func (q ProposalsQ) OrderByDisagreedDesc() ProposalsQ {
	q.selector = q.selector.OrderBy("disagreed_num DESC")
	return q
}

// -------- Пагинация и Count

func (q ProposalsQ) Page(limit, offset uint64) ProposalsQ {
	q.selector = q.selector.Limit(limit).Offset(offset)
	q.counter = q.counter.Limit(limit).Offset(offset)
	return q
}

func (q ProposalsQ) Count(ctx context.Context) (uint64, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for table %s: %w", proposalsTable, err)
	}
	var c uint64
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		err = tx.QueryRowContext(ctx, query, args...).Scan(&c)
	} else {
		err = q.db.QueryRowContext(ctx, query, args...).Scan(&c)
	}
	return c, err
}
