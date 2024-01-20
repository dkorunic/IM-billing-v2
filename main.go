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
	"embed"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/KimMachineGun/automemlimit"
	"github.com/dkorunic/IM-billing-v2/oauth"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/ffyaml"
	"go.uber.org/automaxprocs/maxprocs"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var (
	calendarName, startDate, endDate, searchString *string
	apiTimeout                                     *time.Duration
	helpFlag, dashFlag, includeRecurring           *bool
	startDateFinal, endDateFinal                   time.Time
)

const (
	DefaultApiTimeout  = 60 * time.Second
	DefaultCredentials = "credentials.json"
)

//go:embed credentials.json
var credentialFS embed.FS

func main() {
	undo, _ := maxprocs.Set()
	defer undo()

	parseArgs()

	ctx := context.Background()
	ctxWithCancel, cancelFunction := context.WithCancel(ctx)

	defer func() {
		cancelFunction()
	}()

	// Load Calendar API credentials
	b, err := credentialFS.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read credentials file: %v", err)
	}

	// Parse Calendar API credentials
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// Retrieve Calendar API user token
	client, err := oauth.GetClient(ctxWithCancel, config, "token.json")
	if err != nil {
		log.Fatalf("Unable to retrieve token: %v", err)
	}

	// Initialize Calendar client
	srv, err := calendar.NewService(ctxWithCancel, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	chanCalendar := make(chan struct{}, 1)
	defer close(chanCalendar)

	chanHolidays := make(chan map[string]holidayEvent)
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

	delay := time.NewTimer(*apiTimeout)
	defer delay.Stop()

	// API timeout handler: wait for *apiTimeout duration until erroring out
	select {
	case <-chanCalendar:
	case <-delay.C:
		log.Fatal("Timeout fetching Google calendar API... Exiting.")
	}
}

// parseArgs parses program arguments via ff and does minimal required sanity checking.
func parseArgs() {
	fs := ff.NewFlagSet("IM-billing-v2")

	calendarName = fs.String('c', "calendar", "", "calendar name")
	startDate = fs.String('s', "start", "", "start date (YYYY-MM-DD)")
	endDate = fs.String('e', "end", "", "end date (YYYY-MM-DD)")
	searchString = fs.String('x', "search", "", "search string (substring match in event description)")

	_ = fs.StringLong("config", "", "config file (optional)")

	apiTimeout = fs.Duration('t', "timeout", DefaultApiTimeout, "Google Calendar API timeout")

	helpFlag = fs.Bool('h', "help", "display help")
	dashFlag = fs.Bool('d', "dash", "use dashes when printing totals")
	includeRecurring = fs.Bool('r', "recurring", "include recurring events")

	if err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("IMB"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ffyaml.Parser{}.Parse)); err != nil {
		fmt.Printf("%s\n", ffhelp.Flags(fs))
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}

	if *helpFlag {
		fmt.Printf("%s\n", ffhelp.Flags(fs))

		os.Exit(0)
	}

	// By default, set start date to the 1st of previous month and end date to the 1st of current month
	t := time.Now()
	startDateFinal = time.Date(t.Year(), t.Month()-1, 1, 0, 0, 0, 0, time.Local)
	endDateFinal = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)

	// Use current location timezone from system
	loc, _ := time.LoadLocation("Local")

	// Convert starting date in regard to local timezone
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
