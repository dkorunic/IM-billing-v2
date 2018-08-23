/**
 */
package main

import (
	"io/ioutil"
	"log"
	"time"

	"os"

	"github.com/pborman/getopt/v2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

var calendarName, startDate, endDate, searchString *string
var helpFlag, dashFlag *bool
var startDateFinal, endDateFinal time.Time

func init() {
	calendarName = getopt.StringLong("calendar", 'c', "", "calendar name")
	startDate = getopt.StringLong("start", 's', "", "start date (YYYY-MM-DD)")
	endDate = getopt.StringLong("end", 'e', "", "end date (YYYY-MM-DD)")
	searchString = getopt.StringLong("search", 'x', "", "search string (substring match in event description)")
	helpFlag = getopt.BoolLong("help", 'h', "display help")
	dashFlag = getopt.BoolLong("dash", 'd', "use dashes when printing totals")

	t := time.Now()
	startDateFinal = time.Date(t.Year(), t.Month()-1, 1, 0, 0, 0, 0, time.Local)
	endDateFinal = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
}

func main() {
	parseArgs()

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	eventMap := getCalendarEvents(srv, calendarName)
	printMonthlyStats(eventMap)
}

func parseArgs() {
	getopt.Parse()

	if *helpFlag {
		getopt.PrintUsage(os.Stderr)
		os.Exit(0)
	}

	loc, _ := time.LoadLocation("Local")

	if *startDate != "" {
		t, err := time.ParseInLocation("2006-01-02", *startDate, loc)
		if err != nil {
			log.Fatal("Cannot parse start time", err)
		}
		startDateFinal = t
	}

	if *endDate != "" {
		t, err := time.ParseInLocation("2006-01-02", *endDate, loc)
		if err != nil {
			log.Fatal("Cannot parse end time", err)
		}
		endDateFinal = t
	}
}
