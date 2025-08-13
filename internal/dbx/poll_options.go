package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const pollOptionsTable = "poll_options"

type PollOption struct {
	ID         uuid.UUID `db:"id"`
	PollID     uuid.UUID `db:"poll_id"`
	OptionText string    `db:"option_text"`
	VotesCount int       `db:"votes_count"`
	CreatedAt  time.Time `db:"created_at"`
}

type PollOptionsQ struct {
	db       *sql.DB
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewPollOptionsQ(db *sql.DB) PollOptionsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	selectCols := []string{
		"id",
		"poll_id",
		"option_text",
		"votes_count",
		"created_at",
	}

	return PollOptionsQ{
		db:       db,
		selector: builder.Select(selectCols...).From(pollOptionsTable),
		inserter: builder.Insert(pollOptionsTable),
		updater:  builder.Update(pollOptionsTable),
		deleter:  builder.Delete(pollOptionsTable),
		counter:  builder.Select("COUNT(*) AS count").From(pollOptionsTable),
	}
}

func (q PollOptionsQ) New() PollOptionsQ {
	return NewPollOptionsQ(q.db)
}

// ---- Insert

type InsertPollOptionInput struct {
	ID         uuid.UUID
	PollID     uuid.UUID
	OptionText string
	CreatedAt  time.Time
}

func (q PollOptionsQ) Insert(ctx context.Context, in InsertPollOptionInput) error {
	values := map[string]interface{}{
		"id":          in.ID,
		"poll_id":     in.PollID,
		"option_text": in.OptionText,
		"created_at":  in.CreatedAt,
	}

	query, args, err := q.inserter.SetMap(values).ToSql()
	if err != nil {
		return fmt.Errorf("building inserter query for table %s: %w", pollOptionsTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

func (q PollOptionsQ) Get(ctx context.Context) (PollOption, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return PollOption{}, fmt.Errorf("building selector query for table %s: %w", pollOptionsTable, err)
	}

	var po PollOption
	var row *sql.Row
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		row = tx.QueryRowContext(ctx, query, args...)
	} else {
		row = q.db.QueryRowContext(ctx, query, args...)
	}

	err = row.Scan(
		&po.ID,
		&po.PollID,
		&po.OptionText,
		&po.VotesCount,
		&po.CreatedAt,
	)
	return po, err
}

func (q PollOptionsQ) Select(ctx context.Context) ([]PollOption, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building selector query for table %s: %w", pollOptionsTable, err)
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

	var out []PollOption
	for rows.Next() {
		var po PollOption
		if err := rows.Scan(
			&po.ID,
			&po.PollID,
			&po.OptionText,
			&po.VotesCount,
			&po.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, po)
	}
	return out, nil
}

func (q PollOptionsQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building deleter query for table %s: %w", pollOptionsTable, err)
	}
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

// ---- Filters

func (q PollOptionsQ) FilterID(id uuid.UUID) PollOptionsQ {
	q.selector = q.selector.Where(sq.Eq{"id": id})
	q.counter = q.counter.Where(sq.Eq{"id": id})
	q.updater = q.updater.Where(sq.Eq{"id": id})
	q.deleter = q.deleter.Where(sq.Eq{"id": id})
	return q
}

func (q PollOptionsQ) FilterPollID(pollID uuid.UUID) PollOptionsQ {
	q.selector = q.selector.Where(sq.Eq{"poll_id": pollID})
	q.counter = q.counter.Where(sq.Eq{"poll_id": pollID})
	q.updater = q.updater.Where(sq.Eq{"poll_id": pollID})
	q.deleter = q.deleter.Where(sq.Eq{"poll_id": pollID})
	return q
}

func (q PollOptionsQ) OrderByCreatedAsc() PollOptionsQ {
	q.selector = q.selector.OrderBy("created_at ASC")
	return q
}

func (q PollOptionsQ) OrderByCreatedDesc() PollOptionsQ {
	q.selector = q.selector.OrderBy("created_at DESC")
	return q
}

func (q PollOptionsQ) OrderByVotesDesc() PollOptionsQ {
	q.selector = q.selector.OrderBy("votes_count DESC")
	return q
}

// ---- Пагинация и count

func (q PollOptionsQ) Count(ctx context.Context) (uint64, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for table %s: %w", pollOptionsTable, err)
	}
	var c uint64
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		err = tx.QueryRowContext(ctx, query, args...).Scan(&c)
	} else {
		err = q.db.QueryRowContext(ctx, query, args...).Scan(&c)
	}
	return c, err
}

func (q PollOptionsQ) Page(limit, offset uint64) PollOptionsQ {
	q.selector = q.selector.Limit(limit).Offset(offset)
	q.counter = q.counter.Limit(limit).Offset(offset)
	return q
}
