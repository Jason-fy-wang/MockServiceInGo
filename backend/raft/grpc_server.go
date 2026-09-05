package raft

import (
	"context"
	"mockservice/backend/raft/raftpb"
)

type GrpcNodeServer struct {
	raftpb.UnimplementedRaftServiceServer
	node *Node
}

// implements the Transport interface, to compatable with the existing raft code. It uses gRPC to communicate with other nodes.
func (s *GrpcNodeServer) RequestVote(ctx context.Context, req *raftpb.RequestVoteRequest) (*raftpb.RequestVoteResponse, error) {

	args := RequestVoteArgs{
		Term:         int(req.Term),
		CandidateID:  req.CandidateId,
		LastLogIndex: int(req.LastLogIndex),
		LastLogTerm:  int(req.LastLogTerm),
	}

	var reply RequestVoteReply
	err := s.node.RequestVotes(args, &reply)

	if err != nil {
		return nil, err
	}

	return &raftpb.RequestVoteResponse{
		Term:        int64(reply.Term),
		VoteGranted: reply.VoteGranted,
	}, nil
}

func (s *GrpcNodeServer) AppendEntries(ctx context.Context, req *raftpb.AppendEntriesRequest) (*raftpb.AppendEntriesResponse, error) {

	args := AppendEntriedArgs{
		Term:         int(req.Term),
		LeaderID:     req.LeaderId,
		PrevLogIndex: int(req.PrevLogIndex),
		PrevLogTerm:  int(req.PrevLogTerm),
		Entries:      make([]LogEntry, len(req.Entries)),
		LeaderCommit: int(req.LeaderCommit),
	}

	for i, entry := range req.Entries {
		args.Entries[i] = LogEntry{
			Index:   int(entry.Index),
			Term:    int(entry.Term),
			Command: entry.Command,
		}
	}

	var reply AppendEntriesReply
	err := s.node.AppendEntries(args, &reply)

	if err != nil {
		return nil, err
	}

	return &raftpb.AppendEntriesResponse{
		Term:          int64(reply.Term),
		Success:       reply.Success,
		ConflictTerm:  int64(reply.ConflictTerm),
		ConflictIndex: int64(reply.ConflictIndex),
	}, nil
}

func (s *GrpcNodeServer) Propose(ctx context.Context, req *raftpb.ProposeRequest) (*raftpb.ProposeResponse, error) {
	args := ProposeArgs{
		Commond: req.Command,
	}

	var reply ProposeReply
	err := s.node.Propose(args, &reply)

	if err != nil {
		return nil, err
	}

	return &raftpb.ProposeResponse{
		Success: reply.Success,
		Leader:  reply.Leader,
	}, nil
}
