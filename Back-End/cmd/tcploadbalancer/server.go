package tcploadbalancer

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/janithht/GoStreamBalancer/database"
	"github.com/janithht/GoStreamBalancer/internal/config"
	"github.com/janithht/GoStreamBalancer/internal/helpers"
	"github.com/janithht/GoStreamBalancer/metrics"
	"github.com/janithht/GoStreamBalancer/migrations"
)

func StartLoadBalancers(upstreamMap map[string]*config.IteratorImpl, portMap map[int]string) {
	database.InitDB()
	migrations.Migrate(database.DB)

	for port, upstreamName := range portMap {
		go func(port int, upstreamName string) {
			//log.Printf("Starting load balancer for %s on port %d", upstreamName, port)
			listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err != nil {
				log.Printf("Failed to listen on port %d: %v", port, err)
				return
			}
			defer listener.Close()

			for {
				clientConn, err := listener.Accept()
				if err != nil {
					log.Printf("Failed to accept connection on port %d: %v", port, err)
					continue
				}
				//log.Printf("Accepted connection from %s on port %d", clientConn.RemoteAddr().String(), port)
				go handleConnection(clientConn, upstreamName, upstreamMap)
			}
		}(port, upstreamName)
	}
}

func handleConnection(clientConn net.Conn, upstreamName string, upstreamMap map[string]*config.IteratorImpl) {
	defer clientConn.Close()
	iterator, exists := upstreamMap[strings.ToLower(upstreamName)]
	if !exists {
		log.Printf("Upstream not found: %s", upstreamName)
		fmt.Fprintf(clientConn, "Upstream not found: %s\n", upstreamName)
		return
	}

	server := iterator.NextLeastConServer()
	if server == nil {
		log.Printf("No available servers for upstream: %s", upstreamName)
		fmt.Fprintf(clientConn, "No available servers for upstream: %s\n", upstreamName)
		return
	}

	server.IncrementConnections()
	metrics.SetTCPConnections(upstreamName, float64(server.ActiveConnections))
	metrics.RecordTCPRequest(upstreamName)

	defer func() {
		server.DecrementConnections()
		metrics.SetTCPConnections(upstreamName, float64(server.ActiveConnections))
	}()

	host, port, err := helpers.ParseHostPort(server.Url)
	if err != nil {
		log.Printf("Invalid server URL: %v", err)
		fmt.Fprintf(clientConn, "Invalid server URL: %v\n", err)
		return
	}

	serverConn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		log.Printf("Failed to connect to server: %v", err)
		fmt.Fprintf(clientConn, "Failed to connect to server: %v\n", err)
		return
	}
	defer serverConn.Close()

	fmt.Fprintf(serverConn, "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")
	clientIP := clientConn.RemoteAddr().String()
	database.AddConnection(database.ConnectionData{ClientIP: clientIP, ServerURL: server.Url})

	go helpers.ProxyData(clientConn, serverConn, upstreamName)
	helpers.ProxyData(serverConn, clientConn, upstreamName)
}
