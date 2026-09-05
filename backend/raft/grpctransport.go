package raft

import (
	"context"
	"mockservice/backend/log"
	"mockservice/backend/raft/raftpb"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCTransport struct {
	mu      sync.Mutex
	Clients map[string]*grpc.ClientConn
	Node    *Node
	server  *grpc.Server
}

func NewGRPCTransport() *GRPCTransport {
	return &GRPCTransport{
		Clients: make(map[string]*grpc.ClientConn),
	}
}

func (t *GRPCTransport) getClient(peer string) (raftpb.RaftServiceClient, error) {

	t.mu.Lock()
	defer t.mu.Unlock()

	if conn, ok := t.Clients[peer]; ok {
		return raftpb.NewRaftServiceClient(conn), nil
	}

	conn, err := grpc.NewClient(peer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Get().Error("Failed to create gRPC client for peer: %v", zap.String("peer", peer), zap.Error(err))
		return nil, err
	}
	t.Clients[peer] = conn
	return raftpb.NewRaftServiceClient(conn), nil
}

func (t *GRPCTransport) RequestVote(peer string, req RequestVoteArgs) (RequestVoteReply, error) {
	client, err := t.getClient(peer)
	if err != nil {
		return RequestVoteReply{}, err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		1000*time.Millisecond,
	)

	defer cancel()

	reply, err := client.RequestVote(ctx, &raftpb.RequestVoteRequest{
		Term:         int64(req.Term),
		CandidateId:  req.CandidateID,
		LastLogIndex: int64(req.LastLogIndex),
		LastLogTerm:  int64(req.LastLogTerm),
	})

	if err != nil {
		log.Get().Error("Failed to request vote from peer: %v", zap.String("peer", peer), zap.Error(err))
		return RequestVoteReply{}, err
	}

	return RequestVoteReply{
		Term:        int(reply.Term),
		VoteGranted: reply.VoteGranted,
	}, nil
}

func (t *GRPCTransport) AppendEntries(peer string, args AppendEntriedArgs) (AppendEntriesReply, error) {
	client, err := t.getClient(peer)
	if err != nil {
		return AppendEntriesReply{}, err
	}

	entries := make([]*raftpb.LogEntry, len(args.Entries))
	for i, e := range args.Entries {
		entries[i] = &raftpb.LogEntry{
			Term:    int64(e.Term),
			Index:   int64(e.Index),
			Command: e.Command,
		}
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		1000*time.Millisecond,
	)

	defer cancel()

	reply, err := client.AppendEntries(ctx, &raftpb.AppendEntriesRequest{
		Term:         int64(args.Term),
		LeaderId:     args.LeaderID,
		PrevLogIndex: int64(args.PrevLogIndex),
		PrevLogTerm:  int64(args.PrevLogTerm),
		Entries:      entries,
		LeaderCommit: int64(args.LeaderCommit),
	})

	if err != nil {
		log.Get().Error("Failed to append entries to peer: %v", zap.String("peer", peer), zap.Error(err))
		return AppendEntriesReply{}, err
	}

	return AppendEntriesReply{
		Term:          int(reply.Term),
		Success:       reply.Success,
		ConflictTerm:  int(reply.ConflictTerm),
		ConflictIndex: int(reply.ConflictIndex),
	}, nil
}

func (t *GRPCTransport) Propose(peer string, args ProposeArgs) (ProposeReply, error) {
	client, err := t.getClient(peer)
	if err != nil {
		log.Get().Error("Failed to get gRPC client for peer:", zap.String("peer", peer), zap.Error(err))
		return ProposeReply{}, err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		1000*time.Millisecond,
	)
	defer cancel()

	reply, err := client.Propose(ctx, &raftpb.ProposeRequest{
		Command: args.Commond,
	})
	if err != nil {
		log.Get().Error("propose exception: ", zap.String("peer", peer), zap.Error(err))
		return ProposeReply{}, err
	}

	return ProposeReply{
		Success: reply.Success,
		Leader:  reply.Leader,
	}, nil
}

func (t *GRPCTransport) Listen(addr string, node *Node) error {
	t.Node = node

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Get().Error("Failed to listen on address: ", zap.String("address", addr), zap.Error(err))
		return err
	}

	t.server = grpc.NewServer()

	raftpb.RegisterRaftServiceServer(t.server, &GrpcNodeServer{node: node})

	go func() {
		if err := t.server.Serve(listener); err != nil {
			log.Get().Error("Failed to serve gRPC server:", zap.String("address", addr), zap.Error(err))
		}
	}()

	return nil
}
