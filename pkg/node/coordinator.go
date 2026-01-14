package node

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"2pc-sim/pkg/protocol"
	"2pc-sim/pkg/transport"
)

type Coordinator struct {
	ID           string
	Net          transport.Network
	Participants []string
	Inbox        chan protocol.Message
	Timeout      time.Duration
}

func NewCoordinator(id string, net transport.Network, participants []string, timeout time.Duration) *Coordinator {
	return &Coordinator{
		ID:           id,
		Net:          net,
		Participants: participants,
		Inbox:        make(chan protocol.Message, 100),
		Timeout:      timeout,
	}
}

func (c *Coordinator) Start() {
	c.Net.Register(c.ID, c.Inbox)
}

// RunTransaction executes a 2PC transaction
// Returns true if committed, false if aborted
func (c *Coordinator) RunTransaction() (bool, time.Duration) {
	txID := uuid.New()
	startTime := time.Now()

	log.Printf("[Coordinator] Starting Tx %s", txID)

	// Phase 1: Prepare
	c.broadcast(protocol.MsgPrepare, txID)

	// Wait for votes
	votes := make(map[string]bool)
	aborted := false

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	// We need to collect votes from all participants
	pending := len(c.Participants)

	// Inner loop to read from inbox until decision or timeout
Loop:
	for pending > 0 {
		select {
		case <-ctx.Done():
			log.Printf("[Coordinator] Timeout waiting for votes in Tx %s", txID)
			aborted = true
			break Loop
		case msg := <-c.Inbox:
			if msg.TransactionID != txID {
				continue
			}
			if msg.Type == protocol.MsgVoteNo {
				log.Printf("[Coordinator] Received VoteNo from %s", msg.FromID)
				aborted = true
				break Loop
			} else if msg.Type == protocol.MsgVoteYes {
				if !votes[msg.FromID] {
					votes[msg.FromID] = true
					pending--
				}
			}
		}
	}

	// Phase 2: Decision
	decision := protocol.MsgCommit
	if aborted {
		decision = protocol.MsgAbort
	}

	log.Printf("[Coordinator] Decision for Tx %s: %s", txID, decision)
	c.broadcast(decision, txID)

	// Wait for Acks (Optional for strict blocking measurement, but good for completeness)
	// For this simulation, we'll wait for acks just to ensure protocol completes cleanly,
	// checking strictly for latency impact.
	// Re-using context/timeout remaining or new timeout?
	// Real 2PC might retry. Here we just wait with timeout.

	ackPending := len(c.Participants)
	ctxAck, cancelAck := context.WithTimeout(context.Background(), c.Timeout)
	defer cancelAck()

AckLoop:
	for ackPending > 0 {
		select {
		case <-ctxAck.Done():
			log.Printf("[Coordinator] Timeout waiting for ACKs in Tx %s", txID)
			break AckLoop
		case msg := <-c.Inbox:
			if msg.TransactionID != txID {
				continue
			}
			if msg.Type == protocol.MsgAck {
				ackPending--
			}
		}
	}

	duration := time.Since(startTime)
	return !aborted, duration
}

func (c *Coordinator) broadcast(msgType protocol.MessageType, txID uuid.UUID) {
	for _, pID := range c.Participants {
		msg := protocol.Message{
			Type:          msgType,
			TransactionID: txID,
			FromID:        c.ID,
			ToID:          pID,
		}
		c.Net.Send(msg)
	}
}
