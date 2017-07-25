package mimime

import (
	"errors"
	"github.com/sellleon/mimime/fsm"
)

const (
	startState          fsm.FsmState = iota
	leftParseState      fsm.FsmState = iota
	noLeftParseState    fsm.FsmState = iota
	leftOnlyParseState  fsm.FsmState = iota
	rightOnlyParseState fsm.FsmState = iota
	bothParseState      fsm.FsmState = iota
)

const (
	leftParseInput    fsm.FsmInput = iota
	noLeftParseInput  fsm.FsmInput = iota
	rightParseInput   fsm.FsmInput = iota
	noRightParseInput fsm.FsmInput = iota
)

var parseTransitions fsm.FsmTrans
var acceptingParseStates []fsm.FsmState

func init() {
	startStateTrans := make(map[fsm.FsmInput]fsm.FsmState)
	startStateTrans[leftParseInput] = leftParseState
	startStateTrans[noLeftParseInput] = noLeftParseState
	leftParseStateTrans := make(map[fsm.FsmInput]fsm.FsmState)
	leftParseStateTrans[rightParseInput] = bothParseState
	leftParseStateTrans[noRightParseInput] = leftOnlyParseState
	noLeftParseStateTrans := make(map[fsm.FsmInput]fsm.FsmState)
	noLeftParseStateTrans[rightParseInput] = rightOnlyParseState
	parseTransitions = make(fsm.FsmTrans)
	parseTransitions[startState] = startStateTrans
	parseTransitions[leftParseState] = leftParseStateTrans
	parseTransitions[noLeftParseState] = noLeftParseStateTrans

	acceptingParseStates = []fsm.FsmState{
		leftOnlyParseState,
		rightOnlyParseState,
		bothParseState}
}

func resizeFlagFromState(s fsm.FsmState) (resizeFlag, error) {
	switch s {
	case leftOnlyParseState:
		return rFLeft, nil
	case rightOnlyParseState:
		return rFRight, nil
	case bothParseState:
		return rFBoth, nil
	}
	return -1, errors.New("Invalid size option given.")
}
