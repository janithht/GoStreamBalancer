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

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the upstream name: ")
	upstreamName, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read upstream name: %v", err)
	}

	_, err = fmt.Fprint(conn, upstreamName)
	if err != nil {
		log.Fatalf("Failed to send upstream name to server: %v", err)
	}

	// Read the response from the server
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Printf("Error reading response: %v", err)
		if err.Error() == "EOF" {
			log.Println("The server closed the connection unexpectedly. Please check server logs.")
		}
		return
	}

	fmt.Printf("Received response: %s", response)
}
