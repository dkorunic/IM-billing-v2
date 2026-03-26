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

// White-box tests for the geoip package that need access to unexported fields.
package geoip

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// errCloseBody wraps an io.Reader and returns a pre-configured error from Close.
type errCloseBody struct {
	io.Reader
	closeErr error
}

func (b *errCloseBody) Close() error { return b.closeErr }

// fixedTransport returns a canned HTTP response for any request.
type fixedTransport struct {
	body     string
	status   int
	closeErr error
}

func (tr *fixedTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: tr.status,
		Header:     make(http.Header),
		Body:       &errCloseBody{Reader: strings.NewReader(tr.body), closeErr: tr.closeErr},
	}, nil
}

func newInternalTestClient(body string, status int, closeErr error) *Client {
	u, _ := url.Parse("http://example.com")

	return &Client{
		httpClient: &http.Client{Transport: &fixedTransport{
			body:     body,
			status:   status,
			closeErr: closeErr,
		}},
		URL: u,
		ctx: context.Background(),
	}
}

// TC-10a: when JSON decode succeeds but Close() returns an error, the close error
// must be propagated and the decoded data must still be present in the return value.
func TestGetResponse_CloseErrorPropagated(t *testing.T) {
	closeErr := errors.New("deliberate close error")
	validJSON := `{"ip":"1.2.3.4","country":"Croatia","country_iso":"HR","city":"Zagreb"}`

	c := newInternalTestClient(validJSON, http.StatusOK, closeErr)

	resp, err := c.GetResponse()

	if !errors.Is(err, closeErr) {
		t.Errorf("expected close error to be propagated, got: %v", err)
	}

	if resp.IP != "1.2.3.4" {
		t.Errorf("IP: got %q, want 1.2.3.4 — decoded data must be preserved alongside close error", resp.IP)
	}
}

// TC-10b: when JSON decode fails but Close() returns nil, the decode error must not
// be masked. A nil close error must not overwrite a prior non-nil decode error.
func TestGetResponse_DecodeErrorNotMaskedByNilClose(t *testing.T) {
	c := newInternalTestClient(`{not valid json`, http.StatusOK, nil)

	_, err := c.GetResponse()

	if err == nil {
		t.Error("expected decode error to be returned, got nil — decode error was masked by nil close error")
	}
}
