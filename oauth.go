/**
 * @license
 * Copyright (C) 2018  Dinko Korunic, InfoMAR
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the
 * Free Software Foundation; either version 2 of the License, or (at your
 * option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along
 * with this program; if not, write to the Free Software Foundation, Inc.,
 * 59 Temple Place, Suite 330, Boston, MA  02111-1307 USA
 */
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	uuid "github.com/gofrs/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/phayes/freeport"
	"github.com/pkg/browser"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

const (
	AuthTimeout    = 90 * time.Second
	AuthListenAddr = "127.0.0.1"
	AuthScheme     = "http://"
)

// getClient retrieves a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}

	return config.Client(context.Background(), tok)
}

// getTokenFromWeb requests a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// random UUID as a state
	authReqState, err := uuid.NewV7()
	if err != nil {
		log.Fatalf("Unable to generate UUID V4: %v", err)
	}

	tokChan := make(chan string, 1)

	// get a free random port
	authListenPort, err := freeport.GetFreePort()
	if err != nil {
		log.Fatalf("Unable to get a free port: %v", err)
	}
	// redirect uri listener for auth callback
	authListenHost := net.JoinHostPort(AuthListenAddr, strconv.Itoa(authListenPort))
	// oauth config auth redirect uri
	config.RedirectURL = AuthScheme + authListenHost

	s := http.Server{
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		Addr:              authListenHost,
	}
	defer s.Close()

	// oauth callback handler
	s.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if actualState := r.URL.Query().Get("state"); actualState != authReqState.String() {
			http.Error(w, "Invalid authentication state", http.StatusUnauthorized)

			return
		}

		tokChan <- r.URL.Query().Get("code")
		close(tokChan)
		_, _ = w.Write([]byte("Authentication complete, you can close this window."))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	})

	// oauth callback server
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error starting Web server: %v", err)
		}
	}()

	authCodeURL := config.AuthCodeURL(authReqState.String(), oauth2.AccessTypeOffline)
	fmt.Printf("Opening auth URL through system browser: %v\n", authCodeURL)

	// oauth dialog through system browser
	if err := browser.OpenURL(authCodeURL); err != nil {
		log.Fatalf("Unable to open system browser: %v", err)
	}

	var authCode string

	ticker := time.NewTimer(AuthTimeout)

	select {
	case authCode = <-tokChan:
		ticker.Stop()

		break
	case <-ticker.C:
		log.Fatal("Timeout while waiting for authentication to finish")
	}

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from Web: %v", err)
	}

	return tok
}

// tokenFromFile retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer func() {
		cerr := f.Close()
		if err == nil {
			err = cerr
		}
	}()

	tok := &oauth2.Token{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err = json.NewDecoder(f).Decode(tok)

	return tok, err
}

// saveToken saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}

	defer func() {
		cerr := f.Close()
		if err == nil && cerr != nil {
			log.Fatalf("Unable to cache oauth token: %v", cerr)
		}
	}()

	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		log.Fatalf("Unable to encode oauth token to JSON: %v", err)
	}
}
