package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	service := NewMockService()
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
	service := NewMockService()
	router := service.NewRouter()

	rule := mockRule{
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
	service := NewMockService()
	router := service.NewRouter()

	rule := mockRule{
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
		Mocks []mockRule `json:"mocks"`
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
		Mocks []mockRule `json:"mocks"`
	}
	if err := json.NewDecoder(listAgainResp.Body).Decode(&listAgainBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listAgainBody.Mocks) != 0 {
		t.Fatalf("expected 0 registered mocks after clear, got %d", len(listAgainBody.Mocks))
	}
}
