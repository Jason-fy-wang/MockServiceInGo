package api

import (
	"net/http"
	"strings"

	"mockservice/backend/log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MockSerice struct {
	registry *mockRegistry
	logger *zap.Logger
}

func NewMockService() *MockSerice {
	mockservice := &MockSerice{}
	mockservice.registry = newMockRegistry()
	mockservice.logger = log.Get()
	return mockservice
}

func (s *MockSerice) NewRouter() *gin.Engine {

	router := gin.New()
	{
		v1 := router.Group("/v1")
		router.Use(gin.Recovery(), gin.Logger())

		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "mock service running",
				"routes":  "POST /__mock, GET /__mock, DELETE /__mock, ANY /mock/*path",
			})
		})

		v1.POST("/__mock", s.registerMock)
		v1.GET("/__mock", s.listMocks)
		v1.DELETE("/__mock/all", s.clearMocks)
		v1.DELETE("/__mock/:path/:method", s.deleteMock)

		v1.Any("/", s.mockHandler)
	}
	return router
}

func (s *MockSerice) Run(addr string) error {
	s.logger.Info("starting mock service", zap.String("address", addr))
	return s.NewRouter().Run(addr)
}

func (s *MockSerice) registerMock(c *gin.Context) {
	var rule mockRule
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

	s.registry.add(rule)
	c.JSON(http.StatusCreated, gin.H{"message": "mock registered", "mock": rule})
}

func (s *MockSerice) listMocks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"mocks": s.registry.list()})
}

func (s *MockSerice) clearMocks(c *gin.Context) {
	s.registry.clear()
	c.JSON(http.StatusOK, gin.H{"message": "all mocks cleared"})
}

func (s *MockSerice) deleteMock(c *gin.Context) {
	path := c.Param("path")
	method := c.Param("method")
	s.registry.delete(method, path)
	c.JSON(http.StatusOK, gin.H{"message": "mock deleted", "method": method, "path": path})
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

	// Set response status and body
	if rule.ResponseBody != "" {
		c.String(rule.ResponseStatus, rule.ResponseBody)
	} else {
		c.JSON(rule.ResponseStatus, gin.H{})
	}
}
