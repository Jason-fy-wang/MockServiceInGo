package raft

import (
	"encoding/json"
	"os"
	"sync"
)

// replicated log

// PersistentState is what must be flused to storage before any PRC reply
type PersistentState struct {
	CurrentTerm int
	VotedFor string
	Log []LogEntry
}

type DurableLog struct {
	mu sync.Mutex
	filePath	string
}

func (d *DurableLog) Save(state PersistentState) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	// Write data to temp file, then rename - atomic an POSIX
	tmp := d.filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmp,d.filePath)
}

func (d *DurableLog) Load() (PersistentState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var state PersistentState
	
	data, err := os.ReadFile(d.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}

		return state,err
	}

	return state, json.Unmarshal(data, &state)
}


