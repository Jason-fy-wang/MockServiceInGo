package main

import (
	"fmt"
	"log"
	"mockservice/backend/raft"
	"time"
)


func main() {
	perrs := []string{"localhost:8081", "localhost:8082", "localhost:8083"}

	nodes := make([]*raft.Node, 3)
	csms := make([]*raft.ConfigStateMachine, 3)

	for i, addr := range perrs {
		others := []string{}

		for j, p := range perrs {
			if j != i {
				others = append(others, p)
			}
		}
		transport := &raft.TCPTransport{}
		node := raft.NewNode(addr, others, transport)
		nodes = append(nodes,node)
		transport.Listern(addr, node)

		csms[i] = raft.NewConfigStateMachine((node))
		go node.Run()
	}

	time.Sleep(300 * time.Millisecond)

	err := csms[0].Synchronize("db.host", "postgress-primary.internal")

	if err != nil {
		log.Fatalf("Error", err)
	}

	// Read from any node (all converge to same value after commit)
	time.Sleep(100 * time.Millisecond)

	for i, csm := range csms {
		v, _ := csm.Get("db.host")
		fmt.Printf("Node %s sess db.host = %s", i, v)
	}

}

