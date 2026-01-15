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
	ID            string
	Net           transport.Network
	Participants  []string
	Inbox         chan protocol.Message
	Timeout       time.Duration
	RetryInterval time.Duration
}

func NewCoordinator(id string, net transport.Network, participants []string, timeout time.Duration, retryInterval time.Duration) *Coordinator {
	return &Coordinator{
		ID:            id,
		Net:           net,
		Participants:  participants,
		Inbox:         make(chan protocol.Message, 100),
		Timeout:       timeout,
		RetryInterval: retryInterval,
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

	// Helper closure to send message to a specific participant
	sendTo := func(to string, msgType protocol.MessageType) {
		c.Net.Send(protocol.Message{
			Type:          msgType,
			TransactionID: txID,
			FromID:        c.ID,
			ToID:          to,
		})
	}

	// Phase 1: Prepare
	c.broadcast(protocol.MsgPrepare, txID)

	// Wait for votes
	votes := make(map[string]bool)
	pendingVotes := make(map[string]bool)
	for _, p := range c.Participants {
		pendingVotes[p] = true
	}

	aborted := false

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	// Retry loop for Phase 1
	ticker := time.NewTicker(c.RetryInterval)
	defer ticker.Stop()

Loop:
	for len(pendingVotes) > 0 {
		select {
		case <-ctx.Done():
			log.Printf("[Coordinator] Timeout waiting for votes in Tx %s", txID)
			aborted = true
			break Loop
		case <-ticker.C:
			// Retry Prepare for those who haven't voted
			for pID := range pendingVotes {
				// We don't log every retry to avoid spam
				sendTo(pID, protocol.MsgPrepare)
			}
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
					delete(pendingVotes, msg.FromID)
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

	// Wait for Acks
	pendingAcks := make(map[string]bool)
	for _, p := range c.Participants {
		pendingAcks[p] = true
	}

	ctxAck, cancelAck := context.WithTimeout(context.Background(), c.Timeout)
	defer cancelAck()

	// Reuse ticker for Phase 2 retries
AckLoop:
	for len(pendingAcks) > 0 {
		select {
		case <-ctxAck.Done():
			log.Printf("[Coordinator] Timeout waiting for ACKs in Tx %s", txID)
			break AckLoop
		case <-ticker.C:
			// Retry Decision
			for pID := range pendingAcks {
				sendTo(pID, decision)
			}
		case msg := <-c.Inbox:
			if msg.TransactionID != txID {
				continue
			}
			if msg.Type == protocol.MsgAck {
				delete(pendingAcks, msg.FromID)
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
