package raft

import (
	"math/rand/v2"
	"mockservice/backend/log"
	"sync"
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
	grantVoteCh chan struct{}  // reset election timer on vote grant
	heartbeatCh chan struct{}  // reset election timer on valid heartbeat

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
		grantVoteCh: make(chan struct{},1),
		heartbeatCh: make(chan struct{},1),
		transport: transport,
		heartbeatTimeout: 50 * time.Microsecond,
	}

	n.resetElectionTimeout()
	return n
}


func (n *Node) resetElectionTimeout() {
	
	n.electionTimeout = time.Duration(150 + rand.IntN(150)) * time.Millisecond
}

func (n *Node) lastLogIndex() int {
	return n.log[len(n.log)-1].Index
}

func (n *Node) lastLogTerm() int {
	return n.log[len(n.log)-1].Term
}


// main loop
func (n *Node) Run() {
	for{
		n.mu.Lock()
		state := n.state
		n.mu.Unlock()
		log.Get().Info("Node state ", zap.String("id", n.id), zap.Int("state", int(state)))
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
	timer := time.NewTimer(n.electionTimeout)
	defer timer.Stop()

	for {
		select {
		case <- timer.C:
			// timeout waiting for leader.  start election
			n.mu.Lock()
			n.state = Candidate
			n.mu.Unlock()
			return
		case <-n.heartbeatCh:
			// valid AppendEntries received. reset timer
			timer.Reset(n.electionTimeout)
		case <-n.grantVoteCh:
			// just granted a vote, reset timer
			timer.Reset(n.electionTimeout)
		}
	}
}

func (n *Node) runCandidate() {
	n.mu.Lock()
	n.currentTerm++
	n.votedFor = n.id
	term := n.currentTerm
	lastLogIndex := n.lastLogIndex()
	lastTerm := n.lastLogTerm()
	n.mu.Unlock()

	votes := 1		// vote for self
	voteCh := make(chan bool, len(n.peers))

	// Request votes from all peers in parallel
	for _, peer := range n.peers {
		go func(peer string) {
			log.Get().Info("Sending RequestVote ", zap.String("from", n.id), zap.String("to", peer), zap.Int("term", term))
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
			}
			n.mu.Unlock()
			voteCh <- reply.VoteGranted
			
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

	n.state  = Leader

	for _, peer := range n.peers {
		n.nextIndex[peer] = n.lastLogIndex() + 1
		n.matchIndex[peer] = 0
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
		go n.sendAppendEntries(peer)
	}
}

func (n *Node) sendAppendEntries(peer string) {
	n.mu.Lock()

	nextIndex := n.nextIndex[peer]
	prevLogIndex := func() int {
		if (nextIndex - 1) < 0 {
			return 0
		}
		return nextIndex - 1
	}()
	prevLogTerm := 0

	if prevLogIndex >= 0 && prevLogIndex < len(n.log) {
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
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()

	if reply.Term > n.currentTerm {
		n.currentTerm = reply.Term
		n.state = Follower
		n.votedFor = ""
		select {
			case n.stepDownCh <- struct{}{}:
			default:
		}
		return
	}

	if reply.Success {
		// update nextIndex and matchIndex 
		n.matchIndex[peer] = prevLogIndex + len(entries)
		n.nextIndex[peer] = n.matchIndex[peer] + 1
		n.advanceCommitIndex()
	}else {
		// Conflict : use hint to skip back fase
		n.nextIndex[peer] = reply.ConflictIndex
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
			return
		}
	}
}

func (n *Node) applyEntries() {
	for n.lastApplied < n.commitIndex {
		n.lastApplied++
		entry := n.log[n.lastApplied]
		n.applyCh <- entry
	}
}

// RequestVote RPC handler
func (n *Node) RequestVotes(args RequestVoteArgs, reply *RequestVoteReply) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	log.Get().Info("Received RequestVote ", zap.String("candidate", args.CandidateID), zap.Int("term", args.Term))
	reply.Term = n.currentTerm
	reply.VoteGranted = false

	// reject stale candiates
	if args.Term < n.currentTerm {
		return nil
	}

	if args.Term > n.currentTerm {
		n.currentTerm = args.Term
		n.state = Follower
		n.votedFor = ""
	}

	// Grant vote only if we havn't voted yet and candidates's log is at least as up-to-date
	logOk := args.LastLogIndex > n.lastLogTerm() || 
		   (args.LastLogTerm == n.lastLogTerm() && args.LastLogIndex >= n.lastLogIndex())
	
	if (n.votedFor == "" || n.votedFor == args.CandidateID) && logOk{
		n.votedFor = args.CandidateID
		reply.VoteGranted = true

		select {
			case n.grantVoteCh <- struct{}{}:
			default:
		}
	}

	return nil
}


// AppendEntries RPC handler 
func (n *Node) AppendEntries(args AppendEntriedArgs, reply *AppendEntriesReply) error {       
	n.mu.Lock()
	defer n.mu.Unlock()
	reply.Term = n.currentTerm
	reply.Success = false
	log.Get().Info("Received AppendEntries ", zap.String("leader", args.LeaderID), zap.Int("term", args.Term), zap.Int("entries", len(args.Entries)), zap.Int("prevLogIndex", args.PrevLogIndex))
	// 1. Reject stale leaders
	if args.Term < n.currentTerm {
		return nil
	}
	
	 // valid heartbeat -- reset election timer
	 select {
		case n.heartbeatCh <- struct{}{}:
		default:
	 }

	 if args.Term > n.currentTerm {
		n.currentTerm = args.Term
		n.state = Follower
		n.votedFor = ""
	 }

	 // 2. Consistency check -- does our log contain prevLogIndex at prevLogTerm ?
	 if args.PrevLogIndex >= len(n.log) {
		reply.ConflictIndex = len(n.log)
		reply.ConflictTerm = -1
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

	return nil
}
