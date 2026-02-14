package mikrotik_outbound_adapter

import (
	"context"
	"fmt"
	"sync"
)

// StreamManager manages active streams with context-based cancellation
type StreamManager struct {
	mu      sync.RWMutex
	streams map[string]context.CancelFunc
}

// NewStreamManager creates a new stream manager instance
func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[string]context.CancelFunc),
	}
}

// Start registers a new stream with its cancel function
// Returns false if stream with the same key already exists
func (sm *StreamManager) Start(key string, cancel context.CancelFunc) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.streams[key]; exists {
		return false
	}

	sm.streams[key] = cancel
	return true
}

// Stop cancels and removes a stream by key
// Returns true if stream was found and stopped, false if not found
func (sm *StreamManager) Stop(key string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cancel, exists := sm.streams[key]
	if !exists {
		return false
	}

	cancel()
	delete(sm.streams, key)
	return true
}

// IsActive checks if a stream with given key is currently active
func (sm *StreamManager) IsActive(key string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	_, exists := sm.streams[key]
	return exists
}

// StopAll cancels all active streams
func (sm *StreamManager) StopAll() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for key, cancel := range sm.streams {
		cancel()
		delete(sm.streams, key)
	}
}

// Count returns the number of active streams
func (sm *StreamManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.streams)
}

// GetStreamKey generates a unique key for a stream
func GetStreamKey(prefix, routerID, identifier string) string {
	if identifier == "" {
		return fmt.Sprintf("%s:%s:all", prefix, routerID)
	}
	return fmt.Sprintf("%s:%s:%s", prefix, routerID, identifier)
}
