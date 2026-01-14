package node

import (
	"testing"

	"github.com/google/uuid"

	"2pc-sim/pkg/protocol"
)

// MockNetwork implements transport.Network for testing
type MockNetwork struct {
	Sent []protocol.Message
}

func (m *MockNetwork) Send(msg protocol.Message) {
	m.Sent = append(m.Sent, msg)
}

func (m *MockNetwork) Register(id string, ch chan protocol.Message) {}
func (m *MockNetwork) Unregister(id string)                         {}

func TestParticipant_StateTransitions(t *testing.T) {
	net := &MockNetwork{}
	p := NewParticipant("p1", net, "coord")

	// 1. Initial State
	if p.State != protocol.StateInit {
		t.Errorf("Expected StateInit, got %s", p.State)
	}

	// 2. Prepare -> Ready
	txID := uuid.New()
	prepareMsg := protocol.Message{
		Type:          protocol.MsgPrepare,
		TransactionID: txID,
		FromID:        "coord",
		ToID:          "p1",
	}

	p.handlePrepare(prepareMsg)

	if p.State != protocol.StateReady {
		t.Errorf("Expected StateReady after Prepare, got %s", p.State)
	}
	if len(net.Sent) != 1 || net.Sent[0].Type != protocol.MsgVoteYes {
		t.Errorf("Expected MsgVoteYes sent, got %v", net.Sent)
	}

	// 3. Commit -> Committed
	commitMsg := protocol.Message{
		Type:          protocol.MsgCommit,
		TransactionID: txID,
		FromID:        "coord",
		ToID:          "p1",
	}

	net.Sent = nil // Clear sent
	p.handleCommit(commitMsg)

	if p.State != protocol.StateCommitted {
		t.Errorf("Expected StateCommitted after Commit, got %s", p.State)
	}
	if len(net.Sent) != 1 || net.Sent[0].Type != protocol.MsgAck {
		t.Errorf("Expected MsgAck sent, got %v", net.Sent)
	}
}

func TestParticipant_ForceAbort(t *testing.T) {
	net := &MockNetwork{}
	p := NewParticipant("p1", net, "coord")
	p.ForceVoteNo = true

	txID := uuid.New()
	prepareMsg := protocol.Message{
		Type:          protocol.MsgPrepare,
		TransactionID: txID,
		FromID:        "coord",
		ToID:          "p1",
	}

	p.handlePrepare(prepareMsg)

	if p.State != protocol.StateAborted {
		t.Errorf("Expected StateAborted after Forced VoteNo, got %s", p.State)
	}
	if len(net.Sent) != 1 || net.Sent[0].Type != protocol.MsgVoteNo {
		t.Errorf("Expected MsgVoteNo sent, got %v", net.Sent)
	}
}
