package loadbalancer

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/janithht/GoStreamBalancer/internal/config"
	"github.com/janithht/GoStreamBalancer/internal/helpers"
)

func StartLoadBalancer(upstreams []config.Upstream) {
	upstreamMap := make(map[string]*config.RoundRobinIterator)
	for i := range upstreams {
		upstream := &upstreams[i]
		iterator := config.NewRoundRobinIterator()
		for _, server := range upstream.Servers {
			iterator.Add(server) // Add all servers initially
		}
		upstreamMap[strings.ToLower(upstream.Name)] = iterator
	}

	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatalf("Failed to listen on port 3000: %v", err)
	}
	defer listener.Close()
	fmt.Println()
	log.Println("Load balancer running on port 3000")

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(clientConn, upstreamMap)
	}
}

func handleConnection(clientConn net.Conn, upstreamMap map[string]*config.RoundRobinIterator) {
	defer clientConn.Close()

	reader := bufio.NewReader(clientConn)
	fmt.Fprint(clientConn, "Enter upstream name: ")
	upstreamName, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Failed to read from connection: %v", err)
		return
	}
	upstreamName = strings.TrimSpace(upstreamName)

	iterator, exists := upstreamMap[strings.ToLower(upstreamName)]
	if !exists {
		fmt.Fprintf(clientConn, "Upstream not found: %s\n", upstreamName)
		return
	}

	server := iterator.NextHealthy()
	if server == nil {
		fmt.Fprintf(clientConn, "No available servers for upstream: %s\n", upstreamName)
		return
	}

	host, port, err := helpers.ParseHostPort(server.Url)
	if err != nil {
		fmt.Fprintf(clientConn, "Invalid server URL: %v\n", err)
		return
	}

	serverConn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		fmt.Fprintf(clientConn, "Failed to connect to server: %v\n", err)
		return
	}
	defer clientConn.Close()
	defer serverConn.Close()

	// Sends a HTTP GET request to the server
	fmt.Fprintf(serverConn, "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")

	// Proxy data between the client and the server
	go helpers.ProxyData(clientConn, serverConn)
	helpers.ProxyData(serverConn, clientConn)
}
