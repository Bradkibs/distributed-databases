package transport

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"2pc-sim/pkg/protocol"
)

// Network interface defines how messages are sent
type Network interface {
	Send(msg protocol.Message)
	Register(id string, ch chan protocol.Message)
	Unregister(id string)
}

// SimulatedNetwork implements Network with simulated latency and packet loss
type SimulatedNetwork struct {
	mu           sync.RWMutex
	nodes        map[string]chan protocol.Message
	AverageDelay time.Duration
	DropRate     float64 // 0.0 to 1.0 (0% to 100% loss)
	Jitter       float64 // 0.0 to 1.0 (relative to AverageDelay)
	r            *rand.Rand
}

// NewSimulatedNetwork creates a new simulated network
func NewSimulatedNetwork(delay time.Duration, dropRate float64, jitter float64) *SimulatedNetwork {
	return &SimulatedNetwork{
		nodes:        make(map[string]chan protocol.Message),
		AverageDelay: delay,
		DropRate:     dropRate,
		Jitter:       jitter,
		r:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (n *SimulatedNetwork) Register(id string, ch chan protocol.Message) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.nodes[id] = ch
}

func (n *SimulatedNetwork) Unregister(id string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.nodes, id)
}

func (n *SimulatedNetwork) Send(msg protocol.Message) {
	// 1. Simulate Drop
	if n.DropCheck() {
		log.Printf("[Network] DROPPED message %s from %s to %s", msg.Type, msg.FromID, msg.ToID)
		return
	}

	// 2. Simulate Delay asynchronously
	go func() {
		delay := n.calculateDelay()
		time.Sleep(delay)

		n.mu.RLock()
		ch, ok := n.nodes[msg.ToID]
		n.mu.RUnlock()

		if ok {
			// Non-blocking send or blocking?
			// For simulation, blocking is safer to guarantee delivery if not dropped,
			// but could deadlock if recipient is full.
			// Using buffered channels in nodes is recommended.
			select {
			case ch <- msg:
				// Delivered
			default:
				log.Printf("[Network] Failed to deliver message %s to %s (channel full)", msg.Type, msg.ToID)
			}
		} else {
			log.Printf("[Network] Destination %s not found for message %s from %s", msg.ToID, msg.Type, msg.FromID)
		}
	}()
}

func (n *SimulatedNetwork) DropCheck() bool {
	// n.r is not safe for concurrent use, so we strictly speaking should lock it or use a per-goroutine source.
	// Let's use math/rand global functions which are thread-safe (locked internally).
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.r.Float64() < n.DropRate
}

func (n *SimulatedNetwork) calculateDelay() time.Duration {
	// Simple uniform distribution: Delay +/- Jitter
	jitterRange := float64(n.AverageDelay) * n.Jitter
	// n.r.Float64 returns [0.0, 1.0)
	// We want offset in [-jitterRange, +jitterRange)
	n.mu.Lock()
	defer n.mu.Unlock()
	offset := (n.r.Float64() * 2 * jitterRange) - jitterRange
	delay := time.Duration(float64(n.AverageDelay) + offset)
	if delay < 0 {
		return 0
	}
	return delay
}
