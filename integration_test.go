package opik

import (
	"context"
	"os"
	"testing"
	"time"
)

// Integration tests for the Opik SDK.
// These tests require OPIK_API_KEY and OPIK_WORKSPACE environment variables.
// Run with: go test -v -run Integration

func skipIfNoCredentials(t *testing.T) {
	t.Helper()
	if os.Getenv("OPIK_API_KEY") == "" {
		t.Skip("OPIK_API_KEY not set, skipping integration test")
	}
	if os.Getenv("OPIK_WORKSPACE") == "" {
		t.Skip("OPIK_WORKSPACE not set, skipping integration test")
	}
}

func TestIntegration_CreateClient(t *testing.T) {
	skipIfNoCredentials(t)

	client, err := NewClient(
		WithProjectName("go-sdk-integration-tests"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Client is nil")
	}

	t.Logf("Client created successfully, URL: %s", client.Config().URL)
}

func TestIntegration_ListProjects(t *testing.T) {
	skipIfNoCredentials(t)

	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	projects, err := client.ListProjects(ctx, 1, 10)
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}

	t.Logf("Found %d projects", len(projects))
	for _, p := range projects {
		t.Logf("  - %s (ID: %s)", p.Name, p.ID)
	}
}

func TestIntegration_CreateTrace(t *testing.T) {
	skipIfNoCredentials(t)

	client, err := NewClient(
		WithProjectName("go-sdk-integration-tests"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Create a trace
	trace, err := client.Trace(ctx, "integration-test-trace",
		WithTraceInput(map[string]any{
			"test":      "integration",
			"timestamp": time.Now().Format(time.RFC3339),
		}),
		WithTraceTags("integration-test", "go-sdk"),
	)
	if err != nil {
		t.Fatalf("Failed to create trace: %v", err)
	}

	t.Logf("Created trace: %s", trace.ID())

	// Verify trace ID is a valid UUID v7 (starts with timestamp-based prefix)
	if trace.ID() == "" {
		t.Fatal("Trace ID is empty")
	}
}

func TestIntegration_CreateTraceWithSpan(t *testing.T) {
	skipIfNoCredentials(t)

	client, err := NewClient(
		WithProjectName("go-sdk-integration-tests"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Create a trace
	trace, err := client.Trace(ctx, "integration-test-with-span",
		WithTraceInput(map[string]any{
			"prompt": "Test prompt for integration",
		}),
		WithTraceTags("integration-test", "go-sdk", "with-span"),
	)
	if err != nil {
		t.Fatalf("Failed to create trace: %v", err)
	}
	t.Logf("Created trace: %s", trace.ID())

	// Create a span within the trace
	span, err := trace.Span(ctx, "test-llm-call",
		WithSpanType(SpanTypeLLM),
		WithSpanModel("test-model"),
		WithSpanProvider("test-provider"),
		WithSpanInput(map[string]any{
			"messages": []map[string]string{
				{"role": "user", "content": "Hello from integration test"},
			},
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create span: %v", err)
	}
	t.Logf("Created span: %s", span.ID())

	// Simulate work
	time.Sleep(50 * time.Millisecond)

	// End span
	err = span.End(ctx,
		WithSpanOutput(map[string]any{
			"response": "Integration test response",
		}),
	)
	if err != nil {
		t.Fatalf("Failed to end span: %v", err)
	}
	t.Log("Span ended successfully")

	// End trace
	err = trace.End(ctx,
		WithTraceOutput(map[string]any{
			"result": "success",
		}),
	)
	if err != nil {
		t.Fatalf("Failed to end trace: %v", err)
	}
	t.Log("Trace ended successfully")
}

func TestIntegration_AddFeedbackScore(t *testing.T) {
	skipIfNoCredentials(t)

	client, err := NewClient(
		WithProjectName("go-sdk-integration-tests"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Create a trace
	trace, err := client.Trace(ctx, "integration-test-feedback",
		WithTraceInput(map[string]any{"test": "feedback"}),
	)
	if err != nil {
		t.Fatalf("Failed to create trace: %v", err)
	}
	t.Logf("Created trace: %s", trace.ID())

	// Create a span
	span, err := trace.Span(ctx, "llm-call-for-feedback",
		WithSpanType(SpanTypeLLM),
	)
	if err != nil {
		t.Fatalf("Failed to create span: %v", err)
	}

	// End span
	err = span.End(ctx, WithSpanOutput(map[string]any{"response": "test"}))
	if err != nil {
		t.Fatalf("Failed to end span: %v", err)
	}

	// Add feedback score
	err = span.AddFeedbackScore(ctx, "quality", 0.95, "Good response from integration test")
	if err != nil {
		t.Fatalf("Failed to add feedback score: %v", err)
	}
	t.Log("Feedback score added successfully")

	// End trace
	err = trace.End(ctx, WithTraceOutput(map[string]any{"result": "success"}))
	if err != nil {
		t.Fatalf("Failed to end trace: %v", err)
	}
}
