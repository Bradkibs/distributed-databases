package node

import (
	"testing"

	"github.com/google/uuid"

	"2pc-sim/pkg/protocol"
)

func TestParticipant_StateTransitions(t *testing.T) {
	// Use NewMockNetwork from common_test.go
	net := NewMockNetwork()
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
	if len(net.SentMessages) != 1 || net.SentMessages[0].Type != protocol.MsgVoteYes {
		t.Errorf("Expected MsgVoteYes sent, got %v", net.SentMessages)
	}

	// 3. Commit -> Committed
	commitMsg := protocol.Message{
		Type:          protocol.MsgCommit,
		TransactionID: txID,
		FromID:        "coord",
		ToID:          "p1",
	}

	net.SentMessages = nil // Clear sent
	p.handleCommit(commitMsg)

	if p.State != protocol.StateCommitted {
		t.Errorf("Expected StateCommitted after Commit, got %s", p.State)
	}
	if len(net.SentMessages) != 1 || net.SentMessages[0].Type != protocol.MsgAck {
		t.Errorf("Expected MsgAck sent, got %v", net.SentMessages)
	}
}

func TestParticipant_ForceAbort(t *testing.T) {
	net := NewMockNetwork()
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
	if len(net.SentMessages) != 1 || net.SentMessages[0].Type != protocol.MsgVoteNo {
		t.Errorf("Expected MsgVoteNo sent, got %v", net.SentMessages)
	}
}
