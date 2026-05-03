package raft

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func testNodeForConfig() *Node {
	return &Node{
		log:         []LogEntry{{Index: 0, Term: 0}},
		applyCh:     make(chan LogEntry, 8),
		state:       Follower,
		currentTerm: 1,
	}
}

func TestConfigStateMachineApplySetAndDelete(t *testing.T) {
	t.Parallel()

	n := testNodeForConfig()
	csm := NewConfigStateMachine(n)

	setCmd, _ := json.Marshal(ConfigCommand{Op: "set", Key: "region", Value: "us-east-1"})
	delCmd, _ := json.Marshal(ConfigCommand{Op: "delete", Key: "region"})

	n.applyCh <- LogEntry{Index: 1, Term: 1, Command: setCmd}
	time.Sleep(20 * time.Millisecond)
	v, ok := csm.Get("region")
	if !ok || v != "us-east-1" {
		t.Fatalf("Get(region) = (%q, %v), want (%q, true)", v, ok, "us-east-1")
	}

	n.applyCh <- LogEntry{Index: 2, Term: 1, Command: delCmd}
	time.Sleep(20 * time.Millisecond)
	_, ok = csm.Get("region")
	if ok {
		t.Fatalf("Get(region) found key after delete")
	}
}

func TestConfigStateMachineSynchronizeNotLeader(t *testing.T) {
	t.Parallel()

	n := testNodeForConfig()
	n.votedFor = "node-1"
	csm := NewConfigStateMachine(n)

	err := csm.Synchronize("a", "1")
	if err == nil || !strings.Contains(err.Error(), "not leader") {
		t.Fatalf("Synchronize() error = %v, want not leader error", err)
	}
}

func TestConfigStateMachineSynchronizeLeaderCommitted(t *testing.T) {
	t.Parallel()

	n := testNodeForConfig()
	n.state = Leader
	csm := NewConfigStateMachine(n)

	done := make(chan struct{})
	go func() {
		for {
			n.mu.Lock()
			if len(n.log) > 1 {
				n.commitIndex = n.log[1].Index
				n.mu.Unlock()
				close(done)
				return
			}
			n.mu.Unlock()
			time.Sleep(5 * time.Millisecond)
		}
	}()

	if err := csm.Synchronize("feature_x", "on"); err != nil {
		t.Fatalf("Synchronize() error = %v", err)
	}
	<-done

	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.log) != 2 {
		t.Fatalf("log length = %d, want 2", len(n.log))
	}
	var cmd ConfigCommand
	if err := json.Unmarshal(n.log[1].Command, &cmd); err != nil {
		t.Fatalf("unmarshal command error = %v", err)
	}
	if cmd.Op != "set" || cmd.Key != "feature_x" || cmd.Value != "on" {
		t.Fatalf("command = %+v, want set feature_x=on", cmd)
	}
}
