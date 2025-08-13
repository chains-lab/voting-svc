package entities

import (
	"context"

	"github.com/chains-lab/voting-svc/internal/dbx"
	"github.com/google/uuid"
)

type pollsQ interface {
	New() dbx.PollsQ

	Insert(ctx context.Context, in dbx.InsertPollInput) error
	Update(ctx context.Context, in dbx.UpdatePollInput) error
	Delete(ctx context.Context) error
	Get(ctx context.Context) (dbx.Poll, error)
	Select(ctx context.Context) ([]dbx.Poll, error)

	FilterID(id uuid.UUID) dbx.PollsQ
	FilterCityID(cityID uuid.UUID) dbx.PollsQ
	FilterInitiatorID(initiatorID uuid.UUID) dbx.PollsQ
	FilterStatus(status string) dbx.PollsQ
	TitleLike(s string) dbx.PollsQ

	BBox(minLng, minLat, maxLng, maxLat float64) dbx.PollsQ
	WithinRadius(lng, lat, radiusMeters float64) dbx.PollsQ

	Count(ctx context.Context) (uint64, error)
	Page(limit, offset uint64) dbx.PollsQ
}

type pollOptionsQ interface {
	New() dbx.PollOptionsQ

	Insert(ctx context.Context, in dbx.InsertPollOptionInput) error
	Get(ctx context.Context) (dbx.PollOption, error)
	Select(ctx context.Context) ([]dbx.PollOption, error)
	Delete(ctx context.Context) error

	FilterID(id uuid.UUID) dbx.PollOptionsQ
	FilterPollID(pollID uuid.UUID) dbx.PollOptionsQ

	OrderByCreatedAsc() dbx.PollOptionsQ
	OrderByCreatedDesc() dbx.PollOptionsQ
	OrderByVotesDesc() dbx.PollOptionsQ

	Count(ctx context.Context) (uint64, error)
	Page(limit, offset uint64) dbx.PollOptionsQ
}

type pollVotesQ interface {
	New() dbx.PollVotesQ

	Insert(ctx context.Context, in dbx.InsertPollVoteInput) error
	Get(ctx context.Context) (dbx.PollVote, error)
	Select(ctx context.Context) ([]dbx.PollVote, error)
	Update(ctx context.Context, in dbx.UpdatePollVoteInput) error
	Delete(ctx context.Context) error

	FilterID(id uuid.UUID) dbx.PollVotesQ
	FilterPollID(pollID uuid.UUID) dbx.PollVotesQ
	FilterUserID(userID uuid.UUID) dbx.PollVotesQ
	FilterOptionID(optionID uuid.UUID) dbx.PollVotesQ

	OrderByCreatedAsc() dbx.PollVotesQ
	OrderByCreatedDesc() dbx.PollVotesQ

	Count(ctx context.Context) (uint64, error)
	Page(limit, offset uint64) dbx.PollVotesQ
}
