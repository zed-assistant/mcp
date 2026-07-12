package oauth

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/zed-assistant/mcp/internal/configuration"
)

type PendingAuth struct {
	ID    string
	Query url.Values
	Nonce string
	CSRF  string

	Sub   string
	Email string

	CreatedAt time.Time
}

type PendingStore struct {
	mu           sync.Mutex
	pendingAuths map[string]*PendingAuth
	ttl          time.Duration
	random       Random
}

func NewPendingStore(appConfig *configuration.AppConfig, random Random) *PendingStore {
	return &PendingStore{
		pendingAuths: make(map[string]*PendingAuth),
		ttl:          appConfig.OAuth2.PendingAuthTTL,
		random:       random,
	}
}

func (s *PendingStore) StorePendingAuth(query url.Values) (*PendingAuth, error) {
	id, err := s.random.RandomBytesHex(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random ID for pending auth: %w", err)
	}
	nonce, err := s.random.RandomBytesHex(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random nonce for pending auth: %w", err)
	}
	csrf, err := s.random.RandomBytesHex(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random CSRF token for pending auth: %w", err)
	}

	pa := &PendingAuth{
		ID:        id,
		Query:     query,
		Nonce:     nonce,
		CSRF:      csrf,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.gcLocked()
	s.pendingAuths[id] = pa

	return pa, nil
}

func (p *PendingStore) Get(id string) (*PendingAuth, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	pa, ok := p.pendingAuths[id]
	if !ok || time.Since(pa.CreatedAt) > p.ttl {
		delete(p.pendingAuths, id)
		return nil, false
	}
	return pa, true
}

func (p *PendingStore) Delete(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.pendingAuths, id)
}

func (p *PendingStore) gcLocked() {
	for id, pa := range p.pendingAuths {
		if time.Since(pa.CreatedAt) > p.ttl {
			delete(p.pendingAuths, id)
		}
	}
}
