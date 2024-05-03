package servertcp

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/janithht/GoStreamBalancer/internal/config"
	"github.com/janithht/GoStreamBalancer/internal/helpers"
)

func StartLoadBalancers(upstreamMap map[string]*config.LeastConnectionsIterator, portMap map[int]string) {
	for port, upstreamName := range portMap {
		go func(port int, upstreamName string) {
			listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err != nil {
				log.Printf("Failed to listen on port %d: %v", port, err)
				return
			}
			defer listener.Close()
			log.Printf("Load balancer for %s running on port %d", upstreamName, port)

			for {
				clientConn, err := listener.Accept()
				if err != nil {
					log.Printf("Failed to accept connection on port %d: %v", port, err)
					continue
				}
				go handleConnection(clientConn, upstreamName, upstreamMap)
				fmt.Println("frfrv")
			}
		}(port, upstreamName)
	}
}

func handleConnection(clientConn net.Conn, upstreamName string, upstreamMap map[string]*config.LeastConnectionsIterator) {
	log.Printf("Handling connection for upstream: %s", upstreamName)
	defer clientConn.Close()

	iterator, exists := upstreamMap[strings.ToLower(upstreamName)]
	if !exists {
		fmt.Fprintf(clientConn, "Upstream not found: %s\n", upstreamName)
		return
	}
	log.Printf("Found iterator for %s", upstreamName)

	server := iterator.NextHealthy()
	if server == nil {
		fmt.Fprintf(clientConn, "No available servers for upstream: %s\n", upstreamName)
		return
	}
	log.Printf("Connecting to server: %s", server.Url)

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
	log.Printf("Server connection established: %s", server.Url)
	defer serverConn.Close()

	fmt.Fprintf(serverConn, "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")
	// Proxy data between the client and the server
	go helpers.ProxyData(clientConn, serverConn)
	helpers.ProxyData(serverConn, clientConn)
}
