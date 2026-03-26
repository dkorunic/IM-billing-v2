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
package ics_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dkorunic/IM-billing-v2/ics"
)

// validICS is a minimal RFC 5545 calendar with two date-only events.
// goics accepts LF line endings so no CRLF conversion is needed.
const validICS = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:holiday-1@test
DTSTART;VALUE=DATE:20240101
DTEND;VALUE=DATE:20240102
SUMMARY:New Year's Day
END:VEVENT
BEGIN:VEVENT
UID:holiday-2@test
DTSTART;VALUE=DATE:20240501
DTEND;VALUE=DATE:20240502
SUMMARY:Labour Day
END:VEVENT
END:VCALENDAR
`

// incompleteICS has one event missing a required field (SUMMARY) — it should be skipped.
const incompleteICS = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:good@test
DTSTART;VALUE=DATE:20240301
DTEND;VALUE=DATE:20240302
SUMMARY:Good Friday
END:VEVENT
BEGIN:VEVENT
UID:bad@test
DTSTART;VALUE=DATE:20240601
DTEND;VALUE=DATE:20240602
END:VEVENT
END:VCALENDAR
`

func TestNewClientWithContext(t *testing.T) {
	client, err := ics.NewClientWithContext(context.Background(), "HR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.URL == nil {
		t.Fatal("expected non-nil URL")
	}
}

func TestGetResponse_ValidICS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		_, _ = w.Write([]byte(validICS))
	}))
	defer srv.Close()

	client, _ := ics.NewClientWithContext(context.Background(), "HR")
	client.URL, _ = url.Parse(srv.URL)

	events, err := client.GetResponse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}

	found := false

	for _, ev := range events {
		if ev.Summary == "New Year's Day" {
			found = true

			if ev.Start.Month() != 1 || ev.Start.Day() != 1 || ev.Start.Year() != 2024 {
				t.Errorf("New Year's Day start: got %v, want 2024-01-01", ev.Start)
			}

			break
		}
	}

	if !found {
		t.Error("New Year's Day event not found in parsed events")
	}
}

func TestGetResponse_IncompleteEventSkipped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		_, _ = w.Write([]byte(incompleteICS))
	}))
	defer srv.Close()

	client, _ := ics.NewClientWithContext(context.Background(), "HR")
	client.URL, _ = url.Parse(srv.URL)

	events, err := client.GetResponse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The event missing SUMMARY must be silently skipped.
	if len(events) != 1 {
		t.Errorf("expected 1 event (incomplete event skipped), got %d", len(events))
	}

	if events[0].Summary != "Good Friday" {
		t.Errorf("Summary: got %q, want %q", events[0].Summary, "Good Friday")
	}
}

// TC-12: Event.Start must equal DTSTART and Event.End must equal DTEND — not swapped.
func TestConsumeICal_StartAndEndNotSwapped(t *testing.T) {
	const singleEvent = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//EN
BEGIN:VEVENT
UID:swap-test@test
DTSTART;VALUE=DATE:20240101
DTEND;VALUE=DATE:20240102
SUMMARY:New Year Day
END:VEVENT
END:VCALENDAR
`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(singleEvent))
	}))
	defer srv.Close()

	client, _ := ics.NewClientWithContext(context.Background(), "HR")
	client.URL, _ = url.Parse(srv.URL)

	events, err := client.GetResponse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	ev := events[0]

	// Start must be DTSTART = 2024-01-01
	if ev.Start.Year() != 2024 || ev.Start.Month() != 1 || ev.Start.Day() != 1 {
		t.Errorf("Start: got %v, want 2024-01-01 (DTSTART); Start and End may be swapped", ev.Start)
	}

	// End must be DTEND = 2024-01-02
	if ev.End.Year() != 2024 || ev.End.Month() != 1 || ev.End.Day() != 2 {
		t.Errorf("End: got %v, want 2024-01-02 (DTEND); Start and End may be swapped", ev.End)
	}

	if !ev.Start.Before(ev.End) {
		t.Errorf("Start (%v) must be before End (%v); fields may be swapped", ev.Start, ev.End)
	}
}

func TestGetResponse_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before making the request

	client, _ := ics.NewClientWithContext(ctx, "HR")
	client.URL, _ = url.Parse(srv.URL)

	_, err := client.GetResponse()
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}
