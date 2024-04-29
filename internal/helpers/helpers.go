package helpers

import (
	"io"
	"log"
	"net"
	"net/url"
)

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
