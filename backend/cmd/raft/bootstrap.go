package main

import (
	. "mockservice/backend/common"
	"mockservice/backend/log"
	"mockservice/backend/raft"
	"time"

	"go.uber.org/zap"
)

/*
*

	Raft test bootstrapper. This is a simple bootstrapper for testing the Raft implementation. It creates a cluster of 3 nodes, starts them, and performs some operations to demonstrate the Raft consensus algorithm.
*/
func main() {
	Startconfig := &StarterConfig{
		LogFile:     "./raft.log",
		ServiceAddr: ":8080",
		Image:       true,
		Raft: struct {
			Enabled bool     `json:"enabled"`
			Address string   `json:"address"`
			Peers   []string `json:"peers"`
		}{
			Enabled: true,
			Address: "",
		},
	}
	log.Init(Startconfig)
	perrs := []string{"127.0.0.1:8081", "127.0.0.1:8082", "127.0.0.1:8083"}

	nodes := make([]*raft.Node, 3)
	csms := make([]*raft.ConfigStateMachine, 3)

	for i, addr := range perrs {
		others := []string{}

		for j, p := range perrs {
			if j != i {
				others = append(others, p)
			}
		}
		transport := raft.NewTCPTransport()
		node := raft.NewNode(addr, others, transport)
		nodes[i] = node
		transport.Listen(addr, node)

		csms[i] = raft.NewConfigStateMachine(node)
		go node.Run()
	}

	time.Sleep(2 * time.Second)

	// get leader
	var leader *raft.ConfigStateMachine
	for _, csm := range csms {
		if csm.Node.IsLeader() {
			leader = csm
			log.Get().Info("Leader found at index ", zap.String("node", csm.Node.Id()))
			break
		}
	}

	if leader == nil {
		log.Get().Fatal("no leader found")
	}

	err := leader.Synchronize(OperationAdd, "db.host", "postgress-primary.internal")

	if err != nil {
		log.Get().Error("Error: %v", zap.Error(err))
	}

	// Read from any node (all converge to same value after commit)
	time.Sleep(1500 * time.Millisecond)

	for i, csm := range csms {
		v, _ := csm.Get("db.host")
		log.Get().Info("Node sees db.host ", zap.String("node", csm.Node.Id()), zap.Int("index", i), zap.String("value", v))
	}

}
