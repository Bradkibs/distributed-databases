package node

import (
	"testing"
	"time"

	"2pc-sim/pkg/protocol"
)

func TestCoordinatorCommit(t *testing.T) {
	net := NewMockNetwork()
	pID := "p1"
	pChan := make(chan protocol.Message, 10)
	net.Register(pID, pChan)

	coordID := "coord"
	coord := NewCoordinator(coordID, net, []string{pID}, 1*time.Second, 10*time.Millisecond)
	coord.Start()

	// Simulate Coordinator Logic in a goroutine because it blocks
	done := make(chan bool)
	go func() {
		committed, _ := coord.RunTransaction()
		if !committed {
			t.Error("Transaction failed, expected Commit")
		}
		done <- true
	}()

	// Participant simulation
	// 1. Receive Prepare
	select {
	case msg := <-pChan:
		if msg.Type != protocol.MsgPrepare {
			t.Errorf("Expected Prepare, got %v", msg.Type)
		}
		// 2. Send VoteYes
		coord.Inbox <- protocol.Message{
			Type:          protocol.MsgVoteYes,
			TransactionID: msg.TransactionID,
			FromID:        pID,
			ToID:          coordID,
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Timeout waiting for Prepare")
	}

	// 3. Receive Decision
	select {
	case msg := <-pChan:
		if msg.Type != protocol.MsgCommit {
			// Might be a retry of Prepare if it was fast, but expected Commit next
			if msg.Type == protocol.MsgPrepare {
				// Ignore duplicate Prepare
			} else {
				t.Errorf("Expected Commit, got %v", msg.Type)
			}
		}
		// Send Ack
		coord.Inbox <- protocol.Message{
			Type:          protocol.MsgAck,
			TransactionID: msg.TransactionID,
			FromID:        pID,
			ToID:          coordID,
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Timeout waiting for Commit")
	}

	<-done
}

func TestCoordinatorAbortOnNo(t *testing.T) {
	net := NewMockNetwork()
	pID := "p1"
	pChan := make(chan protocol.Message, 10)
	net.Register(pID, pChan)

	coordID := "coord"
	coord := NewCoordinator(coordID, net, []string{pID}, 1*time.Second, 10*time.Millisecond)
	coord.Start()

	done := make(chan bool)
	go func() {
		committed, _ := coord.RunTransaction()
		if committed {
			t.Error("Transaction committed, expected Abort")
		}
		done <- true
	}()

	// 1. Receive Prepare
	var txID interface{} // To grab UUID
	select {
	case msg := <-pChan:
		txID = msg.TransactionID
		// 2. Send VoteNo
		coord.Inbox <- protocol.Message{
			Type:          protocol.MsgVoteNo,
			TransactionID: msg.TransactionID,
			FromID:        pID,
			ToID:          coordID,
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Timeout waiting for Prepare")
	}

	// 3. Receive Abort
	select {
	case msg := <-pChan:
		if msg.Type != protocol.MsgAbort {
			t.Errorf("Expected Abort, got %v", msg.Type)
		}
		if msg.TransactionID != txID {
			t.Errorf("Expected Abort for Tx %s, got %s", txID, msg.TransactionID)
		}
		// Send Ack
		coord.Inbox <- protocol.Message{
			Type:          protocol.MsgAck,
			TransactionID: msg.TransactionID,
			FromID:        pID,
			ToID:          coordID,
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Timeout waiting for Abort")
	}

	<-done
}

func TestCoordinatorPrepareRetry(t *testing.T) {
	net := NewMockNetwork()
	pID := "p1"
	// Don't register p1, so Send drops it strictly speaking,
	// but MockNetwork captures it in SentMessages.

	coordID := "coord"
	coord := NewCoordinator(coordID, net, []string{pID}, 200*time.Millisecond, 50*time.Millisecond)
	coord.Start()

	// RunTransaction will timeout because we never reply
	go coord.RunTransaction()

	// Allow some time for retries
	time.Sleep(150 * time.Millisecond)

	net.mu.Lock()
	count := 0
	for _, msg := range net.SentMessages {
		if msg.Type == protocol.MsgPrepare {
			count++
		}
	}
	net.mu.Unlock()

	if count < 2 {
		t.Errorf("Expected multiple Prepare messages (retries), got %d", count)
	}
}
