package main

import (
	"mockservice/backend/log"
	"mockservice/backend/raft"
	"time"

	"go.uber.org/zap"
)


func main() {
	log.Init("raft.log")
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
	var leaderIdx int
	for i, csm:= range csms {
		err := csm.Synchronize("_test", "test")
		if err == nil {
			leader = csm
			leaderIdx = i
			log.Get().Info("Leader found at index ", zap.Int("index", leaderIdx))
			break
		}
		csm.Delete("_test")
	}

	if leader == nil {
		log.Get().Fatal("no leader found")
	}

	err := leader.Synchronize("db.host", "postgress-primary.internal")

	if err != nil {
		log.Get().Error("Error: %v", zap.Error(err))
	}

	// Read from any node (all converge to same value after commit)
	time.Sleep(500 * time.Millisecond)

	for i, csm := range csms {
		v, _ := csm.Get("db.host")
		log.Get().Info("Node sees db.host ", zap.Int("index", i), zap.String("value", v))
	}

}

