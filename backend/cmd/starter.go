package main

import (
	"fmt"

	"mockservice/backend/api"
	"mockservice/backend/log"
)

func main() {
	log.Init("mockservice.log")
	logger := log.Get()
	logger.Info("starting application")
	service := api.NewMockService()
	if err := service.Run(":8080"); err != nil {
		fmt.Printf("failed to run server: %v\n", err)
	}
}
