package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"mockservice/backend/api"
	"mockservice/backend/log"

	"github.com/gorilla/websocket"
)

const defaultAddr = ":8080"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		serveCmd(os.Args[2:])
	case "http":
		httpCmd(os.Args[2:])
	case "sse":
		sseCmd(os.Args[2:])
	case "ws":
		wsCmd(os.Args[2:])
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`Usage:
  cmder <command> [options]

Commands:
  serve      Start mock service
  http       Send a generic HTTP request
  sse        Connect to an SSE endpoint and stream events
  ws         Connect to a WebSocket endpoint and receive messages
  help       Show this help message

Examples:
  cmder serve -addr :8080
  cmder http -method GET -url http://localhost:8080/v1/test
  cmder http -method POST -url http://localhost:8080/v1/test -body '{"a":1}' -H 'Content-Type: application/json'
  cmder sse -url http://localhost:8080/v1/sse-test
  cmder ws -url ws://localhost:8080/v1/ws-test -send "hello"
`)
}

func serveCmd(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", defaultAddr, "listen address")
	logPath := fs.String("log", "mockservice.log", "log file path")
	fs.Parse(args)

	log.Init(*logPath)
	service := api.NewMockService("")
	if err := service.Run(*addr); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start server: %v\n", err)
		os.Exit(1)
	}
}

func httpCmd(args []string) {
	fs := flag.NewFlagSet("http", flag.ExitOnError)
	method := fs.String("method", "GET", "HTTP method")
	urlString := fs.String("url", "", "request URL")
	body := fs.String("body", "", "request body")
	contentType := fs.String("content-type", "", "Content-Type header")
	var headers headerFlags
	fs.Var(&headers, "H", "header in 'Key: Value' form, repeatable")
	fs.Parse(args)

	if *urlString == "" {
		fmt.Fprintln(os.Stderr, "url is required")
		fs.Usage()
		os.Exit(1)
	}

	req, err := http.NewRequest(strings.ToUpper(*method), *urlString, strings.NewReader(*body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid request: %v\n", err)
		os.Exit(1)
	}

	for _, header := range headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "invalid header: %s\n", header)
			os.Exit(1)
		}
		req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}

	if *contentType != "" {
		req.Header.Set("Content-Type", *contentType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Println("Headers:")
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read body: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nBody:\n%s\n", string(bodyBytes))
}

func sseCmd(args []string) {
	fs := flag.NewFlagSet("sse", flag.ExitOnError)
	urlString := fs.String("url", "", "SSE endpoint URL")
	timeout := fs.Duration("timeout", 15*time.Second, "read timeout")
	fs.Parse(args)

	if *urlString == "" {
		fmt.Fprintln(os.Stderr, "url is required")
		fs.Usage()
		os.Exit(1)
	}

	parsedURL, err := url.Parse(*urlString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid url: %v\n", err)
		os.Exit(1)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		fmt.Fprintln(os.Stderr, "SSE URL must use http or https")
		os.Exit(1)
	}

	req, err := http.NewRequest(http.MethodGet, *urlString, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: *timeout}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SSE connection failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "unexpected status: %s\n", resp.Status)
		os.Exit(1)
	}

	fmt.Println("Connected to SSE endpoint")
	fmt.Println("Streaming events...")

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		fmt.Println(line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(1)
	}
}

func wsCmd(args []string) {
	fs := flag.NewFlagSet("ws", flag.ExitOnError)
	httpURL := fs.String("url", "", "mock service URL (e.g., http://localhost:8080/v1/ws-test)")
	message := fs.String("message", "hello from ws mock", "message to send from the mock")
	messageDelay := fs.Duration("delay", 0, "delay before sending message")
	count := fs.Int("count", 1, "number of messages to send")
	timeout := fs.Duration("timeout", 15*time.Second, "connection timeout")
	fs.Parse(args)

	if *httpURL == "" {
		fmt.Fprintln(os.Stderr, "url is required")
		fs.Usage()
		os.Exit(1)
	}

	// Convert http URL to ws URL
	wsURL := strings.Replace(*httpURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)

	// Build WebSocket mock registration
	var messages []api.WebSocketMessage
	for i := 0; i < *count; i++ {
		messages = append(messages, api.WebSocketMessage{
			Message: *message,
			Delay:   *messageDelay,
			Type:    "text",
		})
	}

	// Extract path from URL
	// parsedURL, err := url.Parse(*httpURL)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "invalid url: %v\n", err)
	// 	os.Exit(1)
	// }
	//path := parsedURL.Path

	// Register mock via API
	// rule := api.MockRule{
	// 	Method:            strings.ToUpper(*method),
	// 	Path:              path,
	// 	ResponseType:      api.ResponseTypeWebSocket,
	// 	WebSocketMessages: messages,
	// }

	// ruleJSON, _ := json.Marshal(rule)
	// regResp, err := http.Post(
	// 	*serviceAddr+"/v1/__mock",
	// 	"application/json",
	// 	bytes.NewBuffer(ruleJSON),
	// )
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to register mock: %v\n", err)
	// 	os.Exit(1)
	// }
	// if regResp.StatusCode != http.StatusCreated {
	// 	io.ReadAll(regResp.Body)
	// 	regResp.Body.Close()
	// 	fmt.Fprintf(os.Stderr, "failed to register mock: status %d\n", regResp.StatusCode)
	// 	os.Exit(1)
	// }
	// regResp.Body.Close()
	// fmt.Printf("Registered WebSocket mock at %s\n", path)

	// Connect to WebSocket
	dialer := websocket.Dialer{HandshakeTimeout: *timeout}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WebSocket dial failed: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("WebSocket connected, waiting for messages...")

	received := 0
	for {
		if received >= *count {
			break
		}

		if err := conn.SetReadDeadline(time.Now().Add(*timeout)); err != nil {
			fmt.Fprintf(os.Stderr, "read deadline failed: %v\n", err)
			os.Exit(1)
		}

		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				fmt.Println("connection closed by server")
				return
			}
			fmt.Fprintf(os.Stderr, "read failed: %v\n", err)
			os.Exit(1)
		}

		prefix := "TEXT"
		if messageType == websocket.BinaryMessage {
			prefix = "BINARY"
		}
		fmt.Printf("received (%s): %s\n", prefix, string(msg))
		received++
	}
	fmt.Println("All messages received")
}

type headerFlags []string

func (h *headerFlags) String() string {
	return strings.Join(*h, ", ")
}

func (h *headerFlags) Set(value string) error {
	* h = append(*h, value)
	return nil
}
