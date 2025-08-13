package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const petitionsTable = "petitions"

type GeoPoint struct {
	Lat float64
	Lng float64
}

type PetitionModel struct {
	ID          uuid.UUID  `db:"id"`
	CityID      uuid.UUID  `db:"city_id"`
	Title       string     `db:"title"`
	Description string     `db:"description"`
	InitiatorID uuid.UUID  `db:"initiator_id"`
	AddressToID *uuid.UUID `db:"address_to_id"` // NULLable
	Status      string     `db:"status"`        // petition_status
	Signatures  int        `db:"signatures"`
	Goal        int        `db:"goal"`
	EndDate     time.Time  `db:"end_date"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`

	// Прочитанные из ST_Y/ST_X координаты; если location NULL — поля будут nil
	Lat *float64 `db:"lat"`
	Lng *float64 `db:"lng"`
}

type PetitionsQ struct {
	db       *sql.DB
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewPetitionsQ(db *sql.DB) PetitionsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Явно выбираем колонки + вычисляем lat/lng из geometry
	selectCols := []string{
		"id",
		"city_id",
		"title",
		"description",
		"initiator_id",
		"address_to_id",
		"status",
		"signatures",
		"goal",
		"end_date",
		"created_at",
		"updated_at",
		// PostGIS: сначала долгота (X), потом широта (Y), но для читателя удобнее lat/lng
		"ST_Y(location) AS lat",
		"ST_X(location) AS lng",
	}

	return PetitionsQ{
		db:       db,
		selector: builder.Select(selectCols...).From(petitionsTable),
		inserter: builder.Insert(petitionsTable),
		updater:  builder.Update(petitionsTable),
		deleter:  builder.Delete(petitionsTable),
		counter:  builder.Select("COUNT(*) AS count").From(petitionsTable),
	}
}

func (q PetitionsQ) New() PetitionsQ {
	return NewPetitionsQ(q.db)
}

type InsertPetitionInput struct {
	ID          uuid.UUID
	CityID      uuid.UUID
	Title       string
	Description string
	InitiatorID uuid.UUID
	AddressToID *uuid.UUID
	Status      string
	Signatures  int
	Goal        int
	EndDate     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Location    *GeoPoint // nil -> NULL в БД
}

func (q PetitionsQ) Insert(ctx context.Context, in InsertPetitionInput) error {
	values := map[string]interface{}{
		"id":            in.ID,
		"city_id":       in.CityID,
		"title":         in.Title,
		"description":   in.Description,
		"initiator_id":  in.InitiatorID,
		"address_to_id": in.AddressToID,
		"status":        in.Status,
		"signatures":    in.Signatures,
		"goal":          in.Goal,
		"end_date":      in.EndDate,
		"created_at":    in.CreatedAt,
		"updated_at":    in.UpdatedAt,
	}

	if in.Location != nil {
		// location = ST_SetSRID(ST_MakePoint(lng, lat), 4326)
		values["location"] = sq.Expr("ST_SetSRID(ST_MakePoint(?, ?), 4326)", in.Location.Lng, in.Location.Lat)
	} else {
		values["location"] = nil
	}

	query, args, err := q.inserter.SetMap(values).ToSql()
	if err != nil {
		return fmt.Errorf("building inserter query for table %s: %w", petitionsTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

func (q PetitionsQ) Get(ctx context.Context) (PetitionModel, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return PetitionModel{}, fmt.Errorf("building selector query for table %s: %w", petitionsTable, err)
	}

	var p PetitionModel
	var row *sql.Row
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		row = tx.QueryRowContext(ctx, query, args...)
	} else {
		row = q.db.QueryRowContext(ctx, query, args...)
	}

	err = row.Scan(
		&p.ID,
		&p.CityID,
		&p.Title,
		&p.Description,
		&p.InitiatorID,
		&p.AddressToID,
		&p.Status,
		&p.Signatures,
		&p.Goal,
		&p.EndDate,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.Lat,
		&p.Lng,
	)

	return p, err
}

func (q PetitionsQ) Select(ctx context.Context) ([]PetitionModel, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building selector query for table %s: %w", petitionsTable, err)
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

	var out []PetitionModel
	for rows.Next() {
		var p PetitionModel
		if err := rows.Scan(
			&p.ID,
			&p.CityID,
			&p.Title,
			&p.Description,
			&p.InitiatorID,
			&p.AddressToID,
			&p.Status,
			&p.Signatures,
			&p.Goal,
			&p.EndDate,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.Lat,
			&p.Lng,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

type UpdatePetitionInput struct {
	Title       *string
	Description *string
	AddressToID **uuid.UUID // отличаем "поставить NULL" от "не менять": передайте &ptr, где ptr может быть nil
	Status      *string
	Goal        *int
	EndDate     *time.Time
	UpdatedAt   *time.Time
	Location    *GeoPoint // nil -> не менять; Location с нулями не трогаем — это на уровне бизнес-логики решайте
	// Подписи обычно не правят напрямую — инкремент через отдельный метод (см. ниже)
}

func (q PetitionsQ) Update(ctx context.Context, in UpdatePetitionInput) error {
	updates := map[string]interface{}{}

	if in.Title != nil {
		updates["title"] = *in.Title
	}
	if in.Description != nil {
		updates["description"] = *in.Description
	}
	if in.AddressToID != nil {
		// внутренняя ссылка может быть nil => записать NULL
		updates["address_to_id"] = *in.AddressToID
	}
	if in.Status != nil {
		updates["status"] = *in.Status
	}
	if in.Goal != nil {
		updates["goal"] = *in.Goal
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
		return fmt.Errorf("building updater query for table %s: %w", petitionsTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

func (q PetitionsQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building deleter query for table %s: %w", petitionsTable, err)
	}
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

// -------- Фильтры

func (q PetitionsQ) FilterID(id uuid.UUID) PetitionsQ {
	q.selector = q.selector.Where(sq.Eq{"id": id})
	q.counter = q.counter.Where(sq.Eq{"id": id})
	q.updater = q.updater.Where(sq.Eq{"id": id})
	q.deleter = q.deleter.Where(sq.Eq{"id": id})
	return q
}

func (q PetitionsQ) FilterCityID(cityID uuid.UUID) PetitionsQ {
	q.selector = q.selector.Where(sq.Eq{"city_id": cityID})
	q.counter = q.counter.Where(sq.Eq{"city_id": cityID})
	q.updater = q.updater.Where(sq.Eq{"city_id": cityID})
	q.deleter = q.deleter.Where(sq.Eq{"city_id": cityID})
	return q
}

func (q PetitionsQ) FilterInitiatorID(initiatorID uuid.UUID) PetitionsQ {
	q.selector = q.selector.Where(sq.Eq{"initiator_id": initiatorID})
	q.counter = q.counter.Where(sq.Eq{"initiator_id": initiatorID})
	q.updater = q.updater.Where(sq.Eq{"initiator_id": initiatorID})
	q.deleter = q.deleter.Where(sq.Eq{"initiator_id": initiatorID})
	return q
}

func (q PetitionsQ) FilterStatus(status string) PetitionsQ {
	q.selector = q.selector.Where(sq.Eq{"status": status})
	q.counter = q.counter.Where(sq.Eq{"status": status})
	q.updater = q.updater.Where(sq.Eq{"status": status})
	q.deleter = q.deleter.Where(sq.Eq{"status": status})
	return q
}

func (q PetitionsQ) TitleLike(s string) PetitionsQ {
	p := fmt.Sprintf("%%%s%%", s)
	q.selector = q.selector.Where("title ILIKE ?", p)
	q.counter = q.counter.Where("title ILIKE ?", p)
	return q
}

// Геофильтры (PostGIS)

func (q PetitionsQ) BBox(minLng, minLat, maxLng, maxLat float64) PetitionsQ {
	env := sq.Expr("ST_MakeEnvelope(?, ?, ?, ?, 4326)", minLng, minLat, maxLng, maxLat)
	q.selector = q.selector.Where(sq.Expr("location IS NOT NULL AND ST_Intersects(location, ?)", env))
	q.counter = q.counter.Where(sq.Expr("location IS NOT NULL AND ST_Intersects(location, ?)", env))
	return q
}

func (q PetitionsQ) WithinRadius(lng, lat, radiusMeters float64) PetitionsQ {
	pt := sq.Expr("ST_SetSRID(ST_MakePoint(?, ?), 4326)", lng, lat)
	q.selector = q.selector.Where(sq.Expr("location IS NOT NULL AND ST_DWithin(location::geography, (? )::geography, ?)", pt, radiusMeters))
	q.counter = q.counter.Where(sq.Expr("location IS NOT NULL AND ST_DWithin(location::geography, (? )::geography, ?)", pt, radiusMeters))
	return q
}

// Пагинация и счёт

func (q PetitionsQ) Count(ctx context.Context) (uint64, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for table %s: %w", petitionsTable, err)
	}

	var count uint64
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		err = tx.QueryRowContext(ctx, query, args...).Scan(&count)
	} else {
		err = q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	}

	return count, err
}

func (q PetitionsQ) Page(limit, offset uint64) PetitionsQ {
	q.selector = q.selector.Limit(limit).Offset(offset)
	q.counter = q.counter.Limit(limit).Offset(offset)
	return q
}

// Спец-методы для счётчика подписей (если вы храните агрегат)
func (q PetitionsQ) IncrementSignatures(ctx context.Context, delta int) error {
	query, args, err := q.updater.
		Set("signatures", sq.Expr("GREATEST(signatures + ?, 0)", delta)).
		ToSql()
	if err != nil {
		return fmt.Errorf("building increment signatures query: %w", err)
	}
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}
