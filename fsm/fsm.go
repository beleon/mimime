package fsm

import "errors"

type FsmState int
type FsmInput int

type FsmTrans map[FsmState](map[FsmInput]FsmState)

type fsm struct {
    transitions FsmTrans
    initial     FsmState
    accepting   []FsmState
    current     FsmState
    validState  bool
}

func (fsm *fsm) Advance(in FsmInput) {
    if !fsm.validState {
        return
    }

    i, ok := fsm.transitions[fsm.current]
    if ok {
        j, ok := i[in]
        if ok {
            fsm.current = j
            return
        }
    }

    fsm.validState = false
}

func (fsm *fsm) IsAccepting() bool {
    if !fsm.validState {
        return false
    }

    for _, state := range fsm.accepting {
        if fsm.current == state {
            return true
        }
    }

    return false
}

func (fsm *fsm) Finalize() (state FsmState, err error) {
    state = fsm.current

    if !fsm.validState {
        err = errors.New("FSM is in error state.")
        return
    }

    if !fsm.IsAccepting() {
        err = errors.New("FSM is not in an accepting state.")
    }
    return
}

func Initialize(transitions FsmTrans, initial FsmState, accepting []FsmState) fsm {
    return fsm{transitions, initial, accepting, initial, true}
}
