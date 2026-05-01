package raft

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Config state machine on top of Raft

type ConfigCommand struct {
	Op  string  // "set" | "delete"
	Key string
	Value string // only for set
}

type ConfigStateMachine struct {
	mu sync.RWMutex
	store map[string]string
	node *Node
}

func NewConfigStateMachine(node *Node) *ConfigStateMachine {
	csm := &ConfigStateMachine {
		store: make(map[string]string),
		node: node,
	}
	go csm.apply()

	return csm
}

// apply commits log entries to the local key-value store
func (csm *ConfigStateMachine) apply() {

	for entry := range csm.node.applyCh {
		var cmd ConfigCommand

		if err := json.Unmarshal(entry.Command, &cmd); err != nil {
			continue
		}
		csm.mu.Lock()
		switch cmd.Op {
		case "set":
			csm.store[cmd.Key] = cmd.Value
		case "delete":
			delete(csm.store, cmd.Key)
		}
		csm.mu.Unlock()
	}
}


// Synchronize -- the public API:  propose a config change to the cluster
func (csm *ConfigStateMachine) Synchronize(key, value string) error {
	cmd := ConfigCommand{Op: "set", Key: key, Value: value}

	data,err := json.Marshal(cmd)

	if err != nil {
		return err
	}

	csm.node.mu.Lock()

	if csm.node.state != Leader {
		csm.node.mu.Unlock()

		return fmt.Errorf("not leader -- redirect to %s", csm.node.votedFor)
	}

	// append to leader's log -- replicate happends via heartbeat loop
	entry := LogEntry {
		Index: csm.node.lastLogIndex() + 1,
		Term: csm.node.currentTerm,
		Command:  data,
	}

	csm.node.log = append(csm.node.log, entry)
	csm.node.mu.Unlock()

	// wait to commit (poll commitIndex)
	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		csm.node.mu.Lock()
		committed := csm.node.commitIndex >= entry.Index
		csm.node.mu.Unlock()

		if committed {
			return nil
		}

		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timeout: entry not committed")
}


// Read local config (stale read -- use for performance)
func (csm *ConfigStateMachine) Get(key string) (string, bool) {

	csm.mu.RLock()
	defer csm.mu.RLocker().Unlock()

	v, ok := csm.store[key]

	return v,ok
}

