package api

import "sync"


type mockRule struct {
	Method          string            `json:"method"`
	Path            string            `json:"path"`
	RequestHeaders  map[string]string `json:"headers,omitempty"`
	RequestBody     string            `json:"body,omitempty"`
	RequestQuery    map[string]string `json:"query,omitempty"`
	ResponseStatus  int               `json:"status,omitempty"`
	ResponseHeaders map[string]string `json:"responseHeaders,omitempty"`
	ResponseBody    string            `json:"responseBody,omitempty"`
}

type mockRegistry struct {
	lock  sync.RWMutex
	// perisit mock rules with map, speed up search
	rules map[string]mockRule
}

func newMockRegistry() *mockRegistry {
	return &mockRegistry{rules: make(map[string]mockRule)}
}

func (r *mockRegistry) add(rule mockRule) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.rules[rule.Method+rule.Path] = rule
}

func (r *mockRegistry) list() []mockRule {
	r.lock.RLock()
	defer r.lock.RUnlock()
	copyRules := make([]mockRule, len(r.rules))
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
	delete(r.rules, method+path)
}

func (r *mockRegistry) clear() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.rules = make(map[string]mockRule)
}

func (r *mockRegistry) find(method, path string) (*mockRule, bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	rule, exists := r.rules[method+path]

	return &rule, exists
}