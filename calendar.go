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
	"fmt"
	"log"
	"strings"
	"time"

	"sort"

	"google.golang.org/api/calendar/v3"
)

// workEvent holds individual calendar event with aggregate hourly total.
type workEvent struct {
	workDesc   string
	hoursTotal int
}

// holidayEvent holds individual calendar holiday event.
type holidayEvent struct {
	holidayDesc string
}

// dateLayout is a Time format parse layout of "YYYY-MM-DD"
const dateLayout = "2006-01-02"

// calendarMaxResults is a default maximum number of Google API results
const calendarMaxResults = 200

// icsParserTimeout is a default hard timeout for geolocation (ifconfig.co) and ICS parsing (officeholidays.com)
const icsParserTimeout = time.Second * 10

// getCalendarID gets a Google calendar ID out of a symbolic calendar name.
func getCalendarID(srv *calendar.Service, calendarName *string) string {
	// If the calendar name is not specified, use default (primary) calendar
	if *calendarName == "" {
		return "primary"
	}

	// Get calendar listing (paginated) and try to match name
	nextPageToken := ""
	for {
		calendarsCall := srv.CalendarList.List().
			MaxResults(calendarMaxResults).
			PageToken(nextPageToken)

		listCal, err := calendarsCall.Do()
		if err != nil {
			log.Fatalf("Unable to retrieve user's calendar: %v", err)
		}

		// Match calendar name
		for _, item := range listCal.Items {
			if item.Summary == *calendarName {
				return item.Id
			}
		}

		// Handle pagination
		nextPageToken = listCal.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	return ""
}

// getCalendarEvents gets all calendar events for a calendar ID and a date range.
func getCalendarEvents(srv *calendar.Service, calendarName *string) map[string]workEvent {
	// Fetch calendar ID
	calID := getCalendarID(srv, calendarName)
	if calID == "" {
		log.Fatalf("Unable to find calendar ID for %q", *calendarName)
	}

	// Allocate empty map structure corresponding to calendar events
	eventMap := make(map[string]workEvent)

	// Get all calendar events within specified date range (paginated)
	nextPageToken := ""
	for {
		eventsCall := srv.Events.List(calID).
			ShowDeleted(false).
			SingleEvents(true).
			TimeMin(startDateFinal.Format(time.RFC3339)).
			TimeMax(endDateFinal.Format(time.RFC3339)).
			MaxResults(calendarMaxResults).
			OrderBy("startTime").
			PageToken(nextPageToken)

		events, err := eventsCall.Do()
		if err != nil {
			log.Fatalf("Unable to retrieve user's events: %v", err)
		}

		// Use current location timezone from system
		loc, _ := time.LoadLocation("Local")

		// Don't parse event if it's recurring event
		for _, item := range events.Items {
			if item.RecurringEventId != "" {
				continue
			}

			start := item.Start.DateTime
			if start == "" {
				start = item.Start.Date
			}

			end := item.End.DateTime
			if end == "" {
				end = item.End.Date
			}

			// Trim event description/summary whitespace
			desc := strings.TrimSpace(item.Description)
			if desc == "" {
				desc = strings.TrimSpace(item.Summary)
			}

			// Match prefix string if requested
			if *searchString != "" {
				if !strings.HasPrefix(desc, *searchString) {
					continue
				} else {
					desc = strings.TrimSpace(strings.TrimPrefix(desc, *searchString))
				}
			}

			// Parse individual event and update calendar event map
			eventMap = parseCalendarEvent(desc, start, end, loc, eventMap)
		}

		// Handle pagination
		nextPageToken = events.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	return eventMap
}

// parseCalendarEvent parses individual calendar events and returns map with cumulative event hours per day and concatenated event descriptions.
func parseCalendarEvent(desc, start, end string, loc *time.Location, eventMap map[string]workEvent) map[string]workEvent {
	// Parse event starting time in RFC3339 (recurring events do not comply)
	startTime, err := time.ParseInLocation(time.RFC3339, start, loc)
	if err != nil {
		return eventMap
	}

	// Parse event ending time in RFC3339 (recurring events do not comply)
	endTime, err := time.ParseInLocation(time.RFC3339, end, loc)
	if err != nil {
		return eventMap
	}

	dateKey := startTime.Format(dateLayout) // Starting time is an event key
	workDuration := endTime.Sub(startTime)
	hours := int(workDuration.Round(time.Hour).Hours()) // Round to hours

	// Update calendar event map with either adding work hours or creating a new entry
	if temp, ok := eventMap[dateKey]; ok {
		temp.hoursTotal += hours
		temp.workDesc = fmt.Sprintf("%s, %s", temp.workDesc, desc)
		eventMap[dateKey] = temp
	} else {
		eventMap[dateKey] = workEvent{workDesc: desc, hoursTotal: hours}

	}

	return eventMap
}

// printMonthlyStats displays final monthly calendar statistics.
func printMonthlyStats(eventMap map[string]workEvent, holidayMap map[string]holidayEvent) {
	fmt.Printf("Listing work done on %v project from %v to %v\n", *calendarName,
		startDateFinal.Format(dateLayout), endDateFinal.Format(dateLayout))

	var eventKeys []string
	var dayCount int

	// Create temporary sorted slice for sorted map access
	for k := range eventMap {
		eventKeys = append(eventKeys, k)
		dayCount++
	}
	sort.Strings(eventKeys)

	var totalHours int

	// Dash or classic output format
	if *dashFlag {
		fmt.Printf("%10s - Hr - Description\n", "Date")
		for _, k := range eventKeys {
			fmt.Printf("%10s - %dh - %s\n", k, eventMap[k].hoursTotal, eventMap[k].workDesc)
			totalHours += eventMap[k].hoursTotal
		}
	} else {
		fmt.Printf("%10s\tHr\tDescription\n", "Date")
		for _, k := range eventKeys {
			fmt.Printf("%10s\t%2d\t%s\n", k, eventMap[k].hoursTotal, eventMap[k].workDesc)
			totalHours += eventMap[k].hoursTotal
		}

	}

	// Total cumulative statistics
	fmt.Printf("\nTotal workhour sum for given period:\t\t%d hours\nTotal active days for given period:\t\t%d days\n", totalHours, dayCount)

	// Attempt to identify event overlap with public holidays
	var holidayKeys []string
	for k := range holidayMap {
		if _, ok := eventMap[k]; ok {
			holidayKeys = append(holidayKeys, k)
		}
	}

	// Display event overlap with holidays only if we have any results
	if len(holidayKeys) > 0 {
		sort.Strings(holidayKeys)

		fmt.Printf("\nYou have calendar events on following public holidays:\n")
		for _, k := range holidayKeys {
			fmt.Printf("%10s\t%v\n", k, holidayMap[k].holidayDesc)
		}
	}
}

// getHolidayEvents does public IP geolocation (ifconfig.co), identifies country ISO code and gets holiday ICS for this country.
func getHolidayEvents() map[string]holidayEvent {
	c1 := make(chan struct{}, 1)
	defer close(c1)

	holidayMap := make(map[string]holidayEvent)

	go func() {
		// Initialize GeoIP/ifconfig HTTP client
		ifconfigClient, err := NewIfconfigClient()
		if err != nil {
			c1 <- struct{}{}
			return
		}

		// Fetch and parse JSON from ifconfig
		geoIP, err := ifconfigClient.GetIfconfigResponse()
		if err != nil {
			c1 <- struct{}{}
			return
		}

		// Use ISO 3166-1 country code for ICS lookup
		if geoIP.CountryISO == "" {
			c1 <- struct{}{}
			return
		}

		// Initialize ICS HTTP client
		icsClient, err := NewIcsClient(geoIP.CountryISO)
		if err != nil {
			c1 <- struct{}{}
			return
		}

		// Fetch and parse ICS response
		cal, err := icsClient.GetIcsResponse()
		if err != nil {
			c1 <- struct{}{}
			return
		}

		// Extract all holiday events
		for _, event := range cal {
			shortDate := event.Start.Format(dateLayout)
			holidayMap[shortDate] = holidayEvent{holidayDesc: event.Summary}
		}

		c1 <- struct{}{}
	}()

	// GeoIP/ifconfig and ICS parsing timeout handler
	select {
	case <-c1:
	case <-time.After(icsParserTimeout):
	}

	return holidayMap
}
