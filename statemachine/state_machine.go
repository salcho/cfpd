package statemachine

import (
	"fmt"
)

type StateMachine struct {
	CurrentState State
	Context      *Context
}

type State byte

const (
	StateIdle State = iota
	WaitingForDirectoryListingMetadata
	SendingDirectoryListingMetadata
	UploadingDirectoryListing
	SendingDirectoryListingEOF
	WaitingForFileData
	ReceivedDirectoryListing
	StateSending
	StateReceiving
	StateError
)

type Context struct {
	FilePath string
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
		CurrentState: StateIdle,
		Context:      &Context{},
	}
}

func (sm *StateMachine) SetState(newState State) {
	if newState < StateIdle || newState > StateError {
		panic(fmt.Sprintf("Invalid state: %d", newState))
	}
	sm.CurrentState = newState
}
