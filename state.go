package main

import "fmt"

type ProtocolState int

const (
	StateStart ProtocolState = iota
	StateWaitingForJoin
	StateJoining
	StateInGame
	StateResyncing
	StateClosed
)

func (s ProtocolState) String() string {
	return [...]string{
		"Start",
		"WaitingForJoin",
		"Joining",
		"InGame",
		"Resyncing",
		"Closed",
	}[s]
}

// transitionTo validates and sets the next state
func transitionTo(current *ProtocolState, next ProtocolState) {
	valid := map[ProtocolState][]ProtocolState{
		StateStart:         {StateWaitingForJoin},
		StateWaitingForJoin:{StateJoining},
		StateJoining:       {StateInGame, StateClosed},
		StateInGame:        {StateResyncing, StateClosed},
		StateResyncing:     {StateInGame, StateClosed},
	}

	nextStates, ok := valid[*current]
	if !ok {
		panic(fmt.Sprintf("No transitions defined for state %v", *current))
	}

	for _, validState := range nextStates {
		if next == validState {
			fmt.Printf("[DFA] Transition: %v → %v\n", *current, next)
			*current = next
			return
		}
	}

	panic(fmt.Sprintf("[DFA] Invalid transition: %v → %v", *current, next))
}