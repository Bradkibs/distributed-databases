package protocol

import (
	"testing"
)

func TestMessageTypeString(t *testing.T) {
	tests := []struct {
		msgType  MessageType
		expected string
	}{
		{MsgPrepare, "Prepare"},
		{MsgVoteYes, "VoteYes"},
		{MsgVoteNo, "VoteNo"},
		{MsgCommit, "Commit"},
		{MsgAbort, "Abort"},
		{MsgAck, "Ack"},
		{MessageType(999), "Unknown"},
	}

	for _, tc := range tests {
		got := tc.msgType.String()
		if got != tc.expected {
			t.Errorf("MessageType(%d).String() = %s; want %s", tc.msgType, got, tc.expected)
		}
	}
}

func TestStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateInit, "Init"},
		{StateReady, "Ready"},
		{StateCommitted, "Committed"},
		{StateAborted, "Aborted"},
		{State(999), "Unknown"},
	}

	for _, tc := range tests {
		got := tc.state.String()
		if got != tc.expected {
			t.Errorf("State(%d).String() = %s; want %s", tc.state, got, tc.expected)
		}
	}
}
