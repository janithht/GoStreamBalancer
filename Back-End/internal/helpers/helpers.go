package helpers

import (
	"io"
	"log"
	"net"
	"net/http"
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

func ProxyData(dst, src net.Conn) {
	defer dst.Close()
	defer src.Close()

	copyBuffer := make([]byte, 32*1024)

	src.SetDeadline(time.Now().Add(2 * time.Second))
	dst.SetDeadline(time.Now().Add(2 * time.Second))

	_, err := io.CopyBuffer(dst, src, copyBuffer)
	if err != nil {
		log.Printf("Error proxying data: %v", err)
	}
}

func (l *SimpleHealthCheckListener) HealthChecked(server *config.UpstreamServer, time time.Time) {
	//log.Printf("Health check performed for server %s at %s", server.Url, time.Format("2006-01-02T15:04:05Z07:00"))
}

func CreateHttpClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
		ForceAttemptHTTP2:   true,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
}
