package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, req *http.Request)
}

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type loadBalancer struct {
	port            string
	roundRobincount int
	servers         []Server
}

func newLoadBalancer(port string, servers []Server) *loadBalancer { //returns a pointer to the loadbalancer struct
	return &loadBalancer{
		port:            port,
		roundRobincount: 0,
		servers:         servers,
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func (s *simpleServer) IsAlive() bool { return true }

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

func (lb *loadBalancer) getNextServer() Server {
	server := lb.servers[lb.roundRobincount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundRobincount++
		server = lb.servers[lb.roundRobincount%len(lb.servers)]
	}

	lb.roundRobincount++
	return server
}

func (lb *loadBalancer) serveProxy(rw http.ResponseWriter, req *http.Request) {
	targetServer := lb.getNextServer()
	fmt.Printf("forwarding request to %s\n", targetServer.Address())
	targetServer.Serve(rw, req)
}

func main() {
	servers := []Server{
		newSimpleServer("https://www.facebook.com/"),
		newSimpleServer("http://www.youtube.com/"),
		newSimpleServer("https://www.amazon.com/"),
	}

	lb := newLoadBalancer("8080", servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.serveProxy(rw, req)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("Load Balancer listening on 'localhost:%s '\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}

func (s *simpleServer) Address() string {
	return s.addr
}
