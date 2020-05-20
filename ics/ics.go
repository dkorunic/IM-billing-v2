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

// icsURL is a country-local holiday calendar in ICS format.
const icsURL = "https://www.officeholidays.com/ics/ics_country_iso.php?tbl_country=%s"

// defaultIcsTimeout is a default ICS fetch HTTP timeout.
const defaultIcsTimeout = 10 * time.Second

// IcsClient is an ICS HTTP client for remote fetching/parsing ICS calendar.
type IcsClient struct {
	httpClient *http.Client
	URL        *url.URL
	ctx        context.Context
}

// IcsEvent is an individual parsed ICS event for ICS decoder.
type IcsEvent struct {
	Start, End  time.Time
	Id, Summary string
}

// IcsEvents is a collection of parsed ICS events from ICS decoder.
type IcsEvents []IcsEvent

// ConsumeICal consumes/parses an individual ICS event into an IcsEvent structure.
func (e *IcsEvents) ConsumeICal(c *goics.Calendar, err error) error {
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
		d := IcsEvent{
			Start:   dtstart,
			End:     dtend,
			Id:      node["UID"].Val,
			Summary: node["SUMMARY"].Val,
		}
		*e = append(*e, d)
	}
	return nil
}

// NewIcsClient creates a HTTP client structure for ICS fetch/parse.
func NewIcsClient(countryCode string) (*IcsClient, error) {
	ctx := context.Background()
	return NewIcsClientWithContext(ctx, countryCode)
}

// NewIcsClientWithContext creates a HTTP client structure for ICS fetch/parse with ctx Context.
func NewIcsClientWithContext(ctx context.Context, countryCode string) (*IcsClient, error) {
	IcsURL, err := url.Parse(fmt.Sprintf(icsURL, countryCode))
	if err != nil {
		log.Fatal(err)
	}

	c := &IcsClient{httpClient: &http.Client{Timeout: defaultIcsTimeout}, URL: IcsURL, ctx: ctx}
	return c, nil
}

// GetIcsResponse fetches a HTTP response from officeholldays site with country-local ICS as a body.
func (IcsClient *IcsClient) GetIcsResponse() (IcsEvents, error) {
	req, err := http.NewRequestWithContext(IcsClient.ctx, "GET", IcsClient.URL.String(), nil)
	if err != nil {
		return IcsEvents{}, err
	}

	// Do the actual HTTP/HTTPS request
	resp, err := IcsClient.httpClient.Do(req)
	if err != nil {
		return IcsEvents{}, err
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
	evs := IcsEvents{}
	err = d.Decode(&evs)
	if err != nil {
		return IcsEvents{}, err
	}

	return evs, nil
}
