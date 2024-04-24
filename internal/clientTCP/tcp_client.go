package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	serverAddr := "localhost:3000"
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	for {
		// Prompt for the upstream name
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the upstream name: ")
		upstreamName, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read upstream name: %v", err)
		}

		fmt.Fprint(conn, upstreamName)
	}

}
