package statemachine

import (
	"fmt"
	"cfdp/messages"
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
	FilePath     string
	ChecksumType messages.ChecksumType
	FileData     []byte
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

// GetState returns the current state of the state machine.
func (sm *StateMachine) GetState() State {
	return sm.CurrentState
}

// GetContext returns the context of the state machine.
func (sm *StateMachine) GetContext() *Context {
	return sm.Context
}

// HandlePDU handles a PDU according to the current state.
func (sm *StateMachine) HandlePDU(pdu messages.PDU) error {
	// This is a placeholder implementation.
	return nil
}

// StateMachineI defines the interface for the state machine.
type StateMachineI interface {
	SetState(state State)
	GetState() State
	HandlePDU(pdu messages.PDU) error
	GetContext() *Context
}
