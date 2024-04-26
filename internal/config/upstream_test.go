package config

import (
	"testing"
)

func TestSetStatus(t *testing.T) {
	server := &UpstreamServer{
		Url:    "http://localhost:9000",
		Status: false,
	}
	server.SetStatus(true)
	if server.Status != true {
		t.Errorf("Expected server status to be true, got %v", server.Status)
	}
}

func TestGetStatus(t *testing.T) {
	server := &UpstreamServer{
		Url:    "http://localhost:9000",
		Status: true,
	}
	status := server.GetStatus()
	if status != true {
		t.Errorf("Expected server status to be true, got %v", status)
	}
}
