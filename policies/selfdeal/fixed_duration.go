package selfdeal

import (
	"context"

	"github.com/filecoin-project/specs-actors/actors/abi"

	"github.com/filecoin-project/go-storage-miner/apis/node"
)

type Chain interface {
	GetChainHead(ctx context.Context) (node.TipSetToken, abi.ChainEpoch, error)
}

type FixedDurationPolicy struct {
	api Chain

	// self deal start epoch equals chain head epoch plus delay
	delay abi.ChainEpoch

	// self deal expiry epoch equals chain head plus delay plus duration
	duration abi.ChainEpoch
}

// NewFixedDurationPolicy produces a new fixed duration self-deal policy.
func NewFixedDurationPolicy(api Chain, delay abi.ChainEpoch, duration abi.ChainEpoch) FixedDurationPolicy {
	return FixedDurationPolicy{api: api, delay: delay, duration: duration}
}

// Schedule produces the deal terms for this fixed duration self-deal policy.
func (p *FixedDurationPolicy) Schedule(ctx context.Context, pieces ...abi.PieceInfo) (Schedule, error) {
	_, epoch, err := p.api.GetChainHead(ctx)
	if err != nil {
		return Schedule{}, err
	}

	return Schedule{
		StartEpoch:  epoch + p.delay,
		ExpiryEpoch: epoch + p.delay + p.duration,
	}, nil
}
