package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const pollVotesTable = "poll_votes"

type PollVoteModel struct {
	ID        uuid.UUID `db:"id"`
	PollID    uuid.UUID `db:"poll_id"`
	UserID    uuid.UUID `db:"user_id"`
	OptionID  uuid.UUID `db:"option_id"`
	CreatedAt time.Time `db:"created_at"`
}

type PollVotesQ struct {
	db       *sql.DB
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewPollVotesQ(db *sql.DB) PollVotesQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	selectCols := []string{
		"id",
		"poll_id",
		"user_id",
		"option_id",
		"created_at",
	}

	return PollVotesQ{
		db:       db,
		selector: builder.Select(selectCols...).From(pollVotesTable),
		inserter: builder.Insert(pollVotesTable),
		updater:  builder.Update(pollVotesTable),
		deleter:  builder.Delete(pollVotesTable),
		counter:  builder.Select("COUNT(*) AS count").From(pollVotesTable),
	}
}

func (q PollVotesQ) New() PollVotesQ {
	return NewPollVotesQ(q.db)
}

// ---- Insert

type InsertPollVoteInput struct {
	ID        uuid.UUID
	PollID    uuid.UUID
	UserID    uuid.UUID
	OptionID  uuid.UUID
	CreatedAt time.Time
}

func (q PollVotesQ) Insert(ctx context.Context, in InsertPollVoteInput) error {
	values := map[string]interface{}{
		"id":         in.ID,
		"poll_id":    in.PollID,
		"user_id":    in.UserID,
		"option_id":  in.OptionID,
		"created_at": in.CreatedAt,
	}

	query, args, err := q.inserter.SetMap(values).ToSql()
	if err != nil {
		return fmt.Errorf("building inserter query for table %s: %w", pollVotesTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

// ---- Read

func (q PollVotesQ) Get(ctx context.Context) (PollVoteModel, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return PollVoteModel{}, fmt.Errorf("building selector query for table %s: %w", pollVotesTable, err)
	}

	var pv PollVoteModel
	var row *sql.Row
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		row = tx.QueryRowContext(ctx, query, args...)
	} else {
		row = q.db.QueryRowContext(ctx, query, args...)
	}

	err = row.Scan(
		&pv.ID,
		&pv.PollID,
		&pv.UserID,
		&pv.OptionID,
		&pv.CreatedAt,
	)
	return pv, err
}

func (q PollVotesQ) Select(ctx context.Context) ([]PollVoteModel, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building selector query for table %s: %w", pollVotesTable, err)
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

	var out []PollVoteModel
	for rows.Next() {
		var pv PollVoteModel
		if err := rows.Scan(
			&pv.ID,
			&pv.PollID,
			&pv.UserID,
			&pv.OptionID,
			&pv.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, pv)
	}
	return out, nil
}

// ---- Update
// Разрешаем менять option_id и/или poll_id. created_at не трогаем.
// NB: действует UNIQUE(poll_id, user_id) — при смене poll_id возможен конфликт.
type UpdatePollVoteInput struct {
	OptionID *uuid.UUID
}

func (q PollVotesQ) Update(ctx context.Context, in UpdatePollVoteInput) error {
	updates := map[string]interface{}{}
	if in.OptionID != nil {
		updates["option_id"] = *in.OptionID
	}
	if len(updates) == 0 {
		return nil
	}

	query, args, err := q.updater.SetMap(updates).ToSql()
	if err != nil {
		return fmt.Errorf("building updater query for table %s: %w", pollVotesTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

// ---- Delete

func (q PollVotesQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building deleter query for table %s: %w", pollVotesTable, err)
	}
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

// ---- Filters

func (q PollVotesQ) FilterID(id uuid.UUID) PollVotesQ {
	q.selector = q.selector.Where(sq.Eq{"id": id})
	q.counter = q.counter.Where(sq.Eq{"id": id})
	q.updater = q.updater.Where(sq.Eq{"id": id})
	q.deleter = q.deleter.Where(sq.Eq{"id": id})
	return q
}

func (q PollVotesQ) FilterPollID(pollID uuid.UUID) PollVotesQ {
	q.selector = q.selector.Where(sq.Eq{"poll_id": pollID})
	q.counter = q.counter.Where(sq.Eq{"poll_id": pollID})
	q.updater = q.updater.Where(sq.Eq{"poll_id": pollID})
	q.deleter = q.deleter.Where(sq.Eq{"poll_id": pollID})
	return q
}

func (q PollVotesQ) FilterUserID(userID uuid.UUID) PollVotesQ {
	q.selector = q.selector.Where(sq.Eq{"user_id": userID})
	q.counter = q.counter.Where(sq.Eq{"user_id": userID})
	q.updater = q.updater.Where(sq.Eq{"user_id": userID})
	q.deleter = q.deleter.Where(sq.Eq{"user_id": userID})
	return q
}

func (q PollVotesQ) FilterOptionID(optionID uuid.UUID) PollVotesQ {
	q.selector = q.selector.Where(sq.Eq{"option_id": optionID})
	q.counter = q.counter.Where(sq.Eq{"option_id": optionID})
	q.updater = q.updater.Where(sq.Eq{"option_id": optionID})
	q.deleter = q.deleter.Where(sq.Eq{"option_id": optionID})
	return q
}

// ---- Сортировки и пагинация

func (q PollVotesQ) OrderByCreatedAsc() PollVotesQ {
	q.selector = q.selector.OrderBy("created_at ASC")
	return q
}

func (q PollVotesQ) OrderByCreatedDesc() PollVotesQ {
	q.selector = q.selector.OrderBy("created_at DESC")
	return q
}

func (q PollVotesQ) Page(limit, offset uint64) PollVotesQ {
	q.selector = q.selector.Limit(limit).Offset(offset)
	q.counter = q.counter.Limit(limit).Offset(offset)
	return q
}

// ---- Count

func (q PollVotesQ) Count(ctx context.Context) (uint64, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for table %s: %w", pollVotesTable, err)
	}
	var c uint64
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		err = tx.QueryRowContext(ctx, query, args...).Scan(&c)
	} else {
		err = q.db.QueryRowContext(ctx, query, args...).Scan(&c)
	}
	return c, err
}
