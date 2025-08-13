package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const petitionSignaturesTable = "petition_signatures"

type PetitionSignature struct {
	ID         uuid.UUID `db:"id"`
	PetitionID uuid.UUID `db:"petition_id"`
	UserID     uuid.UUID `db:"user_id"`
	CreatedAt  time.Time `db:"created_at"`
}

type PetitionSignaturesQ struct {
	db       *sql.DB
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewPetitionSignaturesQ(db *sql.DB) PetitionSignaturesQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return PetitionSignaturesQ{
		db:       db,
		selector: builder.Select("*").From(petitionSignaturesTable),
		inserter: builder.Insert(petitionSignaturesTable),
		updater:  builder.Update(petitionSignaturesTable),
		deleter:  builder.Delete(petitionSignaturesTable),
		counter:  builder.Select("COUNT(*) AS count").From(petitionSignaturesTable),
	}
}

func (q PetitionSignaturesQ) New() PetitionSignaturesQ {
	return NewPetitionSignaturesQ(q.db)
}

// Insert — строгая вставка; вернёт ошибку при нарушении UNIQUE(petition_id,user_id).
func (q PetitionSignaturesQ) Insert(ctx context.Context, input PetitionSignature) error {
	values := map[string]interface{}{
		"id":          input.ID,
		"petition_id": input.PetitionID,
		"user_id":     input.UserID,
		"created_at":  input.CreatedAt,
	}
	query, args, err := q.inserter.SetMap(values).ToSql()
	if err != nil {
		return fmt.Errorf("building inserter query for table: %s input: %w", petitionSignaturesTable, err)
	}
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

func (q PetitionSignaturesQ) Get(ctx context.Context) (PetitionSignature, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return PetitionSignature{}, fmt.Errorf("building selector query for table: %s: %w", petitionSignaturesTable, err)
	}

	var s PetitionSignature
	var row *sql.Row
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		row = tx.QueryRowContext(ctx, query, args...)
	} else {
		row = q.db.QueryRowContext(ctx, query, args...)
	}

	err = row.Scan(
		&s.ID,
		&s.PetitionID,
		&s.UserID,
		&s.CreatedAt,
	)

	return s, err
}

func (q PetitionSignaturesQ) Select(ctx context.Context) ([]PetitionSignature, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building selector query for table: %s: %w", petitionSignaturesTable, err)
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

	var out []PetitionSignature

	for rows.Next() {
		var s PetitionSignature
		if err := rows.Scan(
			&s.ID,
			&s.PetitionID,
			&s.UserID,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}

	return out, nil
}

func (q PetitionSignaturesQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building deleter query for table: %s: %w", petitionSignaturesTable, err)
	}
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}

	return err
}

// -------- Фильтры

func (q PetitionSignaturesQ) FilterID(id uuid.UUID) PetitionSignaturesQ {
	q.selector = q.selector.Where(sq.Eq{"id": id})
	q.counter = q.counter.Where(sq.Eq{"id": id})
	q.updater = q.updater.Where(sq.Eq{"id": id})
	q.deleter = q.deleter.Where(sq.Eq{"id": id})
	return q
}

func (q PetitionSignaturesQ) FilterPetitionID(petitionID uuid.UUID) PetitionSignaturesQ {
	q.selector = q.selector.Where(sq.Eq{"petition_id": petitionID})
	q.counter = q.counter.Where(sq.Eq{"petition_id": petitionID})
	q.updater = q.updater.Where(sq.Eq{"petition_id": petitionID})
	q.deleter = q.deleter.Where(sq.Eq{"petition_id": petitionID})
	return q
}

func (q PetitionSignaturesQ) FilterUserID(userID uuid.UUID) PetitionSignaturesQ {
	q.selector = q.selector.Where(sq.Eq{"user_id": userID})
	q.counter = q.counter.Where(sq.Eq{"user_id": userID})
	q.updater = q.updater.Where(sq.Eq{"user_id": userID})
	q.deleter = q.deleter.Where(sq.Eq{"user_id": userID})
	return q
}

// Count — для пагинации
func (q PetitionSignaturesQ) Count(ctx context.Context) (uint64, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for table: %s: %w", petitionSignaturesTable, err)
	}

	var count uint64
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		err = tx.QueryRowContext(ctx, query, args...).Scan(&count)
	} else {
		err = q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	}

	return count, err
}

func (q PetitionSignaturesQ) Page(limit, offset uint64) PetitionSignaturesQ {
	q.selector = q.selector.Limit(limit).Offset(offset)
	q.counter = q.counter.Limit(limit).Offset(offset)
	return q
}
