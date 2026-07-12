package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ory/fosite"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type cimdCacheEntry struct {
	client    fosite.Client
	fetchedAt time.Time
}

type CIMDResolver struct {
	mu             sync.Mutex
	cache          map[string]cimdCacheEntry
	httpClient     HttpClient
	getCurrentTime func() time.Time
}

func NewCIMDResolver(httpClient HttpClient, getCurrentTime func() time.Time) *CIMDResolver {
	return &CIMDResolver{
		cache:          make(map[string]cimdCacheEntry),
		httpClient:     httpClient,
		getCurrentTime: getCurrentTime,
	}
}

const (
	cimdCacheTTL = 1 * time.Hour
	cimdMaxBody  = 64 << 10 // 64 KiB
)

type cimdDocument struct {
	ClientID                string   `json:"client_id"`
	ClientName              string   `json:"client_name"`
	ClientURI               string   `json:"client_uri"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	Scope                   string   `json:"scope"`
}

func (r *CIMDResolver) Resolve(ctx context.Context, id string, allowedScopes []string, audience string) (fosite.Client, error) {
	u, err := url.Parse(id)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.Fragment != "" {
		return nil, errors.New("client_id must be a valid https URL without a fragment.")
	}

	r.mu.Lock()
	if entry, ok := r.cache[id]; ok {
		r.mu.Unlock()
		return entry.client, nil
	}
	r.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, id, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create CIMD request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute CIMD request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CIMD request for %s failed with status: %s, body: %s", id, resp.Status, string(respBody))
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		return nil, fmt.Errorf("CIMD request for %s returned unexpected content type: %s", id, ct)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, cimdMaxBody+1))
	if err != nil || len(body) > cimdMaxBody {
		return nil, fmt.Errorf("CIMD request for %s returned too large body or failed to read: %w", id, err)
	}

	var doc cimdDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CIMD document for %s: %w", id, err)
	}

	if doc.ClientID == "" || doc.ClientID != id {
		return nil, fmt.Errorf("CIMD document for %s has invalid or mismatched client_id: %s", id, doc.ClientID)
	}
	if m := doc.TokenEndpointAuthMethod; m != "" && m != "none" {
		return nil, fmt.Errorf("CIMD document for %s has unsupported token_endpoint_auth_method: %s", id, m)
	}
	if err := validateRedirectURIs(doc.RedirectURIs); err != nil {
		return nil, fosite.ErrInvalidClient.WithHint(err.Error())
	}

	grants := doc.GrantTypes
	if len(grants) == 0 {
		grants = []string{"authorization_code", "refresh_token"}
	}
	for _, g := range grants {
		if g != "authorization_code" && g != "refresh_token" {
			return nil, fmt.Errorf("CIMD document for %s has unsupported grant type: %s", id, g)
		}
	}

	client := &fosite.DefaultClient{
		ID:            id,
		RedirectURIs:  doc.RedirectURIs,
		GrantTypes:    grants,
		ResponseTypes: []string{"code"},
		Scopes:        allowedScopes,
		Audience:      []string{audience},
		Public:        true,
	}

	r.mu.Lock()
	r.cache[id] = cimdCacheEntry{
		client:    client,
		fetchedAt: r.getCurrentTime(),
	}
	r.mu.Unlock()

	return client, nil
}

func validateRedirectURIs(uris []string) error {
	if len(uris) == 0 {
		return fmt.Errorf("redirect_uris is required")
	}
	for _, raw := range uris {
		u, err := url.Parse(raw)
		if err != nil {
			return fmt.Errorf("invalid redirect_uri %q", raw)
		}
		switch {
		case u.Scheme == "https":
		case u.Scheme == "http" && isLoopbackHost(u.Hostname()):
			continue
		default:
			return fmt.Errorf("redirect_uri %q must be https or a loopback http URI", raw)
		}
	}
	return nil
}

func isLoopbackHost(h string) bool {
	return h == "localhost" || h == "127.0.0.1" || h == "::1"
}
