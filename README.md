# Performance Analysis of Two-Phase Commit (2PC) in Varying Network Conditions

## Overview
This project implements a lightweight distributed transaction manager in Go to empirically measure the performance overhead and "blocking" behavior of the Two-Phase Commit (2PC) protocol. Unlike theoretical models that assume fixed delays, this simulation incorporates realistic network conditions such as variable latency, jitter, and packet loss.

## Project Goals
- **Empirical Measurement**: Quantify the impact of network latency and scale (number of participants) on transaction completion time.
- **Blocking Analysis**: Measure the duration participants spend in the critical "Ready" state where they are blocked waiting for the Coordinator.
- **Fault Tolerance Simulation**: Observe system behavior under failure scenarios (e.g., node aborts, message drops).

## Implementation Details

The system is designed as an in-memory simulation to run efficient experiments without the overhead of real network sockets or disk I/O, while maintaining architectural accuracy.

### 1. ArchitectureThe project is titled:
Performance Analysis of Two-Phase Commit (2PC) in Varying Network Conditions. The
paper identifies managing distributed transactions as critical for reliability and failure atomicity in
distributed database systems.
However, the authors note that full distributed transaction support (like 2PC) has seen slow or
incomplete commercial adoption because its performance overhead is poorly understood and
often a concern for vendors. Existing performance models often make simplistic assumptions,
such as fixed communication delays, which fail to capture the realities of scaling across different
network loads. This project aims to close this gap by creating a working prototype to empirically
measure the "blocking" overhead inherent in the 2PC protocol.
The core of this project will be to implement a lightweight distributed transaction manager. This
manager will consist of a single Coordinator node and multiple Participant nodes, implementing
the standard Two-Phase Commit logic: a "voting phase" where the coordinator queries
participants, and a "decision phase" where the final commit or abort command is broadcast. To
keep the project scope manageable within 20-30 hours, the implementation will focus on the
in-memory message exchange logic and state transitions (e.g., Ready, Abort, Commit) rather
than complex persistent disk recovery logs.
The simulation consists of three core components:

*   **Coordinator**: The central authority that initiates transactions. It executes the standard 2PC logic:
    *   *Phase 1*: Broadcasts `PREPARE` to all participants.
    *   *Phase 2*: Based on votes or timeout, broadcasts `COMMIT` or `ABORT`.
*   **Participant**: Distributed nodes that manage local transaction resources (simulated). They validate requests, vote `YES`/`NO`, and wait for the final decision.
*   **Network Transport**: A custom simulation layer that sits between nodes. It uses Go channels to deliver messages but injects delays and drops based on configuration.

### 2. Key Features
*   **State Machines**: Strict adherence to 2PC state transitions (Init -> Ready -> Committed/Aborted).
*   **Configurable Latency**: Network delays are modeled with an average latency and random jitter to mimic real-world variance.
*   **Packet Loss**: Support for probabilistic message dropping to test timeout and retry mechanisms (basic timeout implemented).
*   **Fault Injection**: Participants can be configured to randomly vote `NO` to simulate local validation failures or deadlocks.

### 3. Project Structure
```text
.
├── cmd
│   └── 2pc-sim        # Main entry point and CLI runner
├── pkg
│   ├── node           # Logic for Coordinator and Participants
│   ├── protocol       # Definitions of 2PC messages (Prepare, Vote, etc.)
│   └── transport      # Network simulation (Channel-based with delay)
└── README.md
```

## Usage

Build the simulation binary:
```bash
go build -o 2pc-sim cmd/2pc-sim/main.go
```

### Running Experiments

The CLI supports several flags to vary the simulation parameters:

| Flag | Default | Description |
|------|---------|-------------|
| `--participants` | 3 | Number of Participant nodes |
| `--latency` | 10 | Average network one-way latency (ms) |
| `--drop-rate` | 0.0 | Probability of packet loss (0.0 - 1.0) |
| `--abort-rate` | 0.0 | Probability of a participant voting NO |
| `--timeout` | 5 | Transaction timeout (seconds) |

### Scenarios

**1. Baseline Performance**
Measure standard overhead with low latency.
```bash
./2pc-sim --participants 5 --latency 5
```

**2. High Latency / WAN Simulation**
Simulate a geo-distributed database across continents.
```bash
./2pc-sim --latency 150 --participants 3
```

**3. High Failure Rate**
Simulate an unstable environment where half the transactions fail.
```bash
./2pc-sim --abort-rate 0.5
```

## Future Work
- **Persistence**: Add Write-Ahead Logging (WAL) to simulate disk I/O latency.
- **Recovery Protocol**: Implement the full recovery procedure for nodes coming back online after a crash.
- **3PC**: Implement Three-Phase Commit to compare blocking behavior.
