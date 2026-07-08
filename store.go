package main

import (
	"sync"
	"time"
)

type SandboxStatus string

const (
	StatusRunning SandboxStatus = "running"
	StatusStopped SandboxStatus = "stopped"
)

type Sandbox struct {
	ID          string        `json:"id"`
	ContainerID string        `json:"-"`
	Status      SandboxStatus `json:"status"`
	CreatedAt   time.Time 		`json:"created_at"`
}

type Store struct {
	mu sync.RWMutex
	sandboxes map[string]*Sandbox
}

func NewStore() *Store {
	return &Store{sandboxes: make(map[string]*Sandbox)}
}

func (s *Store) Save(sb *Sandbox){
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sandboxes[sb.ID] = sb
}

func (s *Store) Get(id string) (*Sandbox, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sb, ok := s.sandboxes[id]
	return sb, ok
}

func (s *Store) Delete(id string)	{
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sandboxes, id)
}

func (s *Store) List() []*Sandbox {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Sandbox, 0, len(s.sandboxes))
	for _, sb := range s.sandboxes {
		out = append(out, sb)
	}

	return out
}