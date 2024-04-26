package serverTCP

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"

	"github.com/janithht/GoStreamBalancer/internal/config"
)

func StartTCPServer(upstreams []config.Upstream) {
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
		log.Fatalf("Failed to start TCP server: %v", err)
	}
	defer listener.Close()
	log.Println("TCP Load Balancer started on port 3000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleTCPConnection(conn, upstreamMap)
	}
}

func handleTCPConnection(conn net.Conn, upstreamMap map[string]*config.RoundRobinIterator) {
	defer conn.Close()

	upstreamName := readUpstreamName(conn)
	if upstreamName == "" {
		log.Println("Invalid or empty upstream name provided")
		return
	}

	// Lookup the iterator for the given upstream
	iterator, exists := upstreamMap[strings.ToLower(strings.TrimSpace(upstreamName))]
	if !exists || iterator == nil {
		log.Printf("Upstream not found or has no servers: %s", upstreamName)
		conn.Write([]byte("Upstream not found or has no servers\n"))
		return
	}

	// Get the next available healthy server
	server := iterator.NextHealthy()
	if server == nil || !server.GetStatus() {
		log.Printf("No healthy servers available for upstream: %s", upstreamName)
		conn.Write([]byte("No healthy servers available\n"))
		return
	}

	// Proxy the connection to the chosen server
	proxyConnection(conn, server)
}

func proxyConnection(clientConn net.Conn, server *config.UpstreamServer) {

	host, port, err := parseHostPort(server.Url)
	if err != nil {
		log.Printf("Invalid server URL: %v", err)
		return
	}

	address := fmt.Sprintf("%s:%s", host, port)
	serverConn, err := net.Dial("tcp", address)
	if err != nil {
		log.Printf("Failed to connect to server: %v", err)
		return
	}
	log.Println("Successfully connected to backend server:", address)
	fmt.Println()
	defer serverConn.Close()

	go io.Copy(serverConn, clientConn)
	io.Copy(clientConn, serverConn)
}

// Helper function to parse host and port from a URL
func parseHostPort(rawUrl string) (host, port string, err error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return "", "", err
	}
	host, port, err = net.SplitHostPort(u.Host)
	if err != nil {
		return "", "", err
	}
	return host, port, nil
}

func readUpstreamName(conn net.Conn) string {
	reader := bufio.NewReader(conn)
	upstreamName, err := reader.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			log.Printf("Error reading upstream name from connection: %v", err)
		}
		return ""
	}

	upstreamName = strings.TrimSpace(upstreamName)
	log.Printf("Received upstream name: %s", upstreamName)
	return upstreamName
}
