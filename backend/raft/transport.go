package raft

import (
	"net"
	"net/rpc"
	"time"
)

// Network transport layer

type Transport interface {
	RequestVote(peer string, args RequestVoteArgs) (RequestVoteReply, error)
	AppendEntries(peer string, args AppendEntriedArgs) (AppendEntriesReply, error)
}


type TCPTransport struct{}


func (t *TCPTransport) RequestVote(peer string, args RequestVoteArgs) (RequestVoteReply, error) {
	var reply RequestVoteReply

	return reply, t.call(peer,"Node.RequestVote", args, &reply)
}


func (t *TCPTransport) AppendEntries(peer string, args AppendEntriedArgs) (AppendEntriesReply, error) {

	var reply AppendEntriesReply
	
	return reply, t.call(peer,"Node.AppendEntries", args, &reply)
}

func (t *TCPTransport) call (peer, method string, args, reply interface{}) error {
	conn, err := net.DialTimeout("tcp", peer, 100 * time.Microsecond)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := rpc.NewClient(conn)
	return client.Call(method, args, reply)
}


func(t *TCPTransport) Listern(addr string, node *Node) error {
	rpc.Register(node)

	ln, err := net.Listen("tcp", addr)

	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()

	return nil
}
