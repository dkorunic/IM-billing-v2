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
package geoip_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dkorunic/IM-billing-v2/geoip"
)

func TestNewClientWithContext(t *testing.T) {
	client, err := geoip.NewClientWithContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.URL == nil {
		t.Fatal("expected non-nil URL")
	}

	if client.URL.String() != geoip.URL {
		t.Errorf("URL: got %q, want %q", client.URL.String(), geoip.URL)
	}
}

func TestGetResponse_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ip":"1.2.3.4","country":"Croatia","country_iso":"HR","city":"Zagreb","hostname":"host.example.com"}`))
	}))
	defer srv.Close()

	client, _ := geoip.NewClientWithContext(context.Background())
	client.URL, _ = url.Parse(srv.URL)

	resp, err := client.GetResponse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.IP != "1.2.3.4" {
		t.Errorf("IP: got %q, want %q", resp.IP, "1.2.3.4")
	}

	if resp.Country != "Croatia" {
		t.Errorf("Country: got %q, want %q", resp.Country, "Croatia")
	}

	if resp.CountryISO != "HR" {
		t.Errorf("CountryISO: got %q, want %q", resp.CountryISO, "HR")
	}
}

func TestGetResponse_HTTPError(t *testing.T) {
	const errBody = "service unavailable"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(errBody))
	}))
	defer srv.Close()

	client, _ := geoip.NewClientWithContext(context.Background())
	client.URL, _ = url.Parse(srv.URL)

	_, err := client.GetResponse()
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}

	if !strings.Contains(err.Error(), errBody) {
		t.Errorf("error %q does not contain body %q", err.Error(), errBody)
	}
}

func TestGetResponse_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer srv.Close()

	client, _ := geoip.NewClientWithContext(context.Background())
	client.URL, _ = url.Parse(srv.URL)

	_, err := client.GetResponse()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestGetResponse_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before making the request

	client, _ := geoip.NewClientWithContext(ctx)
	client.URL, _ = url.Parse(srv.URL)

	_, err := client.GetResponse()
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

// TC-11: GetResponse must return context.Canceled when the context is cancelled,
// not a generic transport-level error.
func TestGetResponse_CancelledContextReturnsContextError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, _ := geoip.NewClientWithContext(ctx)
	client.URL, _ = url.Parse(srv.URL)

	_, err := client.GetResponse()
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}
