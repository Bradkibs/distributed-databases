package protocol

import "github.com/google/uuid"

// MessageType represents the type of 2PC message
type MessageType int

const (
	MsgPrepare MessageType = iota
	MsgVoteYes
	MsgVoteNo
	MsgCommit
	MsgAbort
	MsgAck
)

func (m MessageType) String() string {
	switch m {
	case MsgPrepare:
		return "Prepare"
	case MsgVoteYes:
		return "VoteYes"
	case MsgVoteNo:
		return "VoteNo"
	case MsgCommit:
		return "Commit"
	case MsgAbort:
		return "Abort"
	case MsgAck:
		return "Ack"
	default:
		return "Unknown"
	}
}

// State represents the state of a transaction participant/coordinator
type State int

const (
	StateInit State = iota
	StateReady
	StateCommitted
	StateAborted
)

func (s State) String() string {
	switch s {
	case StateInit:
		return "Init"
	case StateReady:
		return "Ready"
	case StateCommitted:
		return "Committed"
	case StateAborted:
		return "Aborted"
	default:
		return "Unknown"
	}
}

// Message represents a general 2PC message
type Message struct {
	Type          MessageType
	TransactionID uuid.UUID
	FromID        string
	ToID          string
	// Payload could be added here for real transactions
}
