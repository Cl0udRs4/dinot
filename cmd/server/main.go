package main

import (
	"fmt"
	"os"

	"github.com/Cl0udRs4/dinot/internal/server/cli"
)

func main() {
	// Create and start the server
	server := cli.NewServer()
	
	// Start the server
	if err := server.Start(); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
