package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const proposalVotesTable = "proposal_votes"

type ProposalVoteModel struct {
	ID         uuid.UUID `db:"id"`
	ProposalID uuid.UUID `db:"proposal_id"`
	UserID     uuid.UUID `db:"user_id"`
	Vote       bool      `db:"vote"` // true = agree, false = disagree
	CreatedAt  time.Time `db:"created_at"`
}

type ProposalVotesQ struct {
	db       *sql.DB
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewProposalVotesQ(db *sql.DB) ProposalVotesQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	selectCols := []string{
		"id",
		"proposal_id",
		"user_id",
		"vote",
		"created_at",
	}

	return ProposalVotesQ{
		db:       db,
		selector: builder.Select(selectCols...).From(proposalVotesTable),
		inserter: builder.Insert(proposalVotesTable),
		updater:  builder.Update(proposalVotesTable),
		deleter:  builder.Delete(proposalVotesTable),
		counter:  builder.Select("COUNT(*) AS count").From(proposalVotesTable),
	}
}

func (q ProposalVotesQ) New() ProposalVotesQ {
	return NewProposalVotesQ(q.db)
}

// ---- Insert

type InsertProposalVoteInput struct {
	ID         uuid.UUID
	ProposalID uuid.UUID
	UserID     uuid.UUID
	Vote       bool
	CreatedAt  time.Time
}

func (q ProposalVotesQ) Insert(ctx context.Context, in InsertProposalVoteInput) error {
	values := map[string]interface{}{
		"id":          in.ID,
		"proposal_id": in.ProposalID,
		"user_id":     in.UserID,
		"vote":        in.Vote,
		"created_at":  in.CreatedAt,
	}

	query, args, err := q.inserter.SetMap(values).ToSql()
	if err != nil {
		return fmt.Errorf("building inserter query for table %s: %w", proposalVotesTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

// ---- Read

func (q ProposalVotesQ) Get(ctx context.Context) (ProposalVoteModel, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return ProposalVoteModel{}, fmt.Errorf("building selector query for table %s: %w", proposalVotesTable, err)
	}

	var m ProposalVoteModel
	var row *sql.Row
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		row = tx.QueryRowContext(ctx, query, args...)
	} else {
		row = q.db.QueryRowContext(ctx, query, args...)
	}
	err = row.Scan(
		&m.ID,
		&m.ProposalID,
		&m.UserID,
		&m.Vote,
		&m.CreatedAt,
	)
	return m, err
}

func (q ProposalVotesQ) Select(ctx context.Context) ([]ProposalVoteModel, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building selector query for table %s: %w", proposalVotesTable, err)
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

	var out []ProposalVoteModel
	for rows.Next() {
		var m ProposalVoteModel
		if err := rows.Scan(
			&m.ID,
			&m.ProposalID,
			&m.UserID,
			&m.Vote,
			&m.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

// ---- Update
// Меняем только vote. created_at/ids не трогаем.
// NB: действует UNIQUE(proposal_id, user_id)
type UpdateProposalVoteInput struct {
	Vote *bool
}

func (q ProposalVotesQ) Update(ctx context.Context, in UpdateProposalVoteInput) error {
	updates := map[string]interface{}{}
	if in.Vote != nil {
		updates["vote"] = *in.Vote
	}
	if len(updates) == 0 {
		return nil
	}

	query, args, err := q.updater.SetMap(updates).ToSql()
	if err != nil {
		return fmt.Errorf("building updater query for table %s: %w", proposalVotesTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

// ---- Delete

func (q ProposalVotesQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building deleter query for table %s: %w", proposalVotesTable, err)
	}
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

// ---- Filters

func (q ProposalVotesQ) FilterID(id uuid.UUID) ProposalVotesQ {
	q.selector = q.selector.Where(sq.Eq{"id": id})
	q.counter = q.counter.Where(sq.Eq{"id": id})
	q.updater = q.updater.Where(sq.Eq{"id": id})
	q.deleter = q.deleter.Where(sq.Eq{"id": id})
	return q
}

func (q ProposalVotesQ) FilterProposalID(proposalID uuid.UUID) ProposalVotesQ {
	q.selector = q.selector.Where(sq.Eq{"proposal_id": proposalID})
	q.counter = q.counter.Where(sq.Eq{"proposal_id": proposalID})
	q.updater = q.updater.Where(sq.Eq{"proposal_id": proposalID})
	q.deleter = q.deleter.Where(sq.Eq{"proposal_id": proposalID})
	return q
}

func (q ProposalVotesQ) FilterUserID(userID uuid.UUID) ProposalVotesQ {
	q.selector = q.selector.Where(sq.Eq{"user_id": userID})
	q.counter = q.counter.Where(sq.Eq{"user_id": userID})
	q.updater = q.updater.Where(sq.Eq{"user_id": userID})
	q.deleter = q.deleter.Where(sq.Eq{"user_id": userID})
	return q
}

func (q ProposalVotesQ) FilterVote(v bool) ProposalVotesQ {
	q.selector = q.selector.Where(sq.Eq{"vote": v})
	q.counter = q.counter.Where(sq.Eq{"vote": v})
	q.updater = q.updater.Where(sq.Eq{"vote": v})
	q.deleter = q.deleter.Where(sq.Eq{"vote": v})
	return q
}

// ---- Сортировки и пагинация

func (q ProposalVotesQ) OrderByCreatedAsc() ProposalVotesQ {
	q.selector = q.selector.OrderBy("created_at ASC")
	return q
}

func (q ProposalVotesQ) OrderByCreatedDesc() ProposalVotesQ {
	q.selector = q.selector.OrderBy("created_at DESC")
	return q
}

func (q ProposalVotesQ) Page(limit, offset uint64) ProposalVotesQ {
	q.selector = q.selector.Limit(limit).Offset(offset)
	q.counter = q.counter.Limit(limit).Offset(offset)
	return q
}

// ---- Count

func (q ProposalVotesQ) Count(ctx context.Context) (uint64, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for table %s: %w", proposalVotesTable, err)
	}
	var c uint64
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		err = tx.QueryRowContext(ctx, query, args...).Scan(&c)
	} else {
		err = q.db.QueryRowContext(ctx, query, args...).Scan(&c)
	}
	return c, err
}
