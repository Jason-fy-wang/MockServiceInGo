package raft

import "testing"

func newTestNode() *Node {
	return &Node{
		id:         "n1",
		state:      Follower,
		log:        []LogEntry{{Index: 0, Term: 0}},
		applyCh:    make(chan LogEntry, 16),
		stepDownCh: make(chan struct{}, 1),
		nextIndex:  map[string]int{"n2": 1, "n3": 1},
		matchIndex: map[string]int{"n2": 0, "n3": 0},
		peers:      []string{"n2", "n3"},
	}
}

func TestRequestVotesRejectsStaleTerm(t *testing.T) {
	t.Parallel()

	n := newTestNode()
	n.currentTerm = 3
	n.votedFor = "n1"

	var reply RequestVoteReply
	err := n.RequestVotes(RequestVoteArgs{
		Term:         2,
		CandidateID:  "n2",
		LastLogIndex: 0,
		LastLogTerm:  0,
	}, &reply)
	if err != nil {
		t.Fatalf("RequestVotes() error = %v", err)
	}
	if reply.VoteGranted {
		t.Fatalf("VoteGranted = true, want false")
	}
	if n.votedFor != "n1" || n.currentTerm != 3 {
		t.Fatalf("node mutated on stale request: votedFor=%q term=%d", n.votedFor, n.currentTerm)
	}
}

func TestRequestVotesGrantsForUpToDateCandidate(t *testing.T) {
	t.Parallel()

	n := newTestNode()
	n.currentTerm = 1
	n.log = append(n.log, LogEntry{Index: 1, Term: 1})

	var reply RequestVoteReply
	err := n.RequestVotes(RequestVoteArgs{
		Term:         2,
		CandidateID:  "n2",
		LastLogIndex: 1,
		LastLogTerm:  1,
	}, &reply)
	if err != nil {
		t.Fatalf("RequestVotes() error = %v", err)
	}
	if !reply.VoteGranted {
		t.Fatalf("VoteGranted = false, want true")
	}
	if n.votedFor != "n2" || n.currentTerm != 2 || n.state != Follower {
		t.Fatalf("unexpected node state: votedFor=%q term=%d state=%v", n.votedFor, n.currentTerm, n.state)
	}
}

func TestAppendEntriesRejectsWhenPrevIndexOutOfRange(t *testing.T) {
	t.Parallel()

	n := newTestNode()
	n.currentTerm = 2

	var reply AppendEntriesReply
	err := n.AppendEntries(AppendEntriedArgs{
		Term:         2,
		LeaderID:     "n2",
		PrevLogIndex: 5,
		PrevLogTerm:  2,
	}, &reply)
	if err != nil {
		t.Fatalf("AppendEntries() error = %v", err)
	}
	if reply.Success {
		t.Fatalf("Success = true, want false")
	}
	if reply.ConflictIndex != len(n.log) || reply.ConflictTerm != -1 {
		t.Fatalf("conflict=(%d,%d), want=(%d,-1)", reply.ConflictIndex, reply.ConflictTerm, len(n.log))
	}
}

func TestAppendEntriesAppendsAndCommits(t *testing.T) {
	t.Parallel()

	n := newTestNode()
	n.currentTerm = 3

	entry := LogEntry{Index: 1, Term: 3, Command: []byte("x")}
	var reply AppendEntriesReply
	err := n.AppendEntries(AppendEntriedArgs{
		Term:         3,
		LeaderID:     "n2",
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries:      []LogEntry{entry},
		LeaderCommit: 1,
	}, &reply)
	if err != nil {
		t.Fatalf("AppendEntries() error = %v", err)
	}
	if !reply.Success {
		t.Fatalf("Success = false, want true")
	}
	if len(n.log) != 2 || n.log[1].Index != 1 || n.log[1].Term != 3 {
		t.Fatalf("log = %+v, want appended entry at index 1", n.log)
	}
	if n.commitIndex != 1 || n.lastApplied != 1 {
		t.Fatalf("commitIndex=%d lastApplied=%d, want both 1", n.commitIndex, n.lastApplied)
	}

	select {
	case applied := <-n.applyCh:
		if applied.Index != 1 || applied.Term != 3 {
			t.Fatalf("applied entry = %+v, want index=1 term=3", applied)
		}
	default:
		t.Fatalf("expected one applied log entry")
	}
}
