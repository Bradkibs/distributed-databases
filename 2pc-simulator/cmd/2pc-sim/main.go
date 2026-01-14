package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"2pc-sim/pkg/node"
	"2pc-sim/pkg/transport"
)

func main() {
	var (
		numParticipants int
		latencyMs       int
		dropRate        float64
		voteNoRate      float64
		timeoutSec      int
		jitter          float64
	)

	flag.IntVar(&numParticipants, "participants", 3, "Number of participants")
	flag.IntVar(&latencyMs, "latency", 10, "Average network latency in ms")
	flag.Float64Var(&dropRate, "drop-rate", 0.0, "Packet drop rate (0.0 - 1.0)")
	flag.Float64Var(&voteNoRate, "abort-rate", 0.0, "Probability of a participant voting No (0.0 - 1.0)")
	flag.IntVar(&timeoutSec, "timeout", 5, "Transaction timeout in seconds")
	flag.Float64Var(&jitter, "jitter", 0.2, "Network jitter (0.0 - 1.0)")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	fmt.Printf("--- 2PC Simulation Configuration ---\n")
	fmt.Printf("Participants: %d\n", numParticipants)
	fmt.Printf("Latency: %d ms\n", latencyMs)
	fmt.Printf("Drop Rate: %.2f\n", dropRate)
	fmt.Printf("Abort Rate: %.2f\n", voteNoRate)
	fmt.Printf("Timeout: %d s\n", timeoutSec)
	fmt.Println("------------------------------------")

	// Initialize Network
	net := transport.NewSimulatedNetwork(time.Duration(latencyMs)*time.Millisecond, dropRate, jitter)

	// Initialize Participants
	var pIDs []string
	participants := make([]*node.Participant, numParticipants)
	coordID := "coordinator"

	for i := 0; i < numParticipants; i++ {
		pID := fmt.Sprintf("p-%d", i)
		pIDs = append(pIDs, pID)
		p := node.NewParticipant(pID, net, coordID)

		// Randomly decide if this participant will vote No
		if rand.Float64() < voteNoRate {
			p.ForceVoteNo = true
		}

		participants[i] = p
		p.Start()
	}

	// Initialize Coordinator
	coord := node.NewCoordinator(coordID, net, pIDs, time.Duration(timeoutSec)*time.Second)
	coord.Start()

	// Wait a bit for initialization
	time.Sleep(100 * time.Millisecond)

	// Run Transaction
	fmt.Println("\n>>> Starting Transaction <<<")
	committed, duration := coord.RunTransaction()

	fmt.Println("\n--- Results ---")
	status := "COMMITTED"
	if !committed {
		status = "ABORTED"
	}
	fmt.Printf("Transaction Status: %s\n", status)
	fmt.Printf("Total Duration: %v\n", duration)

	// Collect stats (simulated blocking time)
	// In a real study we would aggregate RTTs, etc.
	// For now, we print detailed participant states if needed.
}
