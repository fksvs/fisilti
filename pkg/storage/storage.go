package storage

import (
	"time"
	"sync"
	"errors"

	"github.com/fksvs/fisilti/pkg/cipher"
)

type Secret struct {
	ExpiresAt time.Time
	Data []byte
}

type Storage struct {
	mu sync.Mutex
	data map[string]Secret
}

func InitStorage() *Storage {
	return &Storage {
		data: make(map[string]Secret),
	}
}

func (s *Storage) CreateEntry(data []byte, duration time.Duration) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash, err := cipher.GenerateHashID()
	if err != nil {
		return "", err
	}

	id := string(hash)
	s.data[id] = Secret{
		ExpiresAt: time.Now().Add(duration),
		Data: data,
	}

	return id, nil
}

func (s *Storage) GetAndDelete(id string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.data[id]
	if !exists {
		return nil, errors.New("not found")
	}

	delete(s.data, id)

	if time.Now().After(val.ExpiresAt) {
		return nil, errors.New("expired")
	}

	return val.Data, nil
}

func (s *Storage) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for {
			<-ticker.C

			s.mu.Lock()
			defer s.mu.Unlock()

			for id, secret := range s.data {
				if time.Now().After(secret.ExpiresAt) {
					delete(s.data, id)
				}
			}
		}
	}()
}
