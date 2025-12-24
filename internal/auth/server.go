package auth

import (
	"fmt"
	"net"
	"net/http"
)

// findAvailablePort finds an available port starting from 54321
func findAvailablePort() (int, error) {
	for port := 54321; port < 54330; port++ {
		addr := fmt.Sprintf("localhost:%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports found in range 8080-8099")
}

// startCallbackServer starts a local HTTP server to handle OAuth callback
func startCallbackServer(port int) (chan string, chan error, *http.Server) {
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no authorization code in callback")
			http.Error(w, "No authorization code received", http.StatusBadRequest)
			return
		}

		codeChan <- code

		// Display success page
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>🚀 Authorization Successful</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .success { color: #4CAF50; font-size: 24px; margin-bottom: 20px; }
        .message { color: #666; font-size: 16px; }
    </style>
</head>
<body>
    <div class="success">🚀 Authorization Successful!</div>
    <div class="message">You can close this window and return to the terminal.</div>
</body>
</html>
`)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("callback server error: %v", err)
		}
	}()

	return codeChan, errChan, server
}
