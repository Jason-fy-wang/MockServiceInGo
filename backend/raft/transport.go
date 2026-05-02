package raft

import (
	"mockservice/backend/log"
	"net"
	"net/rpc"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Network transport layer
type Transport interface {
	RequestVote(peer string, args RequestVoteArgs) (RequestVoteReply, error)
	AppendEntries(peer string, args AppendEntriedArgs) (AppendEntriesReply, error)
}


type TCPTransport struct{
	mu sync.Mutex
	Clients map[string]*rpc.Client
}

func NewTCPTransport() *TCPTransport {
	return &TCPTransport{
		Clients: make(map[string]*rpc.Client),
	}
}

func (t *TCPTransport) RequestVote(peer string, args RequestVoteArgs) (RequestVoteReply, error) {
	var reply RequestVoteReply

	return reply, t.call(peer,"Node.RequestVotes", args, &reply)
}


func (t *TCPTransport) AppendEntries(peer string, args AppendEntriedArgs) (AppendEntriesReply, error) {

	var reply AppendEntriesReply
	
	return reply, t.call(peer,"Node.AppendEntries", args, &reply)
}

func (t *TCPTransport) getClient(peer string) (*rpc.Client, error) {
	if client, ok := t.Clients[peer]; ok {
		return client, nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	log.Get().Info("Dialing to ", zap.String("peer", peer))
	conn, err := net.DialTimeout("tcp", peer, 100 * time.Microsecond)
	if err != nil {
		return nil, err
	}

	client := rpc.NewClient(conn)
	t.Clients[peer] = client
	return client, nil
}

func (t *TCPTransport) call (peer, method string, args, reply interface{}) error {
	client, err := t.getClient(peer)
	if err != nil {
		return err
	}
	return client.Call(method, args, reply)
}


func(t *TCPTransport) Listen(addr string, node *Node) error {
	rpc.Register(node)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Get().Info("Listening on ", zap.String("address", addr))

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Get().Error("Error accepting connection: %v", zap.Error(err))
				continue
			}
			log.Get().Info("Accepted connection from ", zap.String("remote", conn.RemoteAddr().String()))
			go rpc.ServeConn(conn)
		}
	}()

	return nil
}
