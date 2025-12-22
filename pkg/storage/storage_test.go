package storage

import (
	"testing"
	"bytes"
	"time"
)

func TestCreateEntry(t *testing.T) {
	s := InitStorage()

	data := []byte("thisisatestdata")
	duration := 10 * time.Second

	id, err := s.CreateEntry(data, duration)
	if err != nil {
		t.Errorf("CreateEntry(): %v", err)
	}

	secret, exists := s.data[id]
	if !exists {
		t.Errorf("Storage: Entry ID %s was not found in map.", id)
	}
	if !bytes.Equal(secret.Data, data) {
		t.Errorf("CreateEntry(): Expected %s, got %s", data, secret.Data)
	}
}

func TestGetAndDelete(t *testing.T) {
	s := InitStorage()

	data := []byte("thisisatestdata")
	duration := 10 * time.Second

	id, err := s.CreateEntry(data, duration)
	if err != nil {
		t.Errorf("CreateEntry(): %v", err)
	}

	d, err := s.GetAndDelete(id)
	if err != nil {
		t.Errorf("GetAndDelete(): %v", err)
	}
	if !bytes.Equal(d, data) {
		t.Errorf("GetAndDelete(): Expected %s, got %s", data, d)
	}

	_, exists := s.data[id]
	if exists {
		t.Errorf("GetAndDelete(): data is not deleted after reading")
	}

	duration = 1 * time.Second
	id, err = s.CreateEntry(data, duration)
	if err != nil {
		t.Errorf("CreateEntry(): %v", err)
	}

	time.Sleep(2 * time.Second)
	d, err = s.GetAndDelete(id)
	if err == nil {
		t.Errorf("GetAndDelete(): data is not deleted after duration")
	}
}

func TestStartCleanup(t *testing.T) {
	s := InitStorage()
	s.StartCleanup(2 * time.Second)

	id, err := s.CreateEntry([]byte("thisisatestdata"), 1 * time.Second)
	if err != nil {
		t.Errorf("CreateEntry(): %v", err)
	}

	time.Sleep(3 * time.Second)

	_, exists := s.data[id]
	if exists {
		t.Errorf("StartCleanup(): data is not deleted after expiration time")
	}
}
