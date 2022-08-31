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
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/api/option"

	"github.com/pborman/getopt/v2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

var (
	calendarName, startDate, endDate, searchString *string
	apiTimeout                                     *int
	helpFlag, dashFlag                             *bool
	startDateFinal, endDateFinal                   time.Time
)

func init() {
	calendarName = getopt.StringLong("calendar", 'c', "", "calendar name")
	startDate = getopt.StringLong("start", 's', "", "start date (YYYY-MM-DD)")
	endDate = getopt.StringLong("end", 'e', "", "end date (YYYY-MM-DD)")
	searchString = getopt.StringLong("search", 'x', "", "search string (substring match in event description)")
	helpFlag = getopt.BoolLong("help", 'h', "display help")
	dashFlag = getopt.BoolLong("dash", 'd', "use dashes when printing totals")
	apiTimeout = getopt.IntLong("timeout", 't', 120, "Google Calendar API timeout (in seconds)")

	// By default, set start date to the 1st of previous month and end date to the 1st of current month
	t := time.Now()
	startDateFinal = time.Date(t.Year(), t.Month()-1, 1, 0, 0, 0, 0, time.Local)
	endDateFinal = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
}

func main() {
	parseArgs()

	ctx := context.Background()
	ctxWithCancel, cancelFunction := context.WithCancel(ctx)

	defer func() {
		cancelFunction()
	}()

	// Load Calendar API credentials
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// Parse Calendar API credentials
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	// Initialize Calendar client
	srv, err := calendar.NewService(ctxWithCancel, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	chanCalendar := make(chan struct{}, 1)
	chanHolidays := make(chan map[string]holidayEvent)
	defer close(chanCalendar)
	defer close(chanHolidays)

	// Fetch Office holiday events
	go func() {
		chanHolidays <- getHolidayEvents(ctxWithCancel)
	}()

	// Fetch Calendar events and display them
	go func() {
		eventMap := getCalendarEvents(srv, calendarName)
		holidayMap := <-chanHolidays
		printMonthlyStats(eventMap, holidayMap)
		chanCalendar <- struct{}{}
	}()

	// API timeout handler: wait for *apiTimeout duration until erroring out
	t := time.Duration(*apiTimeout) * time.Second
	select {
	case <-chanCalendar:
	case <-time.After(t):
		log.Fatal("Timeout fetching Google calendar API... Exiting.")
	}
}

// parseArgs parses program arguments via getopt and does minimal required sanity checking.
func parseArgs() {
	getopt.Parse()

	// Display usage and exit
	if *helpFlag {
		getopt.PrintUsage(os.Stderr)
		os.Exit(0)
	}

	// Use current location timezone from system
	loc, _ := time.LoadLocation("Local")

	// Convert starting date in regards to local timezone
	if *startDate != "" {
		t, err := time.ParseInLocation(dateLayout, *startDate, loc)
		if err != nil {
			log.Fatalf("Cannot parse start time: %v", err)
		}
		startDateFinal = t
	}

	// Convert ending date in regards to local timezone
	if *endDate != "" {
		t, err := time.ParseInLocation(dateLayout, *endDate, loc)
		if err != nil {
			log.Fatalf("Cannot parse end time: %v", err)
		}
		endDateFinal = t
	}

	// Check if dates are swapped
	if endDateFinal.Sub(startDateFinal) < 0 {
		log.Fatalf("End date (%v) is before start date (%v)\n", endDateFinal, startDateFinal)
	}
}
