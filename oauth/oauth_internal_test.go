// Copyright (C) 2023  Dinko Korunic
//
// SPDX-License-Identifier: MIT

package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

const (
	testTokenType    = "Bearer"
	testRefreshToken = "test-refresh-token"
)

func TestTokenFromFile_Valid(t *testing.T) {
	t.Parallel()

	expiry := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	original := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    testTokenType,
		RefreshToken: testRefreshToken,
		Expiry:       expiry,
	}

	path := filepath.Join(t.TempDir(), "token.json")

	data, err := json.Marshal(original) //nolint:gosec // test fixture token
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	err = os.WriteFile(path, data, DefaultPerms)
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	tok, err := tokenFromFile(path)
	if err != nil {
		t.Fatalf("tokenFromFile: %v", err)
	}

	if tok.AccessToken != original.AccessToken {
		t.Errorf("AccessToken: got %q, want %q", tok.AccessToken, original.AccessToken)
	}

	if tok.RefreshToken != original.RefreshToken {
		t.Errorf("RefreshToken: got %q, want %q", tok.RefreshToken, original.RefreshToken)
	}

	if !tok.Expiry.Equal(expiry) {
		t.Errorf("Expiry: got %v, want %v", tok.Expiry, expiry)
	}
}

func TestTokenFromFile_NotFound(t *testing.T) {
	t.Parallel()

	_, err := tokenFromFile(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestTokenFromFile_InvalidJSON(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "bad.json")

	err := os.WriteFile(path, []byte(`{invalid json`), DefaultPerms)
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err = tokenFromFile(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestSaveToken_Valid(t *testing.T) {
	t.Parallel()

	expiry := time.Date(2099, 6, 15, 12, 0, 0, 0, time.UTC)
	tok := &oauth2.Token{
		AccessToken:  "save-access-token",
		TokenType:    testTokenType,
		RefreshToken: "save-refresh-token",
		Expiry:       expiry,
	}

	path := filepath.Join(t.TempDir(), "token.json")

	err := saveToken(path, tok)
	if err != nil {
		t.Fatalf("saveToken: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var got oauth2.Token

	err = json.Unmarshal(data, &got)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.AccessToken != tok.AccessToken {
		t.Errorf("AccessToken: got %q, want %q", got.AccessToken, tok.AccessToken)
	}

	if got.RefreshToken != tok.RefreshToken {
		t.Errorf("RefreshToken: got %q, want %q", got.RefreshToken, tok.RefreshToken)
	}
}

// TC-14: an expired token must trigger a refresh attempt via the token endpoint.
func TestGetClient_ExpiredTokenIsRefreshed(t *testing.T) {
	t.Parallel()

	const newAccessToken = "refreshed-access-token"

	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"` + newAccessToken + `","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	config := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			TokenURL: tokenSrv.URL,
		},
	}

	expiredTok := &oauth2.Token{
		AccessToken:  "old-access-token",
		RefreshToken: testRefreshToken,
		TokenType:    testTokenType,
		Expiry:       time.Now().Add(-time.Hour),
	}

	tokenPath := filepath.Join(t.TempDir(), "token.json")

	err := saveToken(tokenPath, expiredTok)
	if err != nil {
		t.Fatalf("saveToken: %v", err)
	}

	client, err := GetClient(context.Background(), config, tokenPath)
	if err != nil {
		t.Fatalf("GetClient: %v", err)
	}

	if client == nil {
		t.Fatal("expected non-nil HTTP client after refresh")
	}
}

// TC-15: after a successful token refresh that returns a different access token,
// the new token must be persisted to the token file.
func TestGetClient_RefreshedTokenSavedToFile(t *testing.T) {
	t.Parallel()

	const newAccessToken = "saved-refreshed-token"

	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"` + newAccessToken + `","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	config := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			TokenURL: tokenSrv.URL,
		},
	}

	expiredTok := &oauth2.Token{
		AccessToken:  "old-token-before-refresh",
		RefreshToken: testRefreshToken,
		TokenType:    testTokenType,
		Expiry:       time.Now().Add(-time.Hour),
	}

	tokenPath := filepath.Join(t.TempDir(), "token.json")

	err := saveToken(tokenPath, expiredTok)
	if err != nil {
		t.Fatalf("saveToken: %v", err)
	}

	_, err = GetClient(context.Background(), config, tokenPath)
	if err != nil {
		t.Fatalf("GetClient: %v", err)
	}

	savedTok, err := tokenFromFile(tokenPath)
	if err != nil {
		t.Fatalf("tokenFromFile after refresh: %v", err)
	}

	if savedTok.AccessToken != newAccessToken {
		t.Errorf("AccessToken: got %q, want %q — refreshed token was not saved to file", savedTok.AccessToken, newAccessToken)
	}
}

func TestSaveToken_RoundTrip(t *testing.T) {
	t.Parallel()

	original := &oauth2.Token{
		AccessToken:  "rt-access",
		TokenType:    testTokenType,
		RefreshToken: "rt-refresh",
		Expiry:       time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	path := filepath.Join(t.TempDir(), "roundtrip.json")

	err := saveToken(path, original)
	if err != nil {
		t.Fatalf("saveToken: %v", err)
	}

	loaded, err := tokenFromFile(path)
	if err != nil {
		t.Fatalf("tokenFromFile: %v", err)
	}

	if loaded.AccessToken != original.AccessToken {
		t.Errorf("AccessToken: got %q, want %q", loaded.AccessToken, original.AccessToken)
	}

	if loaded.RefreshToken != original.RefreshToken {
		t.Errorf("RefreshToken: got %q, want %q", loaded.RefreshToken, original.RefreshToken)
	}

	if !loaded.Expiry.Equal(original.Expiry) {
		t.Errorf("Expiry: got %v, want %v", loaded.Expiry, original.Expiry)
	}
}
