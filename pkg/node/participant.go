package node

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"2pc-sim/pkg/protocol"
	"2pc-sim/pkg/transport"
)

// Participant represents a node in the distributed system
type Participant struct {
	ID            string
	State         protocol.State
	Net           transport.Network
	Inbox         chan protocol.Message
	CoordinatorID string
	mu            sync.Mutex
	// Logic hooks for simulation
	ForceVoteNo bool
	// Metrics
	ReadyTime time.Time
}

func NewParticipant(id string, net transport.Network, coordinatorID string) *Participant {
	return &Participant{
		ID:            id,
		State:         protocol.StateInit,
		Net:           net,
		Inbox:         make(chan protocol.Message, 100),
		CoordinatorID: coordinatorID,
	}
}

func (p *Participant) Start() {
	p.Net.Register(p.ID, p.Inbox)
	go p.loop()
}

func (p *Participant) loop() {
	for msg := range p.Inbox {
		p.handleMessage(msg)
	}
}

func (p *Participant) handleMessage(msg protocol.Message) {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Printf("[Participant %s] Rx %s from %s", p.ID, msg.Type, msg.FromID)

	switch msg.Type {
	case protocol.MsgPrepare:
		p.handlePrepare(msg)
	case protocol.MsgCommit:
		p.handleCommit(msg)
	case protocol.MsgAbort:
		p.handleAbort(msg)
	default:
		log.Printf("[Participant %s] Ignoring unexpected message type %s", p.ID, msg.Type)
	}
}

func (p *Participant) handlePrepare(msg protocol.Message) {
	if p.State != protocol.StateInit {
		return
	}

	// Decision logic
	vote := protocol.MsgVoteYes
	if p.ForceVoteNo {
		vote = protocol.MsgVoteNo
		p.State = protocol.StateAborted
	} else {
		p.State = protocol.StateReady
		p.ReadyTime = time.Now()
	}

	reply := protocol.Message{
		Type:          vote,
		TransactionID: msg.TransactionID,
		FromID:        p.ID,
		ToID:          msg.FromID,
	}
	p.Net.Send(reply)
}

func (p *Participant) handleCommit(msg protocol.Message) {
	if p.State == protocol.StateReady {
		p.State = protocol.StateCommitted
		log.Printf("[Participant %s] COMMITTED Tx %s", p.ID, msg.TransactionID)
		p.sendAck(msg.TransactionID, msg.FromID)
	} else {
		log.Printf("[Participant %s] Received Commit but state is %s", p.ID, p.State)
	}
}

func (p *Participant) handleAbort(msg protocol.Message) {
	if p.State == protocol.StateReady || p.State == protocol.StateInit {
		p.State = protocol.StateAborted
		log.Printf("[Participant %s] ABORTED Tx %s", p.ID, msg.TransactionID)
		p.sendAck(msg.TransactionID, msg.FromID)
	}
}

func (p *Participant) sendAck(txID uuid.UUID, to string) {
	ack := protocol.Message{
		Type:          protocol.MsgAck,
		TransactionID: txID,
		FromID:        p.ID,
		ToID:          to,
	}
	p.Net.Send(ack)
}
