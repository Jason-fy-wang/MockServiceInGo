package raft

import (
	"math/rand/v2"
	"mockservice/backend/log"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// core raft node statte machine



type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)


type LogEntry struct {
	Index	int
	Term 	int
	Command []byte		// serialized config key=value
}

type Node struct {
	mu sync.Mutex

	// Identity
	id string
	peers []string  // peer addresses

	// Persistent state on all servers (must servive after crash, flush to disk before responding)
	currentTerm int
	votedFor string
	log []LogEntry

	// Volatile state
	state  NodeState
	commitIndex int	// highest log entry known to be committed
	lastApplied int // highest log entry applied to state machine

	// leader-only valatile state (reinit adter election)
	nextIndex map[string]int   // next log index to send to each other
	matchIndex map[string]int  // highest log index know replicated on peer

	// Channels 
	applyCh   chan LogEntry		// committed entries go here -> config state machine
	stepDownCh chan struct{}   // leader steps down when it sees higher term
	// grantVoteCh chan struct{}  // reset election timer on vote grant
	// heartbeatCh chan struct{}  // reset election timer on valid heartbeat

	lastContact atomic.Int64  // timestamp of last contact with leader or candidate (for election timeout)

	// transport
	transport Transport

	// election timing
	electionTimeout time.Duration
	heartbeatTimeout time.Duration
}

func NewNode(id string, peers []string, transport Transport) *Node {
	n := &Node {
		id: id,
		peers: peers,
		state: Follower,
		votedFor: "",
		log: []LogEntry{{Index: 0, Term: 0}},
		nextIndex: make(map[string]int),
		matchIndex: make(map[string]int),
		applyCh: make(chan LogEntry, 254),
		stepDownCh: make(chan struct{},1),
		transport: transport,
		heartbeatTimeout: 50 * time.Millisecond,
	}
	log.Get().Info("Node created", zap.String("id", n.id), zap.Strings("peers", n.peers))
	n.lastContact.Store(time.Now().UnixNano())
	n.resetElectionTimeout()
	return n
}

func (n *Node) Id() string {
	return n.id
}

func (n *Node) resetElectionTimeout() {
	n.electionTimeout = time.Duration(300 + rand.IntN(300)) * time.Millisecond
}

func (n *Node) lastLogIndex() int {
	return n.log[len(n.log)-1].Index
}

func (n *Node) lastLogTerm() int {
	return n.log[len(n.log)-1].Term
}

func (n *Node) currentState() string {
	switch n.state {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}
}


// main loop
func (n *Node) Run() {
	for{
		n.mu.Lock()
		state := n.state
		n.mu.Unlock()
		log.Get().Info("Node state", zap.String("id", n.id), zap.String("state", n.currentState()))
		switch state {
			case Follower:
				n.runFollower()
			case Candidate:
				n.runCandidate()
			case Leader:
				n.runLeader()
		}
	}
}

func (n *Node) runFollower() {
	n.mu.Lock()
	n.resetElectionTimeout()
	timeout := n.electionTimeout
	n.mu.Unlock()

	if time.Since(time.Unix(0, n.lastContact.Load())) > timeout {
		n.lastContact.Store(time.Now().UnixNano())
	}

	for {
		time.Sleep(10 * time.Millisecond)
		n.mu.Lock()
		timeout = n.electionTimeout
		n.mu.Unlock()
		if time.Since(time.Unix(0, n.lastContact.Load())) > timeout {
			log.Get().Info("Election timeout, starting election", zap.String("node", n.id), zap.Duration("electionTimeout", timeout))
			n.mu.Lock()
			n.state = Candidate
			n.mu.Unlock()
			return
		}
	}
}

func (n *Node) runCandidate() {
	n.mu.Lock()
	n.resetElectionTimeout()
	n.currentTerm++
	n.votedFor = n.id
	term := n.currentTerm
	lastLogIndex := n.lastLogIndex()
	lastTerm := n.lastLogTerm()
	n.mu.Unlock()

	n.lastContact.Store(time.Now().UnixNano())
	votes := 1		// vote for self
	voteCh := make(chan bool, len(n.peers))

	// Request votes from all peers in parallel
	for _, peer := range n.peers {
		go func(peer string) {
			log.Get().Info("Sending RequestVote", zap.String("from", n.id), zap.String("to", peer), zap.Int("term", term),zap.Int("lastLogIndex", lastLogIndex), zap.Int("lastLogTerm", lastTerm))
			args := RequestVoteArgs {
				Term: term,
				CandidateID: n.id,
				LastLogIndex: lastLogIndex,
				LastLogTerm: lastTerm,
			}

			reply, err := n.transport.RequestVote(peer, args)
			if err != nil {
				voteCh <- false
				return
			}

			n.mu.Lock()
			if reply.Term > n.currentTerm {
				n.currentTerm = reply.Term
				n.state = Follower
				n.votedFor = ""
				select {
					case n.stepDownCh <- struct{}{}:
					default:
				}
			}
			n.mu.Unlock()
			voteCh <- reply.VoteGranted
			log.Get().Info("Received RequestVote reply", zap.String("from", peer), zap.Int("term", reply.Term), zap.Bool("voteGranted", reply.VoteGranted))
		}(peer)
	}

	quorum := (len(n.peers) + 1) / 2 + 1
	timer := time.NewTimer(n.electionTimeout)
	defer timer.Stop()

	for {
		select  {
		case granted := <- voteCh:
			if granted {
				votes++
				if votes >= quorum {
					n.becomeLeader()
					return
				}
			}
		
		case <- timer.C:
			// Split vote or timeout, -- restart election
			log.Get().Info("election timeout, restarting...")
			return
		
		case <- n.stepDownCh:
			n.mu.Lock()
			n.state = Follower
			n.mu.Unlock()
			return
		}

	}
}

func (n *Node) becomeLeader(){
	n.mu.Lock()
	defer n.mu.Unlock()
	log.Get().Info("Becoming leader", zap.String("node", n.id), zap.Int("term", n.currentTerm))
	n.state  = Leader

	for _, peer := range n.peers {
		n.nextIndex[peer] = n.lastLogIndex() + 1
		if n.nextIndex[peer] < 1 {
			n.nextIndex[peer] = 1
		}
		n.matchIndex[peer] = 0
	}

	// drain any stepDown signals if any
	select{
		case <- n.stepDownCh:
		default:
	}
}

func (n *Node) runLeader() {
	// Send immediate heartbeats on election win
	n.broadcaseAppendEntries()

	ticker := time.NewTicker(n.heartbeatTimeout)
	defer ticker.Stop()

	for {
		select {
		case <- ticker.C:
			n.broadcaseAppendEntries()

		case <- n.stepDownCh:
			n.mu.Lock()
			n.state = Follower
			n.mu.Unlock()
			return
		}
	}
}

func (n *Node) broadcaseAppendEntries() {
	n.mu.Lock()
	peers := n.peers
	n.mu.Unlock()

	for _, peer := range peers {
		log.Get().Info("sending appendEntries to", zap.String("peer", peer))
		go n.sendAppendEntries(peer)
	}
}

func (n *Node) sendAppendEntries(peer string) {
	n.mu.Lock()

	nextIndex := n.nextIndex[peer]
	prevLogIndex := nextIndex - 1
	if prevLogIndex < 0 {
		prevLogIndex = 0
	}
	prevLogTerm := 0

	if prevLogIndex < len(n.log) {
		prevLogTerm = n.log[prevLogIndex].Term
	}

	// Send all entries from nextIndex onward
	entries := make([]LogEntry, len(n.log[nextIndex:]))
	copy(entries, n.log[nextIndex:])

	args := AppendEntriedArgs {
		Term :  n.currentTerm,
		LeaderID:  n.id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm: prevLogTerm,
		Entries: entries,
		LeaderCommit: n.commitIndex,
	}

	n.mu.Unlock()

	reply, err := n.transport.AppendEntries(peer, args)
	if err != nil {
		 log.Get().Warn("AppendEntries RPC failed",
			zap.String("leader", n.id),
			zap.String("peer", peer),
			zap.Int("nextIndex", args.PrevLogIndex+1),
			zap.Error(err))
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()

	if reply.Term > n.currentTerm {
		n.currentTerm = reply.Term
		n.votedFor = ""
		if n.state == Leader {
			select {
				case n.stepDownCh <- struct{}{}:
				default:
			}
		}
		n.state = Follower
		return
	}

	if reply.Success {
		// update nextIndex and matchIndex 
		n.matchIndex[peer] = prevLogIndex + len(entries)
		n.nextIndex[peer] = n.matchIndex[peer] + 1
		n.advanceCommitIndex()
	}else {
		// Conflict : use hint to skip back fase
		if reply.ConflictIndex >0 {
			n.nextIndex[peer] = reply.ConflictIndex
		}else{
			if n.nextIndex[peer] > 1 {
				n.nextIndex[peer]--
			}
		}
		
	}
}	

// advance commitIndex if a majority have replaceted up to N
func (n *Node) advanceCommitIndex() {
	for idx := n.lastLogIndex();  idx > n.commitIndex; idx-- {
		if n.log[idx].Term != n.currentTerm {
			continue // only commit entries from current term
		}

		replicated := 1

		for _, peer := range n.peers {
			if n.matchIndex[peer] >= idx {
				replicated++
			}
		}

		quorum := (len(n.peers) + 1) / 2 + 1

		if replicated >= quorum {
			n.commitIndex = idx
			n.applyEntries()
			log.Get().Info("CommitIndex advanced", zap.Int("commitIndex", n.commitIndex))
			return
		}
	}
}

func (n *Node) applyEntries() {
	for n.lastApplied < n.commitIndex {
		n.lastApplied++
		entry := n.log[n.lastApplied]
		log.Get().Info("Applying log entry",zap.String("node", n.id),zap.Int("index", entry.Index), zap.Int("term", entry.Term), zap.ByteString("command", entry.Command))
		n.applyCh <- entry
	}
}

// RequestVote RPC handler
func (n *Node) RequestVotes(args RequestVoteArgs, reply *RequestVoteReply) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	log.Get().Info("Received RequestVote", zap.String("candidate", args.CandidateID), zap.Int("term", args.Term), zap.Int("lastLogIndex", args.LastLogIndex), zap.Int("lastLogTerm", args.LastLogTerm))
	reply.Term = n.currentTerm
	reply.VoteGranted = false

	// reject stale candiates
	if args.Term < n.currentTerm {
		log.Get().Info("Rejecting RequestVote from stale candidate", zap.String("candidate", args.CandidateID), zap.Int("term", args.Term), zap.Int("currentTerm", n.currentTerm))
		return nil
	}

	if args.Term > n.currentTerm {
		n.currentTerm = args.Term
		n.votedFor = ""
		if n.state == Leader {
			// Reset the election timeout after get newer term to avoid split vote
			select {
			case n.stepDownCh <- struct{}{}:
			default:
			}
		}
		n.state = Follower
		n.resetElectionTimeout()
	}

	// Grant vote only if we havn't voted yet and candidates's log is at least as up-to-date
	logOk := args.LastLogTerm > n.lastLogTerm() || 
		   (args.LastLogTerm == n.lastLogTerm() && args.LastLogIndex >= n.lastLogIndex())
	
	if (n.votedFor == "" || n.votedFor == args.CandidateID) && logOk{
		n.votedFor = args.CandidateID
		reply.VoteGranted = true
		log.Get().Info("Granting vote to candidate", zap.String("candidate", args.CandidateID), zap.Int("term", args.Term))
		n.lastContact.Store(time.Now().UnixNano())
	}

	return nil
}


// AppendEntries RPC handler 
func (n *Node) AppendEntries(args AppendEntriedArgs, reply *AppendEntriesReply) error {       
	n.mu.Lock()
	defer n.mu.Unlock()
	reply.Term = n.currentTerm
	reply.Success = false
	log.Get().Info("Received AppendEntries ", zap.String("leader", args.LeaderID), zap.String("curNode", n.id),zap.Int("term", args.Term),zap.Int("currentTerm", n.currentTerm), zap.Int("entries", len(args.Entries)), zap.Int("prevLogIndex", args.PrevLogIndex), zap.Int("prevLogTerm", args.PrevLogTerm))
	// 1. Reject stale leaders
	if args.Term < n.currentTerm {
		log.Get().Info("Rejecting AppendEntries from stale leader", zap.String("leader", args.LeaderID), zap.Int("term", args.Term), zap.Int("currentTerm", n.currentTerm))
		return nil
	}
	
	n.lastContact.Store(time.Now().UnixNano())

	 if args.Term > n.currentTerm {
		if n.state == Leader || n.state == Candidate {
			// Step down immediately if we learn about a higher term
			select {
			case n.stepDownCh <- struct{}{}:
			default:
			}
	 	}
		n.currentTerm = args.Term
		n.votedFor = ""
		n.state = Follower
	 }else if n.state == Candidate {
		select {
			case n.stepDownCh <- struct{}{}:
			default:
		}
		n.state = Follower
		log.Get().Info("Stepping down to follower due to AppendEntries from current leader", zap.String("leader", args.LeaderID), zap.Int("term", args.Term))
	 }

	 // 2. Consistency check -- does our log contain prevLogIndex at prevLogTerm ?
	 if args.PrevLogIndex >= len(n.log) {
		reply.ConflictIndex = len(n.log)
		reply.ConflictTerm = -1
		log.Get().Info("Rejecting AppendEntries due to log inconsistency", zap.String("leader", args.LeaderID), zap.Int("term", args.Term), zap.Int("prevLogIndex", args.PrevLogIndex), zap.Int("logLength", len(n.log)))
		return nil
	 }

	 if n.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		// find first index of the conflicting term for fast rollback
		conflictTerm := n.log[args.PrevLogIndex].Term
		reply.ConflictTerm = conflictTerm

		for i:=1; i<len(n.log); i++ {
			if n.log[i].Term == conflictTerm {
				reply.ConflictIndex = i
				break
			}
		}
		log.Get().Info("Rejecting AppendEntries due to term mismatch", zap.String("leader", args.LeaderID), zap.Int("term", args.Term), zap.Int("prevLogIndex", args.PrevLogIndex), zap.Int("prevLogTerm", args.PrevLogTerm))
		return nil
	 }

	 // 3. Append new entries, overwriteing conflicting ones
	 for i,entry := range args.Entries {
		logIdx := args.PrevLogIndex + 1 + i
		if logIdx < len(n.log) {
			if n.log[logIdx].Term != entry.Term {
				// delete conflicting suffix
				n.log = n.log[:logIdx]
				n.log = append(n.log, args.Entries[i:]...)
			}
		}else {
			n.log = append(n.log, args.Entries[i:]...)
			break
		}
	 }

	 // 4. advance commitIndex 
	 if args.LeaderCommit > n.commitIndex {
		n.commitIndex = min(args.LeaderCommit, n.lastLogIndex())
		n.applyEntries()
	 }
	 reply.Success = true
	 n.resetElectionTimeout()
	 log.Get().Info("AppendEntries successful", zap.String("leader", args.LeaderID), zap.Int("term", args.Term))

	return nil
}
