package config

import (
	"reflect"
	"testing"
)

var iterator IteratorImpl

func TestNewRoundRobinIterator(t *testing.T) {
	iterator := NewIterator()
	if reflect.ValueOf(iterator.servers).Kind() != reflect.Slice {
		t.Errorf("Expected item list to be initialized, got nil")
	}
}

func TestAdd(t *testing.T) {
	server := &UpstreamServer{
		Url:    "http://localhost:9000",
		Status: true,
	}
	iterator.Add(server)
	if iterator.servers == nil {
		t.Errorf("Expected 1 item in the iterator, got %d", len(iterator.servers))
	}
}

func TestNext(t *testing.T) {
	server := iterator.Next()
	if server == nil {
		t.Errorf("List is empty, expected a server to be returned")
	}
}