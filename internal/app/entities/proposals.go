package entities

import (
	"context"

	"github.com/chains-lab/voting-svc/internal/dbx"
	"github.com/google/uuid"
)

type proposalsQ interface {
	New() dbx.ProposalsQ

	Insert(ctx context.Context, in dbx.InsertProposalInput) error
	Get(ctx context.Context) (dbx.Proposal, error)
	Select(ctx context.Context) ([]dbx.Proposal, error)
	Update(ctx context.Context, in dbx.UpdateProposalInput) error
	Delete(ctx context.Context) error

	FilterID(id uuid.UUID) dbx.ProposalsQ
	FilterCityID(cityID uuid.UUID) dbx.ProposalsQ
	FilterInitiatorID(initiatorID uuid.UUID) dbx.ProposalsQ
	FilterStatus(status string) dbx.ProposalsQ
	FilterAddressedTo(addressToID uuid.UUID) dbx.ProposalsQ
	FilterAddressedToCityGov() dbx.ProposalsQ

	TitleLike(s string) dbx.ProposalsQ

	BBox(minLng, minLat, maxLng, maxLat float64) dbx.ProposalsQ
	WithinRadius(lng, lat, radiusMeters float64) dbx.ProposalsQ

	OrderByCreatedAsc() dbx.ProposalsQ
	OrderByCreatedDesc() dbx.ProposalsQ
	OrderByAgreedDesc() dbx.ProposalsQ
	OrderByDisagreedDesc() dbx.ProposalsQ

	Page(limit, offset uint64) dbx.ProposalsQ
	Count(ctx context.Context) (uint64, error)
}

type proposalVotesQ interface {
	New() dbx.ProposalVotesQ

	Insert(ctx context.Context, in dbx.InsertProposalVoteInput) error
	Get(ctx context.Context) (dbx.ProposalVote, error)
	Select(ctx context.Context) ([]dbx.ProposalVote, error)
	Update(ctx context.Context, in dbx.UpdateProposalVoteInput) error
	Delete(ctx context.Context) error

	FilterID(id uuid.UUID) dbx.ProposalVotesQ
	FilterProposalID(proposalID uuid.UUID) dbx.ProposalVotesQ
	FilterUserID(userID uuid.UUID) dbx.ProposalVotesQ
	FilterVote(v bool) dbx.ProposalVotesQ

	OrderByCreatedAsc() dbx.ProposalVotesQ
	OrderByCreatedDesc() dbx.ProposalVotesQ

	Page(limit, offset uint64) dbx.ProposalVotesQ
	Count(ctx context.Context) (uint64, error)
}
