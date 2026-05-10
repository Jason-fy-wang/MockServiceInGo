package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"mockservice/backend/api"
	"mockservice/backend/log"
	"mockservice/backend/raft"
	"os"
)

type starterConfig struct {
	ServiceAddr string `json:"serviceAddr"`
	RulesFile   string `json:"rulesFile"`
	LogFile     string `json:"logFile"`
	Raft        struct {
		Enabled bool     `json:"enabled"`
		Address string   `json:"address"`
		Peers   []string `json:"peers"`
	} `json:"raft"`
}

func main() {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	configFile := fs.String("config", "", "path to config file")
	fs.Parse(os.Args[1:])

	if configFile == nil || *configFile == "" {
		fmt.Fprintln(os.Stderr, "config file is required")
		fs.Usage()
		os.Exit(1)
	}

	cfg, err := loadStarterConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logFile := cfg.LogFile
	if logFile == "" {
		logFile = "mockservice.log"
	}
	log.Init(logFile)
	logger := log.Get()
	logger.Info("starting application")

	var csm *raft.ConfigStateMachine
	if cfg.Raft.Enabled {
		transport := raft.NewTCPTransport()
		node := raft.NewNode(cfg.Raft.Address, cfg.Raft.Peers, transport)
		transport.Listen(cfg.Raft.Address, node)
		csm = raft.NewConfigStateMachine(node)
		go node.Run()
		
		// only the leader should start the service, followers will wait until they become leader
		// for{
		// 	if csm != nil && csm.Node.IsLeader() {
		// 		break
		// 	}else{
		// 		// fmt.Printf("waiting to be leader.")
		// 		time.Sleep(2 * time.Second)
		// 	}
		// }
	}

	serviceAddr := cfg.ServiceAddr
	if serviceAddr == "" {
		serviceAddr = ":8080"
	}
	service := api.NewMockServiceWithOptions(api.ServiceOptions{
		FilePath:   cfg.RulesFile,
		RaftConfig: csm,
	})
	if err := service.Run(serviceAddr); err != nil {
		fmt.Printf("failed to run server: %v\n", err)
	}
}

func loadStarterConfig(path string) (*starterConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg starterConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
