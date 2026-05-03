package raft

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Config state machine on top of Raft

type ConfigCommand struct {
	Op    string // "set" | "delete"
	Key   string
	Value string // only for set
}

type ConfigStateMachine struct {
	mu          sync.RWMutex
	store       map[string]string
	Node        *Node
	subscribers []func(ConfigCommand)
}

func NewConfigStateMachine(node *Node) *ConfigStateMachine {
	csm := &ConfigStateMachine{
		store: make(map[string]string),
		Node:  node,
	}
	go csm.apply()

	return csm
}

// apply commits log entries to the local key-value store
func (csm *ConfigStateMachine) apply() {

	for entry := range csm.Node.applyCh {
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
		subs := append([]func(ConfigCommand){}, csm.subscribers...)
		csm.mu.Unlock()
		for _, sub := range subs {
			sub(cmd)
		}
	}
}

// Synchronize -- the public API:  propose a config change to the cluster
func (csm *ConfigStateMachine) Synchronize(key, value string) error {
	cmd := ConfigCommand{Op: "set", Key: key, Value: value}

	data, err := json.Marshal(cmd)

	if err != nil {
		return err
	}

	csm.Node.mu.Lock()

	if csm.Node.state != Leader {
		csm.Node.mu.Unlock()

		return fmt.Errorf("not leader -- redirect to %s", csm.Node.votedFor)
	}

	// append to leader's log -- replicate happends via heartbeat loop
	entry := LogEntry{
		Index:   csm.Node.lastLogIndex() + 1,
		Term:    csm.Node.currentTerm,
		Command: data,
	}

	csm.Node.log = append(csm.Node.log, entry)
	csm.Node.mu.Unlock()

	// wait to commit (poll commitIndex)
	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		csm.Node.mu.Lock()
		committed := csm.Node.commitIndex >= entry.Index
		csm.Node.mu.Unlock()

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

	return v, ok
}

func (csm *ConfigStateMachine) Delete(key string) error {

	csm.mu.Lock()
	defer csm.mu.Unlock()

	delete(csm.store, key)

	return nil
}

func (csm *ConfigStateMachine) Subscribe(fn func(ConfigCommand)) {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	csm.subscribers = append(csm.subscribers, fn)
}
