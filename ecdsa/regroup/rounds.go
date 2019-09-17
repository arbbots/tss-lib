package regroup

import (
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/tss"
)

const (
	TaskName = "regroup"
)

type (
	base struct {
		*tss.ReGroupParameters
		key,
		save *keygen.LocalPartySaveData
		temp   *LocalPartyTempData
		out    chan<- tss.Message
		oldOK, // old committee "ok" tracker
		newOK []bool // `ok` tracks parties which have been verified by Update(); this one is for the new committee
		started bool
		number  int
	}
	round1 struct {
		*base
	}
	round2 struct {
		*round1
	}
	round3 struct {
		*round2
	}
	round4 struct {
		*round3
	}
)

var (
	_ tss.Round = (*round1)(nil)
	_ tss.Round = (*round2)(nil)
	_ tss.Round = (*round3)(nil)
	_ tss.Round = (*round4)(nil)
)

// ----- //

func (round *base) Params() *tss.Parameters {
	return round.ReGroupParameters.Parameters
}

func (round *base) ReGroupParams() *tss.ReGroupParameters {
	return round.ReGroupParameters
}

func (round *base) RoundNumber() int {
	return round.number
}

// CanProceed is inherited by other rounds
func (round *base) CanProceed() bool {
	if !round.started {
		return false
	}
	for _, ok := range append(round.oldOK, round.newOK...) {
		if !ok {
			return false
		}
	}
	return true
}

// WaitingFor is called by a Party for reporting back to the caller
func (round *base) WaitingFor() []*tss.PartyID {
	oldPs := round.OldParties().IDs()
	newPs := round.NewParties().IDs()
	idsMap := make(map[*tss.PartyID]bool)
	ids := make([]*tss.PartyID, 0, len(round.oldOK))
	for j, ok := range round.oldOK {
		if ok {
			continue
		}
		idsMap[oldPs[j]] = true
	}
	for j, ok := range round.newOK {
		if ok {
			continue
		}
		idsMap[newPs[j]] = true
	}
	// consolidate into the list
	for id := range idsMap {
		ids = append(ids, id)
	}
	return ids
}

func (round *base) WrapError(err error, culprits ...*tss.PartyID) *tss.Error {
	pID := round.OldCommitteePartyID()
	if pID == nil {
		pID = round.NewCommitteePartyID()
	}
	return tss.NewError(err, TaskName, round.number, pID, culprits...)
}

// ----- //

// `oldOK` tracks parties which have been verified by Update()
func (round *base) resetOK() {
	for j := range round.oldOK {
		round.oldOK[j] = false
	}
	for j := range round.newOK {
		round.newOK[j] = false
	}
}

// sets all pairings in `oldOK` to true
func (round *base) allOldOK() {
	for j := range round.oldOK {
		round.oldOK[j] = true
	}
}

// sets all pairings in `newOK` to true
func (round *base) allNewOK() {
	for j := range round.newOK {
		round.newOK[j] = true
	}
}
