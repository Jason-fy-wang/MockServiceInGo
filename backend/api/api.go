package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"mockservice/backend/log"
	"mockservice/backend/raft"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type MockSerice struct {
	registry   *mockRegistry
	logger     *zap.Logger
	upgrader   websocket.Upgrader
	raftConfig *raft.ConfigStateMachine
}

const mockRulesRaftKey = "__mock_rules_cmd__"

type ServiceOptions struct {
	FilePath   string
	RaftConfig *raft.ConfigStateMachine
}

type ruleCommand struct {
	Op     string     `json:"op"`
	Rule   *MockRule  `json:"rule,omitempty"`
	Method string     `json:"method,omitempty"`
	Path   string     `json:"path,omitempty"`
	Rules  []MockRule `json:"rules,omitempty"`
}

func NewMockService(filepath string) *MockSerice {
	return NewMockServiceWithOptions(ServiceOptions{FilePath: filepath})
}

func NewMockServiceWithOptions(opts ServiceOptions) *MockSerice {
	mockservice := &MockSerice{}
	mockservice.registry = newMockRegistry()
	mockservice.logger = log.Get()
	mockservice.raftConfig = opts.RaftConfig
	mockservice.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin for mocking purposes
		},
	}

	if mockservice.raftConfig != nil {
		mockservice.raftConfig.Subscribe(mockservice.applyRaftCommand)
	}

	if _, err := os.Stat(opts.FilePath); err == nil {
		mockservice.registry.LoadFromFile(opts.FilePath)
		mockservice.logger.Info("loaded mock rules from file", zap.String("filepath", opts.FilePath))
		go mockservice.SynchronizeLoadedRules()
	}

	mockservice.logger.Info("start schedule saving rules")
	mockservice.registry.StartAutoSave(opts.FilePath)
	return mockservice
}

func (s *MockSerice) SynchronizeLoadedRules() {
	// synchronize all rules to others
	if s.raftConfig != nil && s.registry != nil {
		deadline := time.Now().Add(15 * time.Second)
		time.Sleep(5 * time.Second)
		for time.Now().Before(deadline) {
			if s.raftConfig.Node.IsCandidate() {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			break
		}
		if s.raftConfig.Node.IsCandidate() {
			s.logger.Warn("skip loaded-rules synchronization: election not finished before timeout")
			return
		}
		if !s.raftConfig.Node.IsLeader() {
			s.logger.Info("skip loaded-rules synchronization on follower node")
			return
		}
		s.logger.Info("Synchronize the loaded rules", zap.String("node",s.raftConfig.Node.Id()))
		if err := s.replicateRuleCommand(ruleCommand{
			Op:    "replace",
			Rules: s.registry.list(),
		}); err != nil {
			s.logger.Error("failed to replicate loaded rules", zap.Error(err))
		}
	}
}

func (s *MockSerice) NewRouter() *gin.Engine {

	router := gin.New()
	{
		v1 := router.Group("/v1")
		router.Use(gin.Recovery(), gin.Logger())

		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message":  "mock service running",
				"routes":   "POST /__mock, GET /__mock, DELETE /__mock, ANY /mock/*path",
				"features": "HTTP, SSE, WebSocket mocking",
			})
		})

		v1.POST("/__mock", s.registerMock)
		v1.POST("/__mock/upload", s.UploadConfig)
		v1.GET("/__mock", s.listMocks)
		v1.DELETE("/__mock/all", s.clearMocks)
		v1.DELETE("/__mock/:method", s.deleteMock)
	}

	router.NoRoute(s.mockHandler)
	return router
}

func (s *MockSerice) UploadConfig(c *gin.Context) {
	file, err := c.FormFile("config.json")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get uploaded file"})
		return
	}

	if err := c.SaveUploadedFile(file, file.Filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save uploaded file"})
		return
	}

	s.registry.LoadFromFile(file.Filename)
	s.logger.Info("loaded mock rules from uploaded file", zap.String("filepath", file.Filename))
	if err := s.replicateRuleCommand(ruleCommand{
		Op:    "replace",
		Rules: s.registry.list(),
	}); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "config uploaded and loaded"})
}

func (s *MockSerice) Run(addr string) error {
	s.logger.Info("starting mock service", zap.String("address", addr))
	return s.NewRouter().Run(addr)
}

func (s *MockSerice) registerMock(c *gin.Context) {
	var rule MockRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if rule.Method == "" || rule.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "method and path are required"})
		return
	}

	if rule.ResponseStatus == 0 {
		rule.ResponseStatus = http.StatusOK
	}

	// Set default response type if not specified
	if rule.ResponseType == "" {
		rule.ResponseType = ResponseTypeHTTP
	}

	// Validate response type specific requirements
	switch rule.ResponseType {
	case ResponseTypeSSE:
		if len(rule.SSEEvents) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sseEvents are required for SSE response type"})
			return
		}
	case ResponseTypeWebSocket:
		if len(rule.WebSocketMessages) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "websocketMessages are required for WebSocket response type"})
			return
		}
		rule.Method = "GET" // WebSocket upgrades must be GET requests
	case ResponseTypeHTTP:
		// No additional validation needed
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid responseType. Must be 'http', 'sse', or 'websocket'"})
		return
	}

	if err := s.replicateRuleCommand(ruleCommand{
		Op:   "add",
		Rule: &rule,
	}); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "mock registered", "mock": rule})
}

func (s *MockSerice) listMocks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"mocks": s.registry.list()})
}

func (s *MockSerice) clearMocks(c *gin.Context) {
	if err := s.replicateRuleCommand(ruleCommand{Op: "clear"}); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all mocks cleared"})
}

func (s *MockSerice) deleteMock(c *gin.Context) {
	path := c.Query("path")
	method := c.Param("method")
	s.logger.Info("deleting mock", zap.String("method", method), zap.String("path", path))
	if err := s.replicateRuleCommand(ruleCommand{
		Op:     "delete",
		Method: method,
		Path:   path,
	}); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "mock deleted", "method": method, "path": path})
}

func (s *MockSerice) replicateRuleCommand(cmd ruleCommand) error {
	if s.raftConfig == nil {
		s.applyRuleCommand(cmd)
		return nil
	}

	payload, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	return s.raftConfig.Synchronize(mockRulesRaftKey, string(payload))
}

func (s *MockSerice) applyRaftCommand(cmd raft.ConfigCommand) {
	if cmd.Op != "set" || cmd.Key != mockRulesRaftKey {
		return
	}
	var ruleCmd ruleCommand
	if err := json.Unmarshal([]byte(cmd.Value), &ruleCmd); err != nil {
		s.logger.Error("failed to unmarshal raft mock rule command", zap.Error(err))
		return
	}
	s.applyRuleCommand(ruleCmd)
}

func (s *MockSerice) applyRuleCommand(cmd ruleCommand) {
	switch cmd.Op {
	case "add":
		if cmd.Rule != nil {
			s.registry.add(*cmd.Rule)
		}
	case "delete":
		s.registry.delete(cmd.Method, cmd.Path)
	case "clear":
		s.registry.clear()
	case "replace":
		s.registry.replaceAll(cmd.Rules)
	}
}

func (s *MockSerice) mockHandler(c *gin.Context) {
	method := strings.ToUpper(c.Request.Method)
	urlpath := c.Request.URL.Path

	rule, exists := s.registry.find(method, urlpath)

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "no mock found for this method and path"})
		return
	}

	// Set response headers
	if rule.ResponseHeaders != nil {
		for key, value := range rule.ResponseHeaders {
			c.Header(key, value)
		}
	}

	// Handle different response types
	switch rule.ResponseType {
	case ResponseTypeSSE:
		s.handleSSEResponse(c, rule)
	case ResponseTypeWebSocket:
		s.handleWebSocketResponse(c, rule)
	default: // ResponseTypeHTTP or empty (default to HTTP)
		s.handleHTTPResponse(c, rule)
	}
}

func (s *MockSerice) handleHTTPResponse(c *gin.Context, rule *MockRule) {
	// Set response status and body
	if rule.ResponseBody != "" {
		c.String(rule.ResponseStatus, rule.ResponseBody)
	} else {
		c.JSON(rule.ResponseStatus, gin.H{})
	}
}

func (s *MockSerice) handleSSEResponse(c *gin.Context, rule *MockRule) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Status(http.StatusOK)

	w := c.Writer
	for _, event := range rule.SSEEvents {
		if event.Delay > 0 {
			time.Sleep(event.Delay)
		}

		if event.Event != "" {
			fmt.Fprintf(w, "event: %s\n", event.Event)
		}
		fmt.Fprintf(w, "data: %s\n\n", event.Data)

		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}

func (s *MockSerice) handleWebSocketResponse(c *gin.Context, rule *MockRule) {
	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("Failed to upgrade to WebSocket", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}
	defer conn.Close()

	// Send WebSocket messages
	for _, msg := range rule.WebSocketMessages {
		if msg.Delay > 0 {
			time.Sleep(msg.Delay)
		}

		messageType := websocket.TextMessage
		if msg.Type == "binary" {
			messageType = websocket.BinaryMessage
		}

		err := conn.WriteMessage(messageType, []byte(msg.Message))
		if err != nil {
			s.logger.Error("Failed to send WebSocket message", zap.Error(err))
			break
		}
	}
}
