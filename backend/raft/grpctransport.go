package raft

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"mockservice/backend/common"
	"mockservice/backend/log"
	"mockservice/backend/raft/raftpb"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCTransport struct {
	mu      sync.Mutex
	Clients map[string]*grpc.ClientConn
	Node    *Node
	server  *grpc.Server
	cfg     *common.StarterConfig
}

func NewGRPCTransport(cfg *common.StarterConfig) *GRPCTransport {
	return &GRPCTransport{
		Clients: make(map[string]*grpc.ClientConn),
		cfg:     cfg,
	}
}

func (t *GRPCTransport) getClient(peer string) (raftpb.RaftServiceClient, error) {

	t.mu.Lock()
	defer t.mu.Unlock()

	if conn, ok := t.Clients[peer]; ok {
		return raftpb.NewRaftServiceClient(conn), nil
	}
	idx := strings.Index(peer, ":")
	servername := peer[:idx]

	if t.cfg.Raft.TLS.Enable && t.cfg.Raft.TLS.MTLS {
		caPem, err := os.ReadFile(t.cfg.Raft.TLS.CAFile)
		if err != nil {
			log.Get().Error("failed to load ca pem", zap.Error(err))
			os.Exit(1)
		}
		rootCAs := x509.NewCertPool()
		if !rootCAs.AppendCertsFromPEM(caPem) {
			log.Get().Error("failed to append ca pem")
			os.Exit(1)
		}
		clientCert, err := tls.LoadX509KeyPair(t.cfg.Raft.TLS.CertFile, t.cfg.Raft.TLS.KeyFile)
		if err != nil {
			log.Get().Error("failed to load client cert and key", zap.Error(err))
			os.Exit(1)
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      rootCAs,
			MinVersion:   tls.VersionTLS12,
			ServerName:   servername,
		}

		creds := credentials.NewTLS(tlsConfig)
		opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}
		conn, err := grpc.NewClient(peer, opts...)
		if err != nil {
			log.Get().Error("Failed to create gRPC client for peer:", zap.String("peer", peer), zap.Error(err))
			os.Exit(1)
		}
		t.Clients[peer] = conn
		return raftpb.NewRaftServiceClient(conn), nil

	} else if t.cfg.Raft.TLS.Enable {
		var opts []grpc.DialOption
		if t.cfg.Raft.TLS.CAFile == "" || t.cfg.Raft.TLS.CertFile == "" || t.cfg.Raft.TLS.KeyFile == "" {
			log.Get().Error("TLS is enabled but CAFile, CertFile, or KeyFile is not provided")
			os.Exit(1)
		}
		creds, err := credentials.NewClientTLSFromFile(t.cfg.Raft.TLS.CAFile, servername)
		if err != nil {
			log.Get().Error("Failed to create TLS credentials: ", zap.Error(err))
			os.Exit(1)
		}
		opts = []grpc.DialOption{grpc.WithTransportCredentials(creds)}
		conn, err := grpc.NewClient(peer, opts...)
		if err != nil {
			log.Get().Error("Failed to create gRPC client for peer:", zap.String("peer", peer), zap.Error(err))
			os.Exit(1)
		}
		t.Clients[peer] = conn
		return raftpb.NewRaftServiceClient(conn), nil
	}

	conn, err := grpc.NewClient(peer, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Get().Error("Failed to create gRPC client for peer:", zap.String("peer", peer), zap.Error(err))
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
		log.Get().Error("Failed to request vote from peer:", zap.String("peer", peer), zap.Error(err))
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
		log.Get().Error("Failed to append entries to peer:", zap.String("peer", peer), zap.Error(err))
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
		Command: args.Command,
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

func (t *GRPCTransport) Close() {
	if t.server != nil {
		t.server.GracefulStop()
	}
}

func (t *GRPCTransport) certNormalize() {
	if t.cfg.Raft.TLS.CAFile == "" || t.cfg.Raft.TLS.CertFile == "" || t.cfg.Raft.TLS.KeyFile == "" {
		log.Get().Error("TLS is enabled but CAFile, CertFile, or KeyFile is not provided")
		os.Exit(1)
	}

	capath, err := filepath.Abs(t.cfg.Raft.TLS.CAFile)
	if err != nil {
		log.Get().Error("Failed to resolve CAFile path: ", zap.String("path", t.cfg.Raft.TLS.CAFile), zap.Error(err))
		os.Exit(1)
	}
	t.cfg.Raft.TLS.CAFile = capath

	keypath, err := filepath.Abs(t.cfg.Raft.TLS.KeyFile)
	if err != nil {
		log.Get().Error("Failed to resolve KeyFile path: ", zap.String("path", t.cfg.Raft.TLS.KeyFile), zap.Error(err))
		os.Exit(1)
	}
	t.cfg.Raft.TLS.KeyFile = keypath

	certpath, err := filepath.Abs(t.cfg.Raft.TLS.CertFile)
	if err != nil {
		log.Get().Error("Failed to resolve CertFile path: ", zap.String("path", t.cfg.Raft.TLS.CertFile), zap.Error(err))
		os.Exit(1)
	}
	t.cfg.Raft.TLS.CertFile = certpath

}

func (t *GRPCTransport) Listen(addr string, node *Node) error {
	t.Node = node

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Get().Error("Failed to listen on address: ", zap.String("address", addr), zap.Error(err))
		return err
	}
	var opts []grpc.ServerOption
	if t.cfg.Raft.TLS.Enable && t.cfg.Raft.TLS.MTLS {
		t.certNormalize()
		caPem, err := os.ReadFile(t.cfg.Raft.TLS.CAFile)
		if err != nil {
			log.Get().Error("failed to load ca pem", zap.Error(err))
			os.Exit(1)
		}
		clientCAs := x509.NewCertPool()
		if !clientCAs.AppendCertsFromPEM(caPem) {
			log.Get().Error("failed to append ca pem")
			os.Exit(1)
		}
		serverCert, err := tls.LoadX509KeyPair(t.cfg.Raft.TLS.CertFile, t.cfg.Raft.TLS.KeyFile)
		if err != nil {
			log.Get().Error("failed to load server cert and key", zap.Error(err))
			os.Exit(1)
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{serverCert},
			ClientCAs:    clientCAs,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			MinVersion:   tls.VersionTLS12,
		}
		opts = []grpc.ServerOption{grpc.Creds(credentials.NewTLS(tlsConfig))}

	} else if t.cfg.Raft.TLS.Enable {
		t.certNormalize()
		creds, err := credentials.NewServerTLSFromFile(t.cfg.Raft.TLS.CertFile, t.cfg.Raft.TLS.KeyFile)
		if err != nil {
			log.Get().Error("Failed to create TLS credentials: ", zap.Error(err))
			os.Exit(1)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	t.server = grpc.NewServer(opts...)

	raftpb.RegisterRaftServiceServer(t.server, &GrpcNodeServer{node: node})

	go func() {
		if err := t.server.Serve(listener); err != nil {
			log.Get().Error("Failed to serve gRPC server:", zap.String("address", addr), zap.Error(err))
		}
	}()

	return nil
}
