// Copyright (C) 2023  Dinko Korunic
//
// SPDX-License-Identifier: MIT

package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/renameio/v2/maybe"
	"github.com/google/uuid"
	"github.com/phayes/freeport"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

const (
	AuthTimeout       = 90 * time.Second
	AuthListenAddr    = "127.0.0.1"
	AuthScheme        = "http://"
	DefaultPerms      = 0o600
	ReadTimeout       = 5 * time.Second
	WriteTimeout      = 5 * time.Second
	IdleTimeout       = 60 * time.Second
	ReadHeaderTimeout = 10 * time.Second
)

var (
	ErrOAuthUUID        = errors.New("unable to generate UUID")
	ErrOAuthFreePort    = errors.New("unable to get a free port")
	ErrOAuthHTTPServer  = errors.New("unable to start HTTP server")
	ErrOAuthBrowser     = errors.New("unable to open system browser")
	ErrOAuthTimeout     = errors.New("timeout while waiting for authentication to finish")
	ErrOAuthTokenFetch  = errors.New("unable to retrieve token from Google API")
	ErrOAuthTokenSave   = errors.New("unable to save token to file")
	ErrOAuthTokenEncode = errors.New("unable to encode OAuth token to JSON")
)

// GetClient returns an authenticated HTTP client, loading the token from tokenPath,
// refreshing it if expired, or running the interactive browser flow if needed.
func GetClient(ctx context.Context, config *oauth2.Config, tokenPath string) (*http.Client, error) {
	tok, err := tokenFromFile(tokenPath)
	if err == nil && tok.Valid() {
		return config.Client(ctx, tok), nil
	}

	// we have a token, but it has expired so attempt to refresh it
	if err == nil {
		tok, err = config.TokenSource(ctx, tok).Token()
	}

	// no usable token or refresh failed (e.g. missing or revoked refresh token);
	// fall back to interactive browser flow
	if err != nil {
		tok, err = getTokenFromWeb(ctx, config)
		if err != nil {
			return nil, err
		}
	}

	// always persist a new token to capture updated expiry or rotated refresh token
	err = saveToken(tokenPath, tok)
	if err != nil {
		return nil, err
	}

	return config.Client(ctx, tok), nil
}

// getTokenFromWeb runs the interactive OAuth2 flow: it opens the system browser,
// serves a local callback to capture the auth code, and exchanges it for a token.
func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// random UUID as a state
	authReqState, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOAuthUUID, err)
	}

	tokChan := make(chan string, 1)
	errChan := make(chan error, 1)

	var once sync.Once

	// get a free random port
	authListenPort, err := freeport.GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOAuthFreePort, err)
	}

	// redirect uri listener for auth callback
	authListenHost := net.JoinHostPort(AuthListenAddr, strconv.Itoa(authListenPort))

	// oauth config auth redirect uri
	config.RedirectURL = AuthScheme + authListenHost

	s := http.Server{
		ReadTimeout:       ReadTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       IdleTimeout,
		ReadHeaderTimeout: ReadHeaderTimeout,
		Addr:              authListenHost,
	}

	defer func() {
		// detach from ctx cancellation so shutdown still runs after ctx is done
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()

		_ = s.Shutdown(shutdownCtx)
	}()

	// oauth callback handler
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Reject mismatched state but keep waiting; a stray or CSRF request
		// must not consume the one-shot success path or abort the flow.
		if r.URL.Query().Get("state") != authReqState.String() {
			http.Error(w, "Invalid authentication state", http.StatusUnauthorized)

			return
		}

		once.Do(func() {
			tokChan <- r.URL.Query().Get("code")
		})

		_, _ = io.WriteString(w, "Authentication complete, you can close this window.\n")

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	})

	s.Handler = r

	// oauth callback server
	go func() {
		err := s.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			once.Do(func() { errChan <- fmt.Errorf("%w: %w", ErrOAuthHTTPServer, err) })
		}
	}()

	authCodeURL := config.AuthCodeURL(authReqState.String(), oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	log.Printf("Opening auth URL through system browser: %v", authCodeURL)

	// oauth dialog through system browser
	err = browser.OpenURL(authCodeURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOAuthBrowser, err)
	}

	var authCode string

	ticker := time.NewTimer(AuthTimeout)
	defer ticker.Stop()

	select {
	case authCode = <-tokChan:
		ticker.Stop()
	case handlerErr := <-errChan:
		return nil, handlerErr
	case <-ticker.C:
		return nil, ErrOAuthTimeout
	}

	tok, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOAuthTokenFetch, err)
	}

	return tok, nil
}

// tokenFromFile reads and JSON-decodes an OAuth2 token from tokenPath.
func tokenFromFile(tokenPath string) (*oauth2.Token, error) {
	b, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}

	tok := &oauth2.Token{}

	err = json.Unmarshal(b, tok)

	return tok, err
}

// saveToken atomically writes the JSON-encoded OAuth2 token to tokenPath.
func saveToken(tokenPath string, token *oauth2.Token) error {
	buf := new(bytes.Buffer)

	err := json.NewEncoder(buf).Encode(token) //nolint:gosec // persisting the token to disk is the intent
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOAuthTokenEncode, err)
	}

	err = maybe.WriteFile(tokenPath, buf.Bytes(), DefaultPerms)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOAuthTokenSave, err)
	}

	return nil
}
