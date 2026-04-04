package main

import (
	"fmt"

	"mockservice/backend/api"
)

func main() {
	fmt.Println("Starting mock service on :8080")
	service := api.NewMockService()
	if err := service.Run(":8080"); err != nil {
		fmt.Printf("failed to run server: %v\n", err)
	}
}
