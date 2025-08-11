package services

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type MockStore struct {
	mu    sync.RWMutex
	mocks map[string]string
}

func NewMockStore() *MockStore {
	return &MockStore{
		mocks: make(map[string]string),
	}
}

func (ms *MockStore) CreateMock(jsonBody string) (string, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Generate a short, random, unique ID
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	id := hex.EncodeToString(b)

	ms.mocks[id] = jsonBody
	return id, nil
}

func (ms *MockStore) GetMock(id string) (string, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	mock, found := ms.mocks[id]
	return mock, found
}