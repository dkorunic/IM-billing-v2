package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"sort"

	"google.golang.org/api/calendar/v3"
)

type workEvent struct {
	workDesc   string
	hoursTotal int
}

func getCalendarID(srv *calendar.Service, calendarName *string) *string {
	if *calendarName == "" {
		temp := "primary"
		return &temp
	}

	nextPageToken := ""
	for {
		calendarsCall := srv.CalendarList.List().
			MaxResults(100).
			PageToken(nextPageToken)

		listCal, err := calendarsCall.Do()
		if err != nil {
			log.Fatalf("Unable to retrieve user's calendar: %v", err)
		}

		for _, item := range listCal.Items {
			if item.Summary == *calendarName {
				return &item.Id
			}
		}

		nextPageToken = listCal.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	log.Fatalf("Unable to find calendar ID for %q", *calendarName)
	return nil
}

func getCalendarEvents(srv *calendar.Service, calendarName *string) map[string]workEvent {
	calID := getCalendarID(srv, calendarName)
	eventMap := make(map[string]workEvent)

	nextPageToken := ""
	for {
		eventsCall := srv.Events.List(*calID).
			ShowDeleted(false).
			SingleEvents(true).
			TimeMin(startDateFinal.Format(time.RFC3339)).
			TimeMax(endDateFinal.Format(time.RFC3339)).
			MaxResults(200).
			OrderBy("startTime").
			PageToken(nextPageToken)

		events, err := eventsCall.Do()
		if err != nil {
			log.Fatalf("Unable to retrieve user's events: %v", err)
		}

		loc, _ := time.LoadLocation("Local")

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

			desc := strings.TrimSpace(item.Description)
			if desc == "" {
				desc = strings.TrimSpace(item.Summary)
			}

			if *searchString != "" {
				if !strings.HasPrefix(desc, *searchString) {
					continue
				} else {
					desc = strings.TrimSpace(strings.TrimPrefix(desc, *searchString))
				}
			}

			eventMap = parseCalendarEvent(&desc, &start, &end, loc, eventMap)
		}

		nextPageToken = events.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	return eventMap
}

func parseCalendarEvent(desc, start, end *string, loc *time.Location, eventMap map[string]workEvent) map[string]workEvent {
	startTime, err := time.ParseInLocation(time.RFC3339, *start, loc)
	if err != nil {
		return eventMap
	}

	endTime, err := time.ParseInLocation(time.RFC3339, *end, loc)
	if err != nil {
		return eventMap
	}

	dateKey := startTime.Format("2006-01-02")
	workDuration := endTime.Sub(startTime)
	hours := int(workDuration.Round(time.Hour).Hours())

	if temp, ok := eventMap[dateKey]; ok {
		temp.hoursTotal += hours
		temp.workDesc = fmt.Sprintf("%s, %s", temp.workDesc, *desc)
		eventMap[dateKey] = temp
	} else {
		eventMap[dateKey] = workEvent{*desc, hours}

	}

	return eventMap
}

func printMonthlyStats(eventMap map[string]workEvent) {
	fmt.Printf("Listing work done on %v project from %v to %v\n", *calendarName,
		startDateFinal.Format("2006-01-02"), endDateFinal.Format("2006-01-02"))

	var keys []string
	var dayCount int

	for k := range eventMap {
		keys = append(keys, k)
		dayCount++
	}
	sort.Strings(keys)

	sep := "\t"
	if *dashFlag {
		sep = " - "
	}

	var totalHours int

	fmt.Printf("%10s%sHr%sDescription\n", "Date", sep, sep)
	for _, k := range keys {
		fmt.Printf("%s%s%2d%s%s\n", k, sep, eventMap[k].hoursTotal, sep, eventMap[k].workDesc)
		totalHours += eventMap[k].hoursTotal
	}

	fmt.Printf("\nTotal workhour sum for given period:\t\t%d hours\nTotal active days for given period:\t\t%d days\n", totalHours, dayCount)
}
