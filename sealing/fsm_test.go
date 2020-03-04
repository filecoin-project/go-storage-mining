package sealing

import (
	"testing"

	"github.com/filecoin-project/go-storage-miner/apis/node"

	"github.com/filecoin-project/go-statemachine"
	logging "github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = logging.SetLogLevel("*", "INFO")
}

func (t *test) planSingle(evt interface{}) {
	events := []statemachine.Event{{evt}}
	_, processed, err := t.s.plan(events, t.state)
	require.NoError(t.t, err)
	require.Equal(t.t, processed, uint64(len(events)))
}

type test struct {
	s     *Sealing
	t     *testing.T
	state *SectorInfo

	next func(statemachine.Context, SectorInfo) error
}

func TestHappyPath(t *testing.T) {
	m := test{
		s:     &Sealing{},
		t:     t,
		state: &SectorInfo{State: Packing},
	}

	m.planSingle(SectorPacked{})
	require.Equal(m.t, m.state.State, Unsealed)

	m.planSingle(SectorSealed{})
	require.Equal(m.t, m.state.State, PreCommitting)

	m.planSingle(SectorPreCommitted{})
	require.Equal(m.t, m.state.State, WaitSeed)

	m.planSingle(SectorSeedReady{})
	require.Equal(m.t, m.state.State, Committing)

	m.planSingle(SectorCommitted{})
	require.Equal(m.t, m.state.State, CommitWait)

	m.planSingle(SectorProving{})
	require.Equal(m.t, m.state.State, FinalizeSector)

	m.planSingle(SectorFinalized{})
	require.Equal(m.t, m.state.State, Proving)
}

func TestSeedRevert(t *testing.T) {
	m := test{
		s:     &Sealing{},
		t:     t,
		state: &SectorInfo{State: Packing},
	}

	m.planSingle(SectorPacked{})
	require.Equal(m.t, m.state.State, Unsealed)

	m.planSingle(SectorSealed{})
	require.Equal(m.t, m.state.State, PreCommitting)

	m.planSingle(SectorPreCommitted{})
	require.Equal(m.t, m.state.State, WaitSeed)

	m.planSingle(SectorSeedReady{})
	require.Equal(m.t, m.state.State, Committing)

	_, _, err := m.s.plan([]statemachine.Event{{SectorSeedReady{seed: node.SealSeed{BlockHeight: 5}}}, {SectorCommitted{}}}, m.state)
	require.NoError(t, err)
	require.Equal(m.t, m.state.State, Committing)

	// not changing the seed this time
	_, _, err = m.s.plan([]statemachine.Event{{SectorSeedReady{seed: node.SealSeed{BlockHeight: 5}}}, {SectorCommitted{}}}, m.state)
	require.Equal(m.t, m.state.State, CommitWait)

	m.planSingle(SectorProving{})
	require.Equal(m.t, m.state.State, FinalizeSector)

	m.planSingle(SectorFinalized{})
	require.Equal(m.t, m.state.State, Proving)
}

func TestPlanCommittingHandlesSectorCommitFailed(t *testing.T) {
	m := test{
		s:     &Sealing{},
		t:     t,
		state: &SectorInfo{State: Committing},
	}

	events := []statemachine.Event{{SectorCommitFailed{}}}

	require.NoError(t, planCommitting(events, m.state))

	require.Equal(t, SectorStates[CommitFailed], SectorStates[m.state.State])
}
