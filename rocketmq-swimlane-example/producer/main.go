// Package main implements a RocketMQ producer that accepts HTTP GET requests and sends messages to RocketMQ
// It supports swimlane routing by tagging messages according to ASM swimlane configuration
// It also supports OpenTelemetry trace context propagation
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

const (
	Topic         = "swimlane-test-topic"          // Name of the RocketMQ topic to send messages to
	NameServer    = "127.0.0.1:9876"               // Default RocketMQ nameserver address
	ConsumerGroup = "swimlane-test-consumer-group" // Default consumer group name
)

// Main function initializes and starts the RocketMQ producer with an HTTP interface
// It reads configuration from environment variables, creates the producer,
// sets up an HTTP server to accept GET requests, and sends messages to RocketMQ
func main() {
	// Get configuration from environment variables
	tagName := os.Getenv("TAG_NAME")
	if tagName == "" {
		tagName = "v1" // Default tag name for swimlane routing
	}

	listenPort := os.Getenv("LISTEN_PORT")
	if listenPort == "" {
		listenPort = "8080" // Default port for HTTP server
	}

	nameServer := os.Getenv("NAMESRV_ADDR")
	if nameServer == "" {
		nameServer = NameServer // Use default nameserver if not specified
	}
	nameServer, err := resolveSingleAddress(nameServer)
	if err != nil {
		panic("Failed to resolve nameserver address: " + err.Error())
	}

	fmt.Printf("Starting RocketMQ producer with tag: %s, listening on port: %s, nameserver: %s\n", tagName, listenPort, nameServer)

	// Initialize RocketMQ producer with the specified nameserver
	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{nameServer}),
		producer.WithRetry(2), // Retry sending messages up to 2 times if failed
	)
	if err != nil {
		panic("Failed to create producer: " + err.Error())
	}

	if err := p.Start(); err != nil {
		panic("Failed to start producer: " + err.Error())
	}
	defer p.Shutdown()

	fmt.Println("RocketMQ Producer started successfully")

	// Set up HTTP server to accept GET requests for sending messages
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get message from query parameter, or use default timestamp
		messageContent := r.URL.Query().Get("message")
		if messageContent == "" {
			messageContent = fmt.Sprintf("Message sent at %s", time.Now().Format(time.RFC3339))
		}

		// Get the header name from environment variable for ASM swimlane routing
		httpHeaderName := os.Getenv("HTTP_HEADER_NAME")
		if httpHeaderName == "" {
			httpHeaderName = "x-asm-prefer-tag" // Default header name for ASM swimlane routing
		}

		// Get tag from header or use default - this enables swimlane routing
		requestTag := r.Header.Get(httpHeaderName)
		if requestTag == "" {
			requestTag = tagName
		}

		// Extract trace context headers to propagate tracing information
		traceHeaders := make(map[string]string)
		traceHeaderNames := []string{
			"traceparent",       // W3C Trace Context
			"tracestate",        // W3C Trace Context
			"x-request-id",      // Common request tracing header
			"x-b3-traceid",      // Zipkin B3 trace ID
			"x-b3-spanid",       // Zipkin B3 span ID
			"x-b3-parentspanid", // Zipkin B3 parent span ID
			"x-b3-sampled",      // Zipkin B3 sampled
			"x-b3-flags",        // Zipkin B3 flags
			"x-ot-span-context", // OpenTracing span context
		}

		for _, headerName := range traceHeaderNames {
			if value := r.Header.Get(headerName); value != "" {
				traceHeaders[headerName] = value
			}
		}

		fmt.Printf("Received HTTP request to send message: %s with tag: %s\n", messageContent, requestTag)

		// Send message to RocketMQ with the appropriate tag and trace context
		err = sendToRocketMQ(p, messageContent, tagName, requestTag, traceHeaders)
		if err != nil {
			http.Error(w, "Failed to send message to RocketMQ: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Message sent to RocketMQ %s successfully", tagName)))
	})

	// Start the HTTP server to accept requests
	addr := ":" + listenPort
	fmt.Printf("Starting HTTP server on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// sendToRocketMQ sends a message to RocketMQ with the specified tag for swimlane routing
// It creates a message with both custom properties and message tags for proper routing
// It also includes trace context headers for distributed tracing
func sendToRocketMQ(p rocketmq.Producer, messageContent string, tag string, property string, traceHeaders map[string]string) error {
	// Create a message with topic and body
	msg := &primitive.Message{
		Topic: Topic,
		Body:  []byte(messageContent),
	}

	// Get the tag property name from environment variable for ASM swimlane routing
	tagPropertyName := os.Getenv("TAG_PROPERTY_NAME")
	if tagPropertyName == "" {
		tagPropertyName = "x-asm-prefer-tag" // Default tag property name for ASM swimlane routing
	}

	// Add custom properties to the message for swimlane routing
	msg.WithProperty(tagPropertyName, property)

	// Set the actual message tag for filtering by consumers
	msg.WithTag(tag)

	// Add trace context headers as message properties for distributed tracing
	for headerName, headerValue := range traceHeaders {
		msg.WithProperty(headerName, headerValue)
	}

	fmt.Printf("Sending message: %s with tag: %s\n", messageContent, tag)

	// Send the message synchronously to RocketMQ
	res, err := p.SendSync(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("send message failed: %s", err.Error())
	}

	fmt.Printf("Message sent successfully. Message ID: %s\n", res.MsgID)
	return nil
}

// resolveSingleAddress resolves a hostname to a single IP address
// This is particularly useful when connecting to services in Kubernetes
func resolveSingleAddress(address string) (string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		host = address
		port = "9876"
	}

	if net.ParseIP(host) != nil {
		return net.JoinHostPort(host, port), nil
	}

	// Resolve hostname to IP addresses
	ips, err := net.LookupHost(host)
	if err != nil {
		return "", fmt.Errorf("failed to resolve %s: %w", host, err)
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no IP addresses found for %s", host)
	}

	// Use the first IP address returned
	return net.JoinHostPort(ips[0], port), nil
}
