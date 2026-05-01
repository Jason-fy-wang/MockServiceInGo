package raft

// RPC types (AppendEntries, RequestVote)


type RequestVoteArgs struct {
	Term int // candidate's term
	CandidateID string // candidate requesting vote
	LastLogIndex int // index of candidate's last log entry
	LastLogTerm int // term of candidate's last log entry
}

type RequestVoteReply struct {
	Term int  // currentTerm, for candidate to update itself
	VoteGranted bool // true means candidate received vote
}


type AppendEntriedArgs struct {
	Term int	// leader's term
	LeaderID string
	PrevLogIndex int // index of log entry immediately preceding new ones
	PrevLogTerm int  // term of prevLogIndex entry
	Entries []LogEntry //empty for heartbeat
	LeaderCommit int // leader's commitIndex
}


type AppendEntriesReply struct {
	Term int // currentTerm,, for leader to update itself
	Success bool // true if follower contained entry matching prevLogIndex and prevLogTerm

	// Conflict optimization: conflict hints to skip back faster then one by one
	ConflictTerm int	// term of conflicting entry
	ConflictIndex int   // first index of that conflicting term
}


