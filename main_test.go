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
	"os"
	"testing"
	"time"
)

// TC-18: without explicit --start/--end flags, startDateFinal must be the 1st of
// the previous month and endDateFinal must be the 1st of the current month.
func TestParseArgs_DefaultDates(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	os.Args = []string{"IM-billing-v2"}

	parseArgs()

	// Derive expected values from endDateFinal (always 1st of whatever month parseArgs used)
	// to avoid a time.Now() race if the test runs at a month boundary.
	wantEnd := time.Date(endDateFinal.Year(), endDateFinal.Month(), 1, 0, 0, 0, 0, time.Local)
	wantStart := time.Date(endDateFinal.Year(), endDateFinal.Month()-1, 1, 0, 0, 0, 0, time.Local)

	if !startDateFinal.Equal(wantStart) {
		t.Errorf("startDateFinal: got %v, want %v", startDateFinal, wantStart)
	}

	if !endDateFinal.Equal(wantEnd) {
		t.Errorf("endDateFinal: got %v, want %v", endDateFinal, wantEnd)
	}
}

// TC-19 (valid path): a date range where end > start must be accepted without fataling.
func TestParseArgs_ValidDateRangeAccepted(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	os.Args = []string{"IM-billing-v2", "--start", "2024-01-01", "--end", "2024-02-01"}

	parseArgs()

	wantStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	wantEnd := time.Date(2024, 2, 1, 0, 0, 0, 0, time.Local)

	if !startDateFinal.Equal(wantStart) {
		t.Errorf("startDateFinal: got %v, want %v", startDateFinal, wantStart)
	}

	if !endDateFinal.Equal(wantEnd) {
		t.Errorf("endDateFinal: got %v, want %v", endDateFinal, wantEnd)
	}
}
