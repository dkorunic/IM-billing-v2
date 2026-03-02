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
	"strings"
	"testing"
	"time"
)

// fixed start time used across rounding sub-tests.
const roundingStart = "2024-01-15T09:00:00+00:00"

func TestParseCalendarEvent_NewEvent(t *testing.T) {
	eventMap := make(map[string]workEvent)

	result := parseCalendarEvent(
		"Work on project",
		"2024-01-15T09:00:00+00:00",
		"2024-01-15T17:00:00+00:00",
		time.UTC,
		eventMap,
	)

	if len(result) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result))
	}

	ev, ok := result["2024-01-15"]
	if !ok {
		t.Fatal("key 2024-01-15 not found in result map")
	}

	if ev.hoursTotal != 8 {
		t.Errorf("hoursTotal: got %d, want 8", ev.hoursTotal)
	}

	if ev.workDesc != "Work on project" {
		t.Errorf("workDesc: got %q, want %q", ev.workDesc, "Work on project")
	}
}

func TestParseCalendarEvent_AccumulateSameDay(t *testing.T) {
	eventMap := make(map[string]workEvent)

	eventMap = parseCalendarEvent("Morning", "2024-01-15T09:00:00+00:00", "2024-01-15T13:00:00+00:00", time.UTC, eventMap)
	eventMap = parseCalendarEvent("Afternoon", "2024-01-15T14:00:00+00:00", "2024-01-15T18:00:00+00:00", time.UTC, eventMap)

	if len(eventMap) != 1 {
		t.Fatalf("expected 1 map entry, got %d", len(eventMap))
	}

	ev := eventMap["2024-01-15"]

	if ev.hoursTotal != 8 {
		t.Errorf("hoursTotal: got %d, want 8", ev.hoursTotal)
	}

	if !strings.Contains(ev.workDesc, "Morning") || !strings.Contains(ev.workDesc, "Afternoon") {
		t.Errorf("workDesc missing expected parts: %q", ev.workDesc)
	}
}

func TestParseCalendarEvent_DifferentDays(t *testing.T) {
	eventMap := make(map[string]workEvent)

	eventMap = parseCalendarEvent("Day1", "2024-01-15T09:00:00+00:00", "2024-01-15T17:00:00+00:00", time.UTC, eventMap)
	eventMap = parseCalendarEvent("Day2", "2024-01-16T09:00:00+00:00", "2024-01-16T17:00:00+00:00", time.UTC, eventMap)

	if len(eventMap) != 2 {
		t.Fatalf("expected 2 map entries, got %d", len(eventMap))
	}

	if _, ok := eventMap["2024-01-15"]; !ok {
		t.Error("key 2024-01-15 not found")
	}

	if _, ok := eventMap["2024-01-16"]; !ok {
		t.Error("key 2024-01-16 not found")
	}
}

func TestParseCalendarEvent_InvalidStart(t *testing.T) {
	eventMap := make(map[string]workEvent)

	result := parseCalendarEvent("Work", "not-a-date", "2024-01-15T17:00:00+00:00", time.UTC, eventMap)

	if len(result) != 0 {
		t.Errorf("expected empty map for invalid start, got %d entries", len(result))
	}
}

func TestParseCalendarEvent_InvalidEnd(t *testing.T) {
	eventMap := make(map[string]workEvent)

	result := parseCalendarEvent("Work", "2024-01-15T09:00:00+00:00", "not-a-date", time.UTC, eventMap)

	if len(result) != 0 {
		t.Errorf("expected empty map for invalid end, got %d entries", len(result))
	}
}

func TestParseCalendarEvent_HourRounding(t *testing.T) {
	tests := []struct {
		name      string
		end       string
		wantHours int
	}{
		// 2h00m — exact, no rounding
		{"exact two hours", "2024-01-15T11:00:00+00:00", 2},
		// 2h30m — ties round up per time.Duration.Round semantics
		{"half hour rounds up", "2024-01-15T11:30:00+00:00", 3},
		// 2h29m — below half, rounds down
		{"below half rounds down", "2024-01-15T11:29:00+00:00", 2},
		// 2h31m — above half, rounds up
		{"above half rounds up", "2024-01-15T11:31:00+00:00", 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			eventMap := make(map[string]workEvent)

			result := parseCalendarEvent("Work", roundingStart, tc.end, time.UTC, eventMap)

			ev, ok := result["2024-01-15"]
			if !ok {
				t.Fatal("key 2024-01-15 not found")
			}

			if ev.hoursTotal != tc.wantHours {
				t.Errorf("hoursTotal: got %d, want %d", ev.hoursTotal, tc.wantHours)
			}
		})
	}
}

func TestParseCalendarEvent_DescriptionConcatenation(t *testing.T) {
	eventMap := make(map[string]workEvent)

	eventMap = parseCalendarEvent("First", "2024-01-15T09:00:00+00:00", "2024-01-15T10:00:00+00:00", time.UTC, eventMap)
	eventMap = parseCalendarEvent("Second", "2024-01-15T11:00:00+00:00", "2024-01-15T12:00:00+00:00", time.UTC, eventMap)
	eventMap = parseCalendarEvent("Third", "2024-01-15T13:00:00+00:00", "2024-01-15T14:00:00+00:00", time.UTC, eventMap)

	ev := eventMap["2024-01-15"]
	want := "First, Second, Third"

	if ev.workDesc != want {
		t.Errorf("workDesc: got %q, want %q", ev.workDesc, want)
	}
}
