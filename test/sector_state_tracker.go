package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/go-storage-miner"
	"github.com/filecoin-project/specs-actors/actors/abi"
)

type sectorStateTracker struct {
	actualSequence   []storage.SectorState
	expectedSequence []storage.SectorState
	sectorNum        abi.SectorNumber
	t                *testing.T
}

func begin(t *testing.T, sectorNum abi.SectorNumber, initialState storage.SectorState) *sectorStateTracker {
	return &sectorStateTracker{
		actualSequence:   []storage.SectorState{},
		expectedSequence: []storage.SectorState{initialState},
		sectorNum:        sectorNum,
		t:                t,
	}
}

func (f *sectorStateTracker) then(s storage.SectorState) *sectorStateTracker {
	f.expectedSequence = append(f.expectedSequence, s)
	return f
}

func (f *sectorStateTracker) end() (func(abi.SectorNumber, storage.SectorState), func() string, <-chan struct{}) {
	var indx int
	var last storage.SectorState
	done := make(chan struct{})

	next := func(sectorNum abi.SectorNumber, curr storage.SectorState) {
		if sectorNum != f.sectorNum {
			return
		}

		if indx < len(f.expectedSequence) {
			assert.Equal(f.t, f.expectedSequence[indx], curr, "unexpected transition from %s to %s (expected transition to %s)", storage.SectorStates[last], storage.SectorStates[curr], storage.SectorStates[f.expectedSequence[indx]])
		}

		last = curr
		indx++
		f.actualSequence = append(f.actualSequence, curr)

		// if this is the last value we care about, signal completion
		if f.expectedSequence[len(f.expectedSequence)-1] == curr {
			go func() {
				done <- struct{}{}
			}()
		}
	}

	status := func() string {
		expected := make([]string, len(f.expectedSequence))
		for i, s := range f.expectedSequence {
			expected[i] = storage.SectorStates[s]
		}

		actual := make([]string, len(f.actualSequence))
		for i, s := range f.actualSequence {
			actual[i] = storage.SectorStates[s]
		}

		return fmt.Sprintf("expected transitions: %+v, actual transitions: %+v", expected, actual)
	}

	return next, status, done
}
