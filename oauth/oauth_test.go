// @license
// Copyright (C) 2023  Dinko Korunic
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package oauth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestTokenFromFile_Valid(t *testing.T) {
	expiry := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	original := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       expiry,
	}

	path := filepath.Join(t.TempDir(), "token.json")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if err = os.WriteFile(path, data, DefaultPerms); err != nil {
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
	_, err := tokenFromFile(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestTokenFromFile_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")

	if err := os.WriteFile(path, []byte(`{invalid json`), DefaultPerms); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := tokenFromFile(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestSaveToken_Valid(t *testing.T) {
	expiry := time.Date(2099, 6, 15, 12, 0, 0, 0, time.UTC)
	tok := &oauth2.Token{
		AccessToken:  "save-access-token",
		TokenType:    "Bearer",
		RefreshToken: "save-refresh-token",
		Expiry:       expiry,
	}

	path := filepath.Join(t.TempDir(), "token.json")

	if err := saveToken(path, tok); err != nil {
		t.Fatalf("saveToken: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var got oauth2.Token

	if err = json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.AccessToken != tok.AccessToken {
		t.Errorf("AccessToken: got %q, want %q", got.AccessToken, tok.AccessToken)
	}

	if got.RefreshToken != tok.RefreshToken {
		t.Errorf("RefreshToken: got %q, want %q", got.RefreshToken, tok.RefreshToken)
	}
}

func TestSaveToken_RoundTrip(t *testing.T) {
	original := &oauth2.Token{
		AccessToken:  "rt-access",
		TokenType:    "Bearer",
		RefreshToken: "rt-refresh",
		Expiry:       time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	path := filepath.Join(t.TempDir(), "roundtrip.json")

	if err := saveToken(path, original); err != nil {
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
