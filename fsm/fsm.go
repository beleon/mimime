package fsm

import "errors"

type State int
type Input int

type Transition struct {
	Input Input
	Target State
}

type TransitionFn map[Input]State
type TransitionMap map[State](TransitionFn)

type Builder struct {
	transitions TransitionMap
	initial     State
	accepting   []State
}

type Fsm struct {
	transitions TransitionMap
	initial     State
	accepting   []State
	current     State
	validState  bool
}

/*
 *
 * Using the Builder is typically easier to construct a new FSM.
 */
func NewFsm(transitions TransitionMap, initial State, accepting []State) *Fsm {
	return &Fsm{transitions,initial,accepting,initial,true}
}

func NewBuilder(initial State, accepting ...State) *Builder {
	return &Builder{make(TransitionMap), initial, accepting}
}

func (fsm *Fsm) Advance(in Input) {
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

func (fsm *Fsm) IsAccepting() bool {
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

func (fsm *Fsm) Finalize() (state State, err error) {
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

func (fsm *Fsm) InErrorState() bool {
	return !fsm.validState
}

func (b *Builder) BindTransitions(from State, transitions ...Transition) *Builder {
	transitionFn := make(TransitionFn)
	for _, transition := range transitions {
		transitionFn[transition.Input] = transition.Target
	}
	b.transitions[from] = transitionFn
	return b
}

func (b *Builder) Build() *Fsm {
	return NewFsm(b.transitions, b.initial, b.accepting)
}
