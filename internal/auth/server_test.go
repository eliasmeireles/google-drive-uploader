package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestFindAvailablePort(t *testing.T) {
	port, err := findAvailablePort()
	if err != nil {
		t.Fatalf("findAvailablePort failed: %v", err)
	}

	if port < 54321 || port >= 54330 {
		t.Errorf("Port %d out of expected range [54321, 54330)", port)
	}
}

func TestCallbackServerHandler(t *testing.T) {
	// Pick a port - we don't start the real server but we test the handler logic
	// Actually startCallbackServer returns the server and channels, we can test it fully if we want.
	// But let's test the handler primarily via httptest for better isolation?
	// The handler logic is inside startCallbackServer closure.
	// Ideally we would extract the handler func, but since it's private and inside,
	// let's test the full integration of startCallbackServer.

	port, err := findAvailablePort()
	if err != nil {
		t.Fatal("Could not find port for test")
	}

	codeChan, errChan, server := startCallbackServer(port)
	defer server.Close() // Will be closed by shutdown usually but good for test cleanup

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	baseURL := fmt.Sprintf("http://localhost:%d/callback", port)

	t.Run("Success Case", func(t *testing.T) {
		resp, err := http.Get(baseURL + "?code=test-auth-code")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Verify content type has charset
		contentType := resp.Header.Get("Content-Type")
		if contentType != "text/html; charset=utf-8" {
			t.Errorf("Expected Content-Type 'text/html; charset=utf-8', got '%s'", contentType)
		}

		select {
		case code := <-codeChan:
			if code != "test-auth-code" {
				t.Errorf("Expected code 'test-auth-code', got '%s'", code)
			}
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for code on channel")
		}
	})

	t.Run("Missing Code", func(t *testing.T) {
		// Drain channels first
		select {
		case <-codeChan:
		case <-errChan:
		default:
		}

		resp, err := http.Get(baseURL) // No code param
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for missing code, got %d", resp.StatusCode)
		}

		select {
		case err := <-errChan:
			if err.Error() != "no authorization code in callback" {
				t.Errorf("Unexpected error message: %v", err)
			}
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for error on channel")
		}
	})
}
