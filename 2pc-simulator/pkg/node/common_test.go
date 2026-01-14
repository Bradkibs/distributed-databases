package node

import (
	"sync"

	"2pc-sim/pkg/protocol"
)

// MockNetwork captures sent messages for verification
type MockNetwork struct {
	mu           sync.Mutex
	SentMessages []protocol.Message
	Participants map[string]chan protocol.Message
}

func NewMockNetwork() *MockNetwork {
	return &MockNetwork{
		Participants: make(map[string]chan protocol.Message),
	}
}

func (m *MockNetwork) Send(msg protocol.Message) {
	m.mu.Lock()
	m.SentMessages = append(m.SentMessages, msg)
	ch, ok := m.Participants[msg.ToID]
	m.mu.Unlock()

	if ok {
		// Non-blocking send purely for test speed
		select {
		case ch <- msg:
		default:
		}
	}
}

func (m *MockNetwork) Register(id string, ch chan protocol.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Participants[id] = ch
}

func (m *MockNetwork) Unregister(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Participants, id)
}
