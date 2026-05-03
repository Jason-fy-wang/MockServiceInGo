package raft

import (
	"path/filepath"
	"testing"
)

func TestDurableLogSaveLoadRoundTrip(t *testing.T) {
	t.Parallel()

	d := &DurableLog{filePath: filepath.Join(t.TempDir(), "raft-state.json")}
	want := PersistentState{
		CurrentTerm: 7,
		VotedFor:    "node-2",
		Log: []LogEntry{
			{Index: 0, Term: 0},
			{Index: 1, Term: 7, Command: []byte(`{"k":"v"}`)},
		},
	}

	if err := d.Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := d.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got.CurrentTerm != want.CurrentTerm || got.VotedFor != want.VotedFor || len(got.Log) != len(want.Log) {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}
	if got.Log[1].Index != want.Log[1].Index || got.Log[1].Term != want.Log[1].Term || string(got.Log[1].Command) != string(want.Log[1].Command) {
		t.Fatalf("Load().Log[1] = %+v, want %+v", got.Log[1], want.Log[1])
	}
}

func TestDurableLogLoadMissingFile(t *testing.T) {
	t.Parallel()

	d := &DurableLog{filePath: filepath.Join(t.TempDir(), "does-not-exist.json")}

	got, err := d.Load()
	if err != nil {
		t.Fatalf("Load() unexpected error = %v", err)
	}
	if got.CurrentTerm != 0 || got.VotedFor != "" || len(got.Log) != 0 {
		t.Fatalf("Load() = %+v, want zero value state", got)
	}
}
