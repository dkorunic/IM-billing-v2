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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/pquerna/ffjson/ffjson"
)

// ifconfigURL is a default GeoIP URL with JSON response.
const ifconfigURL = "https://ifconfig.co/json"

// defaultIfconfigTimeout is a default Ifconfig/GeoIP request timeout.
const defaultIfconfigTimeout = 10 * time.Second

// IfconfigResponse is a structure for parsed ifconfig.co JSON response.
type IfconfigResponse struct {
	IP         string      `json:"ip"`
	IPdecimal  json.Number `json:"ip_decimal"`
	Country    string      `json:"country"`
	CountryISO string      `json:"country_iso"`
	City       string      `json:"city"`
	Hostname   string      `json:"hostname"`
}

// IfconfigClient is an Ifconfig client that performs simple geolocation.
type IfconfigClient struct {
	httpClient *http.Client
	URL        *url.URL
}

// NewIfconfigClient prepares HTTP client structure for Ifconfig API request.
func NewIfconfigClient() (*IfconfigClient, error) {
	ifconfigURL, err := url.Parse(ifconfigURL)
	if err != nil {
		log.Fatal(err)
	}

	c := &IfconfigClient{httpClient: &http.Client{Timeout: defaultIfconfigTimeout}, URL: ifconfigURL}
	return c, nil
}

// GetIfconfigResponse fetches a HTTP response with JSON body from ifconfig.co site and parses it.
func (IfconfigClient *IfconfigClient) GetIfconfigResponse() (IfconfigResponse, error) {
	req, err := http.NewRequest("GET", IfconfigClient.URL.String(), nil)
	if err != nil {
		return IfconfigResponse{}, err
	}

	// Do the actual HTTP/HTTPS request
	resp, err := IfconfigClient.httpClient.Do(req)
	if err != nil {
		return IfconfigResponse{}, err
	}

	// Defer body close() with error propagation
	defer func() {
		cerr := resp.Body.Close()
		if err == nil {
			err = cerr
		}
	}()

	// Fetch whole body at once
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return IfconfigResponse{}, err
	}

	// Handle HTTP errors
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		var err = fmt.Errorf(string(body))
		return IfconfigResponse{}, err
	}

	// Parse received JSON
	geoip := IfconfigResponse{}
	err = ffjson.Unmarshal(body, &geoip)
	return geoip, err
}
