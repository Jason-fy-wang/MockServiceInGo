package api

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"
)

type ResponseType string

const (
	ResponseTypeHTTP      ResponseType = "http"
	ResponseTypeSSE       ResponseType = "sse"
	ResponseTypeWebSocket ResponseType = "websocket"
)

type SSEEvent struct {
	Event string        `json:"event,omitempty"`
	Data  string        `json:"data"`
	Delay time.Duration `json:"delay,omitempty"` // delay in milliseconds
}

type WebSocketMessage struct {
	Message string        `json:"message"`
	Delay   time.Duration `json:"delay,omitempty"` // delay in milliseconds
	Type    string        `json:"type,omitempty"`  // "text" or "binary", defaults to "text"
}

type MockRule struct {
	Method            string             `json:"method"`
	Path              string             `json:"path"`
	RequestHeaders    map[string]string  `json:"headers,omitempty"`
	RequestBody       string             `json:"body,omitempty"`
	RequestQuery      map[string]string  `json:"query,omitempty"`
	ResponseStatus    int                `json:"status,omitempty"`
	ResponseHeaders   map[string]string  `json:"responseHeaders,omitempty"`
	ResponseBody      string             `json:"responseBody,omitempty"`
	ResponseType      ResponseType       `json:"responseType,omitempty"` // "http", "sse", "websocket"
	SSEEvents         []SSEEvent         `json:"sseEvents,omitempty"`
	WebSocketMessages []WebSocketMessage `json:"websocketMessages,omitempty"`
}

type mockRegistry struct {
	lock sync.RWMutex
	// persist mock rules with map, speed up search
	rules map[string]MockRule

	// file persistence
	filepath string
	stopChan chan struct{}
}

func newMockRegistry() *mockRegistry {
	return &mockRegistry{rules: make(map[string]MockRule)}
}

func (r *mockRegistry) add(rule MockRule) {
	r.lock.Lock()
	defer r.lock.Unlock()
	m := strings.ToUpper(rule.Method)
	r.rules[m+rule.Path] = rule
}

func (r *mockRegistry) list() []MockRule {
	r.lock.RLock()
	defer r.lock.RUnlock()
	copyRules := make([]MockRule, len(r.rules))
	i := 0
	for _, rule := range r.rules {
		copyRules[i] = rule
		i++
	}
	return copyRules
}

func (r *mockRegistry) delete(method, path string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	method = strings.ToUpper(method)
	delete(r.rules, method+path)
}

func (r *mockRegistry) clear() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.rules = make(map[string]MockRule)
}

func (r *mockRegistry) replaceAll(rules []MockRule) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.rules = make(map[string]MockRule, len(rules))
	for _, rule := range rules {
		m := strings.ToUpper(rule.Method)
		r.rules[m+rule.Path] = rule
	}
}

func (r *mockRegistry) find(method, path string) (*MockRule, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	rule, exists := r.rules[method+path]

	return &rule, exists
}

// SaveToFile saves all mock rules to a JSON file
func (r *mockRegistry) SaveToFile(filepath string) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	// Convert map to slice for JSON marshaling
	rules := make([]MockRule, 0, len(r.rules))
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// LoadFromFile loads mock rules from a JSON file
func (r *mockRegistry) LoadFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, no error
		}
		return err
	}

	var rules []MockRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return err
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	for _, rule := range rules {
		m := strings.ToUpper(rule.Method)
		r.rules[m+rule.Path] = rule
	}

	return nil
}

// StartAutoSave starts a background goroutine that saves registry every 2 seconds
func (r *mockRegistry) StartAutoSave(filepath string) {
	r.filepath = filepath
	r.stopChan = make(chan struct{})

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_ = r.SaveToFile(r.filepath) // ignore errors silently
			case <-r.stopChan:
				return
			}
		}
	}()
}

// StopAutoSave stops the auto-save goroutine
func (r *mockRegistry) StopAutoSave() {
	if r.stopChan != nil {
		close(r.stopChan)
	}
}
