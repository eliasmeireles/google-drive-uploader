package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/eliasferreira/google-drive-uploader/internal/config"
)

func newTokenEndpoint(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func overrideTerminal(t *testing.T, interactive bool) {
	t.Helper()
	original := stdinIsTerminal
	stdinIsTerminal = func() bool { return interactive }
	t.Cleanup(func() { stdinIsTerminal = original })
}

func expiredToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  "expired-access-token",
		RefreshToken: "revoked-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}
}

func oauthConfig(tokenURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Endpoint:     oauth2.Endpoint{TokenURL: tokenURL},
	}
}

func TestGetClientRefreshFailure(t *testing.T) {
	t.Run("given a revoked token without a terminal then it fails fast", func(t *testing.T) {
		overrideTerminal(t, false)
		srv := newTokenEndpoint(t, http.StatusBadRequest, `{"error":"invalid_grant","error_description":"Token has been expired or revoked."}`)
		a := NewAuthenticator(config.Config{TokenPath: filepath.Join(t.TempDir(), "token.json")})

		client, err := a.getClient(context.Background(), oauthConfig(srv.URL), expiredToken())

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrTokenRefreshFailed)
		assert.Contains(t, err.Error(), "invalid_grant")
		assert.Contains(t, err.Error(), "--token-gen")
		assert.Nil(t, client)
	})
}

func TestGetClientRefreshSuccess(t *testing.T) {
	t.Run("given a refreshable token without a terminal then it returns a client and saves the new token", func(t *testing.T) {
		overrideTerminal(t, false)
		srv := newTokenEndpoint(t, http.StatusOK, `{"access_token":"new-access-token","token_type":"Bearer","expires_in":3600}`)
		tokenPath := filepath.Join(t.TempDir(), "token.json")
		saveToken(tokenPath, NewTokenFile("client-id", "client-secret").Refresh(expiredToken()))
		a := NewAuthenticator(config.Config{TokenPath: tokenPath})

		client, err := a.getClient(context.Background(), oauthConfig(srv.URL), expiredToken())

		require.NoError(t, err)
		require.NotNil(t, client)
		saved, _, err := tokenFromFile(tokenPath)
		require.NoError(t, err)
		assert.Equal(t, "new-access-token", saved.AccessToken)
	})

	t.Run("given a valid token then it returns a client without contacting the token endpoint", func(t *testing.T) {
		overrideTerminal(t, false)
		srv := newTokenEndpoint(t, http.StatusInternalServerError, `{}`)
		tokenPath := filepath.Join(t.TempDir(), "token.json")
		valid := &oauth2.Token{AccessToken: "valid", TokenType: "Bearer", Expiry: time.Now().Add(time.Hour)}
		saveToken(tokenPath, NewTokenFile("client-id", "client-secret").Refresh(valid))
		a := NewAuthenticator(config.Config{TokenPath: tokenPath})

		client, err := a.getClient(context.Background(), oauthConfig(srv.URL), valid)

		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}
