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
package ics

import (
	"context"
	"log"
	"net/http"
	"net/url"

	"time"

	"fmt"

	"github.com/jordic/goics"
)

// URL is a country-local holiday calendar in ICS format.
const URL = "https://www.officeholidays.com/ics/ics_country_iso.php?tbl_country=%s"

// DefaultTimeout is a default ICS fetch HTTP timeout.
const DefaultTimeout = 10 * time.Second

// Client is an ICS HTTP client for remote fetching/parsing ICS calendar.
type Client struct {
	httpClient *http.Client
	URL        *url.URL
	ctx        context.Context
}

// Event is an individual parsed ICS event for ICS decoder.
type Event struct {
	Start, End  time.Time
	ID, Summary string
}

// Events is a collection of parsed ICS events from ICS decoder.
type Events []Event

// ConsumeICal consumes/parses an individual ICS event into an Event structure.
func (e *Events) ConsumeICal(c *goics.Calendar, err error) error {
	for _, el := range c.Events {
		node := el.Data
		dtstart, err := node["DTSTART"].DateDecode()
		if err != nil {
			return err
		}
		dtend, err := node["DTEND"].DateDecode()
		if err != nil {
			return err
		}
		d := Event{
			Start:   dtstart,
			End:     dtend,
			ID:      node["UID"].Val,
			Summary: node["SUMMARY"].Val,
		}
		*e = append(*e, d)
	}
	return nil
}

// NewClient creates a HTTP client structure for ICS fetch/parse.
func NewClient(countryCode string) (*Client, error) {
	ctx := context.Background()

	return NewClientWithContext(ctx, countryCode)
}

// NewClientWithContext creates a HTTP client structure for ICS fetch/parse with ctx Context.
func NewClientWithContext(ctx context.Context, countryCode string) (*Client, error) {
	IcsURL, err := url.Parse(fmt.Sprintf(URL, countryCode))
	if err != nil {
		log.Fatal(err)
	}

	c := &Client{httpClient: &http.Client{Timeout: DefaultTimeout}, URL: IcsURL, ctx: ctx}
	return c, nil
}

// GetResponse fetches a HTTP response from officeholldays site with country-local ICS as a body.
func (c *Client) GetResponse() (Events, error) {
	req, err := http.NewRequestWithContext(c.ctx, "GET", c.URL.String(), nil)
	if err != nil {
		return Events{}, err
	}

	// Do the actual HTTP/HTTPS request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		select {
		case <-c.ctx.Done():
			return Events{}, c.ctx.Err()
		default:
			return Events{}, err
		}
	}

	// Defer body close() with error propagation
	defer func() {
		cerr := resp.Body.Close()
		if err == nil {
			err = cerr
		}
	}()

	// Parse received ICS
	d := goics.NewDecoder(resp.Body)
	evs := Events{}
	err = d.Decode(&evs)
	if err != nil {
		return Events{}, err
	}

	return evs, nil
}
