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
package geoip

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// URL is a default GeoIP URL with JSON response.
const URL = "https://ifconfig.co/json"

// DefaultTimeout is a default Ifconfig/GeoIP request timeout.
const DefaultTimeout = 10 * time.Second

// Response is a structure for parsed ifconfig.co JSON response.
type Response struct {
	IP         string      `json:"ip"`
	IPdecimal  json.Number `json:"ip_decimal"`
	Country    string      `json:"country"`
	CountryISO string      `json:"country_iso"`
	City       string      `json:"city"`
	Hostname   string      `json:"hostname"`
}

// Client is an Ifconfig client that performs simple geolocation.
type Client struct {
	httpClient *http.Client
	URL        *url.URL
	ctx        context.Context
}

// NewClient prepares HTTP client structure for Ifconfig API request.
func NewClient() (*Client, error) {
	ctx := context.Background()

	return NewClientWithContext(ctx)
}

// NewClientWithContext prepares HTTP client structure for Ifconfig API request with ctx Context.
func NewClientWithContext(ctx context.Context) (*Client, error) {
	ifconfigURL, err := url.Parse(URL)
	if err != nil {
		log.Fatal(err)
	}

	c := &Client{httpClient: &http.Client{Timeout: DefaultTimeout}, URL: ifconfigURL, ctx: ctx}
	return c, nil
}

// GetResponse fetches a HTTP response with JSON body from ifconfig.co site and parses it.
func (c *Client) GetResponse() (Response, error) {
	req, err := http.NewRequestWithContext(c.ctx, "GET", c.URL.String(), nil)
	if err != nil {
		return Response{}, err
	}

	// Do the actual HTTP/HTTPS request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		select {
		case <-c.ctx.Done():
			return Response{}, c.ctx.Err()
		default:
			return Response{}, err
		}
	}

	// Defer body close() with error propagation
	defer func() {
		cerr := resp.Body.Close()
		if err == nil {
			err = cerr
		}
	}()

	// Fetch whole body at once
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	// Handle HTTP errors
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		err := fmt.Errorf(string(body))

		return Response{}, err
	}

	// Parse received JSON
	geoip := Response{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err = json.Unmarshal(body, &geoip)
	return geoip, err
}
