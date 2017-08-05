package mimime

import (
	"errors"
	"github.com/sellleon/mimime/fsm"
)

const (
	startState          fsm.State = iota
	leftParseState      fsm.State = iota
	noLeftParseState    fsm.State = iota
	leftOnlyParseState  fsm.State = iota
	rightOnlyParseState fsm.State = iota
	bothParseState      fsm.State = iota
)

const (
	leftParseInput    fsm.Input = iota
	noLeftParseInput  fsm.Input = iota
	rightParseInput   fsm.Input = iota
	noRightParseInput fsm.Input = iota
)

func newResizeFsm() *fsm.Fsm {
	return fsm.NewBuilder(startState, leftOnlyParseState, rightOnlyParseState, bothParseState).
		BindTransitions(
			startState,
			fsm.Transition{leftParseInput, leftParseState},
			fsm.Transition{noLeftParseInput, noLeftParseState}).
		BindTransitions(
			leftParseState,
			fsm.Transition{rightParseInput, bothParseState},
			fsm.Transition{noRightParseInput, leftOnlyParseState}).
		BindTransitions(
			noLeftParseState,
			fsm.Transition{rightParseInput, rightOnlyParseState}).
		Build()
}

func resizeFlagFromState(s fsm.State) (resizeFlag, error) {
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
