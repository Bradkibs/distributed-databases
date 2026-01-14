package transport

import (
	"testing"
	"time"

	"2pc-sim/pkg/protocol"
)

func TestNetworkRegistration(t *testing.T) {
	net := NewSimulatedNetwork(0, 0, 0)
	ch := make(chan protocol.Message, 1)
	id := "test-node"

	net.Register(id, ch)

	// Verify internal map
	net.mu.RLock()
	gotCh, ok := net.nodes[id]
	net.mu.RUnlock()

	if !ok {
		t.Fatal("Register failed: node not found in map")
	}
	if gotCh != ch {
		t.Error("Register stored incorrect channel")
	}

	net.Unregister(id)
	net.mu.RLock()
	_, ok = net.nodes[id]
	net.mu.RUnlock()

	if ok {
		t.Error("Unregister failed: node still exists in map")
	}
}

func TestNetworkSend(t *testing.T) {
	// 0 latency, 0 drop
	net := NewSimulatedNetwork(0, 0, 0)
	ch := make(chan protocol.Message, 1)
	id := "receiver"
	net.Register(id, ch)

	msg := protocol.Message{
		Type:   protocol.MsgPrepare,
		ToID:   id,
		FromID: "sender",
	}

	net.Send(msg)

	select {
	case received := <-ch:
		if received.Type != msg.Type {
			t.Errorf("Received wrong message type: got %v, want %v", received.Type, msg.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message")
	}
}

func TestNetworkDrop(t *testing.T) {
	// 100% drop rate
	net := NewSimulatedNetwork(0, 1.0, 0)
	ch := make(chan protocol.Message, 1)
	id := "receiver"
	net.Register(id, ch)

	msg := protocol.Message{
		Type:   protocol.MsgPrepare,
		ToID:   id,
		FromID: "sender",
	}

	net.Send(msg)

	select {
	case <-ch:
		t.Fatal("Message should have been dropped")
	case <-time.After(50 * time.Millisecond):
		// Success: timeout means nothing arrived
	}
}
