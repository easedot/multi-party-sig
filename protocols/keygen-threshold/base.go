package keygen_threshold

import (
	"github.com/taurusgroup/cmp-ecdsa/pkg/party"
	"github.com/taurusgroup/cmp-ecdsa/pkg/round"
)

func NewRound(session *round.Session) (*round1, error) {
	// Create round with a clone of the original secret
	base, err := round.NewBaseRound(session)
	if err != nil {
		return nil, err
	}

	parties := make(map[party.ID]*localParty, base.S.N())
	for j, publicJ := range base.S.Public {
		// Set the public data to a clone of the current data
		parties[j] = &localParty{
			Party: round.NewBaseParty(publicJ),
		}
	}

	return &round1{
		BaseRound: base,
		thisParty: parties[base.S.SelfID()],
		parties:   parties,
	}, nil
}