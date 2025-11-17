// Package main implements a RocketMQ consumer that receives messages and forwards them to an HTTP server
// It supports swimlane routing by tagging messages and forwarding them according to the swimlane configuration
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

const (
	Topic         = "swimlane-test-topic"          // Name of the RocketMQ topic to consume from
	NameServer    = "127.0.0.1:9876"               // Default RocketMQ nameserver address
	ConsumerGroup = "swimlane-test-consumer-group" // Name of the consumer group
)

// Main function initializes and starts the RocketMQ consumer
// It reads configuration from environment variables, creates the consumer,
// subscribes to the topic with tag filtering, and starts listening for messages
func main() {
	// Get configuration from environment variables
	tagName := os.Getenv("TAG_NAME")
	if tagName == "" {
		tagName = "v1" // Default tag name for swimlane routing
	}

	httpServer := os.Getenv("HTTP_SERVER")
	if httpServer == "" {
		httpServer = "http://mocka:8000" // Default HTTP server to forward messages (points to mock service)
	}

	nameServer := os.Getenv("NAMESRV_ADDR")
	if nameServer == "" {
		nameServer = NameServer // Use default nameserver if not specified
	}
	nameServer, err := resolveSingleAddress(nameServer)
	if err != nil {
		panic("Failed to resolve nameserver address: " + err.Error())
	}

	fmt.Printf("Starting RocketMQ consumer with tag: %s, forwarding to: %s, nameserver: %s\n", tagName, httpServer, nameServer)

	// Create a push consumer with the specified nameserver and consumer group
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{nameServer}),
		consumer.WithConsumerModel(consumer.Clustering),   // Use clustering consumer model
		consumer.WithGroupName(ConsumerGroup+"-"+tagName), // Create unique consumer group per tag
	)
	if err != nil {
		panic("Failed to create consumer: " + err.Error())
	}

	// Subscribe to the topic with tag filter and register the message listener
	err = c.Subscribe(Topic, consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: tagName, // Filter messages by the tag name to implement swimlane routing
	}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			// Get the tag property name from environment variable
			tagPropertyName := os.Getenv("TAG_PROPERTY_NAME")
			if tagPropertyName == "" {
				tagPropertyName = "x-asm-prefer-tag" // Default tag property name
			}

			// Extract property from message properties for swimlane routing
			property := msg.GetProperty(tagPropertyName)
			if property == "" {
				// Fallback to default tag if not found in properties
				property = tagName
			}

			messageContent := string(msg.Body)
			// Extract trace context headers from message properties
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
				if value := msg.GetProperty(headerName); value != "" {
					traceHeaders[headerName] = value
				}
			}

			fmt.Printf("Received message: %s with property: %s\n", messageContent, property)

			// Forward the message to the HTTP server with the tag as a header for swimlane routing
			// and include the extracted trace headers for distributed tracing
			err := forwardToHTTPServer(httpServer, messageContent, property, traceHeaders)
			if err != nil {
				fmt.Printf("Error forwarding message to HTTP server: %s\n", err.Error())
				// Return failure to retry message consumption
				return consumer.ConsumeRetryLater, err
			}

			fmt.Println("Message forwarded to HTTP server successfully")
		}

		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		panic("Failed to subscribe: " + err.Error())
	}

	// Start the consumer to begin receiving messages
	err = c.Start()
	if err != nil {
		panic("Failed to start consumer: " + err.Error())
	}

	fmt.Println("RocketMQ Consumer started successfully")
	fmt.Printf("Waiting for messages with tag: %s\n", tagName)

	// Keep the main goroutine running to maintain consumer operation
	select {}
}

// forwardToHTTPServer forwards the received message to an HTTP server with the swimlane tag as a header
// This allows the HTTP server to route the request according to ASM swimlane rules
// It also includes trace headers for distributed tracing
func forwardToHTTPServer(httpServer, messageContent, property string, traceHeaders map[string]string) error {
	client := &http.Client{}

	// Create an HTTP POST request with the message content in the body
	req, err := http.NewRequest("POST", httpServer, bytes.NewBuffer([]byte(messageContent)))
	if err != nil {
		return fmt.Errorf("error creating request: %s", err.Error())
	}

	// Get the header name from environment variable for ASM swimlane routing
	httpHeaderName := os.Getenv("HTTP_HEADER_NAME")
	if httpHeaderName == "" {
		httpHeaderName = "x-asm-prefer-tag" // Default header name for ASM swimlane routing
	}

	// Add the swimlane tag as a header for traffic routing
	req.Header.Set(httpHeaderName, property)
	req.Header.Set("Content-Type", "text/plain")

	// Add trace headers for distributed tracing
	for headerName, headerValue := range traceHeaders {
		req.Header.Set(headerName, headerValue)
	}

	// Send the request to the HTTP server
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err.Error())
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Warning: Could not read response body: %s\n", err.Error())
	}

	fmt.Printf("HTTP request sent with property '%s', response status: %s, body: %s\n", property, resp.Status, string(respBody))

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
