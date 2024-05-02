package helpers

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"time"

	"github.com/janithht/GoStreamBalancer/internal/config"
)

type SimpleHealthCheckListener struct{}

func ParseHostPort(rawUrl string) (host, port string, err error) {
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

func ProxyData(src, dst net.Conn) {
	_, err := io.Copy(dst, src)
	if err != nil {
		log.Printf("Error during data transfer: %v", err)
	}
}

func (l *SimpleHealthCheckListener) HealthChecked(server *config.UpstreamServer, time time.Time) {
	fmt.Println()
	log.Printf("Health check performed for server %s at %s", server.Url, time.Format("2006-01-02T15:04:05Z07:00"))
}
