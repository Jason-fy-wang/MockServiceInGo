package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHealthEndpoint(t *testing.T) {
	service := NewMockService("") // No file persistence for this test
	router := service.NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["message"] != "mock service running" {
		t.Fatalf("unexpected message: %q", body["message"])
	}
	if body["routes"] == "" {
		t.Fatal("expected routes metadata in health response")
	}
}

func TestRegisterAndServeMock(t *testing.T) {
	service := NewMockService("")
	router := service.NewRouter()

	rule := MockRule{
		Method: http.MethodPost,
		Path:   "/v1/",
		RequestHeaders: map[string]string{
			"X-Mock-Request": "true",
		},
		RequestBody:    `{"name":"dynamic"}`,
		ResponseStatus: http.StatusCreated,
		ResponseHeaders: map[string]string{
			"X-Mock":       "yes",
			"Content-Type": "application/json",
		},
		ResponseBody: `{"result":"ok"}`,
	}

	payload, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("marshal registration payload: %v", err)
	}

	registerReq := httptest.NewRequest(http.MethodPost, "/v1/__mock", bytes.NewBuffer(payload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)

	if registerResp.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d", registerResp.Code)
	}

	var registered map[string]any
	if err := json.NewDecoder(registerResp.Body).Decode(&registered); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if registered["message"] != "mock registered" {
		t.Fatalf("unexpected register response: %v", registered["message"])
	}

	testReq := httptest.NewRequest(http.MethodPost, "/v1/", bytes.NewBufferString(rule.RequestBody))
	testReq.Header.Set("X-Mock-Request", "true")
	testResp := httptest.NewRecorder()
	router.ServeHTTP(testResp, testReq)

	if testResp.Code != http.StatusCreated {
		t.Fatalf("expected mock response status 201, got %d", testResp.Code)
	}
	if got := testResp.Header().Get("X-Mock"); got != "yes" {
		t.Fatalf("expected X-Mock header yes, got %q", got)
	}
	if got := testResp.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}
	if body := testResp.Body.String(); body != rule.ResponseBody {
		t.Fatalf("expected response body %q, got %q", rule.ResponseBody, body)
	}
}

func TestListAndClearMocks(t *testing.T) {
	service := NewMockService("")
	router := service.NewRouter()

	rule := MockRule{
		Method: http.MethodGet,
		Path:   "/v1/",
		ResponseStatus: http.StatusOK,
		ResponseHeaders: map[string]string{
			"X-Mock": "yes",
		},
		ResponseBody: "ok",
	}

	payload, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("marshal registration payload: %v", err)
	}

	registerReq := httptest.NewRequest(http.MethodPost, "/v1/__mock", bytes.NewBuffer(payload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)

	if registerResp.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d", registerResp.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/__mock", nil)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)

	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listResp.Code)
	}

	var listBody struct {
		Mocks []MockRule `json:"mocks"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listBody.Mocks) != 1 {
		t.Fatalf("expected 1 registered mock, got %d", len(listBody.Mocks))
	}

	clearReq := httptest.NewRequest(http.MethodDelete, "/v1/__mock/all", nil)
	clearResp := httptest.NewRecorder()
	router.ServeHTTP(clearResp, clearReq)

	if clearResp.Code != http.StatusOK {
		t.Fatalf("expected clear status 200, got %d", clearResp.Code)
	}

	listAgainReq := httptest.NewRequest(http.MethodGet, "/v1/__mock", nil)
	listAgainResp := httptest.NewRecorder()
	router.ServeHTTP(listAgainResp, listAgainReq)

	if listAgainResp.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listAgainResp.Code)
	}

	var listAgainBody struct {
		Mocks []MockRule `json:"mocks"`
	}
	if err := json.NewDecoder(listAgainResp.Body).Decode(&listAgainBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listAgainBody.Mocks) != 0 {
		t.Fatalf("expected 0 registered mocks after clear, got %d", len(listAgainBody.Mocks))
	}
}

func TestConfigUpload(t *testing.T) {
	service := NewMockService("")
	router := service.NewRouter()

	rules := []MockRule{
		        {Method: "GET", Path: "/v1/uploaded", ResponseStatus: http.StatusOK, ResponseBody: "uploaded"},
        {Method: "POST", Path: "/v1/data", ResponseStatus: http.StatusCreated, ResponseBody: `{"status":"created"}`},
	}
	configData, err := json.Marshal(rules)
	if err != nil {
		t.Fatalf("marshal config data: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("config.json", "config.json")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	if _, err := part.Write(configData); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/__mock/upload", bytes.NewBuffer(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp := httptest.NewRecorder()

	router.ServeHTTP(resp,req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, get %d", resp.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result["message"] != "config uploaded and loaded" {
		t.Fatalf("unexpected response message: %q", result["message"])
	}
}

func TestConfigUploadMissingFile(t *testing.T) {
	service := NewMockService("")
	router := service.NewRouter()
	
	req := httptest.NewRequest(http.MethodPost, "/v1/__mock/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)
	
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for missing file, got %d", resp.Code)
	}
	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result["error"] != "failed to get uploaded file" {
		t.Fatalf("unexpected error message: %q", result["error"])
	}
}

func TestSSERegistrationAndMock(t *testing.T) {
	service := NewMockService("")
	router := service.NewRouter()

	rule := MockRule{
		Method: http.MethodGet,
		Path:   "/v1/sse-test",
		ResponseType: ResponseTypeSSE,
		SSEEvents: []SSEEvent{
			{Event: "test", Data: "hello", Delay: 10 * time.Millisecond},
			{Event: "test", Data: "world", Delay: 10 * time.Millisecond},
		},
	}

	payload, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("marshal registration payload: %v", err)
	}

	registerReq := httptest.NewRequest(http.MethodPost, "/v1/__mock", bytes.NewBuffer(payload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)

	if registerResp.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d", registerResp.Code)
	}

	// Test SSE endpoint
	testReq := httptest.NewRequest(http.MethodGet, "/v1/sse-test", nil)
	testResp := httptest.NewRecorder()
	router.ServeHTTP(testResp, testReq)

	if testResp.Code != http.StatusOK {
		t.Fatalf("expected SSE response status 200, got %d", testResp.Code)
	}
	if contentType := testResp.Header().Get("Content-Type"); contentType != "text/event-stream" {
		t.Fatalf("expected Content-Type text/event-stream, got %q", contentType)
	}
}

func TestWebSocketRegistration(t *testing.T) {
	service := NewMockService("")
	router := service.NewRouter()

	rule := MockRule{
		Method: http.MethodGet,
		Path:   "/v1/ws-test",
		ResponseType: ResponseTypeWebSocket,
		WebSocketMessages: []WebSocketMessage{
			{Message: "hello", Delay: 10 * time.Millisecond},
			{Message: "world", Delay: 10 * time.Millisecond},
		},
	}

	payload, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("marshal registration payload: %v", err)
	}

	registerReq := httptest.NewRequest(http.MethodPost, "/v1/__mock", bytes.NewBuffer(payload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)

	if registerResp.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d", registerResp.Code)
	}
}

func TestInvalidResponseType(t *testing.T) {
	service := NewMockService("")
	router := service.NewRouter()

	rule := MockRule{
		Method:       http.MethodGet,
		Path:         "/v1/invalid-test",
		ResponseType: "invalid",
	}

	payload, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("marshal registration payload: %v", err)
	}

	registerReq := httptest.NewRequest(http.MethodPost, "/v1/__mock", bytes.NewBuffer(payload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)

	if registerResp.Code != http.StatusBadRequest {
		t.Fatalf("expected register status 400 for invalid response type, got %d", registerResp.Code)
	}
}

func TestSSERequiresEvents(t *testing.T) {
	service := NewMockService("")
	router := service.NewRouter()

	rule := MockRule{
		Method:       http.MethodGet,
		Path:         "/v1/sse-no-events",
		ResponseType: ResponseTypeSSE,
		// No SSEEvents provided
	}

	payload, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("marshal registration payload: %v", err)
	}

	registerReq := httptest.NewRequest(http.MethodPost, "/v1/__mock", bytes.NewBuffer(payload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)

	if registerResp.Code != http.StatusBadRequest {
		t.Fatalf("expected register status 400 for SSE without events, got %d", registerResp.Code)
	}
}

func TestWebSocketRequiresMessages(t *testing.T) {
	service := NewMockService("")
	router := service.NewRouter()

	rule := MockRule{
		Method:       http.MethodGet,
		Path:         "/v1/ws-no-messages",
		ResponseType: ResponseTypeWebSocket,
		// No WebSocketMessages provided
	}

	payload, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("marshal registration payload: %v", err)
	}

	registerReq := httptest.NewRequest(http.MethodPost, "/v1/__mock", bytes.NewBuffer(payload))
	registerReq.Header.Set("Content-Type", "application/json")
	registerResp := httptest.NewRecorder()
	router.ServeHTTP(registerResp, registerReq)

	if registerResp.Code != http.StatusBadRequest {
		t.Fatalf("expected register status 400 for WebSocket without messages, got %d", registerResp.Code)
	}
}

func TestSaveToFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "mocks.json")
	registry := newMockRegistry()

	rule1 := MockRule{Method: "GET", Path: "/test1", ResponseBody: "test1"}
	rule2 := MockRule{Method: "POST", Path: "/test2", ResponseBody: "test2"}

	registry.add(rule1)
	registry.add(rule2)

	if err := registry.SaveToFile(tmpFile); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Verify file exists and contains the rules
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	var rules []MockRule
	if err := json.Unmarshal(data, &rules); err != nil {
		t.Fatalf("failed to unmarshal saved data: %v", err)
	}

	if len(rules) != 2 {
		t.Fatalf("expected 2 rules saved, got %d", len(rules))
	}
}

func TestLoadFromFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "mocks.json")

	// Create a test file with rules
	rules := []MockRule{
		{Method: "GET", Path: "/test1", ResponseBody: "test1"},
		{Method: "POST", Path: "/test2", ResponseBody: "test2"},
	}
	data, _ := json.Marshal(rules)
	os.WriteFile(tmpFile, data, 0644)

	registry := newMockRegistry()
	if err := registry.LoadFromFile(tmpFile); err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if len(registry.list()) != 2 {
		t.Fatalf("expected 2 rules loaded, got %d", len(registry.list()))
	}

	rule, exists := registry.find("GET", "/test1")
	if !exists {
		t.Fatal("expected to find GET /test1")
	}
	if rule.ResponseBody != "test1" {
		t.Fatalf("expected response body 'test1', got %q", rule.ResponseBody)
	}
}

func TestLoadFromFileMissing(t *testing.T) {
	registry := newMockRegistry()
	// Should not error if file doesn't exist
	if err := registry.LoadFromFile("/nonexistent/mocks.json"); err != nil {
		t.Fatalf("LoadFromFile should not error for missing file, got: %v", err)
	}
}

func TestAutoSave(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "auto_mocks.json")
	registry := newMockRegistry()

	// Start auto-save with 100ms interval for testing
	registry.StartAutoSave(tmpFile)
	defer registry.StopAutoSave()

	// Add a rule
	rule := MockRule{Method: "GET", Path: "/auto-test", ResponseBody: "auto"}
	registry.add(rule)

	// Wait for auto-save to trigger (2 seconds default, but we'll check early)
	time.Sleep(2500 * time.Millisecond)

	// Verify file was created and contains the rule
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("auto-save file not created: %v", err)
	}

	var savedRules []MockRule
	if err := json.Unmarshal(data, &savedRules); err != nil {
		t.Fatalf("failed to unmarshal auto-saved data: %v", err)
	}

	if len(savedRules) != 1 {
		t.Fatalf("expected 1 rule auto-saved, got %d", len(savedRules))
	}
	if savedRules[0].Path != "/auto-test" {
		t.Fatalf("expected path /auto-test, got %s", savedRules[0].Path)
	}
}
