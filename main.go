// Copyright (C) 2018  Dinko Korunic, InfoMAR
//
// SPDX-License-Identifier: GPL-2.0-or-later

package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/KimMachineGun/automemlimit/memlimit"
	"github.com/dkorunic/IM-billing-v2/oauth"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
	"github.com/peterbourgon/ff/v4/ffyaml"
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
	DefaultAPITimeout  = 60 * time.Second
	DefaultCredentials = "assets/credentials.json"
	maxMemRatio        = 0.9
)

//go:embed assets/credentials.json
var credentialFS embed.FS

func main() {
	// configure GOMEMLIMIT to 90% of available memory (Cgroups v2/v1 or system)
	_, _ = memlimit.SetGoMemLimitWithOpts(
		memlimit.WithRatio(maxMemRatio),
		memlimit.WithProvider(
			memlimit.ApplyFallback(
				memlimit.FromCgroup,
				memlimit.FromSystem,
			),
		),
	)

	parseArgs()

	ctx := context.Background()
	ctxWithCancel, cancelFunction := context.WithCancel(ctx)

	defer cancelFunction()

	// Load Calendar API credentials
	b, err := credentialFS.ReadFile(DefaultCredentials)
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

	// Bound API work by the timeout; OAuth stays un-timed so login is excluded.
	// A derived context cancels in-flight requests, unlike a bare timer.
	apiCtx, apiCancel := context.WithTimeout(ctxWithCancel, *apiTimeout)
	defer apiCancel()

	chanCalendar := make(chan struct{}, 1)
	chanHolidays := make(chan map[string]holidayEvent, 1)

	// Fetch Office holiday events
	go func() {
		chanHolidays <- getHolidayEvents(apiCtx)
	}()

	// Fetch Calendar events and display them
	go func() {
		eventMap := getCalendarEvents(apiCtx, srv, calendarName)
		holidayMap := <-chanHolidays
		printMonthlyStats(eventMap, holidayMap)
		chanCalendar <- struct{}{}
	}()

	// Wait for completion or timeout
	select {
	case <-chanCalendar:
	case <-apiCtx.Done():
		log.Fatal("Timeout fetching Google calendar API... Exiting.")
	}
}

// parseArgs parses program arguments via ff and does minimal required sanity checking.
func parseArgs() {
	fs := ff.NewFlagSet("IM-billing-v2")

	calendarName = fs.String('c', "calendar", "", "calendar name")
	startDate = fs.String('s', "start", "", "start date (YYYY-MM-DD)")
	endDate = fs.String('e', "end", "", "end date (YYYY-MM-DD)")
	searchString = fs.String('x', "search", "", "search string (prefix match in event description)")

	_ = fs.StringLong("config", "", "config file (optional)")

	apiTimeout = fs.Duration('t', "timeout", DefaultAPITimeout, "Google Calendar API timeout")

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

	// Convert starting date in regard to local timezone
	if *startDate != "" {
		t, err := time.ParseInLocation(dateLayout, *startDate, time.Local)
		if err != nil {
			log.Fatalf("Cannot parse start time: %v", err)
		}

		startDateFinal = t
	}

	// Convert ending date in regards to local timezone
	if *endDate != "" {
		t, err := time.ParseInLocation(dateLayout, *endDate, time.Local)
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
